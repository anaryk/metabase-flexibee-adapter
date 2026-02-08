package sync

import (
	"context"
	"log/slog"
	"time"

	"github.com/anaryk/metabase-flexibee-adapter/internal/registry"
)

// CleanupConfig controls data retention behavior.
type CleanupConfig struct {
	RetentionDays int
	BatchSize     int
}

// Cleaner handles data retention cleanup.
type Cleaner struct {
	store    SyncStore
	registry *registry.Registry
	config   CleanupConfig
	logger   *slog.Logger
}

// NewCleaner creates a new Cleaner instance.
func NewCleaner(st SyncStore, reg *registry.Registry, cfg CleanupConfig, logger *slog.Logger) *Cleaner {
	return &Cleaner{
		store:    st,
		registry: reg,
		config:   cfg,
		logger:   logger,
	}
}

// Run performs cleanup for all registered evidence types.
func (c *Cleaner) Run(ctx context.Context) error {
	if c.config.RetentionDays <= 0 {
		c.logger.Info("cleanup disabled (retention_days=0)")
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -c.config.RetentionDays)
	c.logger.Info("starting cleanup", "cutoff", cutoff, "retention_days", c.config.RetentionDays)

	for _, ev := range c.registry.All() {
		if ev.IsMasterData {
			c.logger.Debug("skipping master data", "evidence", ev.Slug)
			continue
		}

		deleted, err := c.store.CleanupOldRecords(ctx, ev.Table, cutoff, c.config.BatchSize)
		if err != nil {
			c.logger.Error("cleanup failed", "evidence", ev.Slug, "error", err)
			continue
		}

		if deleted > 0 {
			c.logger.Info("cleaned up records", "evidence", ev.Slug, "deleted", deleted)
			if err := c.store.LogCleanup(ctx, ev.Slug, deleted, &cutoff); err != nil {
				c.logger.Error("failed to log cleanup", "evidence", ev.Slug, "error", err)
			}
		}
	}

	c.logger.Info("cleanup complete")
	return nil
}
