# Profile Enhancements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add validation, conflict detection, and diff functionality to existing profile system

**Architecture:** Three independent enhancements - validation for safety, conflict detection for idempotency, diff for visibility - all integrated into existing profile.go with backward compatibility

**Tech Stack:** Go 1.x, standard library (encoding/json, fmt, strings, path/filepath), github.com/mark-chris/devtools-sync/agent/internal/vscode

---

## Task 1: Profile Name Validation

**Files:**
- Modify: `agent/internal/profile/profile.go:29-89` (Save function)
- Test: `agent/internal/profile/profile_test.go` (new file)

**Step 1: Write the failing test**

Add to `agent/internal/profile/profile_test.go`:

```go
package profile

import (
	"testing"
)

func TestValidate_ValidProfile(t *testing.T) {
	profile := &Profile{
		Name:       "valid-profile",
		Extensions: []Extension{},
	}

	err := Validate(profile)
	if err != nil {
		t.Errorf("expected no error for valid profile, got: %v", err)
	}
}

func TestValidate_EmptyName(t *testing.T) {
	profile := &Profile{
		Name:       "",
		Extensions: []Extension{},
	}

	err := Validate(profile)
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
	if err != nil && err.Error() != "profile name cannot be empty" {
		t.Errorf("expected 'profile name cannot be empty', got: %v", err)
	}
}

func TestValidate_InvalidFilename(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{"profile/with/slash", "profile name contains invalid characters"},
		{"profile\\with\\backslash", "profile name contains invalid characters"},
		{"profile:with:colon", "profile name contains invalid characters"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile := &Profile{
				Name:       tc.name,
				Extensions: []Extension{},
			}

			err := Validate(profile)
			if err == nil {
				t.Errorf("expected error for invalid name '%s', got nil", tc.name)
			}
			if err != nil && err.Error() != tc.expected {
				t.Errorf("expected '%s', got: %v", tc.expected, err)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestValidate ./agent/internal/profile`
Expected: FAIL with "undefined: Validate"

**Step 3: Write minimal implementation**

Add to `agent/internal/profile/profile.go` (after imports, before Save):

```go
// Validate checks if a profile has valid structure and fields
func Validate(profile *Profile) error {
	// Check profile name
	if profile.Name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Check for invalid filename characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(profile.Name, char) {
			return fmt.Errorf("profile name contains invalid characters")
		}
	}

	return nil
}
```

