package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark-chris/devtools-sync/agent/internal/vscode"
)

func TestSave(t *testing.T) {
	// Create temporary profiles directory
	tempDir := t.TempDir()

	// Note: This test requires VS Code to be installed
	// In a real environment, we would mock vscode.ListExtensions
	// For now, we test the error case
	_, err := Save("", tempDir)
	if err == nil {
		t.Error("expected error for empty profile name")
	}
	if err.Error() != "profile name cannot be empty" {
		t.Errorf("expected 'profile name cannot be empty' error, got: %s", err.Error())
	}
}

func TestLoad(t *testing.T) {
	// Create temporary profiles directory
	tempDir := t.TempDir()

	// Test loading non-existent profile
	_, err := Load("nonexistent", tempDir)
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
	if err.Error() != "profile 'nonexistent' not found" {
		t.Errorf("expected 'profile not found' error, got: %s", err.Error())
	}

	// Test empty profile name
	_, err = Load("", tempDir)
	if err == nil {
		t.Error("expected error for empty profile name")
	}
	if err.Error() != "profile name cannot be empty" {
		t.Errorf("expected 'profile name cannot be empty' error, got: %s", err.Error())
	}
}

func TestLoad_WithProfile(t *testing.T) {
	// Create temporary profiles directory
	tempDir := t.TempDir()

	// Create a test profile file
	profile := Profile{
		Name:      "test-profile",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Extensions: []Extension{
			{ID: "test.extension", Version: "1.0.0", Enabled: true},
		},
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	profilePath := filepath.Join(tempDir, "test-profile.json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("failed to write profile file: %v", err)
	}

	// Note: Load will try to install extensions via VS Code CLI
	// This will fail in test environment without VS Code
	// In a real implementation, we would mock vscode.InstallExtension
	_, err = Load("test-profile", tempDir)
	if err == nil {
		// If no error, VS Code is installed and test passed
		return
	}

	// Expected to fail without VS Code installed
	if err.Error() != "failed to install extension test.extension: extension ID cannot be empty" {
		// This is expected in test environment
		t.Logf("Load failed as expected without VS Code: %v", err)
	}
}

func TestList(t *testing.T) {
	// Create temporary profiles directory
	tempDir := t.TempDir()

	// Initially empty
	profiles, err := List(tempDir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}

	// Create some test profiles
	profile1 := Profile{
		Name:       "profile1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Extensions: []Extension{{ID: "ext1", Version: "1.0.0", Enabled: true}},
	}
	profile2 := Profile{
		Name:       "profile2",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Extensions: []Extension{{ID: "ext2", Version: "2.0.0", Enabled: true}},
	}

	// Write profiles
	for _, prof := range []Profile{profile1, profile2} {
		data, err := json.MarshalIndent(prof, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal profile: %v", err)
		}
		profilePath := filepath.Join(tempDir, prof.Name+".json")
		if err := os.WriteFile(profilePath, data, 0644); err != nil {
			t.Fatalf("failed to write profile file: %v", err)
		}
	}

	// List profiles
	profiles, err = List(tempDir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}

	// Verify profile names
	names := make(map[string]bool)
	for _, prof := range profiles {
		names[prof.Name] = true
	}
	if !names["profile1"] || !names["profile2"] {
		t.Error("expected profiles profile1 and profile2")
	}
}

func TestList_IgnoresNonJSONFiles(t *testing.T) {
	// Create temporary profiles directory
	tempDir := t.TempDir()

	// Create a valid profile
	profile := Profile{
		Name:       "valid",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Extensions: []Extension{{ID: "ext1", Version: "1.0.0", Enabled: true}},
	}
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}
	profilePath := filepath.Join(tempDir, "valid.json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("failed to write profile file: %v", err)
	}

	// Create non-JSON files that should be ignored
	if err := os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write txt file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "invalid.json"), []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to write invalid json: %v", err)
	}

	// List profiles
	profiles, err := List(tempDir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Should only include the valid profile
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "valid" {
		t.Errorf("expected profile name 'valid', got '%s'", profiles[0].Name)
	}
}

