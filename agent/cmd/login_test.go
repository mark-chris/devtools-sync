package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark-chris/devtools-sync/agent/internal/keychain"
	"github.com/spf13/cobra"
)

func TestLoginCommand(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/login" {
			t.Errorf("expected /auth/login, got %s", r.URL.Path)
		}

		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Email == "test@example.com" && req.Password == "password123" {
			resp := map[string]any{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			}
			_ = json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	// Initialize config
	configDir := filepath.Join(tempHome, ".devtools-sync")
	setupTestLoginConfig(t, tempHome, server.URL)

	// Ensure profiles directory exists
	if err := os.MkdirAll(filepath.Join(configDir, "profiles"), 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	// Use mock keychain
	mockKC := keychain.NewMockKeychain()
	origFactory := keychainFactory
	keychainFactory = func() keychain.Keychain {
		return mockKC
	}
	defer func() {
		keychainFactory = origFactory
	}()

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(loginCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"login", "--email", "test@example.com", "--password", "password123"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("login command failed: %v", err)
	}

	outStr := output.String()
	if !strings.Contains(outStr, "Login successful") {
		t.Errorf("expected success message, got: %s", outStr)
	}

	// Verify token stored
	token, err := mockKC.Get(keychain.KeyAccessToken)
	if err != nil {
		t.Fatalf("token not stored: %v", err)
	}
	if token != "test-token" {
		t.Errorf("expected 'test-token', got '%s'", token)
	}
}

func TestLoginCommand_InvalidCredentials(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
	}))
	defer server.Close()

	setupTestLoginConfig(t, tempHome, server.URL)

	// Use mock keychain
	mockKC := keychain.NewMockKeychain()
	origFactory := keychainFactory
	keychainFactory = func() keychain.Keychain {
		return mockKC
	}
	defer func() {
		keychainFactory = origFactory
	}()

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(loginCmd)

	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"login", "--email", "wrong@example.com", "--password", "wrongpass"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid credentials, got nil")
	}

	if !strings.Contains(err.Error(), "login failed") {
		t.Errorf("expected 'login failed' in error, got: %v", err)
	}
}

func setupTestLoginConfig(t *testing.T, homeDir, serverURL string) {
	t.Helper()
	configDir := filepath.Join(homeDir, ".devtools-sync")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configYAML := `server:
  url: ` + serverURL + `
profiles:
  directory: ` + filepath.Join(configDir, "profiles") + `
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}
