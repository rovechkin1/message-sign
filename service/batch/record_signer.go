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
	batchSize int) (int, error) {
	log.Printf("INFO: batch size: batch_size: %v",
		batchSize)

	// get available signing keys
	keys, err := c.keyStore.GetKeyIds()
	if err != nil {
		return 0, err
	}

	nRecords, err := store.GetTotalRecords()
	if err != nil {
		return nRecords, err
	}

	// split into batches
	nBatches := nRecords / batchSize
	if nRecords%batchSize > 0 {
		nBatches += 1
	}
	log.Printf("INFO: founds records: %v, will process batches: %v",
		nRecords, nBatches)
	totalSigned := 0
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
			return totalSigned, fmt.Errorf("Error signing batch %v, error: %v",
				iBatch, err)
		}

		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return totalSigned, fmt.Errorf("Error signing batch %v, error: %v",
				iBatch, err)
		}
		nSigned := batchSize
		if iBatch == nBatches-1 && nRecords%batchSize > 0 {
			nSigned = nRecords % batchSize
		}
		totalSigned += nSigned
	}
	log.Printf("INFO: signed records: %v", totalSigned)
	return totalSigned, nil
}