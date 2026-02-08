package flexibee

import (
	"context"
	"fmt"
)

// PageIterator fetches evidence records page by page.
type PageIterator struct {
	client   *Client
	evidence string
	opts     FetchOptions
	done     bool
	total    *int
	fetched  int
}

// IterateEvidence returns a PageIterator for paginated fetching.
func (c *Client) IterateEvidence(ctx context.Context, evidence string, opts FetchOptions) *PageIterator {
	if opts.Limit <= 0 {
		opts.Limit = 100
	}
	opts.AddRowCount = true
	opts.Start = 0
	return &PageIterator{
		client:   c,
		evidence: evidence,
		opts:     opts,
	}
}

// Next returns the next page of records. Returns nil, nil when exhausted.
func (it *PageIterator) Next(ctx context.Context) ([]map[string]any, error) {
	if it.done {
		return nil, nil
	}

	it.opts.Start = it.fetched

	resp, err := it.client.FetchEvidence(ctx, it.evidence, it.opts)
	if err != nil {
		return nil, fmt.Errorf("fetch page at offset %d: %w", it.opts.Start, err)
	}

	if it.total == nil && resp.Winstrom.RowCount != nil {
		it.total = resp.Winstrom.RowCount
	}

	records := resp.Winstrom.Records
	if len(records) == 0 {
		it.done = true
		return nil, nil
	}

	it.fetched += len(records)

	// Done if we got fewer than the page size, or reached total.
	if len(records) < it.opts.Limit {
		it.done = true
	}
	if it.total != nil && it.fetched >= *it.total {
		it.done = true
	}

	return records, nil
}
