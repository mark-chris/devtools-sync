package main

import (
	"fmt"
	"strings"

	"github.com/mark-chris/devtools-sync/agent/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "View and update DevTools Sync Agent configuration settings",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  "Display the current effective configuration including environment variable overrides",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cmd.Printf("Server:\n")
		cmd.Printf("  URL: %s\n", cfg.Server.URL)
		cmd.Printf("\n")
		cmd.Printf("Profiles:\n")
		cmd.Printf("  Directory: %s\n", cfg.Profiles.Directory)
		cmd.Printf("\n")
		cmd.Printf("Logging:\n")
		cmd.Printf("  Level: %s\n", cfg.Logging.Level)

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:               "set <key> <value>",
	Short:             "Update configuration value",
	Long:              "Update a configuration value in the config file. Example: devtools-sync config set server.url https://api.example.com",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: configKeyCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// Load current config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Update the appropriate field
		parts := strings.Split(key, ".")
		if len(parts) != 2 {
			return fmt.Errorf("invalid key format. Expected format: section.field (e.g., server.url)")
		}

		section := parts[0]
		field := parts[1]

		switch section {
		case "server":
			switch field {
			case "url":
				cfg.Server.URL = value
			default:
				return fmt.Errorf("unknown server field: %s", field)
			}
		case "profiles":
			switch field {
			case "directory":
				cfg.Profiles.Directory = value
			default:
				return fmt.Errorf("unknown profiles field: %s", field)
			}
		case "logging":
			switch field {
			case "level":
				cfg.Logging.Level = value
			default:
				return fmt.Errorf("unknown logging field: %s", field)
			}
		default:
			return fmt.Errorf("unknown config section: %s", section)
		}

		// Validate the updated config
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		// Save the config
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		cmd.Printf("Updated %s to: %s\n", key, value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}

// configKeyCompletion provides tab completion for config keys
func configKeyCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// If we already have the key argument, don't provide more completions
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Provide list of valid config keys
	validKeys := []string{
		"server.url\tServer URL for syncing profiles",
		"profiles.directory\tDirectory for storing local profiles",
		"logging.level\tLogging level (info, debug, error)",
	}

	return validKeys, cobra.ShellCompDirectiveNoFileComp
}
