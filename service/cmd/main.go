// build go1.16

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rovechkin1/message-sign/service/config"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rovechkin1/message-sign/service/batch"
	"github.com/rovechkin1/message-sign/service/store"

	"github.com/rovechkin1/message-sign/service/signer"
)

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// initialize objects
	mongoClient, ctxMongo, err := store.NewMongoClient(ctx)
	if err != nil {
		log.Fatalf("Cannot create record-generator client: %v, error: %v", config.GetMongoUrl(), err)
	}
	defer mongoClient.Close(ctxMongo)
	store := store.NewMongoStore(mongoClient)
	keyStore, err := signer.NewFileKeyStore()
	if err != nil {
		log.Fatalf("Canot init key store")
	}

	router := gin.Default()
	// used fro readiness and liveness
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("live"))
	})

	// endpoint to get statistics
	router.GET("/stats", func(c *gin.Context) {
		var err error
		stats, err := batch.GetStats(ctx, store)
		if err != nil {
			log.Printf("ERROR: failed to get stats: %v", err)
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("error to get stats, error: %v", err))
		} else {
			r, _ := json.Marshal(*stats)
			c.String(http.StatusOK, fmt.Sprintf("stats: %s", r))
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetSignerPort()),
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// start periodic signers
	batchSigner, err := batch.NewBatchSigner(store, keyStore, mongoClient.GetMongo())
	if err != nil {
		log.Printf("ERROR: cannot create record signer, error: %v", err)
	} else {
		batchSigner.StartPeriodicBatchSigner(ctx)
	}

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
