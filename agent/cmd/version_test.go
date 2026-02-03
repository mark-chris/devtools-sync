package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCommand(t *testing.T) {
	// Create a new root command for testing
	cmd := &cobra.Command{Use: "devtools-sync"}
	cmd.AddCommand(versionCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"version"})

	// Execute command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	// Verify output contains version
	got := output.String()
	want := "devtools-sync version 0.1.0"
	if got != want+"\n" {
		t.Errorf("version command output:\ngot:  %q\nwant: %q", got, want)
	}
}
