package sync

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/flexibee"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/registry"
	"github.com/tomasmarek/metabase-flexibee-adapter/internal/store"
)

var discardLogger = slog.New(slog.DiscardHandler)

func TestSyncEvidence_FirstSync(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"winstrom": {
				"@version": "1.0",
				"@rowCount": "2",
				"test": [
					{"id": 1, "kod": "T001"},
					{"id": 2, "kod": "T002"}
				]
			}
		}`))
	}))
	t.Cleanup(srv.Close)

	client := flexibee.NewClient(srv.URL, "demo", "user", "pass", discardLogger)
	ms := newMockSyncStore()

	ev := registry.Evidence{
		Slug:       "test",
		Table:      "flexibee_test",
		PrimaryKey: "id",
	}

	err := syncEvidence(context.Background(), client, ms, ev, 100, discardLogger)
	assert.NoError(t, err)

	state := ms.states["test"]
	assert.NotNil(t, state)
	assert.Equal(t, "ok", state.Status)
	assert.Equal(t, int64(2), state.RowCount)
	assert.Equal(t, 2, ms.upsertCount["flexibee_test"])
}

func TestSyncEvidence_IncrementalSync(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")

		w.Header().Set("Content-Type", "application/json")
		if filter != "" {
			_, _ = w.Write([]byte(`{
				"winstrom": {
					"@version": "1.0",
					"@rowCount": "1",
					"test": [{"id": 3, "kod": "T003"}]
				}
			}`))
		} else {
			_, _ = w.Write([]byte(`{
				"winstrom": {
					"@version": "1.0",
					"@rowCount": "0",
					"test": []
				}
			}`))
		}
	}))
	t.Cleanup(srv.Close)

	client := flexibee.NewClient(srv.URL, "demo", "user", "pass", discardLogger)
	ms := newMockSyncStore()

	// Set existing sync state to simulate previous sync
	lastUpdate := time.Now().Add(-time.Hour)
	ms.states["test"] = &store.SyncState{
		Evidence:   "test",
		LastUpdate: &lastUpdate,
		LastSync:   lastUpdate,
		RowCount:   2,
		Status:     "ok",
	}

	ev := registry.Evidence{
		Slug:       "test",
		Table:      "flexibee_test",
		PrimaryKey: "id",
	}

	err := syncEvidence(context.Background(), client, ms, ev, 100, discardLogger)
	assert.NoError(t, err)
	assert.Equal(t, 1, ms.upsertCount["flexibee_test"])

	// Row count should be cumulative
	state := ms.states["test"]
	assert.Equal(t, int64(3), state.RowCount)
}

func TestSyncEvidence_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`bad request`))
	}))
	t.Cleanup(srv.Close)

	client := flexibee.NewClient(srv.URL, "demo", "user", "pass", discardLogger)
	ms := newMockSyncStore()

	ev := registry.Evidence{
		Slug:       "test",
		Table:      "flexibee_test",
		PrimaryKey: "id",
	}

	err := syncEvidence(context.Background(), client, ms, ev, 100, discardLogger)
	assert.Error(t, err)

	state := ms.states["test"]
	if state != nil {
		assert.Equal(t, "error", state.Status)
	}
}

// mockSyncStore implements SyncStore for testing.
type mockSyncStore struct {
	states      map[string]*store.SyncState
	upsertCount map[string]int
	cleanups    map[string]int64
}

func newMockSyncStore() *mockSyncStore {
	return &mockSyncStore{
		states:      make(map[string]*store.SyncState),
		upsertCount: make(map[string]int),
		cleanups:    make(map[string]int64),
	}
}

func (m *mockSyncStore) GetSyncState(_ context.Context, evidence string) (*store.SyncState, error) {
	state, ok := m.states[evidence]
	if !ok {
		return nil, nil
	}
	return state, nil
}

func (m *mockSyncStore) SetSyncState(_ context.Context, evidence string, state store.SyncState) error {
	m.states[evidence] = &state
	return nil
}

func (m *mockSyncStore) UpsertRecords(_ context.Context, table string, records []map[string]any, _ string) (int, error) {
	m.upsertCount[table] += len(records)
	return len(records), nil
}

func (m *mockSyncStore) CleanupOldRecords(_ context.Context, table string, _ time.Time, _ int) (int64, error) {
	deleted := m.cleanups[table]
	return deleted, nil
}

func (m *mockSyncStore) LogCleanup(_ context.Context, _ string, _ int64, _ *time.Time) error {
	return nil
}
