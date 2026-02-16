package main

import (
	"fmt"
	"syscall"

	"github.com/mark-chris/devtools-sync/agent/internal/api"
	"github.com/mark-chris/devtools-sync/agent/internal/config"
	"github.com/mark-chris/devtools-sync/agent/internal/keychain"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// keychainFactory allows injecting a mock keychain in tests
var keychainFactory func() keychain.Keychain = func() keychain.Keychain {
	return keychain.NewSystemKeychain()
}

var (
	loginEmail    string
	loginPassword string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the devtools-sync server",
	Long:  "Login to the devtools-sync server and store authentication token securely.",
	RunE:  runLogin,
}

func init() {
	loginCmd.Flags().StringVar(&loginEmail, "email", "", "Email address")
	loginCmd.Flags().StringVar(&loginPassword, "password", "", "Password (will prompt if not provided)")
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Prompt for email if not provided
	if loginEmail == "" {
		fmt.Fprint(cmd.OutOrStdout(), "Email: ")
		_, err := fmt.Fscanln(cmd.InOrStdin(), &loginEmail)
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
	}

	// Prompt for password if not provided
	if loginPassword == "" {
		fmt.Fprint(cmd.OutOrStdout(), "Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(cmd.OutOrStdout()) // newline after password
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		loginPassword = string(passwordBytes)
	}

	// Create authenticated client
	kc := keychainFactory()
	client := api.NewAuthenticatedClient(cfg.Server.URL, kc)

	// Login
	if err := client.Login(loginEmail, loginPassword); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Login successful! Token stored securely.")

	// Reset flags for reuse in tests
	loginEmail = ""
	loginPassword = ""

	return nil
}
