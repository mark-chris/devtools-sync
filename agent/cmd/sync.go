package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark-chris/devtools-sync/agent/internal/api"
	"github.com/mark-chris/devtools-sync/agent/internal/config"
	"github.com/mark-chris/devtools-sync/agent/internal/profile"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize profiles with server",
	Long:  "Push local profiles to server or pull profiles from server",
}

var syncPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push profiles to server",
	Long:  "Upload all local profiles to the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w\n\nRun 'devtools-sync init' to create the configuration file", err)
		}

		// Create API client
		client := api.NewClient(cfg.Server.URL)

		// List local profiles
		profiles, err := profile.List(cfg.Profiles.Directory)
		if err != nil {
			return fmt.Errorf("failed to list local profiles: %w", err)
		}

		if len(profiles) == 0 {
			cmd.Println("No profiles to push")
			return nil
		}

		// Push each profile
		pushed := make([]string, 0)
		failed := make([]string, 0)

		for _, prof := range profiles {
			// Convert to API profile
			apiProfile := convertToAPIProfile(&prof)

			// Upload to server
			if err := client.UploadProfile(apiProfile); err != nil {
				cmd.Printf("Failed to push profile '%s': %v\n", prof.Name, err)
				failed = append(failed, prof.Name)
				continue
			}

			pushed = append(pushed, prof.Name)
		}

		// Report results
		if len(pushed) > 0 {
			cmd.Printf("Pushed %d profile(s): %v\n", len(pushed), pushed)
		}
		if len(failed) > 0 {
			cmd.Printf("Failed to push %d profile(s): %v\n", len(failed), failed)
		}

		return nil
	},
}

var syncPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull profiles from server",
	Long:  "Download profiles from the server to local storage",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w\n\nRun 'devtools-sync init' to create the configuration file", err)
		}

		// Create API client
		client := api.NewClient(cfg.Server.URL)

		// List server profiles
		serverProfiles, err := client.ListProfiles()
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
				return fmt.Errorf("failed to connect to server at %s: %w\n\nMake sure:\n  1. The server is running\n  2. The server URL is correct (check with 'devtools-sync config show')\n  3. You can reach the server from your network", cfg.Server.URL, err)
			}
			return fmt.Errorf("failed to list server profiles: %w\n\nCheck your server connection with:\n  curl %s/health", err, cfg.Server.URL)
		}

		if len(serverProfiles) == 0 {
			cmd.Println("No profiles on server")
			return nil
		}

		pulled := make([]string, 0)
		skipped := make([]string, 0)
		failed := make([]string, 0)

		// Download each profile
		for _, name := range serverProfiles {
			// Download from server
			apiProfile, err := client.DownloadProfile(name)
			if err != nil {
				cmd.Printf("Failed to download profile '%s': %v\n", name, err)
				failed = append(failed, name)
				continue
			}

			// Check if local profile exists and is newer
			localProfilePath := filepath.Join(cfg.Profiles.Directory, name+".json")
			if _, err := os.Stat(localProfilePath); err == nil {
				// Local profile exists, check if it's newer
				localProfile, err := profile.Get(name, cfg.Profiles.Directory)
				if err == nil && localProfile.UpdatedAt.After(apiProfile.UpdatedAt) {
					cmd.Printf("Skipping '%s' (local version is newer)\n", name)
					skipped = append(skipped, name)
					continue
				}
			}

			// Convert to local profile and save
			localProfile := convertToLocalProfile(apiProfile)

			// Save profile to disk
			if err := saveProfile(localProfile, cfg.Profiles.Directory); err != nil {
				cmd.Printf("Failed to save profile '%s': %v\n", name, err)
				failed = append(failed, name)
				continue
			}

			pulled = append(pulled, name)
		}

		// Report results
		if len(pulled) > 0 {
			cmd.Printf("Pulled %d profile(s): %v\n", len(pulled), pulled)
		}
		if len(skipped) > 0 {
			cmd.Printf("Skipped %d profile(s) (local is newer): %v\n", len(skipped), skipped)
		}
		if len(failed) > 0 {
			cmd.Printf("Failed %d profile(s): %v\n", len(failed), failed)
		}

		return nil
	},
}

func init() {
	syncCmd.AddCommand(syncPushCmd)
	syncCmd.AddCommand(syncPullCmd)
	rootCmd.AddCommand(syncCmd)
}

// Helper functions to convert between local and API profile types

func convertToAPIProfile(p *profile.Profile) *api.Profile {
	extensions := make([]api.Extension, len(p.Extensions))
	for i, ext := range p.Extensions {
		extensions[i] = api.Extension{
			ID:      ext.ID,
			Version: ext.Version,
			Enabled: ext.Enabled,
		}
	}

	return &api.Profile{
		Name:       p.Name,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		Extensions: extensions,
	}
}

func convertToLocalProfile(p *api.Profile) *profile.Profile {
	extensions := make([]profile.Extension, len(p.Extensions))
	for i, ext := range p.Extensions {
		extensions[i] = profile.Extension{
			ID:      ext.ID,
			Version: ext.Version,
			Enabled: ext.Enabled,
		}
	}

	return &profile.Profile{
		Name:       p.Name,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		Extensions: extensions,
	}
}

func saveProfile(p *profile.Profile, profilesDir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Write profile to JSON
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	profilePath := filepath.Join(profilesDir, p.Name+".json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}

	return nil
}
