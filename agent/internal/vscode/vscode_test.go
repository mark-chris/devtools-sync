package vscode

import (
	"os/exec"
	"testing"
)

func TestDetectInstallation(t *testing.T) {
	// This test will vary by platform
	// We just verify it doesn't error
	_, err := DetectInstallation()
	if err != nil {
		t.Errorf("DetectInstallation should not error, got: %v", err)
	}
}

func TestListExtensions(t *testing.T) {
	// Check if VS Code CLI is available
	if !isVSCodeInstalled() {
		t.Skip("VS Code CLI not available, skipping test")
	}

	extensions, err := ListExtensions()
	if err != nil {
		t.Errorf("ListExtensions should not error, got: %v", err)
	}

	if extensions == nil {
		t.Error("expected extensions slice, got nil")
	}
}

func TestInstallExtension(t *testing.T) {
	// Only test the validation error case
	// Don't actually try to install extensions in tests
	err := InstallExtension("")
	if err == nil {
		t.Error("expected error for empty extension ID, got nil")
	}
	if err.Error() != "extension ID cannot be empty" {
		t.Errorf("expected 'extension ID cannot be empty' error, got: %s", err.Error())
	}
}

func TestGetVSCodePaths(t *testing.T) {
	paths := getVSCodePaths()

	// Should return at least one path for common platforms
	if len(paths) == 0 {
		t.Error("expected at least one VS Code path")
	}
}

// Helper function to check if VS Code CLI is available
func isVSCodeInstalled() bool {
	_, err := exec.LookPath("code")
	return err == nil
}
