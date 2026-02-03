package api

import (
	"bytes"
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

	// Send request
	resp, err := c.httpClient.Do(req)
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

	resp, err := c.httpClient.Get(url)
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

	resp, err := c.httpClient.Get(url)
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
