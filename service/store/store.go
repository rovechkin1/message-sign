package store

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
	GetTotalRecords() (int, error)

	// ReadBatch reads messages in batch
	ReadBatch(start int, n int) ([]Record, error)

	// WriteSignaturesBatch writes messages signatures in batch
	WriteSignaturesBatch(signs []Record) error
}
