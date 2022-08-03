package store

// Record describing message to sign
type Record struct {
	// message unique id
	Id string
	// message
	Msg string
	// signature
	Signature []byte
	// Public key to validate
	Pub []byte
}

// MessageStore is an interface to read/write messages
type MessageStore interface{
	// GetTotalRecords records in store
	GetTotalRecords() (int,error)

	// ReadBatch reads messages in batch
	ReadBatch(start int, n int) ([]Record, error)

	// WriteSignaturesBatch writes messages signatures in batch
	WriteSignaturesBatch(signs []Record, n int) error
}
