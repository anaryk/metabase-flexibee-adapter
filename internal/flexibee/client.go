package flexibee

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	maxRetries     = 3
	initialBackoff = 500 * time.Millisecond
)

// Client communicates with the Flexibee REST API.
type Client struct {
	baseURL    string
	company    string
	httpClient *http.Client
	username   string
	password   string
	logger     *slog.Logger
}

// NewClient creates a new Flexibee API client.
func NewClient(baseURL, company, username, password string, logger *slog.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		company: company,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		username: username,
		password: password,
		logger:   logger,
	}
}

// FetchEvidence retrieves records from a single evidence endpoint.
func (c *Client) FetchEvidence(ctx context.Context, evidence string, opts FetchOptions) (*Response, error) {
	u := c.buildURL(evidence, opts)

	body, err := c.doRequest(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("fetch evidence %s: %w", evidence, err)
	}

	resp, err := parseResponse(body, evidence)
	if err != nil {
		return nil, fmt.Errorf("parse evidence %s: %w", evidence, err)
	}

	return resp, nil
}

// FetchEvidenceProperties returns the property definitions for an evidence type.
func (c *Client) FetchEvidenceProperties(ctx context.Context, evidence string) ([]Property, error) {
	u := fmt.Sprintf("%s/c/%s/%s/properties.json", c.baseURL, c.company, evidence)

	body, err := c.doRequest(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("fetch properties for %s: %w", evidence, err)
	}

	props, err := parseProperties(body)
	if err != nil {
		return nil, fmt.Errorf("parse properties for %s: %w", evidence, err)
	}

	return props, nil
}

func (c *Client) buildURL(evidence string, opts FetchOptions) string {
	u := fmt.Sprintf("%s/c/%s/%s.json", c.baseURL, c.company, evidence)

	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Start > 0 {
		params.Set("start", strconv.Itoa(opts.Start))
	}
	if opts.Detail != "" {
		params.Set("detail", opts.Detail)
	}
	if opts.Filter != "" {
		params.Set("filter", opts.Filter)
	}
	if opts.AddRowCount {
		params.Set("add-row-count", "true")
	}

	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return u
}

func (c *Client) doRequest(ctx context.Context, url string) ([]byte, error) {
	var lastErr error

	for attempt := range maxRetries {
		if attempt > 0 {
			backoff := initialBackoff * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.SetBasicAuth(c.username, c.password)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request (attempt %d): %w", attempt+1, err)
			c.logger.Warn("request failed, retrying", "attempt", attempt+1, "error", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read body (attempt %d): %w", attempt+1, err)
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error %d (attempt %d)", resp.StatusCode, attempt+1)
			c.logger.Warn("server error, retrying", "status", resp.StatusCode, "attempt", attempt+1)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}

		return body, nil
	}

	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}
