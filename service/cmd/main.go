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
	"strconv"
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
	store := store.NewMongoStore()
	keyStore, err := signer.NewFileKeyStore()
	if err != nil {
		log.Fatalf("Canot init key store")
	}
	recordSigner := batch.NewRecordSigner(store, keyStore)
	batchSigner := batch.NewBatchSigner(store, keyStore)

	router := gin.Default()
	// endpoint to sign all records in store
	router.GET("/sign/:size", func(c *gin.Context) {
		var err error
		batchSize, err := strconv.Atoi(c.Param("size"))
		if err != nil {
			c.String(http.StatusBadRequest, "invalid batch size")
			return
		}
		err = recordSigner.SignRecords(ctx, store, batchSize)
		if err != nil {
			log.Printf("ERROR: failed to sign in batch: %v", err)
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("error signing records, error: %v", err))
		} else {
			c.String(http.StatusOK, fmt.Sprintf("Success started  signing."))
		}
	})

	// endpoint to get statistics
	router.GET("/stats", func(c *gin.Context) {
		var err error
		stats, err := recordSigner.GetStats(ctx, store)
		if err != nil {
			log.Printf("ERROR: failed to get stats: %v", err)
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("error to get stats, error: %v", err))
		} else {
			r, _ := json.Marshal(*stats)
			c.String(http.StatusOK, fmt.Sprintf("stats: %s", r))
		}
	})

	// endpoint to sing a batch of records
	router.GET("/batch/:offset/:size/:key", func(c *gin.Context) {
		var err error
		offset, err := strconv.Atoi(c.Param("offset"))
		if err != nil {
			c.String(http.StatusBadRequest, "invalid offset")
			return
		}
		batchSize, err := strconv.Atoi(c.Param("size"))
		if err != nil {
			c.String(http.StatusBadRequest, "invalid batch size")
			return
		}
		if len(c.Param("key")) == 0 {
			c.String(http.StatusBadRequest, "invalid keyId")
			return
		}
		err = batchSigner.SignBatch(ctx, offset, batchSize, c.Param("key"))
		if err != nil {
			log.Printf("ERROR: failed to sign in batch: %v", err)
			c.String(http.StatusInternalServerError, fmt.Sprintf("error signing batch, error: %v", err))
		} else {
			c.String(http.StatusOK, fmt.Sprintf("Success started signing"))
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
