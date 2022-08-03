package store

import (
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Mock store implementation

// Interface to read/write messages
type mockStore struct {
}

func NewMockStore() MessageStore {
	return &mockStore{}
}

// GetTotalRecords records in store
func (c *mockStore) GetTotalRecords() (int, error) {
	// generate random number 10k-12k
	s := rand.NewSource(int64(time.Now().Second()))
	r := rand.New(s)
	nRecords := 100000 + r.Intn(20000)
	// simulate store latency
	time.Sleep(time.Duration(500 * time.Millisecond))
	return nRecords, nil
}

// ReadBatch reads messages in batch
func (c *mockStore) ReadBatch(start int, nRecords int) ([]Record, error) {
	var records []Record
	for i := 0; i < nRecords; i += 1 {
		uuidWithHyphen := uuid.New()
		uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
		r := Record{
			Id:  uuid,
			Msg: uuid, // same as guid
		}
		records = append(records, r)
	}
	// simulate request latency
	time.Sleep(1 * time.Second)
	return records, nil
}

// WriteSignaturesBatch writes messages signatures in batch
func (c *mockStore) WriteSignaturesBatch(signs []Record) error {
	return nil
}
