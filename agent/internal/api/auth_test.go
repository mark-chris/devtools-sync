package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark-chris/devtools-sync/agent/internal/keychain"
)

func TestAuthenticatedClient_Login(t *testing.T) {
	kc := keychain.NewMockKeychain()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/login" {
			t.Errorf("expected /auth/login, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Email != "test@example.com" || req.Password != "password123" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		resp := map[string]interface{}{
			"access_token": "test-token-abc123",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthenticatedClient(server.URL, kc)

	err := client.Login("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Verify token stored
	token, err := kc.Get(keychain.KeyAccessToken)
	if err != nil {
		t.Fatalf("token not stored: %v", err)
	}
	if token != "test-token-abc123" {
		t.Errorf("expected token 'test-token-abc123', got '%s'", token)
	}

	// Verify credentials stored
	creds, err := kc.Get(keychain.KeyCredentials)
	if err != nil {
		t.Fatalf("credentials not stored: %v", err)
	}
	if creds != `{"email":"test@example.com","password":"password123"}` {
		t.Errorf("unexpected credentials: %s", creds)
	}
}

func TestAuthenticatedClient_Login_InvalidCredentials(t *testing.T) {
	kc := keychain.NewMockKeychain()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
	}))
	defer server.Close()

	client := NewAuthenticatedClient(server.URL, kc)

	err := client.Login("wrong@example.com", "wrongpass")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAuthenticatedClient_Logout(t *testing.T) {
	kc := keychain.NewMockKeychain()

	// Pre-populate keychain
	_ = kc.Set(keychain.KeyAccessToken, "test-token")
	_ = kc.Set(keychain.KeyCredentials, `{"email":"test@example.com","password":"pass"}`)

	client := NewAuthenticatedClient("http://example.com", kc)

	err := client.Logout()
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Verify token deleted
	_, err = kc.Get(keychain.KeyAccessToken)
	if err == nil {
		t.Error("expected token to be deleted")
	}

	// Verify credentials deleted
	_, err = kc.Get(keychain.KeyCredentials)
	if err == nil {
		t.Error("expected credentials to be deleted")
	}
}

func TestAuthenticatedClient_AuthenticatedRequest(t *testing.T) {
	kc := keychain.NewMockKeychain()
	_ = kc.Set(keychain.KeyAccessToken, "valid-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewAuthenticatedClient(server.URL, kc)

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	resp, err := client.AuthenticatedRequest(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuthenticatedClient_AutoRelogin(t *testing.T) {
	kc := keychain.NewMockKeychain()
	_ = kc.Set(keychain.KeyAccessToken, "expired-token")
	_ = kc.Set(keychain.KeyCredentials, `{"email":"test@example.com","password":"password123"}`)

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		if r.URL.Path == "/auth/login" {
			// Login endpoint
			resp := map[string]interface{}{
				"access_token": "new-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Protected endpoint
		auth := r.Header.Get("Authorization")
		if auth == "Bearer expired-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if auth == "Bearer new-token" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewAuthenticatedClient(server.URL, kc)

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/protected", nil)
	resp, err := client.AuthenticatedRequest(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after re-login, got %d", resp.StatusCode)
	}

	// Verify new token stored
	token, _ := kc.Get(keychain.KeyAccessToken)
	if token != "new-token" {
		t.Errorf("expected new-token, got %s", token)
	}
}

func TestAuthenticatedClient_NoTokenError(t *testing.T) {
	kc := keychain.NewMockKeychain()
	client := NewAuthenticatedClient("http://example.com", kc)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
	_, err := client.AuthenticatedRequest(req)
	if err == nil {
		t.Fatal("expected error when no token, got nil")
	}
	if !errors.Is(err, ErrNotAuthenticated) {
		t.Errorf("expected ErrNotAuthenticated, got: %v", err)
	}
}
