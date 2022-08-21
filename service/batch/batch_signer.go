package batch

import (
	"context"
	"fmt"
	"github.com/rovechkin1/message-sign/service/config"
	"github.com/rovechkin1/message-sign/service/signer"
	"github.com/rovechkin1/message-sign/service/store"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strconv"
	"strings"
	"time"
)

// BatchSigner signs messages in batches
type BatchSigner struct {
	store        store.MessageStore
	keyStore     signer.KeyStore
	mongoClient  *mongo.Client
	signerId     int
	totalSigners int
	batchSize    int
	keyIdx       int
	keys         []string
}

func NewBatchSigner(store store.MessageStore, keyStore signer.KeyStore,
	mongoClient *mongo.Client) (*BatchSigner, error) {
	batchSize := config.GetBatchSize()
	// total signers (size of stateful set)
	totalSigners := config.GetTotalSigners()

	// this signer id in format signer-X
	signerIdString := config.GetMyPodName()
	idParts := strings.Split(signerIdString, "-")
	if len(idParts) < 2 {
		return nil, fmt.Errorf("ERROR: Invalid signer id format: %s, expected signer-0", signerIdString)
	}

	signerId, err := strconv.ParseInt(idParts[len(idParts)-1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Invalid signer id format: %s, expected signer-0", signerIdString)
	}

	if int(signerId) >= totalSigners {
		return nil, fmt.Errorf("ERROR: signerId: %v is grater than totalSigners: %v", signerId, totalSigners)
	}
	log.Printf("INFO: signerId: %v, batch size: batch_size: %v, totalSigners: %v",
		signerId, batchSize, totalSigners)

	// get available signing keys
	keys, err := keyStore.GetKeyIds()
	if err != nil {
		return nil, err
	}

	// number of keys must be more than number of signers
	// otherwise we can't do signing in parallel
	// not enough keys for each signer
	if int(signerId) >= len(keys) {
		return nil, fmt.Errorf("ERROR: not enough keys for each signer: %v, signer %v is inactive",
			len(keys), signerId)
	}

	return &BatchSigner{
		store:        store,
		keyStore:     keyStore,
		mongoClient:  mongoClient,
		signerId:     int(signerId),
		totalSigners: totalSigners,
		batchSize:    batchSize,
		keys:         keys,
	}, nil
}

// StartPeriodicBatchSigner periodically polls available records and signs them
func (c *BatchSigner) StartPeriodicBatchSigner(ctx context.Context) {

	// TODO: use config for timeout
	timer := time.NewTimer(1 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("INFO: BatchSigner is done, batchId: %v", c.signerId)
				return
			case <-timer.C:
				keyIdx := (c.keyIdx*c.totalSigners + c.signerId) % len(c.keys)
				err := c.SignBatch(ctx, c.keys[keyIdx])
				if err != nil {
					log.Printf("ERROR: failed to sign batchId: %v, error: %v", c.signerId, err)
				}
				c.keyIdx += 1
			default:
			}
		}
	}()
}

// SignBatch implements signer for messages
func (c *BatchSigner) SignBatch(ctx context.Context, keyId string) error {
	go func() {
		err := c.signRecords(keyId)
		if err != nil {
			log.Printf("ERROR: failed to sign records for batchId: %v, nRecords: %v, keyId: %v, error: %v",
				c.signerId, c, keyId, err)
		} else {
			log.Printf("INFO: signed %v records for batchId: %v, keyId: %v",
				c.batchSize, c.signerId, keyId)
		}
	}()
	return nil
}
func (c *BatchSigner) signRecords(keyId string) error {
	log.Printf("INFO: sign batch: batchId %v, batchCount: %v",
		c.signerId, c.totalSigners)
	ctx := context.Background()
	if config.GetEnableMongoXact() {
		return c.signRecordsXact(ctx, c.signerId, c.totalSigners, keyId)
	} else {
		log.Printf("INFO: disable mongo xact")
		return c.signRecordsAux(ctx, c.signerId, c.totalSigners, keyId)
	}
}

func (c *BatchSigner) signRecordsXact(ctx context.Context, batchId int, batchCount int, keyId string) error {
	log.Printf("INFO: enabled mongo xact")
	// start transaction
	xact, err := store.NewMongoXact(c.mongoClient)
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
		err := c.signRecordsAux(sessionContext, batchId, batchCount, keyId)
		return nil, err
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
	keyMd, err = c.store.ReadSigningKeyMetadata(ctx, keyId)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if keyMd == nil {
		keyMd = store.NewSigningKeyMetadata(keyId)
	} else {
		fmt.Printf("here\n")
	}

	log.Printf("INFO: start with nonce: %v, keyId: %v, batchId: %v", keyMd.Nonce, keyId, batchId)

	var signedRecords []store.Record
	for _, r := range records {
		// sign here
		r.Salt = fmt.Sprintf("%d", keyMd.Nonce)
		// add random salt if needed
		sign, err := key.Sign(r.Salt + r.Msg)
		if err != nil {
			// ignore error continue signing
			log.Printf("WARN: failed to sign message with key: %v, error: %v", key.KeyId, err)
			continue
		}
		r.Signature = sign
		r.KeyId = key.KeyId
		keyMd.Nonce += 1

		signedRecords = append(signedRecords, r)
	}
	err = c.store.WriteBatch(ctx, signedRecords)

	// write new key metadata, e.g. nonce
	log.Printf("INFO: end with nonce: %v, keyId: %v, batchId: %v", keyMd.Nonce, keyId, batchId)
	err = c.store.WriteSigningKeyMetadata(ctx, keyMd)
	return err
}
