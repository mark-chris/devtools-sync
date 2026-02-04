package vscode

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name        string
		manifest    string
		dirName     string
		wantExt     Extension
		wantErr     bool
		errContains string
	}{
		{
			name: "valid manifest with all fields",
			manifest: `{
				"name": "python",
				"version": "2024.0.0",
				"publisher": "ms-python",
				"displayName": "Python",
				"description": "IntelliSense, linting, debugging"
			}`,
			dirName: "ms-python.python-2024.0.0",
			wantExt: Extension{
				ID:          "ms-python.python",
				Version:     "2024.0.0",
				Enabled:     true,
				DisplayName: "Python",
				Description: "IntelliSense, linting, debugging",
				Publisher:   "ms-python",
			},
			wantErr: false,
		},
		{
			name: "valid manifest with missing optional fields",
			manifest: `{
				"name": "go",
				"version": "0.40.0",
				"publisher": "golang"
			}`,
			dirName: "golang.go-0.40.0",
			wantExt: Extension{
				ID:          "golang.go",
				Version:     "0.40.0",
				Enabled:     true,
				DisplayName: "",
				Description: "",
				Publisher:   "golang",
			},
			wantErr: false,
		},
		{
			name: "missing required field - name",
			manifest: `{
				"version": "1.0.0",
				"publisher": "test"
			}`,
			dirName:     "test.missing-1.0.0",
			wantErr:     true,
			errContains: "missing required field: name",
		},
		{
			name: "invalid JSON",
			manifest: `{
				"name": "broken"
				"version": "1.0.0"
			}`,
			dirName:     "test.broken-1.0.0",
			wantErr:     true,
			errContains: "failed to parse package.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := parseManifest([]byte(tt.manifest), tt.dirName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if ext.ID != tt.wantExt.ID {
				t.Errorf("ID = %v, want %v", ext.ID, tt.wantExt.ID)
			}
			if ext.Version != tt.wantExt.Version {
				t.Errorf("Version = %v, want %v", ext.Version, tt.wantExt.Version)
			}
			if ext.Enabled != tt.wantExt.Enabled {
				t.Errorf("Enabled = %v, want %v", ext.Enabled, tt.wantExt.Enabled)
			}
			if ext.DisplayName != tt.wantExt.DisplayName {
				t.Errorf("DisplayName = %v, want %v", ext.DisplayName, tt.wantExt.DisplayName)
			}
			if ext.Description != tt.wantExt.Description {
				t.Errorf("Description = %v, want %v", ext.Description, tt.wantExt.Description)
			}
			if ext.Publisher != tt.wantExt.Publisher {
				t.Errorf("Publisher = %v, want %v", ext.Publisher, tt.wantExt.Publisher)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || strings.Contains(s, substr))
}

func TestScanExtensionDir(t *testing.T) {
	// Create temporary test directory structure
	tmpDir := t.TempDir()

	// Create valid extension
	ext1Dir := filepath.Join(tmpDir, "ms-python.python-2024.0.0")
	err := os.MkdirAll(ext1Dir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest1 := `{
		"name": "python",
		"displayName": "Python",
		"description": "Python extension",
		"version": "2024.0.0",
		"publisher": "ms-python"
	}`
	err = os.WriteFile(filepath.Join(ext1Dir, "package.json"), []byte(manifest1), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create another valid extension
	ext2Dir := filepath.Join(tmpDir, "golang.go-0.40.0")
	err = os.MkdirAll(ext2Dir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest2 := `{
		"name": "go",
		"version": "0.40.0",
		"publisher": "golang"
	}`
	err = os.WriteFile(filepath.Join(ext2Dir, "package.json"), []byte(manifest2), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create directory without package.json (should be skipped)
	badDir := filepath.Join(tmpDir, "invalid-extension")
	err = os.MkdirAll(badDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create directory with invalid JSON (should be skipped)
	invalidDir := filepath.Join(tmpDir, "invalid-json-1.0.0")
	err = os.MkdirAll(invalidDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(invalidDir, "package.json"), []byte(`{invalid`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Scan directory
	extensions, err := scanExtensionDir(tmpDir)
	if err != nil {
		t.Fatalf("scanExtensionDir failed: %v", err)
	}

	// Should find 2 valid extensions
	if len(extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(extensions))
	}

	// Check that we got the right extensions
	foundPython := false
	foundGo := false
	for _, ext := range extensions {
		if ext.ID == "ms-python.python" && ext.Version == "2024.0.0" {
			foundPython = true
		}
		if ext.ID == "golang.go" && ext.Version == "0.40.0" {
			foundGo = true
		}
	}

	if !foundPython {
		t.Error("did not find Python extension")
	}
	if !foundGo {
		t.Error("did not find Go extension")
	}
}

func TestScanExtensionDirNonexistent(t *testing.T) {
	extensions, err := scanExtensionDir("/nonexistent/directory")
	if err != nil {
		t.Fatalf("expected no error for nonexistent directory, got: %v", err)
	}
	if len(extensions) != 0 {
		t.Errorf("expected empty slice for nonexistent directory, got %d extensions", len(extensions))
	}
}

// Helper function to check if VS Code CLI is available
func isVSCodeInstalled() bool {
	_, err := exec.LookPath("code")
	return err == nil
}
