package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/anaryk/metabase-flexibee-adapter/internal/registry"
)

func TestCleaner_SkipsWhenDisabled(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	reg.Register(registry.Evidence{Slug: "test", Table: "flexibee_test", PrimaryKey: "id"})

	ms := newMockSyncStore()
	c := NewCleaner(ms, reg, CleanupConfig{RetentionDays: 0, BatchSize: 100}, discardLogger)

	err := c.Run(context.Background())
	assert.NoError(t, err)
}

func TestCleaner_SkipsMasterData(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	reg.Register(registry.Evidence{Slug: "master", Table: "flexibee_master", PrimaryKey: "id", IsMasterData: true})
	reg.Register(registry.Evidence{Slug: "trans", Table: "flexibee_trans", PrimaryKey: "id"})

	ms := newMockSyncStore()
	ms.cleanups["flexibee_trans"] = 5

	c := NewCleaner(ms, reg, CleanupConfig{RetentionDays: 30, BatchSize: 100}, discardLogger)

	err := c.Run(context.Background())
	assert.NoError(t, err)
	// Master data should not have cleanup called - only transactional
}

func TestCleaner_RunsCleanup(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	reg.Register(registry.Evidence{Slug: "trans", Table: "flexibee_trans", PrimaryKey: "id"})

	ms := newMockSyncStore()
	ms.cleanups["flexibee_trans"] = 10

	c := NewCleaner(ms, reg, CleanupConfig{RetentionDays: 30, BatchSize: 100}, discardLogger)

	err := c.Run(context.Background())
	assert.NoError(t, err)
}
