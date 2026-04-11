package cmd

import (
	"io/fs"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "armies",
	Short: "Multi-agent coordination engine for Claude Code",
	Long:  "armies — Multi-agent coordination engine for Claude Code",
}

// Execute runs the root command and returns any error.
func Execute() error {
	return rootCmd.Execute()
}

// RegisterSeedCommand wires the seed subcommand into the root using the
// provided FS. Must be called from main() before Execute().
func RegisterSeedCommand(generalsFS fs.FS) {
	rootCmd.AddCommand(NewSeedCommandFS(generalsFS))
}
