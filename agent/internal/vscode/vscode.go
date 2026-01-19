package vscode

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// Extension represents a VS Code extension
type Extension struct {
	ID      string
	Version string
	Enabled bool
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

// ListExtensions returns a list of installed VS Code extensions (stub)
func ListExtensions() ([]Extension, error) {
	// Stub implementation - will be implemented in future iterations
	return []Extension{}, nil
}

// InstallExtension installs a VS Code extension by ID (stub)
func InstallExtension(extensionID string) error {
	// Stub implementation - will be implemented in future iterations
	if extensionID == "" {
		return errors.New("extension ID cannot be empty")
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
