package batch

import (
	"context"
	"github.com/rovechkin1/message-sign/service/store"
	"log"
)

type BatchSigner struct {
	store store.MessageStore
}

func NewBatchSigner(store store.MessageStore) *BatchSigner{
	return &BatchSigner{
		store: store,
	}
}

// SignBatch implements signer for messages
func (c *BatchSigner) SignBatch(ctx context.Context, start int, nRecords int) (int,error) {
	log.Printf("INFO: sign batch: start %v, nRecords: %v",
		start, nRecords)
	return nRecords, nil
}
