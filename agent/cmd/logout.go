package main

import (
	"fmt"

	"github.com/mark-chris/devtools-sync/agent/internal/api"
	"github.com/mark-chris/devtools-sync/agent/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication credentials",
	Long:  "Logout from devtools-sync by removing stored authentication token and credentials.",
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create authenticated client
	kc := keychainFactory()
	client := api.NewAuthenticatedClient(cfg.Server.URL, kc)

	// Logout
	if err := client.Logout(); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Logged out successfully. Authentication credentials removed.")

	return nil
}
