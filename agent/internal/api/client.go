package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"
)

// MaxResponseSize is the maximum allowed response body size (1MB)
// This prevents memory exhaustion from malicious or misconfigured servers
const MaxResponseSize = 1 << 20 // 1MB

// Retry configuration
const (
	MaxRetries    = 3
	InitialDelay  = 1 * time.Second
	MaxDelay      = 30 * time.Second
	BackoffFactor = 2.0
	JitterFactor  = 0.1
)

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

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.retryableRequest(req)
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

// retryableRequest executes an HTTP request with exponential backoff retry
func (c *Client) retryableRequest(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		// Clone request body for retries
		if attempt > 0 && req.Body != nil {
			// For simplicity, we require GetBody to be set for retryable POST/PUT
			if req.GetBody == nil {
				return nil, fmt.Errorf("request body cannot be retried (GetBody not set)")
			}
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				return nil, fmt.Errorf("failed to get request body: %w", bodyErr)
			}
			req.Body = body
		}

		resp, err = c.httpClient.Do(req)

		// Success - return response
		if err == nil && !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Check if error is retryable
		if err != nil && !isRetryableError(err) {
			return nil, err
		}

		// Check if status is retryable
		if err == nil && !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Don't retry after last attempt
		if attempt == MaxRetries {
			if err != nil {
				return nil, err
			}
			return resp, nil
		}

		// Calculate delay with exponential backoff and jitter
		delay := calculateDelay(attempt)
		time.Sleep(delay)

		// Close response body before retry
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	return resp, err
}

// isRetryableError checks if an error should trigger a retry
func isRetryableError(err error) bool {
	// Network errors are retryable
	if _, ok := err.(net.Error); ok {
		return true
	}
	// Check for specific network-related errors
	if err.Error() == "connection refused" ||
	   err.Error() == "no such host" ||
	   err.Error() == "timeout" {
		return true
	}
	return false
}

// isRetryableStatus checks if an HTTP status should trigger a retry
func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests,      // 429
		http.StatusBadGateway,            // 502
		http.StatusServiceUnavailable,    // 503
		http.StatusGatewayTimeout:        // 504
		return true
	default:
		return false
	}
}

// calculateDelay computes the delay with exponential backoff and jitter
func calculateDelay(attempt int) time.Duration {
	// Exponential backoff: InitialDelay * (BackoffFactor ^ attempt)
	delay := float64(InitialDelay) * math.Pow(BackoffFactor, float64(attempt))

	// Cap at MaxDelay
	if delay > float64(MaxDelay) {
		delay = float64(MaxDelay)
	}

	// Add jitter: ±10%
	jitter := delay * JitterFactor * (rand.Float64()*2 - 1)
	delay += jitter

	return time.Duration(delay)
}

// Profile represents an extension profile (matches internal/profile.Profile)
type Profile struct {
	Name       string      `json:"name"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Extensions []Extension `json:"extensions"`
}

// Extension represents a VS Code extension (matches internal/profile.Extension)
type Extension struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

// UploadProfile sends a profile to the server
func (c *Client) UploadProfile(profile *Profile) error {
	url := fmt.Sprintf("%s/api/v1/profiles", c.baseURL)

	// Marshal profile to JSON
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Create POST request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	// Send request with retry
	resp, err := c.retryableRequest(req)
	if err != nil {
		return fmt.Errorf("failed to upload profile: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := readLimitedResponse(resp.Body, MaxResponseSize)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListProfiles retrieves all profile names from server
func (c *Client) ListProfiles() ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/profiles", c.baseURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.retryableRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := readLimitedResponse(resp.Body, MaxResponseSize)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := readLimitedResponse(resp.Body, MaxResponseSize)
	if err != nil {
		return nil, err
	}

	var profiles []string
	if err := json.Unmarshal(body, &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return profiles, nil
}

// DownloadProfile retrieves a specific profile from the server
func (c *Client) DownloadProfile(name string) (*Profile, error) {
	url := fmt.Sprintf("%s/api/v1/profiles/%s", c.baseURL, name)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.retryableRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download profile: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("profile '%s' not found on server", name)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := readLimitedResponse(resp.Body, MaxResponseSize)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := readLimitedResponse(resp.Body, MaxResponseSize)
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &profile, nil
}

// UpdateProfile updates an existing profile on the server
func (c *Client) UpdateProfile(name string, profile *Profile) error {
	url := fmt.Sprintf("%s/api/v1/profiles/%s", c.baseURL, name)

	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	resp, err := c.retryableRequest(req)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := readLimitedResponse(resp.Body, MaxResponseSize)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
