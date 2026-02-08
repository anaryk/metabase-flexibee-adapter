package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_sync_state.sql
var migrationSQL string

// SyncState tracks the last sync state for an evidence type.
type SyncState struct {
	Evidence   string
	LastUpdate *time.Time
	LastSync   time.Time
	RowCount   int64
	Status     string
	ErrorMsg   string
}

// Store manages PostgreSQL operations for synced Flexibee data.
type Store struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewStore creates a new Store with a connection pool.
func NewStore(ctx context.Context, databaseURL string, logger *slog.Logger) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Store{pool: pool, logger: logger}, nil
}

// Pool returns the underlying connection pool (for schema operations).
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// RunMigrations executes the embedded migration SQL.
func (s *Store) RunMigrations(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, migrationSQL); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	s.logger.Info("migrations applied successfully")
	return nil
}

// UpsertRecords inserts or updates records in the given table.
// Returns the number of records upserted.
func (s *Store) UpsertRecords(ctx context.Context, table string, records []map[string]any, primaryKey string) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	safeTable := sanitizeIdentifier(table)
	safePK := sanitizeIdentifier(primaryKey)
	count := 0

	for _, record := range records {
		rawJSON, err := json.Marshal(record)
		if err != nil {
			s.logger.Warn("failed to marshal record", "error", err)
			continue
		}

		id, ok := record[primaryKey]
		if !ok {
			s.logger.Warn("record missing primary key", "key", primaryKey)
			continue
		}

		// Build column names and values for the upsert
		cols := []string{safePK, sanitizeIdentifier("raw_data"), sanitizeIdentifier("synced_at")}
		placeholders := []string{"$1", "$2", "NOW()"}
		updates := []string{
			fmt.Sprintf("%s = $2", sanitizeIdentifier("raw_data")),
			fmt.Sprintf("%s = NOW()", sanitizeIdentifier("synced_at")),
		}
		args := []any{id, rawJSON}

		argIdx := 3
		for k, v := range record {
			if k == primaryKey {
				continue
			}
			safeCol := sanitizeIdentifier(k)
			cols = append(cols, safeCol)
			placeholders = append(placeholders, fmt.Sprintf("$%d", argIdx))
			updates = append(updates, fmt.Sprintf("%s = $%d", safeCol, argIdx))
			args = append(args, v)
			argIdx++
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
			safeTable,
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
			safePK,
			strings.Join(updates, ", "),
		)

		if _, err := s.pool.Exec(ctx, query, args...); err != nil {
			s.logger.Warn("failed to upsert record", "table", table, "id", id, "error", err)
			continue
		}
		count++
	}

	return count, nil
}

// DeleteRecords removes records by their primary key values.
func (s *Store) DeleteRecords(ctx context.Context, table string, ids []any) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	safeTable := sanitizeIdentifier(table)
	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE \"id\" IN (%s)",
		safeTable,
		strings.Join(placeholders, ", "),
	)

	tag, err := s.pool.Exec(ctx, query, ids...)
	if err != nil {
		return 0, fmt.Errorf("delete records from %s: %w", table, err)
	}

	return int(tag.RowsAffected()), nil
}

// GetSyncState returns the sync state for an evidence type.
func (s *Store) GetSyncState(ctx context.Context, evidence string) (*SyncState, error) {
	var state SyncState
	err := s.pool.QueryRow(ctx,
		"SELECT evidence, last_update, last_sync, row_count, status, COALESCE(error_msg, '') FROM sync_state WHERE evidence = $1",
		evidence,
	).Scan(&state.Evidence, &state.LastUpdate, &state.LastSync, &state.RowCount, &state.Status, &state.ErrorMsg)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("get sync state for %s: %w", evidence, err)
	}

	return &state, nil
}

// SetSyncState creates or updates the sync state for an evidence type.
func (s *Store) SetSyncState(ctx context.Context, evidence string, state SyncState) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sync_state (evidence, last_update, last_sync, row_count, status, error_msg)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (evidence) DO UPDATE SET
			last_update = $2, last_sync = $3, row_count = $4, status = $5, error_msg = $6
	`, evidence, state.LastUpdate, state.LastSync, state.RowCount, state.Status, state.ErrorMsg)

	if err != nil {
		return fmt.Errorf("set sync state for %s: %w", evidence, err)
	}
	return nil
}

// CleanupOldRecords deletes records older than the given time in batches.
// Returns total number of deleted rows.
func (s *Store) CleanupOldRecords(ctx context.Context, table string, olderThan time.Time, batchSize int) (int64, error) {
	safeTable := sanitizeIdentifier(table)
	var totalDeleted int64

	for {
		query := fmt.Sprintf(
			`DELETE FROM %s WHERE ctid IN (
				SELECT ctid FROM %s WHERE "synced_at" < $1 LIMIT $2
			)`,
			safeTable, safeTable,
		)

		tag, err := s.pool.Exec(ctx, query, olderThan, batchSize)
		if err != nil {
			return totalDeleted, fmt.Errorf("cleanup %s: %w", table, err)
		}

		deleted := tag.RowsAffected()
		totalDeleted += deleted

		if deleted < int64(batchSize) {
			break
		}
	}

	return totalDeleted, nil
}

// LogCleanup records a cleanup operation.
func (s *Store) LogCleanup(ctx context.Context, evidence string, rowsDeleted int64, oldestKept *time.Time) error {
	_, err := s.pool.Exec(ctx,
		"INSERT INTO cleanup_log (evidence, rows_deleted, oldest_kept) VALUES ($1, $2, $3)",
		evidence, rowsDeleted, oldestKept,
	)
	if err != nil {
		return fmt.Errorf("log cleanup for %s: %w", evidence, err)
	}
	return nil
}

// Close closes the connection pool.
func (s *Store) Close() {
	s.pool.Close()
}
