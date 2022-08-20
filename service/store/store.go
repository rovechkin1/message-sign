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
	// salt
	Salt string
	// Public key id
	KeyId string
}

type SigningKeyMetadata struct {
	Id    string
	Nonce int64
}

func NewSigningKeyMetadata(keyId string) *SigningKeyMetadata {
	return &SigningKeyMetadata{
		Id: keyId,
	}
}

// MessageStore is an interface to read/write messages
type MessageStore interface {
	// GetRecordCount records in store which are signed
	GetRecordCount(ctx context.Context, signed bool) (int, error)

	// ReadBatch reads messages in batch
	ReadBatch(ctx context.Context, batchId int, batchCount int) ([]Record, error)

	// WriteRecord writes a single record
	WriteRecord(ctx context.Context, record Record) error

	// WriteBatch writes records as a batch
	WriteBatch(ctx context.Context, records []Record) error

	// ReadSigningKeyMetadata reads metadata of signing key
	ReadSigningKeyMetadata(ctx context.Context, keyId string) (*SigningKeyMetadata, error)

	// WriteSigningKeyMetadata writes metadata of signing key
	WriteSigningKeyMetadata(ctx context.Context, keyMetadata *SigningKeyMetadata) error
}
