package main

import (
	"fmt"
	"strings"

	"github.com/mark-chris/devtools-sync/agent/internal/config"
	"github.com/mark-chris/devtools-sync/agent/internal/profile"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage extension profiles",
	Long:  "Save, load, and list VS Code extension profiles",
}

var profileSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save current extensions to a profile",
	Long:  "Capture the current VS Code extensions and save them to a named profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Load config to get profiles directory
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Save profile
		prof, err := profile.Save(name, cfg.Profiles.Directory)
		if err != nil {
			if strings.Contains(err.Error(), "VS Code") {
				return fmt.Errorf("failed to save profile: %w\n\nMake sure:\n  1. VS Code is installed\n  2. The 'code' command is available in your PATH\n  3. You can run 'code --version' successfully", err)
			}
			return fmt.Errorf("failed to save profile '%s': %w", name, err)
		}

		cmd.Printf("Saved %d extensions to profile '%s'\n", len(prof.Extensions), name)
		return nil
	},
}

var profileLoadCmd = &cobra.Command{
	Use:               "load <name>",
	Short:             "Load extensions from a profile",
	Long:              "Install VS Code extensions from a saved profile",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: profileNameCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Load config to get profiles directory
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Load profile
		prof, err := profile.Load(name, cfg.Profiles.Directory)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// List available profiles for better UX
				profiles, _ := profile.List(cfg.Profiles.Directory)
				if len(profiles) > 0 {
					names := make([]string, len(profiles))
					for i, p := range profiles {
						names[i] = p.Name
					}
					return fmt.Errorf("profile '%s' not found\n\nAvailable profiles: %s\n\nUse 'devtools-sync profile list' to see all profiles", name, strings.Join(names, ", "))
				}
				return fmt.Errorf("profile '%s' not found\n\nNo profiles available. Create one with:\n  devtools-sync profile save <name>", name)
			}
			if strings.Contains(err.Error(), "VS Code") {
				return fmt.Errorf("failed to load profile: %w\n\nMake sure:\n  1. VS Code is installed\n  2. The 'code' command is available in your PATH", err)
			}
			return fmt.Errorf("failed to load profile '%s': %w", name, err)
		}

		cmd.Printf("Installing %d extensions from profile '%s'...\n", len(prof.Extensions), name)
		cmd.Printf("Done!\n")
		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Long:  "Display all saved extension profiles with their metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config to get profiles directory
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// List profiles
		profiles, err := profile.List(cfg.Profiles.Directory)
		if err != nil {
			return err
		}

		if len(profiles) == 0 {
			cmd.Printf("No profiles found.\n")
			return nil
		}

		// Display profiles in a table format
		cmd.Printf("%-20s %-15s %-25s\n", "NAME", "EXTENSIONS", "LAST UPDATED")
		cmd.Printf("%s\n", strings.Repeat("-", 60))

		for _, prof := range profiles {
			cmd.Printf("%-20s %-15d %-25s\n",
				prof.Name,
				len(prof.Extensions),
				prof.UpdatedAt.Format("2006-01-02 15:04:05"),
			)
		}

		return nil
	},
}

var profileDiffCmd = &cobra.Command{
	Use:               "diff <name>",
	Short:             "Compare a profile with currently installed extensions",
	Long:              "Show which extensions would be installed and which are already installed if loading this profile",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: profileNameCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Load config to get profiles directory
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Compare profile with installed extensions
		result, err := profile.Diff(name, cfg.Profiles.Directory)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// List available profiles for better UX
				profiles, _ := profile.List(cfg.Profiles.Directory)
				if len(profiles) > 0 {
					names := make([]string, len(profiles))
					for i, p := range profiles {
						names[i] = p.Name
					}
					return fmt.Errorf("profile '%s' not found\n\nAvailable profiles: %s\n\nUse 'devtools-sync profile list' to see all profiles", name, strings.Join(names, ", "))
				}
				return fmt.Errorf("profile '%s' not found\n\nNo profiles available. Create one with:\n  devtools-sync profile save <name>", name)
			}
			if strings.Contains(err.Error(), "VS Code") {
				return fmt.Errorf("failed to diff profile: %w\n\nMake sure:\n  1. VS Code is installed\n  2. The 'code' command is available in your PATH", err)
			}
			return fmt.Errorf("failed to diff profile '%s': %w", name, err)
		}

		// Display results in a formatted manner
		cmd.Printf("Profile: %s\n", result.ProfileName)
		cmd.Printf("Total extensions in profile: %d\n\n", result.TotalInProfile)

		if len(result.ToInstall) > 0 {
			cmd.Printf("To Install (%d):\n", len(result.ToInstall))
			for _, ext := range result.ToInstall {
				cmd.Printf("  + %s (%s)\n", ext.ID, ext.Version)
			}
			cmd.Printf("\n")
		}

		if len(result.AlreadyInstalled) > 0 {
			cmd.Printf("Already Installed (%d):\n", len(result.AlreadyInstalled))
			for _, ext := range result.AlreadyInstalled {
				cmd.Printf("  = %s (%s)\n", ext.ID, ext.Version)
			}
			cmd.Printf("\n")
		}

		if len(result.ToInstall) == 0 && len(result.AlreadyInstalled) == result.TotalInProfile {
			cmd.Printf("All extensions from this profile are already installed.\n")
		} else if len(result.ToInstall) > 0 {
			cmd.Printf("Run 'devtools-sync profile load %s' to install missing extensions.\n", name)
		}

		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileSaveCmd)
	profileCmd.AddCommand(profileLoadCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileDiffCmd)
	rootCmd.AddCommand(profileCmd)
}

// profileNameCompletion provides tab completion for profile names
func profileNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Load config to get profiles directory
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// List available profiles
	profiles, err := profile.List(cfg.Profiles.Directory)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Extract profile names
	names := make([]string, len(profiles))
	for i, prof := range profiles {
		names[i] = prof.Name
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
