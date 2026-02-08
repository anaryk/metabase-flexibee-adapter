package sync

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tomasmarek/metabase-flexibee-adapter/internal/flexibee"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/registry"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/store"
)

// Engine orchestrates syncing Flexibee data to PostgreSQL.
type Engine struct {
	client    *flexibee.Client
	store     *store.Store
	syncStore SyncStore
	registry  *registry.Registry
	cleaner   *Cleaner
	logger    *slog.Logger

	syncInterval    time.Duration
	cleanupInterval time.Duration
	batchSize       int
	concurrency     int
}

// EngineConfig holds the engine's configuration values.
type EngineConfig struct {
	SyncInterval    time.Duration
	CleanupInterval time.Duration
	BatchSize       int
	Concurrency     int
}

// NewEngine creates a new sync engine.
func NewEngine(client *flexibee.Client, st *store.Store, reg *registry.Registry, cleaner *Cleaner, cfg EngineConfig, logger *slog.Logger) *Engine {
	return &Engine{
		client:          client,
		store:           st,
		syncStore:       st,
		registry:        reg,
		cleaner:         cleaner,
		logger:          logger,
		syncInterval:    cfg.SyncInterval,
		cleanupInterval: cfg.CleanupInterval,
		batchSize:       cfg.BatchSize,
		concurrency:     cfg.Concurrency,
	}
}

// Start runs the sync engine until the context is cancelled.
// It runs migrations, ensures tables, performs an initial sync,
// then runs periodic sync and cleanup.
func (e *Engine) Start(ctx context.Context) error {
	// Run migrations
	e.logger.Info("running migrations")
	if err := e.store.RunMigrations(ctx); err != nil {
		return err
	}

	// Ensure tables exist for all registered evidence types
	e.logger.Info("ensuring tables for registered evidence types")
	if err := e.ensureTables(ctx); err != nil {
		return err
	}

	// Run initial sync
	e.logger.Info("running initial sync")
	if err := e.RunOnce(ctx); err != nil {
		e.logger.Error("initial sync failed", "error", err)
		// Don't return - continue with periodic sync
	}

	// Start periodic sync and cleanup
	syncTicker := time.NewTicker(e.syncInterval)
	defer syncTicker.Stop()

	cleanupTicker := time.NewTicker(e.cleanupInterval)
	defer cleanupTicker.Stop()

	e.logger.Info("engine started", "sync_interval", e.syncInterval, "cleanup_interval", e.cleanupInterval)

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("engine shutting down")
			return nil
		case <-syncTicker.C:
			e.logger.Info("starting periodic sync")
			if err := e.RunOnce(ctx); err != nil {
				e.logger.Error("periodic sync failed", "error", err)
			}
		case <-cleanupTicker.C:
			e.logger.Info("starting periodic cleanup")
			if err := e.cleaner.Run(ctx); err != nil {
				e.logger.Error("periodic cleanup failed", "error", err)
			}
		}
	}
}

// RunOnce performs a single sync pass across all registered evidence types
// with bounded concurrency.
func (e *Engine) RunOnce(ctx context.Context) error {
	evidences := e.registry.All()
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.concurrency)

	for _, ev := range evidences {
		g.Go(func() error {
			return syncEvidence(ctx, e.client, e.syncStore, ev, e.batchSize, e.logger)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	e.logger.Info("sync pass complete", "evidence_count", len(evidences))
	return nil
}

func (e *Engine) ensureTables(ctx context.Context) error {
	for _, ev := range e.registry.All() {
		props, err := e.client.FetchEvidenceProperties(ctx, ev.Slug)
		if err != nil {
			e.logger.Warn("failed to fetch properties, creating table with base columns only",
				"evidence", ev.Slug, "error", err)
			props = nil
		}

		if err := store.EnsureTable(ctx, e.store.Pool(), ev.Table, props, e.logger); err != nil {
			return err
		}
	}
	return nil
}