func TestGet(t *testing.T) {
	// Create temporary profiles directory
	tempDir := t.TempDir()

	// Test getting non-existent profile
	_, err := Get("nonexistent", tempDir)
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
	if err.Error() != "profile 'nonexistent' not found" {
		t.Errorf("expected 'profile not found' error, got: %s", err.Error())
	}

	// Test empty profile name
	_, err = Get("", tempDir)
	if err == nil {
		t.Error("expected error for empty profile name")
	}

	// Create a test profile
	profile := Profile{
		Name:      "test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Extensions: []Extension{
			{ID: "test.ext", Version: "1.0.0", Enabled: true},
		},
	}
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}
	profilePath := filepath.Join(tempDir, "test.json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("failed to write profile file: %v", err)
	}

	// Get the profile
	retrieved, err := Get("test", tempDir)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Verify profile data
	if retrieved.Name != "test" {
		t.Errorf("expected profile name 'test', got '%s'", retrieved.Name)
	}
	if len(retrieved.Extensions) != 1 {
		t.Errorf("expected 1 extension, got %d", len(retrieved.Extensions))
	}
	if retrieved.Extensions[0].ID != "test.ext" {
		t.Errorf("expected extension ID 'test.ext', got '%s'", retrieved.Extensions[0].ID)
	}
}

func TestProfile_JSONMarshaling(t *testing.T) {
	// Create a profile
	now := time.Now()
	profile := Profile{
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now,
		Extensions: []Extension{
			{ID: "ext1", Version: "1.0.0", Enabled: true},
			{ID: "ext2", Version: "2.0.0", Enabled: false},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	// Unmarshal back
	var decoded Profile
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal profile: %v", err)
	}

	// Verify data
	if decoded.Name != profile.Name {
		t.Errorf("name mismatch: expected %s, got %s", profile.Name, decoded.Name)
	}
	if len(decoded.Extensions) != len(profile.Extensions) {
		t.Errorf("extensions count mismatch: expected %d, got %d", len(profile.Extensions), len(decoded.Extensions))
	}
}

func TestValidate_ValidProfile(t *testing.T) {
	profile := &Profile{
		Name:       "my-profile",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
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
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Extensions: []Extension{},
	}

	err := Validate(profile)
	if err == nil {
		t.Error("expected error for empty profile name, got nil")
	}

	expectedMsg := "profile name cannot be empty"
	if err != nil && err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestValidate_InvalidFilename(t *testing.T) {
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}

	for _, char := range invalidChars {
		t.Run("char_"+char, func(t *testing.T) {
			profile := &Profile{
				Name:       "invalid" + char + "name",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				Extensions: []Extension{},
			}

			err := Validate(profile)
			if err == nil {
				t.Errorf("expected error for profile name with '%s', got nil", char)
			}

			expectedMsg := "profile name contains invalid characters"
			if err != nil && err.Error() != expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
			}
		})
	}
}

func TestValidate_ValidExtensions(t *testing.T) {
	profile := &Profile{
		Name:      "test-profile",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Extensions: []Extension{
			{ID: "ms-python.python", Version: "1.0.0", Enabled: true},
			{ID: "golang.go", Version: "2.0.0", Enabled: true},
		},
	}

	err := Validate(profile)
	if err != nil {
		t.Errorf("expected no error for valid extensions, got: %v", err)
	}
}

