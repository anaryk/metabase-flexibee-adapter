package store

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/flexibee"
)

// FlexibeeTypeToPG maps a Flexibee property type to a PostgreSQL column type.
func FlexibeeTypeToPG(prop flexibee.Property) string {
	switch prop.Type {
	case "integer":
		return "BIGINT"
	case "numeric":
		return "NUMERIC"
	case "date":
		return "DATE"
	case "datetime":
		return "TIMESTAMPTZ"
	case "logic":
		return "BOOLEAN"
	case "string":
		return "TEXT"
	case "relation":
		return "TEXT"
	default:
		return "TEXT"
	}
}

// EnsureTable creates a table if it doesn't exist and adds any new columns
// based on the Flexibee property definitions.
func EnsureTable(ctx context.Context, pool *pgxpool.Pool, table string, properties []flexibee.Property, logger *slog.Logger) error {
	// Sanitize table name
	safeTable := sanitizeIdentifier(table)

	// Create table with base columns if it doesn't exist
	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT PRIMARY KEY,
			raw_data JSONB,
			synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`, safeTable)

	if _, err := pool.Exec(ctx, createSQL); err != nil {
		return fmt.Errorf("create table %s: %w", table, err)
	}

	// Get existing columns
	existing, err := getExistingColumns(ctx, pool, table)
	if err != nil {
		return fmt.Errorf("get columns for %s: %w", table, err)
	}

	// Add missing columns
	for _, prop := range properties {
		colName := sanitizeIdentifier(prop.Name)
		if existing[prop.Name] {
			continue
		}
		// Skip the id column (already created as primary key)
		if prop.Name == "id" {
			continue
		}

		pgType := FlexibeeTypeToPG(prop)
		alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s", safeTable, colName, pgType)
		if _, err := pool.Exec(ctx, alterSQL); err != nil {
			logger.Warn("failed to add column", "table", table, "column", prop.Name, "error", err)
			continue
		}
		logger.Debug("added column", "table", table, "column", prop.Name, "type", pgType)
	}

	return nil
}

func getExistingColumns(ctx context.Context, pool *pgxpool.Pool, table string) (map[string]bool, error) {
	rows, err := pool.Query(ctx,
		"SELECT column_name FROM information_schema.columns WHERE table_name = $1",
		table,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		cols[name] = true
	}
	return cols, rows.Err()
}

// sanitizeIdentifier ensures an identifier is safe for use in SQL.
// It wraps the identifier in double quotes to handle reserved words and special characters.
func sanitizeIdentifier(name string) string {
	// Replace any double quotes in the name to prevent SQL injection
	safe := strings.ReplaceAll(name, "\"", "")
	// Replace hyphens with underscores for PostgreSQL compatibility
	safe = strings.ReplaceAll(safe, "-", "_")
	return fmt.Sprintf("%q", safe)
}
