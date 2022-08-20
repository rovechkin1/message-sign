package batch

import (
	"context"
	"github.com/google/uuid"
	"github.com/rovechkin1/message-sign/service/config"
	"github.com/rovechkin1/message-sign/service/signer"
	"github.com/rovechkin1/message-sign/service/store"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

// BatchSigner signs messages in batches
type BatchSigner struct {
	store    store.MessageStore
	keyStore signer.KeyStore
	mongo    *mongo.Client
}

func NewBatchSigner(store store.MessageStore, keyStore signer.KeyStore,
	mongo *mongo.Client) *BatchSigner {
	return &BatchSigner{
		store:    store,
		keyStore: keyStore,
		mongo:    mongo,
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
	ctx := context.Background()
	// query records
	records, err := c.store.ReadBatch(ctx, batchId, batchCount)
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
	}

	if config.GetEnableMongoXact() {
		log.Printf("INFO: enabled mongo xact")
		// start transaction
		xact, err := store.NewMongoXact(c.mongo)
		if err != nil {
			return err
		}
		defer xact.Close(ctx)

		// write as batch
		// if failed , then fail the whole batch it will be retried later
		writeBatch := func(sessionContext mongo.SessionContext) (interface{}, error) {
			err = c.store.WriteBatch(sessionContext, signedRecords)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		_, err = xact.WithTransaction(ctx, writeBatch)

		return err
	}

	log.Printf("INFO: disable mongo xact")
	err = c.store.WriteBatch(ctx, signedRecords)
	if err != nil {
		return err
	}

	return nil
}
