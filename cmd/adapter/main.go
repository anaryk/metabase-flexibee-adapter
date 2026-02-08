package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tomasmarek/metabase-flexibee-adapter/internal/config"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/flexibee"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/registry"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/store"
	adaptersync "github.com/tomasmarek/metabase-flexibee-adapter/internal/sync"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.LogLevel, cfg.LogFormat)

	logger.Info("starting metabase-flexibee-adapter",
		"flexibee_url", cfg.FlexibeeURL,
		"company", cfg.FlexibeeCompany,
		"sync_interval", cfg.SyncInterval,
		"retention_days", cfg.RetentionDays,
	)

	// Create context with signal handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize Flexibee client
	client := flexibee.NewClient(
		cfg.FlexibeeURL,
		cfg.FlexibeeCompany,
		cfg.FlexibeeUsername,
		cfg.FlexibeePassword,
		logger,
	)

	// Initialize PostgreSQL store
	st, err := store.NewStore(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer st.Close()

	// Initialize evidence registry
	reg := registry.NewDefault()
	logger.Info("registered evidence types", "count", reg.Len())

	// Initialize cleanup
	cleaner := adaptersync.NewCleaner(st, reg, adaptersync.CleanupConfig{
		RetentionDays: cfg.RetentionDays,
		BatchSize:     cfg.CleanupBatchSize,
	}, logger)

	// Initialize and start sync engine
	engine := adaptersync.NewEngine(client, st, reg, cleaner, adaptersync.EngineConfig{
		SyncInterval:    cfg.SyncInterval,
		CleanupInterval: cfg.CleanupInterval,
		BatchSize:       cfg.SyncBatchSize,
		Concurrency:     cfg.SyncConcurrency,
	}, logger)

	if err := engine.Start(ctx); err != nil {
		logger.Error("engine stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("adapter stopped gracefully")
}

func setupLogger(level, format string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
