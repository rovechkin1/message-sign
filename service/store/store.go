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
	// GetTotalRecords records in store
	GetTotalRecords(ctx context.Context) (int, error)

	// GetTotalSignedRecords records in store which are signed
	GetTotalSignedRecords(ctx context.Context) (int, error)

	// ReadBatch reads messages in batch
	ReadBatch(ctx context.Context, start int, n int) ([]Record, error)

	// WriteSignaturesBatch writes messages signatures in batch
	WriteSignaturesBatch(ctx context.Context, signs []Record) error
}
