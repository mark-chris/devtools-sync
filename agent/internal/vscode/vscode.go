package vscode

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
