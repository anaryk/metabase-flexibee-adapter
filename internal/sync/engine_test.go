package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEngineConfig_Defaults(t *testing.T) {
	t.Parallel()

	cfg := EngineConfig{
		SyncInterval:    5 * time.Minute,
		CleanupInterval: 24 * time.Hour,
		BatchSize:       100,
		Concurrency:     4,
	}

	assert.Equal(t, 5*time.Minute, cfg.SyncInterval)
	assert.Equal(t, 24*time.Hour, cfg.CleanupInterval)
	assert.Equal(t, 100, cfg.BatchSize)
	assert.Equal(t, 4, cfg.Concurrency)
}
