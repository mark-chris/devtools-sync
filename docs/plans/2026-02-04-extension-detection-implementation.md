# VS Code Extension Detection Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement comprehensive VS Code extension detection with CLI-first approach and directory parsing fallback, supporting both VS Code stable and Insiders.

**Architecture:** Extend existing `agent/internal/vscode` package with directory parsing capabilities, state detection, and version-aware merging. The CLI method remains primary; directory parsing activates on CLI failure.

**Tech Stack:** Go 1.21+, `golang.org/x/mod/semver` for version comparison, standard library for file I/O and JSON parsing.

---

## Task 1: Update Extension struct with new metadata fields

**Files:**
- Modify: `agent/internal/vscode/vscode.go:14-18`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test for new Extension fields**

Add to `agent/internal/vscode/vscode_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run TestExtensionStruct`
Expected: Compilation error - fields don't exist

**Step 3: Update Extension struct**

Modify `agent/internal/vscode/vscode.go:14-18`:

```go
// Extension represents a VS Code extension
type Extension struct {
	ID          string
	Version     string
	Enabled     bool
	DisplayName string
	Description string
	Publisher   string
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run TestExtensionStruct`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add DisplayName, Description, Publisher to Extension struct

- Added DisplayName for user-friendly extension names
- Added Description for extension purpose
- Added Publisher for extension author
- Added test for new fields

Related to #11"
```

---

## Task 2: Add function to get extension directories

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
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

		found := false
		for _, dir := range dirs {
			if dir == expectedStable || dir == expectedInsiders {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected Linux paths not found in %v", dirs)
		}
	case "darwin":
		expectedStable := filepath.Join(home, "Library", "Application Support", "Code", "extensions")
		found := false
		for _, dir := range dirs {
			if dir == expectedStable {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected macOS path not found in %v", dirs)
		}
	case "windows":
		appdata := os.Getenv("USERPROFILE")
		expectedStable := filepath.Join(appdata, ".vscode", "extensions")
		found := false
		for _, dir := range dirs {
			if dir == expectedStable {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected Windows path not found in %v", dirs)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run "TestGetExtensionDirs"`
Expected: Compilation error - function doesn't exist

**Step 3: Implement getExtensionDirs**

Add to `agent/internal/vscode/vscode.go` after `getVSCodePaths()`:

```go
// getExtensionDirs returns extension directory paths for VS Code and Insiders
func getExtensionDirs() []string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows
	}

	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Application Support", "Code", "extensions"),
			filepath.Join(home, "Library", "Application Support", "Code - Insiders", "extensions"),
		}
	case "windows":
		return []string{
			filepath.Join(home, ".vscode", "extensions"),
			filepath.Join(home, ".vscode-insiders", "extensions"),
		}
	case "linux":
		return []string{
			filepath.Join(home, ".vscode", "extensions"),
			filepath.Join(home, ".vscode-insiders", "extensions"),
		}
	default:
		return []string{}
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run "TestGetExtensionDirs"`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add getExtensionDirs for VS Code and Insiders

- Returns extension directory paths for both stable and Insiders
- Platform-specific paths for macOS, Windows, Linux
- Added comprehensive tests for all platforms

Related to #11"
```

---

## Task 3: Add manifest parsing types and function

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
func TestParseManifest(t *testing.T) {
	tests := []struct {
		name        string
		manifestJSON string
		dirName     string
		want        Extension
		wantErr     bool
	}{
		{
			name: "valid manifest",
			manifestJSON: `{
				"name": "python",
				"displayName": "Python",
				"description": "IntelliSense, linting, debugging",
				"version": "2024.0.0",
				"publisher": "ms-python"
			}`,
			dirName: "ms-python.python-2024.0.0",
			want: Extension{
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
			name: "missing optional fields",
			manifestJSON: `{
				"name": "test",
				"version": "1.0.0",
				"publisher": "test-pub"
			}`,
			dirName: "test-pub.test-1.0.0",
			want: Extension{
				ID:          "test-pub.test",
				Version:     "1.0.0",
				Enabled:     true,
				DisplayName: "",
				Description: "",
				Publisher:   "test-pub",
			},
			wantErr: false,
		},
		{
			name:         "missing required fields",
			manifestJSON: `{"description": "test"}`,
			dirName:      "invalid-dir",
			want:         Extension{},
			wantErr:      true,
		},
		{
			name:         "invalid json",
			manifestJSON: `{invalid json}`,
			dirName:      "test-dir",
			want:         Extension{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseManifest([]byte(tt.manifestJSON), tt.dirName)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID != tt.want.ID {
					t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.Version != tt.want.Version {
					t.Errorf("Version = %v, want %v", got.Version, tt.want.Version)
				}
				if got.DisplayName != tt.want.DisplayName {
					t.Errorf("DisplayName = %v, want %v", got.DisplayName, tt.want.DisplayName)
				}
				if got.Description != tt.want.Description {
					t.Errorf("Description = %v, want %v", got.Description, tt.want.Description)
				}
				if got.Publisher != tt.want.Publisher {
					t.Errorf("Publisher = %v, want %v", got.Publisher, tt.want.Publisher)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run TestParseManifest`
Expected: Compilation error - function doesn't exist

**Step 3: Implement parseManifest**

Add to `agent/internal/vscode/vscode.go`:

