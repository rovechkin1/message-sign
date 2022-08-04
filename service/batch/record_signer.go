package batch

import (
	"context"
	"fmt"
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
	TotalRecords    int `json:"total_records"`
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

	// get available signing keys
	keys, err := c.keyStore.GetKeyIds()
	if err != nil {
		return err
	}

	nRecords, err := store.GetTotalRecords()
	if err != nil {
		return err
	}

	// split into batches
	nBatches := nRecords / batchSize
	if nRecords%batchSize > 0 {
		nBatches += 1
	}
	log.Printf("INFO: founds records: %v, will process batches: %v",
		nRecords, nBatches)
	for iBatch := 0; iBatch < nBatches; iBatch += 1 {
		// spawn batch signing
		// if failure is encountered, return reporting how many
		// batches were spawn
		iKey := iBatch % len(keys)
		log.Printf("Batch: %v, key_id: %v", iBatch, keys[iKey])
		url := fmt.Sprintf("http://localhost:8080/batch/%d/%d/%s",
			iBatch*batchSize, batchSize, keys[iKey])
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
	stats.TotalRecords, err = store.GetTotalRecords()
	if err != nil {
		return nil, err
	}

	stats.SignedRecords, err = store.GetTotalSignedRecords()
	if err != nil {
		return nil, err
	}

	stats.UnsignedRecords = stats.TotalRecords - stats.SignedRecords

	return stats, nil
}