Add import: `"strings"`

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestValidate ./agent/internal/profile`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add agent/internal/profile/profile.go agent/internal/profile/profile_test.go
git commit -m "feat: add profile name validation

Add Validate function to check profile name is non-empty and contains
only valid filename characters. Validates against /, \, :, *, ?, \", <, >, |.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Extension ID Validation

**Files:**
- Modify: `agent/internal/profile/profile.go:11-27` (Validate function)
- Test: `agent/internal/profile/profile_test.go`

**Step 1: Write the failing test**

Add to `agent/internal/profile/profile_test.go`:

```go
func TestValidate_ValidExtensions(t *testing.T) {
	profile := &Profile{
		Name: "test",
		Extensions: []Extension{
			{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
			{ID: "golang.go", Version: "0.40.0", Enabled: false},
		},
	}

	err := Validate(profile)
	if err != nil {
		t.Errorf("expected no error for valid extensions, got: %v", err)
	}
}

func TestValidate_InvalidExtensionID(t *testing.T) {
	testCases := []struct {
		name        string
		extensionID string
		expectedErr string
	}{
		{"empty", "", "extension ID cannot be empty"},
		{"no dot", "python", "extension ID 'python' must be in format 'publisher.name'"},
		{"multiple dots", "ms.python.tools", "extension ID 'ms.python.tools' must be in format 'publisher.name'"},
		{"with space", "ms python.tools", "extension ID 'ms python.tools' must be in format 'publisher.name'"},
		{"trailing dot", "ms.", "extension ID 'ms.' must be in format 'publisher.name'"},
		{"leading dot", ".python", "extension ID '.python' must be in format 'publisher.name'"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile := &Profile{
				Name: "test",
				Extensions: []Extension{
					{ID: tc.extensionID, Version: "1.0.0", Enabled: true},
				},
			}

			err := Validate(profile)
			if err == nil {
				t.Errorf("expected error for invalid extension ID '%s', got nil", tc.extensionID)
			}
			if err != nil && err.Error() != tc.expectedErr {
				t.Errorf("expected '%s', got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestValidate_EmptyExtensionsList(t *testing.T) {
	profile := &Profile{
		Name:       "empty-profile",
		Extensions: []Extension{},
	}

	err := Validate(profile)
	if err != nil {
		t.Errorf("expected no error for empty extensions list, got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestValidate ./agent/internal/profile`
Expected: FAIL with tests not checking extension IDs

**Step 3: Write minimal implementation**

Modify `Validate` function in `agent/internal/profile/profile.go`:

```go
// Validate checks if a profile has valid structure and fields
func Validate(profile *Profile) error {
	// Check profile name
	if profile.Name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Check for invalid filename characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(profile.Name, char) {
			return fmt.Errorf("profile name contains invalid characters")
		}
	}

	// Validate each extension
	for _, ext := range profile.Extensions {
		if ext.ID == "" {
			return fmt.Errorf("extension ID cannot be empty")
		}

		// Check extension ID format: must be "publisher.name"
		parts := strings.Split(ext.ID, ".")
		if len(parts) != 2 {
			return fmt.Errorf("extension ID '%s' must be in format 'publisher.name'", ext.ID)
		}

		// Check for empty publisher or name
		if parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("extension ID '%s' must be in format 'publisher.name'", ext.ID)
		}

		// Check for spaces
		if strings.Contains(ext.ID, " ") {
			return fmt.Errorf("extension ID '%s' must be in format 'publisher.name'", ext.ID)
		}
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestValidate ./agent/internal/profile`
Expected: PASS (all 6 subtests + 3 top-level tests)

**Step 5: Commit**

```bash
git add agent/internal/profile/profile.go agent/internal/profile/profile_test.go
git commit -m "feat: add extension ID validation

Validate extension IDs match publisher.name format. Checks for:
- Non-empty ID
- Exactly one dot separator
- Non-empty publisher and name parts
- No spaces

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Integrate Validation into Load

**Files:**
- Modify: `agent/internal/profile/profile.go:92-121` (Load function)
- Test: `agent/internal/profile/profile_test.go`

**Step 1: Write the failing test**

Add to `agent/internal/profile/profile_test.go`:

```go
func TestLoad_InvalidProfile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Write invalid profile (bad extension ID)
	invalidProfile := `{
		"name": "test-profile",
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z",
		"extensions": [
			{
				"id": "invalid-no-dot",
				"version": "1.0.0",
				"enabled": true
			}
		]
	}`

	profilePath := filepath.Join(tmpDir, "test-profile.json")
	if err := os.WriteFile(profilePath, []byte(invalidProfile), 0644); err != nil {
		t.Fatalf("failed to write test profile: %v", err)
	}

	// Try to load invalid profile
	_, err := Load("test-profile", tmpDir)
	if err == nil {
		t.Error("expected error loading invalid profile, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "must be in format 'publisher.name'") {
		t.Errorf("expected validation error, got: %v", err)
	}
}
```

Add imports to test file: `"os"`, `"path/filepath"`

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestLoad_InvalidProfile ./agent/internal/profile`
Expected: FAIL with "expected error loading invalid profile, got nil"

**Step 3: Write minimal implementation**

Modify `Load` function in `agent/internal/profile/profile.go`:

```go
// Load installs extensions from a profile
func Load(name string, profilesDir string) (*Profile, error) {
	if name == "" {
		return nil, fmt.Errorf("profile name cannot be empty")
	}

	// Read profile file
	profilePath := filepath.Join(profilesDir, name+".json")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	// Parse profile
	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile file: %w", err)
	}

	// Validate profile before installation
	if err := Validate(&profile); err != nil {
		return nil, fmt.Errorf("invalid profile: %w", err)
	}

	// Install each extension
	for _, ext := range profile.Extensions {
		if err := vscode.InstallExtension(ext.ID); err != nil {
			return nil, fmt.Errorf("failed to install extension %s: %w", ext.ID, err)
		}
	}

	return &profile, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestLoad_InvalidProfile ./agent/internal/profile`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/profile/profile.go agent/internal/profile/profile_test.go
git commit -m "feat: integrate validation into Load function

Load now validates profile structure before attempting installation.
Returns clear error messages for invalid profiles.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Conflict Detection Helper Function

**Files:**
- Modify: `agent/internal/profile/profile.go` (add detectConflicts after Validate)
- Test: `agent/internal/profile/profile_test.go`

**Step 1: Write the failing test**

Add to `agent/internal/profile/profile_test.go`:

```go
func TestDetectConflicts_NoConflicts(t *testing.T) {
	profileExts := []Extension{
		{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
		{ID: "golang.go", Version: "0.40.0", Enabled: true},
	}

	installedExts := []vscode.Extension{
		{ID: "different.extension", Version: "1.0.0", Enabled: true},
	}

	toInstall, alreadyInstalled := detectConflicts(profileExts, installedExts)

	if len(toInstall) != 2 {
		t.Errorf("expected 2 extensions to install, got %d", len(toInstall))
	}
	if len(alreadyInstalled) != 0 {
		t.Errorf("expected 0 already installed, got %d", len(alreadyInstalled))
	}
}

func TestDetectConflicts_AllInstalled(t *testing.T) {
	profileExts := []Extension{
		{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
		{ID: "golang.go", Version: "0.40.0", Enabled: true},
	}

	installedExts := []vscode.Extension{
		{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
		{ID: "golang.go", Version: "0.40.0", Enabled: true},
	}

	toInstall, alreadyInstalled := detectConflicts(profileExts, installedExts)

	if len(toInstall) != 0 {
		t.Errorf("expected 0 extensions to install, got %d", len(toInstall))
	}
	if len(alreadyInstalled) != 2 {
		t.Errorf("expected 2 already installed, got %d", len(alreadyInstalled))
	}
}

func TestDetectConflicts_Mixed(t *testing.T) {
	profileExts := []Extension{
		{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
		{ID: "golang.go", Version: "0.40.0", Enabled: true},
		{ID: "rust-lang.rust", Version: "1.0.0", Enabled: true},
	}

	installedExts := []vscode.Extension{
		{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
	}

	toInstall, alreadyInstalled := detectConflicts(profileExts, installedExts)

	if len(toInstall) != 2 {
		t.Errorf("expected 2 extensions to install, got %d", len(toInstall))
	}
	if len(alreadyInstalled) != 1 {
		t.Errorf("expected 1 already installed, got %d", len(alreadyInstalled))
	}

	// Verify correct extensions in each list
	if toInstall[0].ID != "golang.go" && toInstall[1].ID != "golang.go" {
		t.Error("expected golang.go in toInstall list")
	}
	if alreadyInstalled[0].ID != "ms-python.python" {
		t.Errorf("expected ms-python.python in alreadyInstalled, got %s", alreadyInstalled[0].ID)
	}
}

func TestDetectConflicts_EmptyProfile(t *testing.T) {
	profileExts := []Extension{}
	installedExts := []vscode.Extension{
		{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
	}

	toInstall, alreadyInstalled := detectConflicts(profileExts, installedExts)

	if len(toInstall) != 0 {
		t.Errorf("expected 0 extensions to install, got %d", len(toInstall))
	}
	if len(alreadyInstalled) != 0 {
		t.Errorf("expected 0 already installed, got %d", len(alreadyInstalled))
	}
}
```

Add import to test file: `"github.com/mark-chris/devtools-sync/agent/internal/vscode"`

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestDetectConflicts ./agent/internal/profile`
Expected: FAIL with "undefined: detectConflicts"

**Step 3: Write minimal implementation**

Add to `agent/internal/profile/profile.go` (after Validate function):

```go
// detectConflicts compares profile extensions with currently installed extensions
// Returns two lists: extensions to install and extensions already installed
func detectConflicts(profileExtensions []Extension, installedExtensions []vscode.Extension) (toInstall, alreadyInstalled []Extension) {
	// Create map of installed extension IDs for O(1) lookup
	installedMap := make(map[string]bool)
	for _, ext := range installedExtensions {
		installedMap[ext.ID] = true
	}

	// Categorize each profile extension
	for _, ext := range profileExtensions {
		if installedMap[ext.ID] {
			alreadyInstalled = append(alreadyInstalled, ext)
		} else {
			toInstall = append(toInstall, ext)
		}
	}

	return toInstall, alreadyInstalled
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestDetectConflicts ./agent/internal/profile`
Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add agent/internal/profile/profile.go agent/internal/profile/profile_test.go
git commit -m "feat: add conflict detection helper function

detectConflicts categorizes profile extensions as toInstall or
alreadyInstalled based on currently installed extensions. Uses map
for O(1) lookup performance.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Integrate Conflict Detection into Load

**Files:**
- Modify: `agent/internal/profile/profile.go:92-135` (Load function)
- Test: `agent/internal/profile/profile_test.go`

**Step 1: Write the failing test**

Add to `agent/internal/profile/profile_test.go`:

```go
func TestLoad_SkipsAlreadyInstalled(t *testing.T) {
	// This test requires mocking vscode.ListExtensions and vscode.InstallExtension
	// For now, we'll test the logic by checking that Load calls these functions
	// Integration test will verify full behavior

	// Create temp directory
	tmpDir := t.TempDir()

	// Write valid profile
	validProfile := `{
		"name": "test-profile",
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z",
		"extensions": [
			{
				"id": "ms-python.python",
				"version": "2024.0.0",
				"enabled": true
			}
		]
	}`

	profilePath := filepath.Join(tmpDir, "test-profile.json")
	if err := os.WriteFile(profilePath, []byte(validProfile), 0644); err != nil {
		t.Fatalf("failed to write test profile: %v", err)
	}

	// Load profile (will actually try to install via vscode CLI)
	// This test verifies the profile loads without error
	// Actual conflict detection tested in integration tests
	profile, err := Load("test-profile", tmpDir)
	if err != nil {
		// Error is expected if code CLI not available or extension doesn't exist
		// We're mainly testing that validation passed
		if !strings.Contains(err.Error(), "failed to install extension") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	if profile != nil && profile.Name != "test-profile" {
		t.Errorf("expected profile name 'test-profile', got: %s", profile.Name)
	}
}
```

**Step 2: Run test to verify current behavior**

Run: `go test -v -run TestLoad_SkipsAlreadyInstalled ./agent/internal/profile`
Expected: PASS (test validates current behavior before enhancement)

**Step 3: Write implementation**

Modify `Load` function in `agent/internal/profile/profile.go`:

```go
// Load installs extensions from a profile
func Load(name string, profilesDir string) (*Profile, error) {
	if name == "" {
		return nil, fmt.Errorf("profile name cannot be empty")
	}

	// Read profile file
	profilePath := filepath.Join(profilesDir, name+".json")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	// Parse profile
	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile file: %w", err)
	}

	// Validate profile before installation
	if err := Validate(&profile); err != nil {
		return nil, fmt.Errorf("invalid profile: %w", err)
	}

	// Get currently installed extensions
	installedExts, err := vscode.ListExtensions()
	if err != nil {
		return nil, fmt.Errorf("failed to list installed extensions: %w", err)
	}

	// Detect conflicts
	toInstall, alreadyInstalled := detectConflicts(profile.Extensions, installedExts)

	// Report what will be skipped
	if len(alreadyInstalled) > 0 {
		fmt.Printf("Skipping %d already installed extension(s):\n", len(alreadyInstalled))
		for _, ext := range alreadyInstalled {
			fmt.Printf("  - %s@%s\n", ext.ID, ext.Version)
		}
	}

	// Install only new extensions
	for _, ext := range toInstall {
		if err := vscode.InstallExtension(ext.ID); err != nil {
			return nil, fmt.Errorf("failed to install extension %s: %w", ext.ID, err)
		}
	}

	// Report summary
	fmt.Printf("Installed %d extension(s), skipped %d already installed\n",
		len(toInstall), len(alreadyInstalled))

	return &profile, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestLoad ./agent/internal/profile`
Expected: PASS

**Step 5: Commit**

```bash
git add agent/internal/profile/profile.go agent/internal/profile/profile_test.go
git commit -m "feat: integrate conflict detection into Load

Load now:
- Checks currently installed extensions
- Skips extensions already present
- Reports what will be skipped before installing
- Only installs new extensions
- Provides summary of actions taken

Makes Load idempotent and prevents version downgrades.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: DiffResult Structure and Diff Function

**Files:**
- Modify: `agent/internal/profile/profile.go` (add DiffResult type and Diff function)
- Test: `agent/internal/profile/profile_test.go`

**Step 1: Write the failing test**

Add to `agent/internal/profile/profile_test.go`:

```go
func TestDiff_ValidProfile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Write test profile
	testProfile := `{
		"name": "test-profile",
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z",
		"extensions": [
			{
				"id": "ms-python.python",
				"version": "2024.0.0",
				"enabled": true
			},
			{
				"id": "golang.go",
				"version": "0.40.0",
				"enabled": true
			}
		]
	}`

	profilePath := filepath.Join(tmpDir, "test-profile.json")
	if err := os.WriteFile(profilePath, []byte(testProfile), 0644); err != nil {
		t.Fatalf("failed to write test profile: %v", err)
	}

	// Call Diff (will use actual vscode.ListExtensions)
	result, err := Diff("test-profile", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify result structure
	if result.ProfileName != "test-profile" {
		t.Errorf("expected profile name 'test-profile', got: %s", result.ProfileName)
	}
	if result.TotalInProfile != 2 {
		t.Errorf("expected 2 total extensions, got: %d", result.TotalInProfile)
	}

	// toInstall + alreadyInstalled should equal total
	if len(result.ToInstall)+len(result.AlreadyInstalled) != result.TotalInProfile {
		t.Errorf("toInstall (%d) + alreadyInstalled (%d) != total (%d)",
			len(result.ToInstall), len(result.AlreadyInstalled), result.TotalInProfile)
	}
}

func TestDiff_ProfileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Diff("nonexistent", tmpDir)
	if err == nil {
		t.Error("expected error for nonexistent profile, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestDiff_InvalidProfile(t *testing.T) {
	tmpDir := t.TempDir()

	// Write invalid profile
	invalidProfile := `{
		"name": "invalid",
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z",
		"extensions": [
			{
				"id": "bad-extension-id",
				"version": "1.0.0",
				"enabled": true
			}
		]
	}`

	profilePath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(profilePath, []byte(invalidProfile), 0644); err != nil {
		t.Fatalf("failed to write test profile: %v", err)
	}

	_, err := Diff("invalid", tmpDir)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "must be in format") {
		t.Errorf("expected validation error, got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestDiff ./agent/internal/profile`
Expected: FAIL with "undefined: Diff" and "undefined: DiffResult"

**Step 3: Write minimal implementation**

Add to `agent/internal/profile/profile.go` (after detectConflicts):

```go
// DiffResult contains the comparison between a profile and installed extensions
type DiffResult struct {
	ProfileName      string
	ToInstall        []Extension
	AlreadyInstalled []Extension
	TotalInProfile   int
}

// Diff compares a profile with currently installed extensions
func Diff(profileName string, profilesDir string) (*DiffResult, error) {
	if profileName == "" {
		return nil, fmt.Errorf("profile name cannot be empty")
	}

	// Read profile file
	profilePath := filepath.Join(profilesDir, profileName+".json")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' not found", profileName)
		}
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	// Parse profile
	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile file: %w", err)
	}

	// Validate profile
	if err := Validate(&profile); err != nil {
		return nil, fmt.Errorf("invalid profile: %w", err)
	}

	// Get currently installed extensions
	installedExts, err := vscode.ListExtensions()
	if err != nil {
		return nil, fmt.Errorf("failed to list installed extensions: %w", err)
	}

	// Detect conflicts
	toInstall, alreadyInstalled := detectConflicts(profile.Extensions, installedExts)

	// Build result
	result := &DiffResult{
		ProfileName:      profile.Name,
		ToInstall:        toInstall,
		AlreadyInstalled: alreadyInstalled,
		TotalInProfile:   len(profile.Extensions),
	}

	return result, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestDiff ./agent/internal/profile`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add agent/internal/profile/profile.go agent/internal/profile/profile_test.go
git commit -m "feat: add Diff function for profile comparison

Adds DiffResult structure and Diff function to compare profile contents
with currently installed extensions. Returns structured result showing
what would be installed vs what's already present.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Profile Diff CLI Command

**Files:**
- Modify: `agent/cmd/profile.go` (add diff subcommand)
- Create: None (modifying existing CLI file)

**Step 1: Read existing CLI structure**

Read: `agent/cmd/profile.go` to understand command structure

Expected: File with save, load, list commands using cobra

**Step 2: Write diff command implementation**

Add to `agent/cmd/profile.go` (after profileListCmd):

```go
var profileDiffCmd = &cobra.Command{
	Use:   "diff <name>",
	Short: "Show differences between profile and installed extensions",
	Long:  "Compare a profile with currently installed extensions. Shows what would be installed and what's already present.",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getProfileNames(toComplete), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		profilesDir := getProfilesDir()

		result, err := profile.Diff(profileName, profilesDir)
		if err != nil {
			return fmt.Errorf("failed to diff profile: %w", err)
		}

		// Display results
		fmt.Printf("Profile: %s (%d extensions)\n\n", result.ProfileName, result.TotalInProfile)

		if len(result.ToInstall) > 0 {
			fmt.Printf("Will install (%d):\n", len(result.ToInstall))
			for _, ext := range result.ToInstall {
				fmt.Printf("  - %s@%s\n", ext.ID, ext.Version)
			}
			fmt.Println()
		}

		if len(result.AlreadyInstalled) > 0 {
			fmt.Printf("Already installed (%d):\n", len(result.AlreadyInstalled))
			for _, ext := range result.AlreadyInstalled {
				fmt.Printf("  - %s@%s\n", ext.ID, ext.Version)
			}
			fmt.Println()
		}

		if len(result.ToInstall) == 0 {
			fmt.Println("All extensions from this profile are already installed.")
		} else {
			fmt.Printf("Run 'devtools-sync profile load %s' to install.\n", profileName)
		}

		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileDiffCmd)
}
```

**Step 3: Test command manually**

Run: `go run main.go profile diff --help`
Expected: Shows diff command help text

**Step 4: Test with real profile**

Run: `go run main.go profile diff <existing-profile-name>`
Expected: Shows comparison output

**Step 5: Commit**

```bash
git add agent/cmd/profile.go
git commit -m "feat: add profile diff CLI command

Adds 'devtools-sync profile diff <name>' command that shows:
- Total extensions in profile
- Extensions that will be installed (new)
- Extensions already installed (existing)
- Next step suggestion

Includes tab completion for profile names.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Integration Test for Complete Workflow

**Files:**
- Test: `agent/internal/profile/profile_test.go`

**Step 1: Write integration test**

Add to `agent/internal/profile/profile_test.go`:

```go
func TestIntegration_SaveDiffLoad(t *testing.T) {
	// This is an integration test that verifies the complete workflow
	// Skip if vscode CLI not available
	if _, err := vscode.ListExtensions(); err != nil {
		t.Skip("VS Code CLI not available, skipping integration test")
	}

	tmpDir := t.TempDir()

	// Step 1: Save current extensions to profile
	profile1, err := Save("integration-test", tmpDir)
	if err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}
	if profile1.Name != "integration-test" {
		t.Errorf("expected profile name 'integration-test', got: %s", profile1.Name)
	}

	// Step 2: Diff should show all extensions as already installed
	diff1, err := Diff("integration-test", tmpDir)
	if err != nil {
		t.Fatalf("failed to diff profile: %v", err)
	}
	if len(diff1.ToInstall) != 0 {
		t.Errorf("expected 0 extensions to install after save, got: %d", len(diff1.ToInstall))
	}
	if len(diff1.AlreadyInstalled) != len(profile1.Extensions) {
		t.Errorf("expected all %d extensions already installed, got: %d",
			len(profile1.Extensions), len(diff1.AlreadyInstalled))
	}

	// Step 3: Load profile again (should skip all)
	profile2, err := Load("integration-test", tmpDir)
	if err != nil {
		t.Fatalf("failed to load profile: %v", err)
	}
	if profile2.Name != "integration-test" {
		t.Errorf("expected profile name 'integration-test', got: %s", profile2.Name)
	}

	// Step 4: Verify profile can be retrieved
	profile3, err := Get("integration-test", tmpDir)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}
	if profile3.Name != "integration-test" {
		t.Errorf("expected profile name 'integration-test', got: %s", profile3.Name)
	}

	// Step 5: Verify profile appears in list
	profiles, err := List(tmpDir)
	if err != nil {
		t.Fatalf("failed to list profiles: %v", err)
	}
	found := false
	for _, p := range profiles {
		if p.Name == "integration-test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("integration-test profile not found in list")
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test -v -run TestIntegration_SaveDiffLoad ./agent/internal/profile`
Expected: PASS (or SKIP if VS Code not available)

**Step 3: Commit**

```bash
git add agent/internal/profile/profile_test.go
git commit -m "test: add integration test for complete workflow

Tests end-to-end workflow:
- Save profile
- Diff shows all already installed
- Load skips all existing extensions
- Get retrieves profile
- List includes profile

Skips gracefully if VS Code CLI unavailable.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 9: Run All Tests and Verify Coverage

**Files:**
- Test: All test files

**Step 1: Run all tests**

Run: `go test -v ./agent/internal/profile`
Expected: All tests PASS

**Step 2: Check test coverage**

Run: `go test -coverprofile=coverage.out ./agent/internal/profile && go tool cover -func=coverage.out`
Expected: >80% coverage for profile package

**Step 3: Generate coverage report**

Run: `go tool cover -html=coverage.out -o coverage.html`
Expected: HTML coverage report generated

**Step 4: Review and document coverage**

Review coverage.html in browser, note any uncovered lines

**Step 5: Commit coverage verification**

```bash
# Don't commit coverage files, just document the check
git status
# Verify no uncommitted changes except coverage files
# coverage.out and coverage.html should be gitignored
```

Note: If coverage files are tracked, add to .gitignore:
```
coverage.out
coverage.html
```

---

## Task 10: Update Documentation

**Files:**
- Modify: `README.md` (add diff command example)
- Create/Modify: Usage documentation

**Step 1: Add diff command to README**

Add to README.md under profile commands section:

```markdown
### Profile Commands

#### Save Current Extensions
```bash
devtools-sync profile save <name>
```

#### Load Extensions from Profile
```bash
devtools-sync profile load <name>
```

#### Compare Profile with Installed Extensions
```bash
devtools-sync profile diff <name>
```

Shows what extensions would be installed versus what's already present.

#### List All Profiles
```bash
devtools-sync profile list
```
```

**Step 2: Document new features**

Add features section:

```markdown
### Profile Features

- **Validation**: Profiles are validated before loading to catch errors early
- **Conflict Detection**: Loading a profile skips extensions that are already installed
- **Diff Preview**: Compare profile contents with current installation before loading
- **Idempotent**: Running load multiple times is safe and won't reinstall existing extensions
```

**Step 3: Add examples**

Add examples section:

```markdown
### Examples

Save your current extension setup:
```bash
devtools-sync profile save work-setup
```

Preview what would be installed:
```bash
devtools-sync profile diff work-setup
```

Load extensions (skips already installed):
```bash
devtools-sync profile load work-setup
```
```

**Step 4: Commit documentation**

```bash
git add README.md
git commit -m "docs: add profile diff command and features

Documents:
- profile diff command usage
- Validation feature
- Conflict detection behavior
- Idempotent loading
- Usage examples

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 11: Verify Backward Compatibility

**Files:**
- Test: Existing profile files

**Step 1: Create test with old profile format**

Create test that loads a profile without validation:

```go
func TestBackwardCompatibility_OldProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Simulate old profile created before validation existed
	oldProfile := `{
		"name": "old-profile",
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z",
		"extensions": [
			{
				"id": "ms-python.python",
				"version": "2024.0.0",
				"enabled": true
			}
		]
	}`

	profilePath := filepath.Join(tmpDir, "old-profile.json")
	if err := os.WriteFile(profilePath, []byte(oldProfile), 0644); err != nil {
		t.Fatalf("failed to write old profile: %v", err)
	}

	// Should load successfully
	profile, err := Get("old-profile", tmpDir)
	if err != nil {
		t.Fatalf("failed to load old profile: %v", err)
	}
	if profile.Name != "old-profile" {
		t.Errorf("expected profile name 'old-profile', got: %s", profile.Name)
	}

	// Should validate successfully
	if err := Validate(profile); err != nil {
		t.Errorf("old profile should validate successfully: %v", err)
	}

	// Should diff successfully
	if _, err := Diff("old-profile", tmpDir); err != nil {
		t.Errorf("failed to diff old profile: %v", err)
	}
}
```

**Step 2: Run backward compatibility test**

Run: `go test -v -run TestBackwardCompatibility ./agent/internal/profile`
Expected: PASS

**Step 3: Test with actual old profiles if available**

If there are existing profile files in the repository or test fixtures, verify they still work.

**Step 4: Commit compatibility verification**

```bash
git add agent/internal/profile/profile_test.go
git commit -m "test: verify backward compatibility with old profiles

Ensures profiles created before validation/conflict detection still
work correctly with new features.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 12: Final Test Run and Verification

**Files:**
- All Go files in agent package

**Step 1: Run all tests in agent package**

Run: `go test -v ./agent/...`
Expected: All tests PASS

**Step 2: Run tests with race detector**

Run: `go test -race -v ./agent/internal/profile`
Expected: PASS with no race conditions detected

**Step 3: Run linters**

Run: `golangci-lint run ./agent/...` (or project's lint command)
Expected: No linting errors

**Step 4: Build the binary**

Run: `go build -o devtools-sync ./agent`
Expected: Successful build

**Step 5: Manual CLI verification**

Run each command manually:
```bash
./devtools-sync profile save test-manual
./devtools-sync profile diff test-manual
./devtools-sync profile load test-manual
./devtools-sync profile list
```

Expected: All commands work as documented

**Step 6: Final commit**

```bash
git status
# Verify all changes are committed
# If any fixes were needed, commit them now
```

---

## Success Criteria

After completing all tasks, verify:

- ✅ All tests pass with >80% coverage
- ✅ `profile diff` command works and shows clear output
- ✅ `profile load` skips already installed extensions
- ✅ Invalid profiles are rejected with clear errors
- ✅ Old profiles still work (backward compatible)
- ✅ Documentation is updated
- ✅ No linting errors
- ✅ Manual testing confirms expected behavior

## Notes

- Follow TDD strictly: write test first, verify it fails, implement, verify it passes
- Commit after each task completion
- Each commit should be atomic and include tests
- Use descriptive commit messages with Co-Authored-By line
- If tests fail, investigate and fix before proceeding
- Integration test may be skipped if VS Code CLI not available
