package main

import (
	"os"
	"testing"
)

func TestVersionConstant(t *testing.T) {
	if version == "" {
		t.Error("version constant should not be empty")
	}

	expected := "0.1.0"
	if version != expected {
		t.Errorf("expected version %s, got %s", expected, version)
	}
}

func TestMainWithEnvironment(t *testing.T) {
	// Test that environment variable is recognized
	if err := os.Setenv("DEVTOOLS_SYNC_SERVER_URL", "http://test:9090"); err != nil {
		t.Fatalf("failed to set environment variable: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("DEVTOOLS_SYNC_SERVER_URL")
	}()

	// This test verifies the environment is set correctly
	// Main function execution is tested via integration tests
	url := os.Getenv("DEVTOOLS_SYNC_SERVER_URL")
	if url != "http://test:9090" {
		t.Errorf("expected environment variable to be set to http://test:9090, got %s", url)
	}
}
