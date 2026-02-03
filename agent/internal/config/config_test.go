package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantURL   string
		wantError bool
	}{
		{
			name:      "default URL when env not set",
			envValue:  "",
			wantURL:   "http://localhost:8080",
			wantError: false,
		},
		{
			name:      "custom URL from environment",
			envValue:  "http://custom:9000",
			wantURL:   "http://custom:9000",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				if err := os.Setenv("DEVTOOLS_SYNC_SERVER_URL", tt.envValue); err != nil {
					t.Fatalf("failed to set environment variable: %v", err)
				}
				defer func() {
					_ = os.Unsetenv("DEVTOOLS_SYNC_SERVER_URL")
				}()
			} else {
				_ = os.Unsetenv("DEVTOOLS_SYNC_SERVER_URL")
			}

			cfg, err := Load()

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if !tt.wantError && cfg.Server.URL != tt.wantURL {
				t.Errorf("expected ServerURL %s, got %s", tt.wantURL, cfg.Server.URL)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		wantError bool
	}{
		{
			name:      "valid http URL",
			serverURL: "http://localhost:8080",
			wantError: false,
		},
		{
			name:      "valid https URL",
			serverURL: "https://api.example.com",
			wantError: false,
		},
		{
			name:      "valid https URL with port",
			serverURL: "https://api.example.com:8443",
			wantError: false,
		},
		{
			name:      "empty server URL",
			serverURL: "",
			wantError: true,
		},
		{
			name:      "invalid scheme ftp",
			serverURL: "ftp://example.com",
			wantError: true,
		},
		{
			name:      "missing scheme",
			serverURL: "example.com:8080",
			wantError: true,
		},
		{
			name:      "missing host",
			serverURL: "http://",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			cfg.Server.URL = tt.serverURL
			err := cfg.Validate()

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestIsInsecure(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		insecure  bool
	}{
		{
			name:      "http localhost is secure",
			serverURL: "http://localhost:8080",
			insecure:  false,
		},
		{
			name:      "http 127.0.0.1 is secure",
			serverURL: "http://127.0.0.1:8080",
			insecure:  false,
		},
		{
			name:      "http ::1 is secure",
			serverURL: "http://[::1]:8080",
			insecure:  false,
		},
		{
			name:      "https remote is secure",
			serverURL: "https://api.example.com",
			insecure:  false,
		},
		{
			name:      "http remote is insecure",
			serverURL: "http://api.example.com",
			insecure:  true,
		},
		{
			name:      "http remote with port is insecure",
			serverURL: "http://api.example.com:8080",
			insecure:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			cfg.Server.URL = tt.serverURL
			result := cfg.IsInsecure()

			if result != tt.insecure {
				t.Errorf("expected IsInsecure()=%v, got %v", tt.insecure, result)
			}
		})
	}
}
