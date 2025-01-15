package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type APIClient struct {
	BaseURL string
	Client  *http.Client
}

// NewAPIClient initializes and returns a new APIClient.
func NewAPIClient(baseURL string, timeout time.Duration) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchData makes a GET request to the specified endpoint and parses the response.
func (c *APIClient) FetchData(endpoint string, result interface{}) error {
	url := fmt.Sprintf("http://%s%s", c.BaseURL, endpoint)
	slog.Info("Fetching data", slog.String("url", url))

	resp, err := c.Client.Get(url)
	if err != nil {
		slog.Error("Failed to fetch data", slog.String("url", url), slog.Any("error", err))
		return fmt.Errorf("failed to fetch data from %s: %w", url, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", slog.Any("error", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("Non-200 status code", slog.String("url", url), slog.Int("status_code", resp.StatusCode))
		return fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body) // Use io.ReadAll instead of ioutil.ReadAll
	if err != nil {
		slog.Error("Failed to read response body", slog.String("url", url), slog.Any("error", err))
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		slog.Error("Failed to parse JSON response", slog.String("url", url), slog.Any("error", err))
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	slog.Info("Successfully fetched data", slog.String("url", url))
	return nil
}
