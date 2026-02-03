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
	"time"

	"github.com/mark-chris/devtools-sync/agent/internal/api"
	"github.com/mark-chris/devtools-sync/agent/internal/profile"
	"github.com/spf13/cobra"
)

func TestSyncPushCommand_NoProfiles(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Initialize config
	configDir := filepath.Join(tempHome, ".devtools-sync")
	profilesDir := filepath.Join(configDir, "profiles")
	setupTestConfig(t, tempHome, "http://localhost:8080", profilesDir)

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(syncCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"sync", "push"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync push command failed: %v", err)
	}

	// Verify output
	got := output.String()
	if !strings.Contains(got, "No profiles to push") {
		t.Errorf("expected 'No profiles to push', got: %s", got)
	}
}

func TestSyncPushCommand_WithProfiles(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create mock server
	pushedProfiles := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/profiles" {
			t.Errorf("expected path /api/v1/profiles, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		var prof api.Profile
		if err := json.NewDecoder(r.Body).Decode(&prof); err != nil {
			t.Fatalf("failed to decode profile: %v", err)
		}

		pushedProfiles = append(pushedProfiles, prof.Name)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Initialize config
	configDir := filepath.Join(tempHome, ".devtools-sync")
	profilesDir := filepath.Join(configDir, "profiles")
	setupTestConfig(t, tempHome, server.URL, profilesDir)

	// Create test profiles
	createTestProfile(t, profilesDir, "work", 2)
	createTestProfile(t, profilesDir, "personal", 1)

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(syncCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"sync", "push"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync push command failed: %v", err)
	}

	// Verify output
	got := output.String()
	if !strings.Contains(got, "Pushed 2 profile(s)") {
		t.Errorf("expected 'Pushed 2 profile(s)', got: %s", got)
	}

	// Verify profiles were pushed
	if len(pushedProfiles) != 2 {
		t.Errorf("expected 2 profiles to be pushed, got %d", len(pushedProfiles))
	}
}

func TestSyncPullCommand_NoProfiles(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create mock server that returns empty profile list
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/profiles" {
			t.Errorf("expected path /api/v1/profiles, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode([]string{}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Initialize config
	configDir := filepath.Join(tempHome, ".devtools-sync")
	profilesDir := filepath.Join(configDir, "profiles")
	setupTestConfig(t, tempHome, server.URL, profilesDir)

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(syncCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"sync", "pull"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync pull command failed: %v", err)
	}

	// Verify output
	got := output.String()
	if !strings.Contains(got, "No profiles on server") {
		t.Errorf("expected 'No profiles on server', got: %s", got)
	}
}

func TestSyncPullCommand_WithProfiles(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/api/v1/profiles" {
			// List profiles
			if err := json.NewEncoder(w).Encode([]string{"test-profile"}); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		} else if strings.HasPrefix(r.URL.Path, "/api/v1/profiles/") {
			// Download specific profile
			prof := api.Profile{
				Name:      "test-profile",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Extensions: []api.Extension{
					{ID: "ext1", Version: "1.0.0", Enabled: true},
				},
			}
			if err := json.NewEncoder(w).Encode(prof); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}
	}))
	defer server.Close()

	// Initialize config
	configDir := filepath.Join(tempHome, ".devtools-sync")
	profilesDir := filepath.Join(configDir, "profiles")
	setupTestConfig(t, tempHome, server.URL, profilesDir)

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(syncCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"sync", "pull"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync pull command failed: %v", err)
	}

	// Verify output
	got := output.String()
	if !strings.Contains(got, "Pulled 1 profile(s)") {
		t.Errorf("expected 'Pulled 1 profile(s)', got: %s", got)
	}

	// Verify profile was saved locally
	profilePath := filepath.Join(profilesDir, "test-profile.json")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		t.Errorf("expected profile to be saved at %s", profilePath)
	}
}

func TestSyncPullCommand_SkipsNewerLocal(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create mock server with older profile
	oldTime := time.Now().Add(-24 * time.Hour)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/api/v1/profiles" {
			if err := json.NewEncoder(w).Encode([]string{"test-profile"}); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		} else if strings.HasPrefix(r.URL.Path, "/api/v1/profiles/") {
			prof := api.Profile{
				Name:      "test-profile",
				CreatedAt: oldTime,
				UpdatedAt: oldTime,
				Extensions: []api.Extension{
					{ID: "ext1", Version: "1.0.0", Enabled: true},
				},
			}
			if err := json.NewEncoder(w).Encode(prof); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}
	}))
	defer server.Close()

	// Initialize config
	configDir := filepath.Join(tempHome, ".devtools-sync")
	profilesDir := filepath.Join(configDir, "profiles")
	setupTestConfig(t, tempHome, server.URL, profilesDir)

	// Create newer local profile
	createTestProfile(t, profilesDir, "test-profile", 2)

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(syncCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"sync", "pull"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync pull command failed: %v", err)
	}

	// Verify output
	got := output.String()
	if !strings.Contains(got, "Skipped 1 profile(s)") {
		t.Errorf("expected 'Skipped 1 profile(s)', got: %s", got)
	}
	if !strings.Contains(got, "local is newer") {
		t.Errorf("expected 'local is newer' message, got: %s", got)
	}
}

// Helper functions

func setupTestConfig(t *testing.T, homeDir, serverURL, profilesDir string) {
	configDir := filepath.Join(homeDir, ".devtools-sync")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	configYAML := `server:
  url: ` + serverURL + `
profiles:
  directory: ` + profilesDir + `
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

func createTestProfile(t *testing.T, profilesDir, name string, extCount int) {
	extensions := make([]profile.Extension, extCount)
	for i := 0; i < extCount; i++ {
		extensions[i] = profile.Extension{
			ID:      "ext" + string(rune('1'+i)),
			Version: "1.0.0",
			Enabled: true,
		}
	}

	prof := profile.Profile{
		Name:       name,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Extensions: extensions,
	}

	data, err := json.MarshalIndent(prof, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	profilePath := filepath.Join(profilesDir, name+".json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("failed to write profile file: %v", err)
	}
}
