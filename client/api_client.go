package client

import (
	"encoding/json"
	"fmt"
	"io"
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
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	resp, err := c.Client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch data from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body) // Use io.ReadAll instead of ioutil.ReadAll
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}
