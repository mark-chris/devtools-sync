package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mark-chris/devtools-sync/agent/internal/keychain"
)

// ErrNotAuthenticated is returned when no access token is available
var ErrNotAuthenticated = errors.New("not authenticated: please run 'devtools-sync login' first")

// AuthenticatedClient wraps Client with authentication
type AuthenticatedClient struct {
	client   *Client
	keychain keychain.Keychain
}

// NewAuthenticatedClient creates a new authenticated API client
func NewAuthenticatedClient(baseURL string, kc keychain.Keychain) *AuthenticatedClient {
	return &AuthenticatedClient{
		client:   NewClient(baseURL),
		keychain: kc,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// StoredCredentials represents cached credentials for auto re-login
type StoredCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates with the server and stores the token
func (ac *AuthenticatedClient) Login(email, password string) error {
	// Prepare request
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	data, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := fmt.Sprintf("%s/auth/login", ac.client.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	// Send request with retry
	resp, err := ac.client.retryableRequest(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := readLimitedResponse(resp.Body, MaxResponseSize)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := readLimitedResponse(resp.Body, MaxResponseSize)
	if err != nil {
		return err
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	// Store token
	if err := ac.keychain.Set(keychain.KeyAccessToken, loginResp.AccessToken); err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	// Store credentials for auto re-login
	creds := StoredCredentials{
		Email:    email,
		Password: password,
	}
	credsJSON, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := ac.keychain.Set(keychain.KeyCredentials, string(credsJSON)); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	return nil
}
// AuthenticatedRequest executes an HTTP request with authentication and auto re-login on 401
func (ac *AuthenticatedClient) AuthenticatedRequest(req *http.Request) (*http.Response, error) {
	// Get access token
	token, err := ac.keychain.Get(keychain.KeyAccessToken)
	if err != nil {
		if errors.Is(err, keychain.ErrNotFound) {
			return nil, ErrNotAuthenticated
		}
		return nil, fmt.Errorf("failed to retrieve access token: %w", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Make request
	resp, err := ac.client.retryableRequest(req)
	if err != nil {
		return nil, err
	}

	// If 401, attempt auto re-login
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()

		// Try to get stored credentials
		credsJSON, err := ac.keychain.Get(keychain.KeyCredentials)
		if err != nil {
			return nil, fmt.Errorf("session expired: please run 'devtools-sync login' again")
		}

		var creds StoredCredentials
		if err := json.Unmarshal([]byte(credsJSON), &creds); err != nil {
			return nil, fmt.Errorf("failed to parse stored credentials: %w", err)
		}

		// Attempt re-login
		if err := ac.Login(creds.Email, creds.Password); err != nil {
			return nil, fmt.Errorf("auto re-login failed: %w", err)
		}

		// Retry original request with new token
		token, err = ac.keychain.Get(keychain.KeyAccessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve new access token: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Reset request body if needed
		if req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("failed to reset request body: %w", err)
			}
			req.Body = body
		}

		resp, err = ac.client.retryableRequest(req)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// UploadProfile uploads a profile with authentication
func (ac *AuthenticatedClient) UploadProfile(profile *Profile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/profiles", ac.client.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	resp, err := ac.AuthenticatedRequest(req)
	if err != nil {
		return fmt.Errorf("failed to upload profile: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := readLimitedResponse(resp.Body, MaxResponseSize)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListProfiles retrieves all profile names with authentication
func (ac *AuthenticatedClient) ListProfiles() ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/profiles", ac.client.baseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ac.AuthenticatedRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

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

// DownloadProfile retrieves a specific profile with authentication
func (ac *AuthenticatedClient) DownloadProfile(name string) (*Profile, error) {
	url := fmt.Sprintf("%s/api/v1/profiles/%s", ac.client.baseURL, name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ac.AuthenticatedRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download profile: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

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

// Logout removes stored credentials from keychain
func (ac *AuthenticatedClient) Logout() error {
	// Delete access token
	if err := ac.keychain.Delete(keychain.KeyAccessToken); err != nil {
		return fmt.Errorf("failed to delete access token: %w", err)
	}

	// Delete stored credentials
	if err := ac.keychain.Delete(keychain.KeyCredentials); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	return nil
}
