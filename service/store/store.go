package store

import "context"

// Record describing message to sign
type Record struct {
	// message unique id
	Id string
	// message
	Msg string
	// signature
	Signature string
	// Public key id
	KeyId string
}

// MessageStore is an interface to read/write messages
type MessageStore interface {
	// GetRecordCount records in store which are signed
	GetRecordCount(ctx context.Context, signed bool) (int, error)

	// ReadBatch reads messages in batch
	ReadBatch(ctx context.Context, batchId int, batchCount int) ([]Record, error)

	// WriteRecord writes a single record
	WriteRecord(ctx context.Context, record Record) error
}
