package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the agent configuration
type Config struct {
	Server struct {
		URL string `yaml:"url"`
	} `yaml:"server"`
	Profiles struct {
		Directory string `yaml:"directory"`
	} `yaml:"profiles"`
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".devtools-sync", "config.yaml")
}

// GetConfigDir returns the path to the config directory
func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".devtools-sync")
}

// Load reads configuration from YAML file and applies environment variable overrides
func Load() (*Config, error) {
	cfg := &Config{}

	// Set defaults
	cfg.Server.URL = "http://localhost:8080"
	cfg.Profiles.Directory = filepath.Join(GetConfigDir(), "profiles")
	cfg.Logging.Level = "info"

	// Try to read config file
	configPath := GetConfigPath()
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}
	// If file doesn't exist, we just use defaults (not an error)

	// Apply environment variable overrides
	if serverURL := os.Getenv("DEVTOOLS_SYNC_SERVER_URL"); serverURL != "" {
		cfg.Server.URL = serverURL
	}
	if logLevel := os.Getenv("DEVTOOLS_SYNC_LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.URL == "" {
		return errors.New("server URL cannot be empty")
	}

	parsedURL, err := url.Parse(c.Server.URL)
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
	parsedURL, err := url.Parse(c.Server.URL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" {
		return false
	}

	host := parsedURL.Hostname()
	return host != "localhost" && host != "127.0.0.1" && host != "::1"
}

// Save writes the configuration to the YAML file
func (c *Config) Save() error {
	configPath := GetConfigPath()

	// Ensure config directory exists
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
