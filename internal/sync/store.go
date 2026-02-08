package sync

import (
	"context"
	"time"

	"github.com/anaryk/metabase-flexibee-adapter/internal/store"
)

// SyncStore defines the store operations needed by the sync engine.
type SyncStore interface {
	GetSyncState(ctx context.Context, evidence string) (*store.SyncState, error)
	SetSyncState(ctx context.Context, evidence string, state store.SyncState) error
	UpsertRecords(ctx context.Context, table string, records []map[string]any, primaryKey string) (int, error)
	CleanupOldRecords(ctx context.Context, table string, olderThan time.Time, batchSize int) (int64, error)
	LogCleanup(ctx context.Context, evidence string, rowsDeleted int64, oldestKept *time.Time) error
}
