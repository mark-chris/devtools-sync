package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MaxResponseSize is the maximum allowed response body size (1MB)
// This prevents memory exhaustion from malicious or misconfigured servers
const MaxResponseSize = 1 << 20 // 1MB

// ErrResponseTooLarge is returned when a server response exceeds MaxResponseSize
var ErrResponseTooLarge = errors.New("response body exceeds maximum allowed size")

// Client handles communication with the devtools-sync server
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// HealthResponse represents the server health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Health checks if the server is healthy
func (c *Client) Health() (*HealthResponse, error) {
	url := fmt.Sprintf("%s/health", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	body, err := readLimitedResponse(resp.Body, MaxResponseSize)
	if err != nil {
		return nil, err
	}

	var health HealthResponse
	if err := json.Unmarshal(body, &health); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &health, nil
}

// readLimitedResponse reads up to maxSize bytes from the reader.
// Returns ErrResponseTooLarge if the response exceeds the limit.
func readLimitedResponse(r io.Reader, maxSize int64) ([]byte, error) {
	limited := io.LimitReader(r, maxSize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if int64(len(body)) > maxSize {
		return nil, ErrResponseTooLarge
	}

	return body, nil
}
