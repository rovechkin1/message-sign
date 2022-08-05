package batch

import (
	"context"
	"github.com/rovechkin1/message-sign/service/signer"
	"github.com/rovechkin1/message-sign/service/store"
	"log"
)

// BatchSigner signs messages in batches
type BatchSigner struct {
	store    store.MessageStore
	keyStore signer.KeyStore
}

func NewBatchSigner(store store.MessageStore, keyStore signer.KeyStore) *BatchSigner {
	return &BatchSigner{
		store:    store,
		keyStore: keyStore,
	}
}

// SignBatch implements signer for messages
func (c *BatchSigner) SignBatch(ctx context.Context, batchId int, batchCount int, keyId string) error {
	go func() {
		err := c.signRecords(batchId, batchCount, keyId)
		if err != nil {
			log.Printf("ERROR: failed to sign records for batchId: %v, nRecords: %v, keyId: %v, error: %v",
				batchId, batchCount, keyId, err)
		} else {
			log.Printf("INFO: signed %v records for batchId: %v, keyId: %v",
				batchCount, batchId, keyId)
		}
	}()
	return nil
}

func (c *BatchSigner) signRecords(batchId int, batchCount int, keyId string) error {
	log.Printf("INFO: sign batch: batchId %v, batchCount: %v",
		batchId, batchCount)
	// query records
	records, err := c.store.ReadBatch(context.Background(), batchId, batchCount)
	if err != nil {
		return err
	}

	// get key
	key, err := c.keyStore.GetKeyById(keyId)
	if err != nil {
		return err
	}

	for _, r := range records {
		// sign here
		sign, err := key.Sign(r.Msg)
		if err != nil {
			// ignore error continue signing
			log.Printf("WARN: failed to sign message with key: %v, error: %v", key.KeyId, err)
			continue
		}
		r.Signature = sign
		r.KeyId = key.KeyId
		// write records back
		err = c.store.WriteRecord(context.Background(), r)
		if err != nil {
			log.Printf("WARN: failed to write record: %v, error: %v, skip it", r.Id, err)
			continue
		}
	}

	return nil
}
