package vscode

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"golang.org/x/mod/semver"
)

// Extension represents a VS Code extension
type Extension struct {
	ID          string
	Version     string
	Enabled     bool
	DisplayName string
	Description string
	Publisher   string
}

// packageManifest represents the structure of a package.json file
type packageManifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Publisher   string `json:"publisher"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// DetectInstallation checks if VS Code is installed on the system
func DetectInstallation() (bool, error) {
	paths := getVSCodePaths()

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true, nil
		}
	}

	return false, nil
}

// ListExtensions returns a list of installed VS Code extensions
func ListExtensions() ([]Extension, error) {
	// Execute code --list-extensions --show-versions
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

// InstallExtension installs a VS Code extension by ID
func InstallExtension(extensionID string) error {
	if extensionID == "" {
		return errors.New("extension ID cannot be empty")
	}

	// Execute code --install-extension <id>
	cmd := exec.Command("code", "--install-extension", extensionID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install extension %s: %w (output: %s)", extensionID, err, string(output))
	}

	return nil
}

// getVSCodePaths returns common VS Code installation paths by platform
func getVSCodePaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/Visual Studio Code.app",
			filepath.Join(os.Getenv("HOME"), "Applications/Visual Studio Code.app"),
		}
	case "windows":
		return []string{
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Microsoft VS Code"),
			filepath.Join(os.Getenv("PROGRAMFILES"), "Microsoft VS Code"),
		}
	case "linux":
		return []string{
			"/usr/share/code",
			"/usr/bin/code",
			"/snap/bin/code",
		}
	default:
		return []string{}
	}
}

// getExtensionDirs returns extension directory paths for VS Code and Insiders
func getExtensionDirs() []string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows
	}

	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, ".vscode", "extensions"),
			filepath.Join(home, ".vscode-insiders", "extensions"),
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

// parseManifest parses a package.json file and returns an Extension
func parseManifest(data []byte, dirName string) (Extension, error) {
	var manifest packageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Extension{}, fmt.Errorf("failed to parse package.json: %w", err)
	}

	// Validate required fields
	if manifest.Name == "" {
		return Extension{}, errors.New("missing required field: name")
	}
	if manifest.Version == "" {
		return Extension{}, errors.New("missing required field: version")
	}
	if manifest.Publisher == "" {
		return Extension{}, errors.New("missing required field: publisher")
	}

	// Build extension ID from publisher.name
	extensionID := fmt.Sprintf("%s.%s", manifest.Publisher, manifest.Name)

	return Extension{
		ID:          extensionID,
		Version:     manifest.Version,
		Enabled:     true,
		DisplayName: manifest.DisplayName,
		Description: manifest.Description,
		Publisher:   manifest.Publisher,
	}, nil
}

// compareVersions compares two semantic version strings.
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
// Version strings are normalized to include 'v' prefix for semver package.
func compareVersions(v1, v2 string) int {
	// Ensure versions have 'v' prefix for semver package
	if !strings.HasPrefix(v1, "v") {
		v1 = "v" + v1
	}
	if !strings.HasPrefix(v2, "v") {
		v2 = "v" + v2
	}

	return semver.Compare(v1, v2)
}

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

// mergeExtensions combines multiple sets of extensions, deduplicates by ID,
// and keeps the highest version for each extension.
func mergeExtensions(sets ...[]Extension) []Extension {
	// Use map for deduplication
	extMap := make(map[string]Extension)

	for _, set := range sets {
		for _, ext := range set {
			if existing, found := extMap[ext.ID]; found {
				// Extension already exists, compare versions
				cmp := compareVersions(ext.Version, existing.Version)
				if cmp > 0 {
					// New version is higher
					log.Printf("Deduplicating %s: keeping v%s over v%s", ext.ID, ext.Version, existing.Version)
					extMap[ext.ID] = ext
				} else if cmp < 0 {
					// Existing version is higher
					log.Printf("Deduplicating %s: keeping v%s over v%s", ext.ID, existing.Version, ext.Version)
				}
				// If equal (cmp == 0), keep existing
			} else {
				extMap[ext.ID] = ext
			}
		}
	}

	// Convert map to slice
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
