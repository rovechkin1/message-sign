
// build go1.16

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rovechkin1/message-sign/service/batch"
	"github.com/rovechkin1/message-sign/service/store"
)

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// create mock store
	store := store.NewMockStore()
	recordSigner := batch.NewRecordSigner(store)
	batchSigner := batch.NewBatchSigner(store)

	router := gin.Default()
	// endpoint to sign all records in store
	router.GET("/sign/:size", func(c *gin.Context) {
		var err error
		batchSize, err := strconv.Atoi(c.Param("size"))
		if err != nil {
			c.String(http.StatusBadRequest, "invalid batch size")
			return
		}
		nRecords,err := recordSigner.SignRecords(ctx, store, batchSize)
		if err != nil {
			log.Printf("ERROR: failed to sign in batch: %v", err)
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("error signing records, error: %v",err))
		} else {
			c.String(http.StatusOK, fmt.Sprintf("Success signing. Records signed: %v",nRecords))
		}
	})

	// endpoint to sing a batch of records
	router.GET("/batch/:offset/:size", func(c *gin.Context) {
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
		nRecords,err := batchSigner.SignBatch(ctx, offset, batchSize)
		if err != nil {
			log.Printf("ERROR: failed to sign in batch: %v", err)
			c.String(http.StatusInternalServerError, fmt.Sprintf("error signing batch, error: %v",err))
		} else {
			c.String(http.StatusOK, fmt.Sprintf("Success signing. Records signed: %v",nRecords))
		}
	})


	srv := &http.Server{
		Addr:    ":8080",
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
