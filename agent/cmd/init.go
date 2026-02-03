package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark-chris/devtools-sync/agent/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  "Create the configuration file and required directories for DevTools Sync Agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := config.GetConfigDir()
		configPath := config.GetConfigPath()

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("configuration already exists at %s\n\nTo reconfigure, either:\n  1. Edit the file directly, or\n  2. Delete it and run 'devtools-sync init' again, or\n  3. Use 'devtools-sync config set <key> <value>' to update specific values", configPath)
		}

		// Create config directory
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Create profiles directory
		profilesDir := filepath.Join(configDir, "profiles")
		if err := os.MkdirAll(profilesDir, 0755); err != nil {
			return fmt.Errorf("failed to create profiles directory: %w", err)
		}

		// Create default config
		cfg := &config.Config{}
		cfg.Server.URL = "http://localhost:8080"
		cfg.Profiles.Directory = profilesDir
		cfg.Logging.Level = "info"

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		cmd.Printf("Configuration initialized at %s\n", configDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
