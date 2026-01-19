package vscode

import (
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
	extensions, err := ListExtensions()
	if err != nil {
		t.Errorf("ListExtensions should not error, got: %v", err)
	}

	// Stub implementation returns empty list
	if extensions == nil {
		t.Error("expected extensions slice, got nil")
	}
}

func TestInstallExtension(t *testing.T) {
	tests := []struct {
		name        string
		extensionID string
		wantError   bool
	}{
		{
			name:        "valid extension ID",
			extensionID: "ms-vscode.go",
			wantError:   false,
		},
		{
			name:        "empty extension ID",
			extensionID: "",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InstallExtension(tt.extensionID)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestGetVSCodePaths(t *testing.T) {
	paths := getVSCodePaths()

	// Should return at least one path for common platforms
	if len(paths) == 0 {
		t.Error("expected at least one VS Code path")
	}
}
