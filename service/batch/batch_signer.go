package batch

import (
	"context"
	"fmt"
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
	if config.GetEnableMongoXact() {
		return c.signRecordsXact(ctx, batchId, batchCount, keyId)
	} else {
		log.Printf("INFO: disable mongo xact")
		return c.signRecordsAux(ctx, batchId, batchCount, keyId)
	}
}

func (c *BatchSigner) signRecordsXact(ctx context.Context, batchId int, batchCount int, keyId string) error {
	log.Printf("INFO: enabled mongo xact")
	// start transaction
	xact, err := store.NewMongoXact(c.mongo)
	if err != nil {
		return err
	}
	defer xact.Close(ctx)
	// Run reading and writing of records
	// as a single transaction
	// this ensures:
	// 1. no new records are inserted as we do signing
	// 2. keys nonce is properly incremented
	// 3. BulkWrite happens atomically
	// if failed , then fail the whole batch it will be retried later
	writeBatch := func(sessionContext mongo.SessionContext) (interface{}, error) {
	err := c.signRecordsAux(sessionContext, batchId, batchCount,keyId)
	return nil,err
	}
	_, err = xact.WithTransaction(ctx, writeBatch)
	return err
}

func (c *BatchSigner) signRecordsAux(ctx context.Context, batchId int, batchCount int, keyId string) error {
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

	// read key metadata which contains nonce
	var keyMd *store.SigningKeyMetadata
	keyMd, err = c.store.ReadSigningKeyMetadata(ctx,keyId)
	if err != nil && err != mongo.ErrNoDocuments{
		return err
	}
	if keyMd == nil {
		keyMd = store.NewSigningKeyMetadata(keyId)
	} else{
		fmt.Printf("here\n")
	}

	log.Printf("INFO: start with nonce: %v, keyId: %v, batchId: %v", keyMd.Nonce, keyId, batchId)

	var signedRecords []store.Record
	for _, r := range records {
		// sign here
		r.Salt = fmt.Sprintf("%d",keyMd.Nonce)
		// add random salt if needed
		sign, err := key.Sign(r.Salt + r.Msg)
		if err != nil {
			// ignore error continue signing
			log.Printf("WARN: failed to sign message with key: %v, error: %v", key.KeyId, err)
			continue
		}
		r.Signature = sign
		r.KeyId = key.KeyId
		keyMd.Nonce +=1

		signedRecords = append(signedRecords, r)
	}
	err = c.store.WriteBatch(ctx, signedRecords)

	// write new key metadata, e.g. nonce
	log.Printf("INFO: end with nonce: %v, keyId: %v, batchId: %v", keyMd.Nonce, keyId, batchId)
	err = c.store.WriteSigningKeyMetadata(ctx,keyMd)
	return err
}
