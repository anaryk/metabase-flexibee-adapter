package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/anaryk/metabase-flexibee-adapter/internal/flexibee"
	"github.com/anaryk/metabase-flexibee-adapter/internal/registry"
	"github.com/anaryk/metabase-flexibee-adapter/internal/store"
)

// syncEvidence performs a single sync pass for one evidence type.
// It uses incremental sync based on the lastUpdate timestamp.
func syncEvidence(ctx context.Context, client *flexibee.Client, st SyncStore, ev registry.Evidence, batchSize int, logger *slog.Logger) error {
	logger = logger.With("evidence", ev.Slug, "table", ev.Table)

	// Get current sync state
	state, err := st.GetSyncState(ctx, ev.Slug)
	if err != nil {
		return fmt.Errorf("get sync state: %w", err)
	}

	// Build fetch options
	opts := flexibee.FetchOptions{
		Limit:  batchSize,
		Detail: "full",
	}

	// Incremental sync: only fetch records modified since last sync
	if state != nil && state.LastUpdate != nil {
		opts.Filter = fmt.Sprintf("lastUpdate > '%s'", state.LastUpdate.Format(time.RFC3339))
		logger.Info("incremental sync", "since", state.LastUpdate)
	} else {
		logger.Info("full sync (first run)")
	}

	// Iterate through all pages
	it := client.IterateEvidence(ctx, ev.Slug, opts)
	totalUpserted := 0

	for {
		records, err := it.Next(ctx)
		if err != nil {
			// Save error state
			saveErrorState(ctx, st, ev.Slug, state, err, logger)
			return fmt.Errorf("fetch page: %w", err)
		}
		if records == nil {
			break
		}

		upserted, err := st.UpsertRecords(ctx, ev.Table, records, ev.PrimaryKey)
		if err != nil {
			saveErrorState(ctx, st, ev.Slug, state, err, logger)
			return fmt.Errorf("upsert records: %w", err)
		}
		totalUpserted += upserted
		logger.Debug("upserted batch", "count", upserted)
	}

	// Update sync state
	now := time.Now()
	newState := store.SyncState{
		Evidence:   ev.Slug,
		LastUpdate: &now,
		LastSync:   now,
		RowCount:   int64(totalUpserted),
		Status:     "ok",
	}
	if state != nil {
		newState.RowCount = state.RowCount + int64(totalUpserted)
	}

	if err := st.SetSyncState(ctx, ev.Slug, newState); err != nil {
		return fmt.Errorf("set sync state: %w", err)
	}

	logger.Info("sync complete", "upserted", totalUpserted)
	return nil
}

func saveErrorState(ctx context.Context, st SyncStore, evidence string, current *store.SyncState, syncErr error, logger *slog.Logger) {
	state := store.SyncState{
		Evidence: evidence,
		LastSync: time.Now(),
		Status:   "error",
		ErrorMsg: syncErr.Error(),
	}
	if current != nil {
		state.LastUpdate = current.LastUpdate
		state.RowCount = current.RowCount
	}
	if err := st.SetSyncState(ctx, evidence, state); err != nil {
		logger.Error("failed to save error state", "error", err)
	}
}
