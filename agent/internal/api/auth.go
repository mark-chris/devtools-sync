package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mark-chris/devtools-sync/agent/internal/keychain"
)

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
