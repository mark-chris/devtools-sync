package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func TestInitCommand(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(initCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"init"})

	// Execute command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify output
	got := output.String()
	configDir := filepath.Join(tempHome, ".devtools-sync")
	want := "Configuration initialized at " + configDir + "\n"
	if got != want {
		t.Errorf("init command output:\ngot:  %q\nwant: %q", got, want)
	}

	// Verify config file exists
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file was not created at %s", configPath)
	}

	// Verify profiles directory exists
	profilesDir := filepath.Join(configDir, "profiles")
	if info, err := os.Stat(profilesDir); os.IsNotExist(err) {
		t.Errorf("profiles directory was not created at %s", profilesDir)
	} else if !info.IsDir() {
		t.Errorf("%s exists but is not a directory", profilesDir)
	}

	// Verify config file contents
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config file: %v", err)
	}

	// Check server URL
	server, ok := config["server"].(map[string]interface{})
	if !ok {
		t.Fatal("server section not found in config")
	}
	if server["url"] != "http://localhost:8080" {
		t.Errorf("expected server.url to be http://localhost:8080, got %v", server["url"])
	}

	// Check logging level
	logging, ok := config["logging"].(map[string]interface{})
	if !ok {
		t.Fatal("logging section not found in config")
	}
	if logging["level"] != "info" {
		t.Errorf("expected logging.level to be info, got %v", logging["level"])
	}

}

func TestInitCommand_AlreadyExists(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()

	// Create config directory and file
	configDir := filepath.Join(tempHome, ".devtools-sync")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test: data\n"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(initCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"init"})

	// Execute command (should fail)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("init command should have failed when config already exists")
	}

	// Verify error message contains the key information
	if got := err.Error(); !strings.Contains(got, "configuration already exists at "+configPath) {
		t.Errorf("expected error about existing config, got: %s", got)
	}
	// Verify error includes helpful guidance
	if got := err.Error(); !strings.Contains(got, "To reconfigure") {
		t.Errorf("expected error to include guidance, got: %s", got)
	}
}
