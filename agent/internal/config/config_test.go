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
				os.Setenv("DEVTOOLS_SYNC_SERVER_URL", tt.envValue)
				defer os.Unsetenv("DEVTOOLS_SYNC_SERVER_URL")
			} else {
				os.Unsetenv("DEVTOOLS_SYNC_SERVER_URL")
			}

			cfg, err := Load()

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if !tt.wantError && cfg.ServerURL != tt.wantURL {
				t.Errorf("expected ServerURL %s, got %s", tt.wantURL, cfg.ServerURL)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name:      "valid http URL",
			config:    &Config{ServerURL: "http://localhost:8080"},
			wantError: false,
		},
		{
			name:      "valid https URL",
			config:    &Config{ServerURL: "https://api.example.com"},
			wantError: false,
		},
		{
			name:      "valid https URL with port",
			config:    &Config{ServerURL: "https://api.example.com:8443"},
			wantError: false,
		},
		{
			name:      "empty server URL",
			config:    &Config{ServerURL: ""},
			wantError: true,
		},
		{
			name:      "invalid scheme ftp",
			config:    &Config{ServerURL: "ftp://example.com"},
			wantError: true,
		},
		{
			name:      "missing scheme",
			config:    &Config{ServerURL: "example.com:8080"},
			wantError: true,
		},
		{
			name:      "missing host",
			config:    &Config{ServerURL: "http://"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

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
		name     string
		config   *Config
		insecure bool
	}{
		{
			name:     "http localhost is secure",
			config:   &Config{ServerURL: "http://localhost:8080"},
			insecure: false,
		},
		{
			name:     "http 127.0.0.1 is secure",
			config:   &Config{ServerURL: "http://127.0.0.1:8080"},
			insecure: false,
		},
		{
			name:     "http ::1 is secure",
			config:   &Config{ServerURL: "http://[::1]:8080"},
			insecure: false,
		},
		{
			name:     "https remote is secure",
			config:   &Config{ServerURL: "https://api.example.com"},
			insecure: false,
		},
		{
			name:     "http remote is insecure",
			config:   &Config{ServerURL: "http://api.example.com"},
			insecure: true,
		},
		{
			name:     "http remote with port is insecure",
			config:   &Config{ServerURL: "http://api.example.com:8080"},
			insecure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsInsecure()

			if result != tt.insecure {
				t.Errorf("expected IsInsecure()=%v, got %v", tt.insecure, result)
			}
		})
	}
}
