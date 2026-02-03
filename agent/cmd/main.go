package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "devtools-sync",
	Short: "DevTools Sync Agent - Synchronize your development tools",
	Long: `DevTools Sync Agent helps you manage and synchronize your development tool extensions
and configurations across multiple machines.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
