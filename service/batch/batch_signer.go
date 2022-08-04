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
func (c *BatchSigner) SignBatch(ctx context.Context, start int, nRecords int, keyId string) error {
	go func() {
		nr, err := c.signRecords(start, nRecords, keyId)
		if err != nil {
			log.Printf("ERROR: failed to sign records in range: %v, nRecords: %v, keyId: %v, error: %v",
				start, nRecords, keyId, err)
		} else {
			log.Printf("INFO: signed %v records in range: %v, nRecords: %v, keyId: %v",
				nRecords, start, nr, keyId)
		}
	}()
	return nil
}

func (c *BatchSigner) signRecords(start int, nRecords int, keyId string) (int, error) {
	log.Printf("INFO: sign batch: start %v, nRecords: %v",
		start, nRecords)
	// query records
	records, err := c.store.ReadBatch(start, nRecords)
	if err != nil {
		return 0, err
	}

	// get key
	key, err := c.keyStore.GetKeyById(keyId)
	if err != nil {
		return 0, err
	}

	var signedRecords []store.Record
	for _, r := range records {
		// sign record if not signed
		if len(r.Signature) == 0 {
			// sign here
			sign, err := key.Sign(r.Msg)
			if err != nil {
				// ignore error continue signing
				log.Printf("WARN: failed to sign message with key: %v, error: %v", key.KeyId, err)
				continue
			}
			r.Signature = sign
			r.KeyId = key.KeyId
			signedRecords = append(signedRecords, r)
		}
	}
	// write records back
	err = c.store.WriteSignaturesBatch(signedRecords)
	if err != nil {
		return 0, err
	}
	return nRecords, nil
}
