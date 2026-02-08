package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Flexibee
	FlexibeeURL      string
	FlexibeeCompany  string
	FlexibeeUsername string
	FlexibeePassword string

	// PostgreSQL
	DatabaseURL string

	// Sync
	SyncInterval    time.Duration
	SyncBatchSize   int
	SyncConcurrency int

	// Cleanup / Data Retention
	RetentionDays    int
	CleanupInterval  time.Duration
	CleanupBatchSize int

	// Logging
	LogLevel  string
	LogFormat string
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Define flags with defaults
	flag.StringVar(&cfg.FlexibeeURL, "flexibee-url", "", "Flexibee base URL")
	flag.StringVar(&cfg.FlexibeeCompany, "flexibee-company", "", "Flexibee company code")
	flag.StringVar(&cfg.FlexibeeUsername, "flexibee-username", "", "Flexibee username")
	flag.StringVar(&cfg.FlexibeePassword, "flexibee-password", "", "Flexibee password")
	flag.StringVar(&cfg.DatabaseURL, "database-url", "", "PostgreSQL connection URL")
	flag.DurationVar(&cfg.SyncInterval, "sync-interval", 5*time.Minute, "Sync interval")
	flag.IntVar(&cfg.SyncBatchSize, "sync-batch-size", 100, "Records per page when fetching")
	flag.IntVar(&cfg.SyncConcurrency, "sync-concurrency", 4, "Max concurrent evidence syncs")
	flag.IntVar(&cfg.RetentionDays, "retention-days", 365, "Data retention in days (0=disabled)")
	flag.DurationVar(&cfg.CleanupInterval, "cleanup-interval", 24*time.Hour, "Cleanup check interval")
	flag.IntVar(&cfg.CleanupBatchSize, "cleanup-batch-size", 1000, "Rows to delete per batch")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.LogFormat, "log-format", "json", "Log format (json, text)")

	flag.Parse()

	// Env vars override flags only when the flag was not explicitly set
	applyEnv(&cfg.FlexibeeURL, "FLEXIBEE_URL")
	applyEnv(&cfg.FlexibeeCompany, "FLEXIBEE_COMPANY")
	applyEnv(&cfg.FlexibeeUsername, "FLEXIBEE_USERNAME")
	applyEnv(&cfg.FlexibeePassword, "FLEXIBEE_PASSWORD")
	applyEnv(&cfg.DatabaseURL, "DATABASE_URL")
	applyEnvDuration(&cfg.SyncInterval, "SYNC_INTERVAL")
	applyEnvInt(&cfg.SyncBatchSize, "SYNC_BATCH_SIZE")
	applyEnvInt(&cfg.SyncConcurrency, "SYNC_CONCURRENCY")
	applyEnvInt(&cfg.RetentionDays, "RETENTION_DAYS")
	applyEnvDuration(&cfg.CleanupInterval, "CLEANUP_INTERVAL")
	applyEnvInt(&cfg.CleanupBatchSize, "CLEANUP_BATCH_SIZE")
	applyEnv(&cfg.LogLevel, "LOG_LEVEL")
	applyEnv(&cfg.LogFormat, "LOG_FORMAT")

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var errs []error

	if c.FlexibeeURL == "" {
		errs = append(errs, fmt.Errorf("flexibee URL is required (FLEXIBEE_URL or --flexibee-url)"))
	}
	if c.FlexibeeCompany == "" {
		errs = append(errs, fmt.Errorf("flexibee company is required (FLEXIBEE_COMPANY or --flexibee-company)"))
	}
	if c.FlexibeeUsername == "" {
		errs = append(errs, fmt.Errorf("flexibee username is required (FLEXIBEE_USERNAME or --flexibee-username)"))
	}
	if c.FlexibeePassword == "" {
		errs = append(errs, fmt.Errorf("flexibee password is required (FLEXIBEE_PASSWORD or --flexibee-password)"))
	}
	if c.DatabaseURL == "" {
		errs = append(errs, fmt.Errorf("database URL is required (DATABASE_URL or --database-url)"))
	}
	if c.SyncInterval <= 0 {
		errs = append(errs, fmt.Errorf("sync interval must be positive"))
	}
	if c.SyncBatchSize <= 0 {
		errs = append(errs, fmt.Errorf("sync batch size must be positive"))
	}
	if c.SyncConcurrency <= 0 {
		errs = append(errs, fmt.Errorf("sync concurrency must be positive"))
	}
	if c.RetentionDays < 0 {
		errs = append(errs, fmt.Errorf("retention days must be non-negative"))
	}
	if c.CleanupInterval <= 0 {
		errs = append(errs, fmt.Errorf("cleanup interval must be positive"))
	}
	if c.CleanupBatchSize <= 0 {
		errs = append(errs, fmt.Errorf("cleanup batch size must be positive"))
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		errs = append(errs, fmt.Errorf("log level must be one of: debug, info, warn, error"))
	}

	switch c.LogFormat {
	case "json", "text":
	default:
		errs = append(errs, fmt.Errorf("log format must be one of: json, text"))
	}

	return errors.Join(errs...)
}

func applyEnv(dst *string, key string) {
	if v := os.Getenv(key); v != "" && *dst == "" {
		*dst = v
	}
}

func applyEnvDuration(dst *time.Duration, key string) {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			*dst = d
		}
	}
}

func applyEnvInt(dst *int, key string) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*dst = n
		}
	}
}
