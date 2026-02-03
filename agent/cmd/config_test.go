package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func TestConfigShowCommand(t *testing.T) {
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
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configYAML := `server:
  url: http://test:9090
profiles:
  directory: /test/profiles
logging:
  level: debug
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(configCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"config", "show"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config show command failed: %v", err)
	}

	// Verify output
	got := output.String()
	expectedParts := []string{
		"Server:",
		"URL: http://test:9090",
		"Profiles:",
		"Directory: /test/profiles",
		"Logging:",
		"Level: debug",
	}

	for _, part := range expectedParts {
		if !contains(got, part) {
			t.Errorf("output missing expected part: %s\nGot:\n%s", part, got)
		}
	}
}

func TestConfigShowCommand_WithEnvOverride(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalURL := os.Getenv("DEVTOOLS_SYNC_SERVER_URL")
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	if err := os.Setenv("DEVTOOLS_SYNC_SERVER_URL", "https://override.com"); err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
		if originalURL != "" {
			_ = os.Setenv("DEVTOOLS_SYNC_SERVER_URL", originalURL)
		} else {
			_ = os.Unsetenv("DEVTOOLS_SYNC_SERVER_URL")
		}
	}()

	// Initialize config with default URL
	configDir := filepath.Join(tempHome, ".devtools-sync")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configYAML := `server:
  url: http://localhost:8080
profiles:
  directory: /tmp/profiles
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(configCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"config", "show"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config show command failed: %v", err)
	}

	// Verify env var override is shown
	got := output.String()
	if !contains(got, "https://override.com") {
		t.Errorf("expected override URL in output, got:\n%s", got)
	}
}

func TestConfigSetCommand(t *testing.T) {
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
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configYAML := `server:
  url: http://localhost:8080
profiles:
  directory: /tmp/profiles
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(configCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"config", "set", "server.url", "https://api.example.com"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config set command failed: %v", err)
	}

	// Verify output
	got := output.String()
	expected := "Updated server.url to: https://api.example.com\n"
	if got != expected {
		t.Errorf("expected output:\n%s\ngot:\n%s", expected, got)
	}

	// Verify config file was updated
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config file: %v", err)
	}

	server := config["server"].(map[string]interface{})
	if server["url"] != "https://api.example.com" {
		t.Errorf("expected server.url to be updated to https://api.example.com, got %v", server["url"])
	}
}

func TestConfigSetCommand_InvalidURL(t *testing.T) {
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
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configYAML := `server:
  url: http://localhost:8080
profiles:
  directory: /tmp/profiles
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(configCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"config", "set", "server.url", "ftp://invalid.com"})

	// Execute command (should fail)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("config set should have failed with invalid URL")
	}

	// Verify error message
	if !contains(err.Error(), "server URL must use http or https scheme") {
		t.Errorf("expected validation error, got: %s", err.Error())
	}

	// Verify config file was NOT updated
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config file: %v", err)
	}

	server := config["server"].(map[string]interface{})
	if server["url"] != "http://localhost:8080" {
		t.Errorf("expected server.url to remain unchanged, got %v", server["url"])
	}
}

func TestConfigSetCommand_InvalidKey(t *testing.T) {
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
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configYAML := `server:
  url: http://localhost:8080
profiles:
  directory: /tmp/profiles
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	tests := []struct {
		name        string
		key         string
		value       string
		expectedErr string
	}{
		{
			name:        "invalid section",
			key:         "invalid.field",
			value:       "value",
			expectedErr: "unknown config section: invalid",
		},
		{
			name:        "invalid server field",
			key:         "server.invalid",
			value:       "value",
			expectedErr: "unknown server field: invalid",
		},
		{
			name:        "invalid profiles field",
			key:         "profiles.invalid",
			value:       "value",
			expectedErr: "unknown profiles field: invalid",
		},
		{
			name:        "invalid logging field",
			key:         "logging.invalid",
			value:       "value",
			expectedErr: "unknown logging field: invalid",
		},
		{
			name:        "invalid key format",
			key:         "invalidkey",
			value:       "value",
			expectedErr: "invalid key format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new root command for testing
			cmd := &cobra.Command{Use: "devtools-sync"}
			cmd.AddCommand(configCmd)

			// Capture output
			output := &bytes.Buffer{}
			cmd.SetOut(output)
			cmd.SetErr(output)
			cmd.SetArgs([]string{"config", "set", tt.key, tt.value})

			// Execute command (should fail)
			err := cmd.Execute()
			if err == nil {
				t.Fatal("config set should have failed with invalid key")
			}

			// Verify error message
			if !contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing %q, got: %s", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestConfigSetCommand_AllFields(t *testing.T) {
	tests := []struct {
		key   string
		value string
	}{
		{"server.url", "https://test.com"},
		{"profiles.directory", "/custom/profiles"},
		{"logging.level", "debug"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
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
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatalf("failed to create config dir: %v", err)
			}

			configYAML := `server:
  url: http://localhost:8080
profiles:
  directory: /tmp/profiles
logging:
  level: info
`
			configPath := filepath.Join(configDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			// Create a new root command for testing
			cmd := &cobra.Command{Use: "devtools-sync"}
			cmd.AddCommand(configCmd)

			// Capture output
			output := &bytes.Buffer{}
			cmd.SetOut(output)
			cmd.SetErr(output)
			cmd.SetArgs([]string{"config", "set", tt.key, tt.value})

			// Execute command
			if err := cmd.Execute(); err != nil {
				t.Fatalf("config set command failed: %v", err)
			}

			// Verify output
			got := output.String()
			expected := "Updated " + tt.key + " to: " + tt.value + "\n"
			if got != expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", expected, got)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
