package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark-chris/devtools-sync/agent/internal/keychain"
	"github.com/spf13/cobra"
)

func TestLogoutCommand(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	setupTestLoginConfig(t, tempHome, "http://localhost:8080")

	// Ensure profiles directory exists
	configDir := filepath.Join(tempHome, ".devtools-sync")
	if err := os.MkdirAll(filepath.Join(configDir, "profiles"), 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	// Use mock keychain with pre-populated data
	mockKC := keychain.NewMockKeychain()
	_ = mockKC.Set(keychain.KeyAccessToken, "test-token")
	_ = mockKC.Set(keychain.KeyCredentials, `{"email":"test@example.com","password":"pass"}`)

	origFactory := keychainFactory
	keychainFactory = func() keychain.Keychain {
		return mockKC
	}
	defer func() {
		keychainFactory = origFactory
	}()

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(logoutCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"logout"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("logout command failed: %v", err)
	}

	// Verify token deleted
	_, err = mockKC.Get(keychain.KeyAccessToken)
	if err == nil {
		t.Error("expected token to be deleted")
	}

	// Verify credentials deleted
	_, err = mockKC.Get(keychain.KeyCredentials)
	if err == nil {
		t.Error("expected credentials to be deleted")
	}

	outStr := output.String()
	if !strings.Contains(outStr, "Logged out successfully") {
		t.Errorf("expected success message, got: %s", outStr)
	}
}
