package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark-chris/devtools-sync/agent/internal/profile"
	"github.com/spf13/cobra"
)

func TestProfileListCommand_Empty(t *testing.T) {
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
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	configYAML := `server:
  url: http://localhost:8080
profiles:
  directory: ` + profilesDir + `
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(profileCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"profile", "list"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("profile list command failed: %v", err)
	}

	// Verify output
	got := output.String()
	if !strings.Contains(got, "No profiles found") {
		t.Errorf("expected 'No profiles found', got: %s", got)
	}
}

func TestProfileListCommand_WithProfiles(t *testing.T) {
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
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	configYAML := `server:
  url: http://localhost:8080
profiles:
  directory: ` + profilesDir + `
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create test profiles
	now := time.Now()
	profiles := []profile.Profile{
		{
			Name:       "work",
			CreatedAt:  now,
			UpdatedAt:  now,
			Extensions: []profile.Extension{{ID: "ext1", Version: "1.0.0", Enabled: true}},
		},
		{
			Name:       "personal",
			CreatedAt:  now,
			UpdatedAt:  now,
			Extensions: []profile.Extension{{ID: "ext2", Version: "2.0.0", Enabled: true}},
		},
	}

	for _, prof := range profiles {
		data, err := json.MarshalIndent(prof, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal profile: %v", err)
		}
		profilePath := filepath.Join(profilesDir, prof.Name+".json")
		if err := os.WriteFile(profilePath, data, 0644); err != nil {
			t.Fatalf("failed to write profile file: %v", err)
		}
	}

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(profileCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"profile", "list"})

	// Execute command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("profile list command failed: %v", err)
	}

	// Verify output contains profile names
	got := output.String()
	if !strings.Contains(got, "work") {
		t.Errorf("expected output to contain 'work', got: %s", got)
	}
	if !strings.Contains(got, "personal") {
		t.Errorf("expected output to contain 'personal', got: %s", got)
	}
	if !strings.Contains(got, "NAME") || !strings.Contains(got, "EXTENSIONS") {
		t.Errorf("expected table headers in output, got: %s", got)
	}
}

func TestProfileSaveCommand_ValidationError(t *testing.T) {
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
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	configYAML := `server:
  url: http://localhost:8080
profiles:
  directory: ` + profilesDir + `
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(profileCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"profile", "save"})

	// Execute command (should fail - missing profile name)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing profile name")
	}
}

func TestProfileLoadCommand_NotFound(t *testing.T) {
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
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	configYAML := `server:
  url: http://localhost:8080
profiles:
  directory: ` + profilesDir + `
logging:
  level: info
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(profileCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"profile", "load", "nonexistent"})

	// Execute command (should fail - profile not found)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent profile")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %s", err.Error())
	}
}

func TestProfileCommands_RequireArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"save without name", []string{"profile", "save"}},
		{"load without name", []string{"profile", "load"}},
		{"save with extra args", []string{"profile", "save", "name1", "name2"}},
		{"load with extra args", []string{"profile", "load", "name1", "name2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new root command for testing
			cmd := &cobra.Command{Use: "devtools-sync"}
			cmd.AddCommand(profileCmd)

			// Capture output
			output := &bytes.Buffer{}
			cmd.SetOut(output)
			cmd.SetErr(output)
			cmd.SetArgs(tt.args)

			// Execute command (should fail)
			err := cmd.Execute()
			if err == nil {
				t.Errorf("expected error for invalid arguments")
			}
		})
	}
}
