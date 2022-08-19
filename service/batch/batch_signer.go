package batch

import (
	"context"
	"github.com/google/uuid"
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
	if len(records) == 0 {
		log.Printf("INFO: no records to sign. BatchId : %v\n", batchId)
		return nil
	}

	// get key
	key, err := c.keyStore.GetKeyById(keyId)
	if err != nil {
		return err
	}

	var signedRecords []store.Record
	for _, r := range records {
		// sign here
		r.Salt = uuid.New().String()
		sign, err := key.Sign(r.Salt + r.Msg)
		if err != nil {
			// ignore error continue signing
			log.Printf("WARN: failed to sign message with key: %v, error: %v", key.KeyId, err)
			continue
		}
		r.Signature = sign
		r.KeyId = key.KeyId

		signedRecords = append(signedRecords, r)
		// try to write batch

		//// write records back
		//err = c.store.WriteRecord(context.Background(), r)
		//if err != nil {
		//	log.Printf("WARN: failed to write record: %v, error: %v, skip it", r.Id, err)
		//	continue
		//}
	}

	// write as batch first
	err = c.store.WriteBatch(context.Background(), signedRecords)
	if err == nil {
		return nil
	}

	// batch failed
	// write individually
	log.Printf("WARN: failed to write records in bulk, error: %v, write individually", err)
	for _, r := range signedRecords {
		err = c.store.WriteRecord(context.Background(), r)
		if err != nil {
			log.Printf("WARN: failed to write record: %v, error: %v, skip it", r.Id, err)
			continue
		}
	}

	return nil
}
