package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
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

	parsedURL, err := url.Parse(c.ServerURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("server URL must use http or https scheme, got: %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return errors.New("server URL must include a host")
	}

	return nil
}

// IsInsecure returns true if the server URL uses HTTP on a non-localhost host
func (c *Config) IsInsecure() bool {
	parsedURL, err := url.Parse(c.ServerURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" {
		return false
	}

	host := strings.Split(parsedURL.Host, ":")[0]
	return host != "localhost" && host != "127.0.0.1" && host != "::1"
}
