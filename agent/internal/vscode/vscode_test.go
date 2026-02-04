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

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{
			name: "v1 greater than v2",
			v1:   "2.0.0",
			v2:   "1.0.0",
			want: 1,
		},
		{
			name: "v1 less than v2",
			v1:   "1.0.0",
			v2:   "2.0.0",
			want: -1,
		},
		{
			name: "v1 equals v2",
			v1:   "1.0.0",
			v2:   "1.0.0",
			want: 0,
		},
		{
			name: "patch version difference",
			v1:   "1.0.1",
			v2:   "1.0.0",
			want: 1,
		},
		{
			name: "minor version difference",
			v1:   "1.1.0",
			v2:   "1.2.0",
			want: -1,
		},
		{
			name: "pre-release version",
			v1:   "1.0.0-alpha",
			v2:   "1.0.0",
			want: -1,
		},
		{
			name: "versions with v prefix",
			v1:   "v1.2.3",
			v2:   "v1.2.4",
			want: -1,
		},
		{
			name: "mixed prefix formats",
			v1:   "1.0.0",
			v2:   "v1.0.0",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestMergeExtensions(t *testing.T) {
	tests := []struct {
		name string
		sets [][]Extension
		want []Extension
	}{
		{
			name: "no duplicates",
			sets: [][]Extension{
				{
					{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
					{ID: "golang.go", Version: "0.40.0", Enabled: true},
				},
				{
					{ID: "ms-vscode.cpptools", Version: "1.15.0", Enabled: true},
				},
			},
			want: []Extension{
				{ID: "golang.go", Version: "0.40.0", Enabled: true},
				{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
				{ID: "ms-vscode.cpptools", Version: "1.15.0", Enabled: true},
			},
		},
		{
			name: "duplicates with newer version",
			sets: [][]Extension{
				{
					{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
				},
				{
					{ID: "ms-python.python", Version: "2024.1.0", Enabled: true},
				},
			},
			want: []Extension{
				{ID: "ms-python.python", Version: "2024.1.0", Enabled: true},
			},
		},
		{
			name: "stable over older insider",
			sets: [][]Extension{
				{
					{ID: "golang.go", Version: "0.40.0", Enabled: true},
				},
				{
					{ID: "golang.go", Version: "0.39.0", Enabled: true},
				},
			},
			want: []Extension{
				{ID: "golang.go", Version: "0.40.0", Enabled: true},
			},
		},
		{
			name: "multiple duplicates",
			sets: [][]Extension{
				{
					{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
					{ID: "golang.go", Version: "0.39.0", Enabled: true},
				},
				{
					{ID: "ms-python.python", Version: "2024.1.0", Enabled: true},
					{ID: "ms-vscode.cpptools", Version: "1.15.0", Enabled: true},
				},
				{
					{ID: "golang.go", Version: "0.40.0", Enabled: true},
					{ID: "ms-python.python", Version: "2023.0.0", Enabled: true},
				},
			},
			want: []Extension{
				{ID: "golang.go", Version: "0.40.0", Enabled: true},
				{ID: "ms-python.python", Version: "2024.1.0", Enabled: true},
				{ID: "ms-vscode.cpptools", Version: "1.15.0", Enabled: true},
			},
		},
		{
			name: "empty sets",
			sets: [][]Extension{
				{},
				{},
			},
			want: []Extension{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeExtensions(tt.sets...)

			if len(got) != len(tt.want) {
				t.Errorf("mergeExtensions() returned %d extensions, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if i >= len(tt.want) {
					break
				}
				if got[i].ID != tt.want[i].ID {
					t.Errorf("extension[%d].ID = %v, want %v", i, got[i].ID, tt.want[i].ID)
				}
				if got[i].Version != tt.want[i].Version {
					t.Errorf("extension[%d].Version = %v, want %v", i, got[i].Version, tt.want[i].Version)
				}
				if got[i].Enabled != tt.want[i].Enabled {
					t.Errorf("extension[%d].Enabled = %v, want %v", i, got[i].Enabled, tt.want[i].Enabled)
				}
			}
		})
	}
}

func TestListExtensionsFromDirs(t *testing.T) {
	// Create temporary test directories
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	// Setup first directory (stable) with Python extension
	ext1Dir := filepath.Join(tmpDir1, "ms-python.python-2024.0.0")
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

	// Add Go extension to first directory
	ext2Dir := filepath.Join(tmpDir1, "golang.go-0.40.0")
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

	// Setup second directory (insiders) with duplicate Python extension (newer version)
	ext3Dir := filepath.Join(tmpDir2, "ms-python.python-2024.1.0")
	err = os.MkdirAll(ext3Dir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest3 := `{
		"name": "python",
		"displayName": "Python",
		"description": "Python extension",
		"version": "2024.1.0",
		"publisher": "ms-python"
	}`
	err = os.WriteFile(filepath.Join(ext3Dir, "package.json"), []byte(manifest3), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Add C++ extension to second directory
	ext4Dir := filepath.Join(tmpDir2, "ms-vscode.cpptools-1.15.0")
	err = os.MkdirAll(ext4Dir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest4 := `{
		"name": "cpptools",
		"version": "1.15.0",
		"publisher": "ms-vscode"
	}`
	err = os.WriteFile(filepath.Join(ext4Dir, "package.json"), []byte(manifest4), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test with both directories
	dirs := []string{tmpDir1, tmpDir2}
	extensions, err := listExtensionsFromDirs(dirs)
	if err != nil {
		t.Fatalf("listExtensionsFromDirs failed: %v", err)
	}

	// Should find 3 unique extensions (Python merged to higher version)
	if len(extensions) != 3 {
		t.Errorf("expected 3 extensions, got %d", len(extensions))
	}

	// Verify extensions and versions
	extMap := make(map[string]string)
	for _, ext := range extensions {
		extMap[ext.ID] = ext.Version
	}

	// Check Python extension has the higher version
	if version, found := extMap["ms-python.python"]; !found {
		t.Error("Python extension not found")
	} else if version != "2024.1.0" {
		t.Errorf("expected Python version 2024.1.0, got %s", version)
	}

	// Check Go extension
	if version, found := extMap["golang.go"]; !found {
		t.Error("Go extension not found")
	} else if version != "0.40.0" {
		t.Errorf("expected Go version 0.40.0, got %s", version)
	}

	// Check C++ extension
	if version, found := extMap["ms-vscode.cpptools"]; !found {
		t.Error("C++ extension not found")
	} else if version != "1.15.0" {
		t.Errorf("expected C++ version 1.15.0, got %s", version)
	}
}

func TestListExtensionsFromDirsWithErrors(t *testing.T) {
	// Create one valid directory and one nonexistent
	tmpDir := t.TempDir()

	// Add valid extension
	extDir := filepath.Join(tmpDir, "golang.go-0.40.0")
	err := os.MkdirAll(extDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"name": "go",
		"version": "0.40.0",
		"publisher": "golang"
	}`
	err = os.WriteFile(filepath.Join(extDir, "package.json"), []byte(manifest), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test with valid and nonexistent directories
	dirs := []string{tmpDir, "/nonexistent/directory"}
	extensions, err := listExtensionsFromDirs(dirs)

	// Should not error (continues on directory errors)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Should still find the valid extension
	if len(extensions) != 1 {
		t.Errorf("expected 1 extension, got %d", len(extensions))
	}

	if len(extensions) > 0 && extensions[0].ID != "golang.go" {
		t.Errorf("expected golang.go, got %s", extensions[0].ID)
	}
}

func TestListExtensionsFromDirsEmpty(t *testing.T) {
	// Test with empty slice
	extensions, err := listExtensionsFromDirs([]string{})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(extensions) != 0 {
		t.Errorf("expected 0 extensions, got %d", len(extensions))
	}
}

func TestListExtensionsWithFallback(t *testing.T) {
	// This test verifies the fallback behavior
	// We can't easily mock exec.Command, but we can test the logic
	// by temporarily making 'code' unavailable

	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer func() {
		os.Setenv("PATH", originalPath)
	}()

	// Set PATH to empty to make 'code' unavailable
	os.Setenv("PATH", "")

	// Create a temporary extension directory
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "extensions")
	err := os.MkdirAll(extDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create test extension
	testExtDir := filepath.Join(extDir, "test.ext-1.0.0")
	err = os.MkdirAll(testExtDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"name": "ext",
		"displayName": "Test Extension",
		"version": "1.0.0",
		"publisher": "test"
	}`
	err = os.WriteFile(filepath.Join(testExtDir, "package.json"), []byte(manifest), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Temporarily override getExtensionDirs to return our test directory
	originalGetExtensionDirs := getExtensionDirs
	getExtensionDirs = func() []string {
		return []string{extDir}
	}
	defer func() {
		getExtensionDirs = originalGetExtensionDirs
	}()

	// Call ListExtensions - should fall back to directory parsing
	extensions, err := ListExtensions()
	if err != nil {
		t.Fatalf("ListExtensions failed: %v", err)
	}

	// Should find our test extension
	if len(extensions) != 1 {
		t.Errorf("expected 1 extension, got %d", len(extensions))
	}

	if len(extensions) > 0 {
		if extensions[0].ID != "test.ext" {
			t.Errorf("expected ID 'test.ext', got '%s'", extensions[0].ID)
		}
		if extensions[0].Version != "1.0.0" {
			t.Errorf("expected version '1.0.0', got '%s'", extensions[0].Version)
		}
	}
}

func TestGetStatePaths(t *testing.T) {
	paths := getStatePaths()

	// Should return at least 2 paths (VS Code and Insiders)
	if len(paths) < 2 {
		t.Errorf("expected at least 2 state paths, got %d", len(paths))
	}

	// All paths should be absolute
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			t.Errorf("expected absolute path, got '%s'", path)
		}
	}

	// All paths should end with storage.json
	for _, path := range paths {
		if !strings.HasSuffix(path, "storage.json") {
			t.Errorf("expected path to end with storage.json, got '%s'", path)
		}
	}
}

func TestGetStatePathsByOS(t *testing.T) {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows
	}

	paths := getStatePaths()

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support/Code/User/globalStorage/storage.json
		expectedStable := filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "storage.json")
		expectedInsiders := filepath.Join(home, "Library", "Application Support", "Code - Insiders", "User", "globalStorage", "storage.json")

		foundStable := false
		foundInsiders := false
		for _, path := range paths {
			if path == expectedStable {
				foundStable = true
			}
			if path == expectedInsiders {
				foundInsiders = true
			}
		}
		if !foundStable {
			t.Errorf("expected macOS stable path %s not found in %v", expectedStable, paths)
		}
		if !foundInsiders {
			t.Errorf("expected macOS insiders path %s not found in %v", expectedInsiders, paths)
		}

	case "windows":
		// Windows: %APPDATA%/Code/User/globalStorage/storage.json
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			t.Skip("APPDATA not set, skipping Windows-specific test")
		}
		expectedStable := filepath.Join(appdata, "Code", "User", "globalStorage", "storage.json")
		expectedInsiders := filepath.Join(appdata, "Code - Insiders", "User", "globalStorage", "storage.json")

		foundStable := false
		foundInsiders := false
		for _, path := range paths {
			if path == expectedStable {
				foundStable = true
			}
			if path == expectedInsiders {
				foundInsiders = true
			}
		}
		if !foundStable {
			t.Errorf("expected Windows stable path %s not found in %v", expectedStable, paths)
		}
		if !foundInsiders {
			t.Errorf("expected Windows insiders path %s not found in %v", expectedInsiders, paths)
		}

	case "linux":
		// Linux: ~/.config/Code/User/globalStorage/storage.json
		expectedStable := filepath.Join(home, ".config", "Code", "User", "globalStorage", "storage.json")
		expectedInsiders := filepath.Join(home, ".config", "Code - Insiders", "User", "globalStorage", "storage.json")

		foundStable := false
		foundInsiders := false
		for _, path := range paths {
			if path == expectedStable {
				foundStable = true
			}
			if path == expectedInsiders {
				foundInsiders = true
			}
		}
		if !foundStable {
			t.Errorf("expected Linux stable path %s not found in %v", expectedStable, paths)
		}
		if !foundInsiders {
			t.Errorf("expected Linux insiders path %s not found in %v", expectedInsiders, paths)
		}
	}
}

func TestLoadDisabledExtensions(t *testing.T) {
	tests := []struct {
		name        string
		storageJSON string
		want        map[string]bool
		wantErr     bool
	}{
		{
			name: "valid storage.json with disabled extensions",
			storageJSON: `{
				"extensionsIdentifiers/disabled": [
					{"id": "ms-python.python"},
					{"id": "golang.go"}
				]
			}`,
			want: map[string]bool{
				"ms-python.python": true,
				"golang.go":        true,
			},
			wantErr: false,
		},
		{
			name: "valid storage.json with empty disabled array",
			storageJSON: `{
				"extensionsIdentifiers/disabled": []
			}`,
			want:    map[string]bool{},
			wantErr: false,
		},
		{
			name:        "valid storage.json without disabled field",
			storageJSON: `{"otherField": "value"}`,
			want:        map[string]bool{},
			wantErr:     false,
		},
		{
			name:        "empty JSON object",
			storageJSON: `{}`,
			want:        map[string]bool{},
			wantErr:     false,
		},
		{
			name:        "invalid JSON",
			storageJSON: `{invalid json}`,
			want:        nil,
			wantErr:     true,
		},
		{
			name: "disabled extensions with extra fields",
			storageJSON: `{
				"extensionsIdentifiers/disabled": [
					{"id": "ms-python.python", "uuid": "some-uuid"},
					{"id": "golang.go"}
				],
				"otherField": "value"
			}`,
			want: map[string]bool{
				"ms-python.python": true,
				"golang.go":        true,
			},
			wantErr: false,
		},
		{
			name: "disabled extensions with missing id field",
			storageJSON: `{
				"extensionsIdentifiers/disabled": [
					{"id": "ms-python.python"},
					{"uuid": "some-uuid"}
				]
			}`,
			want: map[string]bool{
				"ms-python.python": true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with test data
			tmpFile := filepath.Join(t.TempDir(), "storage.json")
			err := os.WriteFile(tmpFile, []byte(tt.storageJSON), 0644)
			if err != nil {
				t.Fatal(err)
			}

			got, err := loadDisabledExtensions(tmpFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("got %d disabled extensions, want %d", len(got), len(tt.want))
			}

			for id := range tt.want {
				if !got[id] {
					t.Errorf("expected %s to be disabled", id)
				}
			}

			for id := range got {
				if !tt.want[id] {
					t.Errorf("unexpected disabled extension: %s", id)
				}
			}
		})
	}
}

func TestLoadDisabledExtensionsMissingFile(t *testing.T) {
	// Test with non-existent file - should return empty map, no error
	got, err := loadDisabledExtensions("/nonexistent/storage.json")
	if err != nil {
		t.Errorf("expected no error for missing file, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map for missing file, got %d entries", len(got))
	}
}

// Helper function to check if VS Code CLI is available
func isVSCodeInstalled() bool {
	_, err := exec.LookPath("code")
	return err == nil
}