```go
import (
	"encoding/json"
	// ... existing imports
)

// packageManifest represents the structure of a VS Code extension's package.json
type packageManifest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Publisher   string `json:"publisher"`
}

// parseManifest parses extension manifest data and returns an Extension
func parseManifest(data []byte, dirName string) (Extension, error) {
	var manifest packageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Extension{}, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	// Validate required fields
	if manifest.Name == "" || manifest.Version == "" || manifest.Publisher == "" {
		return Extension{}, fmt.Errorf("missing required fields (name, version, or publisher)")
	}

	// Build extension ID from publisher and name
	id := fmt.Sprintf("%s.%s", manifest.Publisher, manifest.Name)

	return Extension{
		ID:          id,
		Version:     manifest.Version,
		Enabled:     true, // Default to enabled, will be updated later
		DisplayName: manifest.DisplayName,
		Description: manifest.Description,
		Publisher:   manifest.Publisher,
	}, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run TestParseManifest`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add manifest parsing for extension metadata

- Parse package.json to extract extension info
- Validate required fields (name, version, publisher)
- Handle missing optional fields gracefully
- Comprehensive tests for valid and invalid manifests

Related to #11"
```

---

## Task 4: Add function to scan extension directory

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run "TestScanExtensionDir"`
Expected: Compilation error - function doesn't exist

**Step 3: Implement scanExtensionDir**

Add to `agent/internal/vscode/vscode.go`:

```go
import (
	"io"
	"log"
	// ... existing imports
)

// scanExtensionDir scans a directory for installed extensions
func scanExtensionDir(dir string) ([]Extension, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []Extension{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read extension directory: %w", err)
	}

	extensions := make([]Extension, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Read package.json from extension directory
		manifestPath := filepath.Join(dir, entry.Name(), "package.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			// Skip extensions without package.json
			log.Printf("Warning: skipping %s - no package.json found", entry.Name())
			continue
		}

		// Parse manifest
		ext, err := parseManifest(data, entry.Name())
		if err != nil {
			// Skip extensions with invalid manifests
			log.Printf("Warning: skipping %s - %v", entry.Name(), err)
			continue
		}

		extensions = append(extensions, ext)
	}

	return extensions, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run "TestScanExtensionDir"`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add scanExtensionDir to discover installed extensions

- Scans directory for extension subdirectories
- Reads and parses package.json for each extension
- Skips invalid extensions with warnings
- Returns empty slice for nonexistent directories
- Added comprehensive tests with temp directory fixtures

Related to #11"
```

---

## Task 5: Add version comparison function

**Files:**
- Modify: `agent/go.mod` (add dependency)
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Add semver dependency**

Run: `cd agent && go get golang.org/x/mod/semver`

**Step 2: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int // -1 if v1 < v2, 0 if equal, 1 if v1 > v2
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"v1 greater", "2.0.0", "1.0.0", 1},
		{"v2 greater", "1.0.0", "2.0.0", -1},
		{"patch difference", "1.0.1", "1.0.0", 1},
		{"minor difference", "1.1.0", "1.0.9", 1},
		{"prerelease vs release", "2.0.0-beta", "2.0.0", -1},
		{"two prereleases", "2.0.0-beta.2", "2.0.0-beta.1", 1},
		{"insider vs stable", "2024.1.0-insider", "2024.0.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%s, %s) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}
```

**Step 3: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run TestCompareVersions`
Expected: Compilation error - function doesn't exist

**Step 4: Implement compareVersions**

Add to `agent/internal/vscode/vscode.go`:

```go
import (
	"golang.org/x/mod/semver"
	// ... existing imports
)

// compareVersions compares two semantic versions
// Returns -1 if v1 < v2, 0 if equal, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Ensure versions start with 'v' for semver package
	if !strings.HasPrefix(v1, "v") {
		v1 = "v" + v1
	}
	if !strings.HasPrefix(v2, "v") {
		v2 = "v" + v2
	}

	return semver.Compare(v1, v2)
}
```

**Step 5: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run TestCompareVersions`
Expected: PASS

**Step 6: Update go.mod and go.sum**

Run: `cd agent && go mod tidy`

**Step 7: Commit**

```bash
git add agent/go.mod agent/go.sum agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add version comparison using semver

- Use golang.org/x/mod/semver for reliable version comparison
- Handle prerelease versions correctly
- Added comprehensive test cases
- Added dependency to go.mod

Related to #11"
```

---

