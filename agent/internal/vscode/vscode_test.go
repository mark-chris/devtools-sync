package vscode

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func TestExtensionStruct(t *testing.T) {
	ext := Extension{
		ID:          "ms-python.python",
		Version:     "2024.0.0",
		Enabled:     true,
		DisplayName: "Python",
		Description: "IntelliSense, linting, debugging",
		Publisher:   "ms-python",
	}

	if ext.DisplayName != "Python" {
		t.Errorf("expected DisplayName 'Python', got '%s'", ext.DisplayName)
	}
	if ext.Description != "IntelliSense, linting, debugging" {
		t.Errorf("expected Description, got '%s'", ext.Description)
	}
	if ext.Publisher != "ms-python" {
		t.Errorf("expected Publisher 'ms-python', got '%s'", ext.Publisher)
	}
}

func TestGetExtensionDirs(t *testing.T) {
	dirs := getExtensionDirs()

	// Should return at least 2 paths (VS Code and Insiders)
	if len(dirs) < 2 {
		t.Errorf("expected at least 2 extension directories, got %d", len(dirs))
	}

	// All paths should be absolute
	for _, dir := range dirs {
		if !filepath.IsAbs(dir) {
			t.Errorf("expected absolute path, got '%s'", dir)
		}
	}
}

func TestGetExtensionDirsByOS(t *testing.T) {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows
	}

	dirs := getExtensionDirs()

	switch runtime.GOOS {
	case "linux":
		expectedStable := filepath.Join(home, ".vscode", "extensions")
		expectedInsiders := filepath.Join(home, ".vscode-insiders", "extensions")

		foundStable := false
		foundInsiders := false
		for _, dir := range dirs {
			if dir == expectedStable {
				foundStable = true
			}
			if dir == expectedInsiders {
				foundInsiders = true
			}
		}
		if !foundStable {
			t.Errorf("expected Linux stable path %s not found in %v", expectedStable, dirs)
		}
		if !foundInsiders {
			t.Errorf("expected Linux insiders path %s not found in %v", expectedInsiders, dirs)
		}
	case "darwin":
		expectedStable := filepath.Join(home, ".vscode", "extensions")
		expectedInsiders := filepath.Join(home, ".vscode-insiders", "extensions")

		foundStable := false
		foundInsiders := false
		for _, dir := range dirs {
			if dir == expectedStable {
				foundStable = true
			}
			if dir == expectedInsiders {
				foundInsiders = true
			}
		}
		if !foundStable {
			t.Errorf("expected macOS stable path %s not found in %v", expectedStable, dirs)
		}
		if !foundInsiders {
			t.Errorf("expected macOS insiders path %s not found in %v", expectedInsiders, dirs)
		}
	case "windows":
		appdata := os.Getenv("USERPROFILE")
		expectedStable := filepath.Join(appdata, ".vscode", "extensions")
		expectedInsiders := filepath.Join(appdata, ".vscode-insiders", "extensions")

		foundStable := false
		foundInsiders := false
		for _, dir := range dirs {
			if dir == expectedStable {
				foundStable = true
			}
			if dir == expectedInsiders {
				foundInsiders = true
			}
		}
		if !foundStable {
			t.Errorf("expected Windows stable path %s not found in %v", expectedStable, dirs)
		}
		if !foundInsiders {
			t.Errorf("expected Windows insiders path %s not found in %v", expectedInsiders, dirs)
		}
	}
}

// Helper function to check if VS Code CLI is available
func isVSCodeInstalled() bool {
	_, err := exec.LookPath("code")
	return err == nil
}
