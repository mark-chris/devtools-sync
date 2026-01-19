package config

import (
	"errors"
	"os"
)

// Config holds the agent configuration
type Config struct {
	ServerURL string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	serverURL := os.Getenv("DEVTOOLS_SYNC_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	cfg := &Config{
		ServerURL: serverURL,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ServerURL == "" {
		return errors.New("server URL cannot be empty")
	}
	return nil
}
