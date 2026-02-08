package flexibee

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var discardLogger = slog.New(slog.DiscardHandler)

func TestFetchEvidence_BasicAuth(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok, "basic auth must be set")
		assert.Equal(t, "testuser", user)
		assert.Equal(t, "testpass", pass)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"winstrom":{"@version":"1.0","prodejka":[{"id":1}]}}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "testuser", "testpass", discardLogger)
	resp, err := c.FetchEvidence(context.Background(), "prodejka", FetchOptions{})
	require.NoError(t, err)
	assert.Len(t, resp.Winstrom.Records, 1)
}

func TestFetchEvidence_QueryParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "50", q.Get("limit"))
		assert.Equal(t, "10", q.Get("start"))
		assert.Equal(t, "full", q.Get("detail"))
		assert.Equal(t, "true", q.Get("add-row-count"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"winstrom":{"@version":"1.0","prodejka":[]}}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "user", "pass", discardLogger)
	_, err := c.FetchEvidence(context.Background(), "prodejka", FetchOptions{
		Limit:       50,
		Start:       10,
		Detail:      "full",
		AddRowCount: true,
	})
	require.NoError(t, err)
}

func TestFetchEvidence_ParsesRecords(t *testing.T) {
	t.Parallel()

	body := `{
		"winstrom": {
			"@version": "1.0",
			"@rowCount": "3",
			"faktura-vydana": [
				{"id": 1, "kod": "FV-001"},
				{"id": 2, "kod": "FV-002"},
				{"id": 3, "kod": "FV-003"}
			]
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "user", "pass", discardLogger)
	resp, err := c.FetchEvidence(context.Background(), "faktura-vydana", FetchOptions{})
	require.NoError(t, err)
	assert.Len(t, resp.Winstrom.Records, 3)
	assert.Equal(t, "FV-001", resp.Winstrom.Records[0]["kod"])
}

func TestFetchEvidence_RetriesOnServerError(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"winstrom":{"@version":"1.0","prodejka":[{"id":1}]}}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "user", "pass", discardLogger)
	resp, err := c.FetchEvidence(context.Background(), "prodejka", FetchOptions{})
	require.NoError(t, err)
	assert.Len(t, resp.Winstrom.Records, 1)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestFetchEvidence_FailsOnClientError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`unauthorized`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "user", "pass", discardLogger)
	_, err := c.FetchEvidence(context.Background(), "prodejka", FetchOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestFetchEvidence_ContextCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewClient(srv.URL, "demo", "user", "pass", discardLogger)
	_, err := c.FetchEvidence(ctx, "prodejka", FetchOptions{})
	require.Error(t, err)
}

func TestFetchEvidenceProperties(t *testing.T) {
	t.Parallel()

	body := `{
		"properties": {
			"property": [
				{"propertyName": "id", "type": "integer", "maxLength": 0, "mandatory": true, "isReadOnly": true},
				{"propertyName": "kod", "type": "string", "maxLength": 20, "mandatory": true, "isReadOnly": false}
			]
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/properties.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "user", "pass", discardLogger)
	props, err := c.FetchEvidenceProperties(context.Background(), "prodejka")
	require.NoError(t, err)
	assert.Len(t, props, 2)
	assert.Equal(t, "id", props[0].Name)
	assert.Equal(t, "integer", props[0].Type)
	assert.Equal(t, "kod", props[1].Name)
	assert.Equal(t, 20, props[1].MaxLength)
}