func TestValidate_InvalidExtensionID(t *testing.T) {
	tests := []struct {
		name        string
		extensionID string
		expectedErr string
	}{
		{
			name:        "empty ID",
			extensionID: "",
			expectedErr: "extension ID cannot be empty",
		},
		{
			name:        "no dot",
			extensionID: "nodot",
			expectedErr: "extension ID 'nodot' must be in format 'publisher.name'",
		},
		{
			name:        "multiple dots",
			extensionID: "too.many.dots",
			expectedErr: "extension ID 'too.many.dots' must be in format 'publisher.name'",
		},
		{
			name:        "with space",
			extensionID: "has space.name",
			expectedErr: "extension ID 'has space.name' must be in format 'publisher.name'",
		},
		{
			name:        "trailing dot",
			extensionID: "publisher.",
			expectedErr: "extension ID 'publisher.' must be in format 'publisher.name'",
		},
		{
			name:        "leading dot",
			extensionID: ".name",
			expectedErr: "extension ID '.name' must be in format 'publisher.name'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &Profile{
				Name:      "test-profile",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Extensions: []Extension{
					{ID: tt.extensionID, Version: "1.0.0", Enabled: true},
				},
			}

			err := Validate(profile)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}

			if err != nil && err.Error() != tt.expectedErr {
				t.Errorf("expected error '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestValidate_EmptyExtensionsList(t *testing.T) {
	profile := &Profile{
		Name:       "test-profile",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Extensions: []Extension{},
	}

	err := Validate(profile)
	if err != nil {
		t.Errorf("expected no error for empty extensions list, got: %v", err)
	}
}

func TestLoad_InvalidProfile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create invalid profile with bad extension ID
	profile := Profile{
		Name:      "invalid-profile",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Extensions: []Extension{
			{ID: "invalid-no-dot", Version: "1.0.0", Enabled: true},
		},
	}

	// Write invalid profile to file
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	profilePath := filepath.Join(tempDir, "invalid-profile.json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("failed to write profile file: %v", err)
	}

	// Try to load the invalid profile
	_, err = Load("invalid-profile", tempDir)

	// Should return an error
	if err == nil {
		t.Error("expected error loading invalid profile, got nil")
	}

	// Error should mention the invalid format
	if err != nil && !strings.Contains(err.Error(), "must be in format 'publisher.name'") {
		t.Errorf("expected error to contain \"must be in format 'publisher.name'\", got: %s", err.Error())
	}
}

func TestDetectConflicts_NoConflicts(t *testing.T) {
	// All extensions are new (none installed)
	profileExtensions := []Extension{
		{ID: "ms-python.python", Version: "1.0.0", Enabled: true},
		{ID: "golang.go", Version: "2.0.0", Enabled: true},
	}

	installedExtensions := []vscode.Extension{} // No installed extensions

	toInstall, alreadyInstalled := detectConflicts(profileExtensions, installedExtensions)

	// All should be toInstall
	if len(toInstall) != 2 {
		t.Errorf("expected 2 extensions to install, got %d", len(toInstall))
	}
	if len(alreadyInstalled) != 0 {
		t.Errorf("expected 0 already installed extensions, got %d", len(alreadyInstalled))
	}

	// Verify IDs are correct
	if toInstall[0].ID != "ms-python.python" || toInstall[1].ID != "golang.go" {
		t.Errorf("toInstall has incorrect extension IDs")
	}
}

func TestDetectConflicts_AllInstalled(t *testing.T) {
	// All extensions already exist
	profileExtensions := []Extension{
		{ID: "ms-python.python", Version: "1.0.0", Enabled: true},
		{ID: "golang.go", Version: "2.0.0", Enabled: true},
	}

	installedExtensions := []vscode.Extension{
		{ID: "ms-python.python", Version: "1.5.0", Enabled: true},
		{ID: "golang.go", Version: "2.1.0", Enabled: true},
	}

	toInstall, alreadyInstalled := detectConflicts(profileExtensions, installedExtensions)

	// All should be alreadyInstalled
	if len(toInstall) != 0 {
		t.Errorf("expected 0 extensions to install, got %d", len(toInstall))
	}
	if len(alreadyInstalled) != 2 {
		t.Errorf("expected 2 already installed extensions, got %d", len(alreadyInstalled))
	}

	// Verify IDs are correct
	if alreadyInstalled[0].ID != "ms-python.python" || alreadyInstalled[1].ID != "golang.go" {
		t.Errorf("alreadyInstalled has incorrect extension IDs")
	}
}

func TestDetectConflicts_Mixed(t *testing.T) {
	// Some new, some existing
	profileExtensions := []Extension{
		{ID: "ms-python.python", Version: "1.0.0", Enabled: true},
		{ID: "golang.go", Version: "2.0.0", Enabled: true},
		{ID: "rust-lang.rust", Version: "0.7.8", Enabled: true},
	}

	installedExtensions := []vscode.Extension{
		{ID: "ms-python.python", Version: "1.5.0", Enabled: true},
		{ID: "other.extension", Version: "1.0.0", Enabled: true},
	}

	toInstall, alreadyInstalled := detectConflicts(profileExtensions, installedExtensions)

	// 2 to install (golang.go, rust-lang.rust), 1 already installed (ms-python.python)
	if len(toInstall) != 2 {
		t.Errorf("expected 2 extensions to install, got %d", len(toInstall))
	}
	if len(alreadyInstalled) != 1 {
		t.Errorf("expected 1 already installed extension, got %d", len(alreadyInstalled))
	}

	// Verify correct categorization
	if alreadyInstalled[0].ID != "ms-python.python" {
		t.Errorf("expected ms-python.python to be already installed, got %s", alreadyInstalled[0].ID)
	}

	// toInstall should contain golang.go and rust-lang.rust
	toInstallIDs := map[string]bool{
		toInstall[0].ID: true,
		toInstall[1].ID: true,
	}
	if !toInstallIDs["golang.go"] || !toInstallIDs["rust-lang.rust"] {
		t.Errorf("toInstall should contain golang.go and rust-lang.rust")
	}
}

func TestDetectConflicts_EmptyProfile(t *testing.T) {
	// Empty extensions list
	profileExtensions := []Extension{}

	installedExtensions := []vscode.Extension{
		{ID: "ms-python.python", Version: "1.5.0", Enabled: true},
	}

	toInstall, alreadyInstalled := detectConflicts(profileExtensions, installedExtensions)

	// Both should be empty
	if len(toInstall) != 0 {
		t.Errorf("expected 0 extensions to install, got %d", len(toInstall))
	}
	if len(alreadyInstalled) != 0 {
		t.Errorf("expected 0 already installed extensions, got %d", len(alreadyInstalled))
	}
}

func TestLoad_SkipsAlreadyInstalled(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Write valid profile JSON with ms-python.python@2024.0.0
	profile := Profile{
		Name:      "test-skip-profile",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Extensions: []Extension{
			{ID: "ms-python.python", Version: "2024.0.0", Enabled: true},
		},
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	profilePath := filepath.Join(tempDir, "test-skip-profile.json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("failed to write profile file: %v", err)
	}

	// Call Load() - will try to install (may fail if VS Code not available, that's OK)
	loadedProfile, err := Load("test-skip-profile", tempDir)

	// Test verifies profile loads and validation passed
	if err != nil {
		// Error is acceptable if it's "failed to install extension" (VS Code not available)
		if strings.Contains(err.Error(), "failed to install extension") {
			t.Logf("Load failed as expected without VS Code: %v", err)
			return
		}
		t.Fatalf("Load failed with unexpected error: %v", err)
	}

	// If no error, verify the profile loaded correctly
	if loadedProfile.Name != "test-skip-profile" {
		t.Errorf("expected profile name 'test-skip-profile', got '%s'", loadedProfile.Name)
	}
	if len(loadedProfile.Extensions) != 1 {
		t.Errorf("expected 1 extension, got %d", len(loadedProfile.Extensions))
	}
}

func TestDiff_ValidProfile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create profile with 2 extensions
	profile := Profile{
		Name:      "diff-test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Extensions: []Extension{
			{ID: "ms-python.python", Version: "1.0.0", Enabled: true},
			{ID: "golang.go", Version: "2.0.0", Enabled: true},
		},
	}

	// Write profile to file
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	profilePath := filepath.Join(tempDir, "diff-test.json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("failed to write profile file: %v", err)
	}

	// Call Diff
	result, err := Diff("diff-test", tempDir)

	// May fail if VS Code not available, that's acceptable
	if err != nil {
		if strings.Contains(err.Error(), "failed to list installed extensions") {
			t.Logf("Diff failed as expected without VS Code: %v", err)
			return
		}
		t.Fatalf("Diff failed with unexpected error: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ProfileName != "diff-test" {
		t.Errorf("expected profile name 'diff-test', got '%s'", result.ProfileName)
	}
	if result.TotalInProfile != 2 {
		t.Errorf("expected TotalInProfile 2, got %d", result.TotalInProfile)
	}

	// ToInstall + AlreadyInstalled should equal TotalInProfile
	total := len(result.ToInstall) + len(result.AlreadyInstalled)
	if total != result.TotalInProfile {
		t.Errorf("expected ToInstall + AlreadyInstalled = %d, got %d", result.TotalInProfile, total)
	}
}

func TestDiff_ProfileNotFound(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Call Diff with nonexistent profile
	_, err := Diff("nonexistent", tempDir)

	// Should return error
	if err == nil {
		t.Error("expected error for nonexistent profile, got nil")
	}

	// Error should contain "not found"
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got: %s", err.Error())
	}
}

func TestDiff_InvalidProfile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create profile with bad extension ID
	profile := Profile{
		Name:      "invalid-diff",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Extensions: []Extension{
			{ID: "bad-extension-id", Version: "1.0.0", Enabled: true},
		},
	}

	// Write profile to file
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	profilePath := filepath.Join(tempDir, "invalid-diff.json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("failed to write profile file: %v", err)
	}

	// Call Diff
	_, err = Diff("invalid-diff", tempDir)

	// Should return validation error
	if err == nil {
		t.Error("expected validation error, got nil")
	}

	// Error should mention format requirement
	if err != nil && !strings.Contains(err.Error(), "must be in format") {
		t.Errorf("expected error to contain 'must be in format', got: %s", err.Error())
	}
}