## Task 6: Add function to merge and deduplicate extensions

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
func TestMergeExtensions(t *testing.T) {
	tests := []struct {
		name  string
		sets  [][]Extension
		want  []Extension
	}{
		{
			name: "no duplicates",
			sets: [][]Extension{
				{
					{ID: "ms-python.python", Version: "2024.0.0", DisplayName: "Python", Publisher: "ms-python"},
				},
				{
					{ID: "golang.go", Version: "0.40.0", DisplayName: "Go", Publisher: "golang"},
				},
			},
			want: []Extension{
				{ID: "golang.go", Version: "0.40.0", DisplayName: "Go", Publisher: "golang"},
				{ID: "ms-python.python", Version: "2024.0.0", DisplayName: "Python", Publisher: "ms-python"},
			},
		},
		{
			name: "duplicate - keep newer version",
			sets: [][]Extension{
				{
					{ID: "ms-python.python", Version: "2024.0.0", DisplayName: "Python", Publisher: "ms-python"},
				},
				{
					{ID: "ms-python.python", Version: "2024.1.0", DisplayName: "Python Insiders", Publisher: "ms-python"},
				},
			},
			want: []Extension{
				{ID: "ms-python.python", Version: "2024.1.0", DisplayName: "Python Insiders", Publisher: "ms-python"},
			},
		},
		{
			name: "duplicate - keep stable over older insider",
			sets: [][]Extension{
				{
					{ID: "test.ext", Version: "2.0.0", DisplayName: "Test", Publisher: "test"},
				},
				{
					{ID: "test.ext", Version: "1.9.0-insider", DisplayName: "Test Insider", Publisher: "test"},
				},
			},
			want: []Extension{
				{ID: "test.ext", Version: "2.0.0", DisplayName: "Test", Publisher: "test"},
			},
		},
		{
			name: "multiple duplicates and unique",
			sets: [][]Extension{
				{
					{ID: "ext1.a", Version: "1.0.0", DisplayName: "A", Publisher: "ext1"},
					{ID: "ext2.b", Version: "2.0.0", DisplayName: "B", Publisher: "ext2"},
				},
				{
					{ID: "ext1.a", Version: "1.1.0", DisplayName: "A New", Publisher: "ext1"},
					{ID: "ext3.c", Version: "3.0.0", DisplayName: "C", Publisher: "ext3"},
				},
			},
			want: []Extension{
				{ID: "ext1.a", Version: "1.1.0", DisplayName: "A New", Publisher: "ext1"},
				{ID: "ext2.b", Version: "2.0.0", DisplayName: "B", Publisher: "ext2"},
				{ID: "ext3.c", Version: "3.0.0", DisplayName: "C", Publisher: "ext3"},
			},
		},
		{
			name: "empty sets",
			sets: [][]Extension{},
			want: []Extension{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeExtensions(tt.sets...)

			// Sort both for comparison (mergeExtensions should return sorted)
			sortExtensions := func(exts []Extension) {
				sort.Slice(exts, func(i, j int) bool {
					return exts[i].ID < exts[j].ID
				})
			}
			sortExtensions(got)
			sortExtensions(tt.want)

			if len(got) != len(tt.want) {
				t.Errorf("got %d extensions, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].ID != tt.want[i].ID {
					t.Errorf("extension[%d].ID = %s, want %s", i, got[i].ID, tt.want[i].ID)
				}
				if got[i].Version != tt.want[i].Version {
					t.Errorf("extension[%d].Version = %s, want %s", i, got[i].Version, tt.want[i].Version)
				}
				if got[i].DisplayName != tt.want[i].DisplayName {
					t.Errorf("extension[%d].DisplayName = %s, want %s", i, got[i].DisplayName, tt.want[i].DisplayName)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run TestMergeExtensions`
Expected: Compilation error - function doesn't exist, also need to import "sort"

**Step 3: Implement mergeExtensions**

Add to `agent/internal/vscode/vscode.go`:

```go
import (
	"sort"
	// ... existing imports
)

// mergeExtensions merges multiple sets of extensions, deduplicating by ID
// and keeping the extension with the highest version number
func mergeExtensions(sets ...[]Extension) []Extension {
	if len(sets) == 0 {
		return []Extension{}
	}

	// Map of extension ID to extension
	extMap := make(map[string]Extension)

	// Collect all extensions
	for _, set := range sets {
		for _, ext := range set {
			existing, exists := extMap[ext.ID]
			if !exists {
				extMap[ext.ID] = ext
				continue
			}

			// Compare versions and keep the newer one
			if compareVersions(ext.Version, existing.Version) > 0 {
				log.Printf("Deduplicating %s: keeping v%s over v%s", ext.ID, ext.Version, existing.Version)
				extMap[ext.ID] = ext
			}
		}
	}

	// Convert map to sorted slice
	result := make([]Extension, 0, len(extMap))
	for _, ext := range extMap {
		result = append(result, ext)
	}

	// Sort by ID for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run TestMergeExtensions`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add mergeExtensions to deduplicate by version

- Merge multiple extension sets into one
- Keep extension with highest version on duplicates
- Sort output by ID for consistency
- Log deduplication decisions
- Comprehensive tests for various scenarios

Related to #11"
```

---

## Task 7: Add function to list extensions from directories

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
func TestListExtensionsFromDirs(t *testing.T) {
	// Create temporary test structure with two "installations"
	tmpDir := t.TempDir()

	// Stable installation
	stableDir := filepath.Join(tmpDir, "stable", "extensions")
	err := os.MkdirAll(stableDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create extension in stable
	ext1Dir := filepath.Join(stableDir, "ms-python.python-2024.0.0")
	err = os.MkdirAll(ext1Dir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest1 := `{
		"name": "python",
		"displayName": "Python",
		"version": "2024.0.0",
		"publisher": "ms-python"
	}`
	err = os.WriteFile(filepath.Join(ext1Dir, "package.json"), []byte(manifest1), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Insiders installation
	insidersDir := filepath.Join(tmpDir, "insiders", "extensions")
	err = os.MkdirAll(insidersDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create newer version of same extension in insiders
	ext2Dir := filepath.Join(insidersDir, "ms-python.python-2024.1.0")
	err = os.MkdirAll(ext2Dir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest2 := `{
		"name": "python",
		"displayName": "Python Insiders",
		"version": "2024.1.0",
		"publisher": "ms-python"
	}`
	err = os.WriteFile(filepath.Join(ext2Dir, "package.json"), []byte(manifest2), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create unique extension in insiders
	ext3Dir := filepath.Join(insidersDir, "golang.go-0.40.0")
	err = os.MkdirAll(ext3Dir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest3 := `{
		"name": "go",
		"displayName": "Go",
		"version": "0.40.0",
		"publisher": "golang"
	}`
	err = os.WriteFile(filepath.Join(ext3Dir, "package.json"), []byte(manifest3), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// List extensions
	extensions, err := listExtensionsFromDirs([]string{stableDir, insidersDir})
	if err != nil {
		t.Fatalf("listExtensionsFromDirs failed: %v", err)
	}

	// Should have 2 extensions (python deduplicated, go unique)
	if len(extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(extensions))
	}

	// Check python - should have newer version
	foundPython := false
	for _, ext := range extensions {
		if ext.ID == "ms-python.python" {
			foundPython = true
			if ext.Version != "2024.1.0" {
				t.Errorf("expected Python version 2024.1.0, got %s", ext.Version)
			}
			if ext.DisplayName != "Python Insiders" {
				t.Errorf("expected DisplayName 'Python Insiders', got '%s'", ext.DisplayName)
			}
		}
	}
	if !foundPython {
		t.Error("Python extension not found")
	}

	// Check go extension exists
	foundGo := false
	for _, ext := range extensions {
		if ext.ID == "golang.go" {
			foundGo = true
		}
	}
	if !foundGo {
		t.Error("Go extension not found")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run TestListExtensionsFromDirs`
Expected: Compilation error - function doesn't exist

**Step 3: Implement listExtensionsFromDirs**

Add to `agent/internal/vscode/vscode.go`:

```go
// listExtensionsFromDirs scans multiple directories and returns merged extensions
func listExtensionsFromDirs(dirs []string) ([]Extension, error) {
	extensionSets := make([][]Extension, 0, len(dirs))

	for _, dir := range dirs {
		extensions, err := scanExtensionDir(dir)
		if err != nil {
			log.Printf("Warning: failed to scan directory %s: %v", dir, err)
			continue
		}
		if len(extensions) > 0 {
			extensionSets = append(extensionSets, extensions)
		}
	}

	return mergeExtensions(extensionSets...), nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run TestListExtensionsFromDirs`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add listExtensionsFromDirs to scan multiple directories

- Scan all provided extension directories
- Merge and deduplicate results
- Continue on errors with warnings
- Added integration test with multiple installations

Related to #11"
```

---

## Task 8: Update ListExtensions to use directory fallback

**Files:**
- Modify: `agent/internal/vscode/vscode.go:34-75`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run TestListExtensionsWithFallback`
Expected: Test fails - getExtensionDirs is not a variable, can't be overridden

**Step 3: Refactor to make getExtensionDirs testable and update ListExtensions**

Modify `agent/internal/vscode/vscode.go`:

First, make getExtensionDirs a variable so it can be overridden in tests:

```go
// Variable to allow overriding in tests
var getExtensionDirs = getExtensionDirsImpl

// getExtensionDirsImpl returns extension directory paths for VS Code and Insiders
func getExtensionDirsImpl() []string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows
	}

	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Application Support", "Code", "extensions"),
			filepath.Join(home, "Library", "Application Support", "Code - Insiders", "extensions"),
		}
	case "windows":
		return []string{
			filepath.Join(home, ".vscode", "extensions"),
			filepath.Join(home, ".vscode-insiders", "extensions"),
		}
	case "linux":
		return []string{
			filepath.Join(home, ".vscode", "extensions"),
			filepath.Join(home, ".vscode-insiders", "extensions"),
		}
	default:
		return []string{}
	}
}
```

Then update the ListExtensions function:

```go
// ListExtensions returns a list of installed VS Code extensions
// Tries CLI first, falls back to directory parsing on failure
func ListExtensions() ([]Extension, error) {
	// Try CLI method first
	extensions, err := listExtensionsViaCLI()
	if err == nil {
		return extensions, nil
	}

	// CLI failed, log and fall back to directory parsing
	log.Printf("CLI method failed (%v), falling back to directory parsing", err)
	return listExtensionsFromDirs(getExtensionDirs())
}

// listExtensionsViaCLI lists extensions using the VS Code CLI
func listExtensionsViaCLI() ([]Extension, error) {
	cmd := exec.Command("code", "--list-extensions", "--show-versions")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("VS Code CLI error: %w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute VS Code CLI: %w", err)
	}

	// Parse output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	extensions := make([]Extension, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: "publisher.name@version"
		parts := strings.Split(line, "@")
		if len(parts) != 2 {
			// If no version, treat entire line as ID
			extensions = append(extensions, Extension{
				ID:      line,
				Version: "",
				Enabled: true,
			})
			continue
		}

		extensions = append(extensions, Extension{
			ID:      parts[0],
			Version: parts[1],
			Enabled: true,
		})
	}

	return extensions, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run TestListExtensionsWithFallback`
Expected: PASS

**Step 5: Run all vscode tests to ensure nothing broke**

Run: `cd agent && go test ./internal/vscode -v`
Expected: All tests PASS

**Step 6: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: implement CLI-first with directory parsing fallback

- ListExtensions now tries CLI first, falls back to directories
- Extracted listExtensionsViaCLI for separation of concerns
- Made getExtensionDirs testable via variable override
- Added test for fallback behavior
- Logs fallback activation for debugging

Related to #11"
```

---

## Task 9: Add state file paths function

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
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

	// Paths should contain "storage.json"
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
		if home == "" {
			t.Skip("HOME or USERPROFILE not set")
		}
	}

	paths := getStatePaths()

	switch runtime.GOOS {
	case "linux":
		expectedStable := filepath.Join(home, ".config", "Code", "User", "globalStorage", "storage.json")
		found := false
		for _, path := range paths {
			if path == expectedStable {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected Linux stable path not found in %v", paths)
		}
	case "darwin":
		expectedStable := filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "storage.json")
		found := false
		for _, path := range paths {
			if path == expectedStable {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected macOS stable path not found in %v", paths)
		}
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			t.Skip("APPDATA not set")
		}
		expectedStable := filepath.Join(appdata, "Code", "User", "globalStorage", "storage.json")
		found := false
		for _, path := range paths {
			if path == expectedStable {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected Windows stable path not found in %v", paths)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run "TestGetStatePaths"`
Expected: Compilation error - function doesn't exist

**Step 3: Implement getStatePaths**

Add to `agent/internal/vscode/vscode.go`:

```go
// getStatePaths returns state file paths for VS Code and Insiders
func getStatePaths() []string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows
	}

	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "storage.json"),
			filepath.Join(home, "Library", "Application Support", "Code - Insiders", "User", "globalStorage", "storage.json"),
		}
	case "windows":
		appdata := os.Getenv("APPDATA")
		return []string{
			filepath.Join(appdata, "Code", "User", "globalStorage", "storage.json"),
			filepath.Join(appdata, "Code - Insiders", "User", "globalStorage", "storage.json"),
		}
	case "linux":
		return []string{
			filepath.Join(home, ".config", "Code", "User", "globalStorage", "storage.json"),
			filepath.Join(home, ".config", "Code - Insiders", "User", "globalStorage", "storage.json"),
		}
	default:
		return []string{}
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run "TestGetStatePaths"`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add getStatePaths for extension state files

- Returns state file paths for VS Code and Insiders
- Platform-specific paths for macOS, Windows, Linux
- Points to storage.json in globalStorage directory
- Added comprehensive tests for all platforms

Related to #11"
```

---

## Task 10: Add function to load disabled extensions from state

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
func TestLoadDisabledExtensions(t *testing.T) {
	tests := []struct {
		name     string
		stateJSON string
		want     map[string]bool
		wantErr  bool
	}{
		{
			name: "valid state with disabled extensions",
			stateJSON: `{
				"extensionsIdentifiers/disabled": [
					"publisher1.ext1",
					"publisher2.ext2"
				]
			}`,
			want: map[string]bool{
				"publisher1.ext1": true,
				"publisher2.ext2": true,
			},
			wantErr: false,
		},
		{
			name: "empty disabled list",
			stateJSON: `{
				"extensionsIdentifiers/disabled": []
			}`,
			want:    map[string]bool{},
			wantErr: false,
		},
		{
			name:      "missing disabled key",
			stateJSON: `{"someOtherKey": "value"}`,
			want:      map[string]bool{},
			wantErr:   false,
		},
		{
			name:      "invalid json",
			stateJSON: `{invalid json}`,
			want:      map[string]bool{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with state JSON
			tmpFile := filepath.Join(t.TempDir(), "storage.json")
			err := os.WriteFile(tmpFile, []byte(tt.stateJSON), 0644)
			if err != nil {
				t.Fatal(err)
			}

			got, err := loadDisabledExtensions(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadDisabledExtensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("got %d disabled extensions, want %d", len(got), len(tt.want))
				}
				for id := range tt.want {
					if !got[id] {
						t.Errorf("expected %s to be disabled", id)
					}
				}
			}
		})
	}
}

func TestLoadDisabledExtensionsNonexistent(t *testing.T) {
	disabled, err := loadDisabledExtensions("/nonexistent/file.json")
	if err != nil {
		t.Errorf("expected no error for nonexistent file, got: %v", err)
	}
	if len(disabled) != 0 {
		t.Errorf("expected empty map for nonexistent file, got %d entries", len(disabled))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run "TestLoadDisabledExtensions"`
Expected: Compilation error - function doesn't exist

**Step 3: Implement loadDisabledExtensions**

Add to `agent/internal/vscode/vscode.go`:

```go
// loadDisabledExtensions loads the list of disabled extensions from a state file
func loadDisabledExtensions(statePath string) (map[string]bool, error) {
	// Check if file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return map[string]bool{}, nil
	}

	// Read state file
	data, err := os.ReadFile(statePath)
	if err != nil {
		return map[string]bool{}, fmt.Errorf("failed to read state file: %w", err)
	}

	// Parse JSON
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return map[string]bool{}, fmt.Errorf("failed to parse state JSON: %w", err)
	}

	// Extract disabled extensions list
	disabledKey := "extensionsIdentifiers/disabled"
	disabledList, exists := state[disabledKey]
	if !exists {
		return map[string]bool{}, nil
	}

	// Convert to map
	disabled := make(map[string]bool)
	if list, ok := disabledList.([]interface{}); ok {
		for _, item := range list {
			if id, ok := item.(string); ok {
				disabled[id] = true
			}
		}
	}

	return disabled, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run "TestLoadDisabledExtensions"`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add loadDisabledExtensions to parse state files

- Parse storage.json to extract disabled extensions
- Return map for fast lookup
- Handle missing files gracefully
- Handle invalid JSON with error
- Added comprehensive tests with fixtures

Related to #11"
```

---

## Task 11: Add function to apply enabled state to extensions

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
func TestApplyEnabledState(t *testing.T) {
	tests := []struct {
		name       string
		extensions []Extension
		disabled   map[string]bool
		want       []Extension
	}{
		{
			name: "no disabled extensions",
			extensions: []Extension{
				{ID: "ext1.a", Version: "1.0.0", Enabled: true},
				{ID: "ext2.b", Version: "2.0.0", Enabled: true},
			},
			disabled: map[string]bool{},
			want: []Extension{
				{ID: "ext1.a", Version: "1.0.0", Enabled: true},
				{ID: "ext2.b", Version: "2.0.0", Enabled: true},
			},
		},
		{
			name: "some disabled extensions",
			extensions: []Extension{
				{ID: "ext1.a", Version: "1.0.0", Enabled: true},
				{ID: "ext2.b", Version: "2.0.0", Enabled: true},
				{ID: "ext3.c", Version: "3.0.0", Enabled: true},
			},
			disabled: map[string]bool{
				"ext2.b": true,
			},
			want: []Extension{
				{ID: "ext1.a", Version: "1.0.0", Enabled: true},
				{ID: "ext2.b", Version: "2.0.0", Enabled: false},
				{ID: "ext3.c", Version: "3.0.0", Enabled: true},
			},
		},
		{
			name: "all disabled",
			extensions: []Extension{
				{ID: "ext1.a", Version: "1.0.0", Enabled: true},
			},
			disabled: map[string]bool{
				"ext1.a": true,
			},
			want: []Extension{
				{ID: "ext1.a", Version: "1.0.0", Enabled: false},
			},
		},
		{
			name:       "empty extensions",
			extensions: []Extension{},
			disabled:   map[string]bool{"ext1.a": true},
			want:       []Extension{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyEnabledState(tt.extensions, tt.disabled)

			if len(got) != len(tt.want) {
				t.Errorf("got %d extensions, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].ID != tt.want[i].ID {
					t.Errorf("extension[%d].ID = %s, want %s", i, got[i].ID, tt.want[i].ID)
				}
				if got[i].Enabled != tt.want[i].Enabled {
					t.Errorf("extension[%d].Enabled = %v, want %v", i, got[i].Enabled, tt.want[i].Enabled)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run TestApplyEnabledState`
Expected: Compilation error - function doesn't exist

**Step 3: Implement applyEnabledState**

Add to `agent/internal/vscode/vscode.go`:

```go
// applyEnabledState updates extension enabled status based on disabled map
func applyEnabledState(extensions []Extension, disabled map[string]bool) []Extension {
	result := make([]Extension, len(extensions))

	for i, ext := range extensions {
		result[i] = ext
		if disabled[ext.ID] {
			result[i].Enabled = false
		}
	}

	return result
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run TestApplyEnabledState`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add applyEnabledState to set extension status

- Update extension Enabled field based on disabled map
- Fast O(1) lookup using map
- Returns new slice without modifying input
- Added comprehensive tests for various scenarios

Related to #11"
```

---

## Task 12: Integrate state detection into directory listing

**Files:**
- Modify: `agent/internal/vscode/vscode.go`
- Test: `agent/internal/vscode/vscode_test.go`

**Step 1: Write failing test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
func TestListExtensionsFromDirsWithState(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()

	// Extensions directory
	extDir := filepath.Join(tmpDir, "extensions")
	err := os.MkdirAll(extDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create two test extensions
	ext1Dir := filepath.Join(extDir, "test.ext1-1.0.0")
	err = os.MkdirAll(ext1Dir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest1 := `{
		"name": "ext1",
		"version": "1.0.0",
		"publisher": "test"
	}`
	err = os.WriteFile(filepath.Join(ext1Dir, "package.json"), []byte(manifest1), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ext2Dir := filepath.Join(extDir, "test.ext2-2.0.0")
	err = os.MkdirAll(ext2Dir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	manifest2 := `{
		"name": "ext2",
		"version": "2.0.0",
		"publisher": "test"
	}`
	err = os.WriteFile(filepath.Join(ext2Dir, "package.json"), []byte(manifest2), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// State directory
	stateDir := filepath.Join(tmpDir, "state")
	err = os.MkdirAll(stateDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create state file with ext2 disabled
	stateFile := filepath.Join(stateDir, "storage.json")
	stateJSON := `{
		"extensionsIdentifiers/disabled": ["test.ext2"]
	}`
	err = os.WriteFile(stateFile, []byte(stateJSON), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// List extensions with state
	extensions, err := listExtensionsFromDirsWithState([]string{extDir}, []string{stateFile})
	if err != nil {
		t.Fatalf("listExtensionsFromDirsWithState failed: %v", err)
	}

	// Should have 2 extensions
	if len(extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(extensions))
	}

	// Check ext1 is enabled
	foundExt1 := false
	for _, ext := range extensions {
		if ext.ID == "test.ext1" {
			foundExt1 = true
			if !ext.Enabled {
				t.Error("expected ext1 to be enabled")
			}
		}
	}
	if !foundExt1 {
		t.Error("ext1 not found")
	}

	// Check ext2 is disabled
	foundExt2 := false
	for _, ext := range extensions {
		if ext.ID == "test.ext2" {
			foundExt2 = true
			if ext.Enabled {
				t.Error("expected ext2 to be disabled")
			}
		}
	}
	if !foundExt2 {
		t.Error("ext2 not found")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd agent && go test ./internal/vscode -v -run TestListExtensionsFromDirsWithState`
Expected: Compilation error - function doesn't exist

**Step 3: Implement listExtensionsFromDirsWithState**

Add to `agent/internal/vscode/vscode.go`:

```go
// listExtensionsFromDirsWithState scans directories and applies state
func listExtensionsFromDirsWithState(extensionDirs []string, statePaths []string) ([]Extension, error) {
	// Get extensions from directories
	extensions, err := listExtensionsFromDirs(extensionDirs)
	if err != nil {
		return nil, err
	}

	// Load disabled extensions from all state files
	allDisabled := make(map[string]bool)
	for _, statePath := range statePaths {
		disabled, err := loadDisabledExtensions(statePath)
		if err != nil {
			log.Printf("Warning: failed to load state from %s: %v", statePath, err)
			continue
		}
		// Merge disabled maps
		for id := range disabled {
			allDisabled[id] = true
		}
	}

	// Apply enabled state
	return applyEnabledState(extensions, allDisabled), nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd agent && go test ./internal/vscode -v -run TestListExtensionsFromDirsWithState`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/vscode/vscode.go agent/internal/vscode/vscode_test.go
git commit -m "feat: add state integration to directory listing

- listExtensionsFromDirsWithState combines scanning and state
- Loads disabled extensions from multiple state files
- Applies enabled/disabled status to extensions
- Added integration test with fixtures

Related to #11"
```

---

## Task 13: Update ListExtensions fallback to include state

**Files:**
- Modify: `agent/internal/vscode/vscode.go`

**Step 1: Update ListExtensions function**

Modify the `ListExtensions` function in `agent/internal/vscode/vscode.go`:

```go
// ListExtensions returns a list of installed VS Code extensions
// Tries CLI first, falls back to directory parsing with state detection on failure
func ListExtensions() ([]Extension, error) {
	// Try CLI method first
	extensions, err := listExtensionsViaCLI()
	if err == nil {
		return extensions, nil
	}

	// CLI failed, log and fall back to directory parsing with state
	log.Printf("CLI method failed (%v), falling back to directory parsing", err)
	return listExtensionsFromDirsWithState(getExtensionDirs(), getStatePaths())
}
```

**Step 2: Run all tests**

Run: `cd agent && go test ./internal/vscode -v`
Expected: All tests PASS

**Step 3: Commit**

```bash
git add agent/internal/vscode/vscode.go
git commit -m "feat: integrate state detection into fallback mechanism

- ListExtensions fallback now includes state detection
- Provides complete extension info even without CLI
- Maintains backward compatibility with existing code

Related to #11"
```

---

## Task 14: Update existing ListExtensions test

**Files:**
- Modify: `agent/internal/vscode/vscode_test.go:17-31`

**Step 1: Update test to handle new fields**

Modify the existing `TestListExtensions` in `agent/internal/vscode/vscode_test.go`:

```go
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

	// If any extensions found, verify structure
	for _, ext := range extensions {
		if ext.ID == "" {
			t.Error("extension ID should not be empty")
		}
		// Version may be empty for CLI output without versions
		// DisplayName, Description, Publisher may be empty from CLI
		// (only populated when using directory parsing)
	}
}
```

**Step 2: Run test**

Run: `cd agent && go test ./internal/vscode -v -run TestListExtensions`
Expected: PASS (or SKIP if VS Code not installed)

**Step 3: Commit**

```bash
git add agent/internal/vscode/vscode_test.go
git commit -m "test: update ListExtensions test for new fields

- Handle new Extension struct fields
- Note that CLI method doesn't populate all fields
- Verify basic structure regardless of detection method

Related to #11"
```

---

## Task 15: Add integration test for end-to-end functionality

**Files:**
- Modify: `agent/internal/vscode/vscode_test.go`

**Step 1: Write integration test**

Add to `agent/internal/vscode/vscode_test.go`:

```go
func TestIntegrationFullExtensionDetection(t *testing.T) {
	// Create comprehensive test environment
	tmpDir := t.TempDir()

	// Create two "installations" (stable and insiders)
	stableExtDir := filepath.Join(tmpDir, "stable", "extensions")
	insidersExtDir := filepath.Join(tmpDir, "insiders", "extensions")
	err := os.MkdirAll(stableExtDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(insidersExtDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Stable: Python 2024.0.0 (enabled), Go 0.39.0 (disabled)
	pythonStableDir := filepath.Join(stableExtDir, "ms-python.python-2024.0.0")
	err = os.MkdirAll(pythonStableDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	pythonManifest := `{
		"name": "python",
		"displayName": "Python",
		"description": "Python language support",
		"version": "2024.0.0",
		"publisher": "ms-python"
	}`
	err = os.WriteFile(filepath.Join(pythonStableDir, "package.json"), []byte(pythonManifest), 0644)
	if err != nil {
		t.Fatal(err)
	}

	goStableDir := filepath.Join(stableExtDir, "golang.go-0.39.0")
	err = os.MkdirAll(goStableDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	goManifest := `{
		"name": "go",
		"displayName": "Go",
		"description": "Go language support",
		"version": "0.39.0",
		"publisher": "golang"
	}`
	err = os.WriteFile(filepath.Join(goStableDir, "package.json"), []byte(goManifest), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Insiders: Python 2024.1.0 (newer), Rust 1.0.0 (unique)
	pythonInsidersDir := filepath.Join(insidersExtDir, "ms-python.python-2024.1.0")
	err = os.MkdirAll(pythonInsidersDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	pythonInsidersManifest := `{
		"name": "python",
		"displayName": "Python Insiders",
		"description": "Python language support (Insiders)",
		"version": "2024.1.0",
		"publisher": "ms-python"
	}`
	err = os.WriteFile(filepath.Join(pythonInsidersDir, "package.json"), []byte(pythonInsidersManifest), 0644)
	if err != nil {
		t.Fatal(err)
	}

	rustInsidersDir := filepath.Join(insidersExtDir, "rust-lang.rust-1.0.0")
	err = os.MkdirAll(rustInsidersDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	rustManifest := `{
		"name": "rust",
		"displayName": "Rust",
		"description": "Rust language support",
		"version": "1.0.0",
		"publisher": "rust-lang"
	}`
	err = os.WriteFile(filepath.Join(rustInsidersDir, "package.json"), []byte(rustManifest), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create state files
	stableStateDir := filepath.Join(tmpDir, "stable", "state")
	err = os.MkdirAll(stableStateDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	stableStateFile := filepath.Join(stableStateDir, "storage.json")
	stableState := `{
		"extensionsIdentifiers/disabled": ["golang.go"]
	}`
	err = os.WriteFile(stableStateFile, []byte(stableState), 0644)
	if err != nil {
		t.Fatal(err)
	}

	insidersStateDir := filepath.Join(tmpDir, "insiders", "state")
	err = os.MkdirAll(insidersStateDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	insidersStateFile := filepath.Join(insidersStateDir, "storage.json")
	insidersState := `{
		"extensionsIdentifiers/disabled": []
	}`
	err = os.WriteFile(insidersStateFile, []byte(insidersState), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Run detection
	extensions, err := listExtensionsFromDirsWithState(
		[]string{stableExtDir, insidersExtDir},
		[]string{stableStateFile, insidersStateFile},
	)
	if err != nil {
		t.Fatalf("detection failed: %v", err)
	}

	// Verify results
	// Should have 3 extensions: Python (newer from Insiders), Go (from stable, disabled), Rust (from Insiders)
	if len(extensions) != 3 {
		t.Errorf("expected 3 extensions, got %d", len(extensions))
	}

	// Check Python - should be Insiders version
	var pythonExt *Extension
	for i := range extensions {
		if extensions[i].ID == "ms-python.python" {
			pythonExt = &extensions[i]
			break
		}
	}
	if pythonExt == nil {
		t.Fatal("Python extension not found")
	}
	if pythonExt.Version != "2024.1.0" {
		t.Errorf("expected Python v2024.1.0, got v%s", pythonExt.Version)
	}
	if pythonExt.DisplayName != "Python Insiders" {
		t.Errorf("expected DisplayName 'Python Insiders', got '%s'", pythonExt.DisplayName)
	}
	if !pythonExt.Enabled {
		t.Error("expected Python to be enabled")
	}

	// Check Go - should be stable version and disabled
	var goExt *Extension
	for i := range extensions {
		if extensions[i].ID == "golang.go" {
			goExt = &extensions[i]
			break
		}
	}
	if goExt == nil {
		t.Fatal("Go extension not found")
	}
	if goExt.Version != "0.39.0" {
		t.Errorf("expected Go v0.39.0, got v%s", goExt.Version)
	}
	if goExt.Enabled {
		t.Error("expected Go to be disabled")
	}

	// Check Rust - should be from Insiders and enabled
	var rustExt *Extension
	for i := range extensions {
		if extensions[i].ID == "rust-lang.rust" {
			rustExt = &extensions[i]
			break
		}
	}
	if rustExt == nil {
		t.Fatal("Rust extension not found")
	}
	if rustExt.Version != "1.0.0" {
		t.Errorf("expected Rust v1.0.0, got v%s", rustExt.Version)
	}
	if !rustExt.Enabled {
		t.Error("expected Rust to be enabled")
	}
}
```

**Step 2: Run test**

Run: `cd agent && go test ./internal/vscode -v -run TestIntegrationFullExtensionDetection`
Expected: PASS

**Step 3: Commit**

```bash
git add agent/internal/vscode/vscode_test.go
git commit -m "test: add comprehensive integration test

- Tests complete workflow: scan, merge, deduplicate, state
- Multiple installations with overlapping extensions
- Version comparison and selection
- State detection and application
- Validates all design requirements

Related to #11"
```

---

## Task 16: Run full test suite and verify coverage

**Files:**
- None (running tests)

**Step 1: Run all vscode tests**

Run: `cd agent && go test ./internal/vscode -v -cover`

**Step 2: Verify coverage is > 80%**

If coverage is below 80%, identify untested code paths and add tests.

**Step 3: Run tests for entire agent**

Run: `cd agent && go test ./... -v`

**Step 4: Document results**

Create a summary of test results:
- Total tests run
- Coverage percentage
- Any failures or issues

---

## Task 17: Update documentation

**Files:**
- Modify: `docs/plans/2026-02-04-vscode-extension-detection-design.md`

**Step 1: Update design document status**

Modify the header of the design document:

```markdown
**Status:** ✅ Implemented
```

Add implementation notes section at the end:

```markdown
## Implementation Notes

**Completed:** 2026-02-04

### Key Implementation Details

- CLI fallback is transparent to callers
- Version comparison uses `golang.org/x/mod/semver`
- State detection works for both stable and Insiders
- Comprehensive test coverage with fixtures
- All error conditions handled gracefully

### Test Coverage

- Unit tests: 100% of public functions
- Integration tests: End-to-end scenarios
- Platform-specific tests: All OS paths verified
- Edge cases: Invalid JSON, missing files, permission errors

### Known Limitations

- State detection only handles global state (not workspace-specific)
- Assumes standard VS Code directory structure
- Requires read access to extension and state directories
```

**Step 2: Commit documentation update**

```bash
git add docs/plans/2026-02-04-vscode-extension-detection-design.md
git commit -m "docs: mark extension detection design as implemented

- Updated status to completed
- Added implementation notes
- Documented test coverage
- Listed known limitations

Related to #11"
```

---

## Task 18: Final commit and push

**Files:**
- All modified files

**Step 1: Review all changes**

Run: `cd /home/mark/Projects/devtools-sync/.worktrees/feature/issue-11-extension-detection && git log --oneline origin/main..HEAD`

**Step 2: Run final test suite**

Run: `cd agent && go test ./... -v`

**Step 3: Create final summary commit if needed**

If there are any uncommitted changes, commit them.

**Step 4: Push branch**

Run: `git push -u origin feature/issue-11-extension-detection`

---

## Success Criteria Verification

After implementation, verify these criteria are met:

- ✅ All installed extensions are detected from both VS Code and Insiders
- ✅ Extension metadata (ID, version, displayName, description, publisher) is accurate
- ✅ Works on Windows, macOS, and Linux
- ✅ Gracefully handles edge cases (corrupted files, permission errors)
- ✅ Falls back to directory parsing when CLI unavailable
- ✅ Test coverage > 80%
- ✅ No breaking changes to existing code

## Notes

- Use TDD approach: write test, see it fail, implement, see it pass, commit
- Keep commits small and focused
- Run tests after each step
- Log warnings for skip-and-continue scenarios
- Follow existing code style in the vscode package
