package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark-chris/devtools-sync/agent/internal/vscode"
)

// Extension represents a VS Code extension in a profile
type Extension struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

// Profile represents a saved VS Code configuration
type Profile struct {
	Name       string      `json:"name"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Extensions []Extension `json:"extensions"`
}

// Validate checks if the profile has valid data
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

	// Check extension IDs
	for _, ext := range profile.Extensions {
		// Check if extension ID is empty
		if ext.ID == "" {
			return fmt.Errorf("extension ID cannot be empty")
		}

		// Check if extension ID contains spaces
		if strings.Contains(ext.ID, " ") {
			return fmt.Errorf("extension ID '%s' must be in format 'publisher.name'", ext.ID)
		}

		// Split by dot to check format
		parts := strings.Split(ext.ID, ".")

		// Must have exactly 2 parts (publisher.name)
		if len(parts) != 2 {
			return fmt.Errorf("extension ID '%s' must be in format 'publisher.name'", ext.ID)
		}

		// Both parts must be non-empty
		if parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("extension ID '%s' must be in format 'publisher.name'", ext.ID)
		}
	}

	return nil
}

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

// Save captures current VS Code extensions to a profile
func Save(name string, profilesDir string) (*Profile, error) {
	if name == "" {
		return nil, fmt.Errorf("profile name cannot be empty")
	}

	// Ensure profiles directory exists
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Get current VS Code extensions
	vscodeExts, err := vscode.ListExtensions()
	if err != nil {
		return nil, fmt.Errorf("failed to list VS Code extensions: %w", err)
	}

	// Convert to profile extensions
	extensions := make([]Extension, len(vscodeExts))
	for i, ext := range vscodeExts {
		extensions[i] = Extension{
			ID:      ext.ID,
			Version: ext.Version,
			Enabled: ext.Enabled,
		}
	}

	// Create or update profile
	profilePath := filepath.Join(profilesDir, name+".json")
	now := time.Now()

	profile := &Profile{
		Name:       name,
		UpdatedAt:  now,
		Extensions: extensions,
	}

	// If profile exists, preserve created_at timestamp
	if existingData, err := os.ReadFile(profilePath); err == nil {
		var existing Profile
		if err := json.Unmarshal(existingData, &existing); err == nil {
			profile.CreatedAt = existing.CreatedAt
		}
	}

	// Set created_at if this is a new profile
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}

	// Write profile to file
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write profile file: %w", err)
	}

	return profile, nil
}

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

	// Validate profile before attempting installation
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

// List returns all local profiles
func List(profilesDir string) ([]Profile, error) {
	// Ensure profiles directory exists
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Read all .json files from profiles directory
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	profiles := make([]Profile, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .json files
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		// Read and parse profile
		profilePath := filepath.Join(profilesDir, entry.Name())
		data, err := os.ReadFile(profilePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var profile Profile
		if err := json.Unmarshal(data, &profile); err != nil {
			continue // Skip files that can't be parsed
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// Get retrieves a specific profile by name
func Get(name string, profilesDir string) (*Profile, error) {
	if name == "" {
		return nil, fmt.Errorf("profile name cannot be empty")
	}

	profilePath := filepath.Join(profilesDir, name+".json")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile file: %w", err)
	}

	return &profile, nil
}
