package batch

import (
	"context"
	"github.com/rovechkin1/message-sign/service/store"
)

type SignerStats struct {
	SignedRecords   int `json:"signed_records"`
	UnsignedRecords int `json:"unsigned_records"`
}

// SignRecords signs records in bulk
func GetStats(ctx context.Context, store store.MessageStore) (*SignerStats, error) {
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
