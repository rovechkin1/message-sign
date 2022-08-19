package batch

import (
	"context"
	"fmt"
	"github.com/rovechkin1/message-sign/service/config"
	"github.com/rovechkin1/message-sign/service/signer"
	"github.com/rovechkin1/message-sign/service/store"
	"io/ioutil"
	"log"
	"net/http"
)

// RecordSigner signs all records available in store
type RecordSigner struct {
	store    store.MessageStore
	keyStore signer.KeyStore
}

type SignerStats struct {
	SignedRecords   int `json:"signed_records"`
	UnsignedRecords int `json:"unsigned_records"`
}

func NewRecordSigner(store store.MessageStore,
	keyStore signer.KeyStore) *RecordSigner {
	return &RecordSigner{
		store:    store,
		keyStore: keyStore,
	}
}

// SignRecords signs records in bulk
func (c *RecordSigner) SignRecords(ctx context.Context,
	store store.MessageStore,
	batchSize int) error {
	log.Printf("INFO: batch size: batch_size: %v",
		batchSize)

	nRecords, err := store.GetRecordCount(ctx, false)
	if err != nil {
		return err
	}

	// split into batches
	batchCount := nRecords / batchSize
	if nRecords%batchSize > 0 {
		batchCount += 1
	}
	log.Printf("INFO: founds records: %v, will process batches: %v",
		nRecords, batchCount)

	// get available signing keys
	keys, err := c.keyStore.GetKeyIds()
	if err != nil {
		return err
	}

	for iBatch := 0; iBatch < batchCount; iBatch += 1 {
		// spawn batch signing
		// if failure is encountered, return reporting how many
		// batches were spawn
		iKey := iBatch % len(keys)
		log.Printf("Batch: %v, key_id: %v, url: %v", iBatch, keys[iKey],
			config.GetMsgSignerUrl())
		url := fmt.Sprintf("%s/batch/%d/%d/%s",
			config.GetMsgSignerUrl(), iBatch, batchCount, keys[iKey])
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("Error signing batch %v, error: %v",
				iBatch, err)
		}

		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error signing batch %v, error: %v",
				iBatch, err)
		}
	}
	log.Printf("INFO: started signed records")
	return nil
}

// SignRecords signs records in bulk
func (c *RecordSigner) GetStats(ctx context.Context,
	store store.MessageStore) (*SignerStats, error) {

	var err error
	stats := &SignerStats{}
	stats.UnsignedRecords, err = store.GetRecordCount(ctx, false)
	if err != nil {
		return nil, err
	}

	stats.SignedRecords, err = store.GetRecordCount(ctx, true)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
