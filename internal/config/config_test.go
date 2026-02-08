package config

import (
	"testing"
	"time"
)

func TestValidate_AllFieldsSet(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidate_MissingRequired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"missing FlexibeeURL", func(c *Config) { c.FlexibeeURL = "" }},
		{"missing FlexibeeCompany", func(c *Config) { c.FlexibeeCompany = "" }},
		{"missing FlexibeeUsername", func(c *Config) { c.FlexibeeUsername = "" }},
		{"missing FlexibeePassword", func(c *Config) { c.FlexibeePassword = "" }},
		{"missing DatabaseURL", func(c *Config) { c.DatabaseURL = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := validConfig()
			tt.mutate(cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidate_InvalidValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"negative sync interval", func(c *Config) { c.SyncInterval = -1 }},
		{"zero batch size", func(c *Config) { c.SyncBatchSize = 0 }},
		{"zero concurrency", func(c *Config) { c.SyncConcurrency = 0 }},
		{"negative retention", func(c *Config) { c.RetentionDays = -1 }},
		{"bad log level", func(c *Config) { c.LogLevel = "trace" }},
		{"bad log format", func(c *Config) { c.LogFormat = "yaml" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := validConfig()
			tt.mutate(cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidate_ZeroRetentionAllowed(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.RetentionDays = 0
	if err := cfg.Validate(); err != nil {
		t.Fatalf("zero retention should be allowed, got: %v", err)
	}
}

func TestApplyEnv(t *testing.T) {
	t.Run("sets empty value", func(t *testing.T) {
		t.Setenv("TEST_APPLY_ENV_1", "hello")
		var dst string
		applyEnv(&dst, "TEST_APPLY_ENV_1")
		if dst != "hello" {
			t.Fatalf("expected hello, got %s", dst)
		}
	})

	t.Run("does not override existing", func(t *testing.T) {
		t.Setenv("TEST_APPLY_ENV_2", "world")
		dst := "existing"
		applyEnv(&dst, "TEST_APPLY_ENV_2")
		if dst != "existing" {
			t.Fatalf("expected existing, got %s", dst)
		}
	})
}

func TestApplyEnvDuration(t *testing.T) {
	t.Setenv("TEST_DUR", "10s")
	dst := 5 * time.Second
	applyEnvDuration(&dst, "TEST_DUR")
	if dst != 10*time.Second {
		t.Fatalf("expected 10s, got %v", dst)
	}
}

func TestApplyEnvInt(t *testing.T) {
	t.Setenv("TEST_INT", "42")
	dst := 10
	applyEnvInt(&dst, "TEST_INT")
	if dst != 42 {
		t.Fatalf("expected 42, got %d", dst)
	}
}

func validConfig() *Config {
	return &Config{
		FlexibeeURL:      "https://demo.flexibee.eu",
		FlexibeeCompany:  "demo",
		FlexibeeUsername:  "winstrom",
		FlexibeePassword: "winstrom",
		DatabaseURL:      "postgres://user:pass@localhost:5432/flexibee",
		SyncInterval:     5 * time.Minute,
		SyncBatchSize:    100,
		SyncConcurrency:  4,
		RetentionDays:    365,
		CleanupInterval:  24 * time.Hour,
		CleanupBatchSize: 1000,
		LogLevel:         "info",
		LogFormat:        "json",
	}
}
