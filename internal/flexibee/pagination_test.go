package flexibee

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageIterator_MultiplePages(t *testing.T) {
	t.Parallel()

	// 5 records total, page size 2 → 3 pages (2, 2, 1)
	allRecords := []map[string]any{
		{"id": float64(1)}, {"id": float64(2)},
		{"id": float64(3)}, {"id": float64(4)},
		{"id": float64(5)},
	}
	total := len(allRecords)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		start := 0
		if s := q.Get("start"); s != "" {
			start, _ = strconv.Atoi(s)
		}
		limit := 2
		if l := q.Get("limit"); l != "" {
			limit, _ = strconv.Atoi(l)
		}

		end := start + limit
		if end > total {
			end = total
		}
		page := allRecords[start:end]

		records := "["
		for i, r := range page {
			if i > 0 {
				records += ","
			}
			records += fmt.Sprintf(`{"id":%v}`, r["id"])
		}
		records += "]"

		body := fmt.Sprintf(`{"winstrom":{"@version":"1.0","@rowCount":%d,"test":%s}}`, total, records)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "user", "pass", slog.New(slog.DiscardHandler))
	ctx := context.Background()
	it := c.IterateEvidence(ctx, "test", FetchOptions{Limit: 2})

	var collected []map[string]any
	for {
		page, err := it.Next(ctx)
		require.NoError(t, err)
		if page == nil {
			break
		}
		collected = append(collected, page...)
	}

	assert.Len(t, collected, 5)
}

func TestPageIterator_EmptyEvidence(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"winstrom":{"@version":"1.0","@rowCount":0,"test":[]}}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "user", "pass", slog.New(slog.DiscardHandler))
	ctx := context.Background()
	it := c.IterateEvidence(ctx, "test", FetchOptions{Limit: 10})

	page, err := it.Next(ctx)
	require.NoError(t, err)
	assert.Nil(t, page)

	// Subsequent calls also return nil
	page, err = it.Next(ctx)
	require.NoError(t, err)
	assert.Nil(t, page)
}

func TestPageIterator_ErrorMidPage(t *testing.T) {
	t.Parallel()

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"winstrom":{"@version":"1.0","@rowCount":10,"test":[{"id":1},{"id":2}]}}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`bad request`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "user", "pass", slog.New(slog.DiscardHandler))
	ctx := context.Background()
	it := c.IterateEvidence(ctx, "test", FetchOptions{Limit: 2})

	page, err := it.Next(ctx)
	require.NoError(t, err)
	assert.Len(t, page, 2)

	_, err = it.Next(ctx)
	require.Error(t, err)
}

func TestPageIterator_SinglePage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"winstrom":{"@version":"1.0","@rowCount":2,"test":[{"id":1},{"id":2}]}}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "demo", "user", "pass", slog.New(slog.DiscardHandler))
	ctx := context.Background()
	it := c.IterateEvidence(ctx, "test", FetchOptions{Limit: 10})

	page, err := it.Next(ctx)
	require.NoError(t, err)
	assert.Len(t, page, 2)

	page, err = it.Next(ctx)
	require.NoError(t, err)
	assert.Nil(t, page)
}
