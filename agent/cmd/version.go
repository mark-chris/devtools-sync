package main

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version number of DevTools Sync Agent",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("devtools-sync version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
