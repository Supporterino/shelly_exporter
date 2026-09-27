package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const defaultTimeout = 10 * time.Second

// Fetcher fetches and decodes a Shelly RPC endpoint into result.
type Fetcher interface {
	FetchData(ctx context.Context, endpoint string, result any) error
}

// Option configures an APIClient.
type Option func(*APIClient)

// WithHTTPClient injects a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *APIClient) {
		if httpClient != nil {
			c.Client = httpClient
		}
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *APIClient) {
		c.Client.Timeout = timeout
	}
}

// WithCredentials configures the device credentials used for authentication.
func WithCredentials(username, password string) Option {
	return func(c *APIClient) {
		c.username = username
		c.password = password
	}
}

// APIClient is an HTTP client for the Shelly RPC API.
type APIClient struct {
	BaseURL string
	Client  *http.Client

	username string
	password string
}

// NewAPIClient initializes an APIClient for the given host.
func NewAPIClient(baseURL string, opts ...Option) *APIClient {
	c := &APIClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}

	base := c.Client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	if c.username != "" || c.password != "" {
		c.Client.Transport = newAuthTransport(base, c.username, c.password)
	}

	return c
}

// FetchData makes a GET request to the specified endpoint and parses the response.
func (c *APIClient) FetchData(ctx context.Context, endpoint string, result any) error {
	url := fmt.Sprintf("http://%s%s", c.BaseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		slog.Debug("Failed to fetch data", slog.String("url", url), slog.Any("error", err))
		return fmt.Errorf("failed to fetch data from %s: %w", url, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Debug("Failed to close response body", slog.Any("error", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Debug("Non-200 status code", slog.String("url", url), slog.Int("status_code", resp.StatusCode))
		return fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse JSON response from %s: %w", url, err)
	}

	slog.Debug("Successfully fetched data", slog.String("url", url))
	return nil
}
