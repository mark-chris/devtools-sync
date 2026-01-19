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
			name:      "valid config",
			config:    &Config{ServerURL: "http://localhost:8080"},
			wantError: false,
		},
		{
			name:      "empty server URL",
			config:    &Config{ServerURL: ""},
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
