package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrationSQL_Embedded(t *testing.T) {
	t.Parallel()
	// Verify the migration SQL is properly embedded and contains expected statements
	assert.Contains(t, migrationSQL, "CREATE TABLE IF NOT EXISTS sync_state")
	assert.Contains(t, migrationSQL, "CREATE TABLE IF NOT EXISTS cleanup_log")
	assert.Contains(t, migrationSQL, "evidence    TEXT PRIMARY KEY")
	assert.Contains(t, migrationSQL, "rows_deleted BIGINT NOT NULL")
}
