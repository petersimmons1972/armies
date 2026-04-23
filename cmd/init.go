package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/petersimmons1972/armies/internal/config"
	"github.com/petersimmons1972/armies/internal/gitops"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var subdirs = []string{"profiles", "accountability", "service-records", "teams"}

// newInitCommand builds the init cobra.Command for internal registration.
func newInitCommand() *cobra.Command {
	return NewInitCommand()
}

// NewInitCommand returns the init cobra.Command, exported so tests can call it directly.
func NewInitCommand() *cobra.Command {
	var armiesDir string
	var remoteURL string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the ~/.armies/ directory structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve armiesDir: flag overrides default
			dir := armiesDir
			if dir == "" {
				d, err := config.ArmiesDir()
				if err != nil {
					return err
				}
				dir = d
			}

			// Create subdirectories
			for _, sub := range subdirs {
				full := filepath.Join(dir, sub)
				if err := os.MkdirAll(full, 0o700); err != nil {
					return fmt.Errorf("cannot create %s: %w", full, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "✓ %s\n", full)
			}

			// Write config.yaml if missing
			cfgPath := filepath.Join(dir, "config.yaml")
			if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
				cfgData := config.Config{
					RemoteURL:    remoteURL,
					DefaultModel: "sonnet",
					ProfilesDir:  filepath.Join(dir, "profiles"),
				}
				data, err := yaml.Marshal(cfgData)
				if err != nil {
					return err
				}
				if err := os.WriteFile(cfgPath, data, 0o600); err != nil {
					return fmt.Errorf("cannot write config.yaml: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "✓ %s\n", cfgPath)

				if remoteURL != "" {
					if err := runInitGit(cmd, dir, remoteURL); err != nil {
						return err
					}
				}
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "config.yaml already exists — skipping\n")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nDone. %s is ready.\n", dir)
			return nil
		},
	}

	cmd.Flags().StringVar(&armiesDir, "armies-dir", "", "Override default ~/.armies location")
	cmd.Flags().StringVar(&remoteURL, "remote-url", "", "GitHub remote URL for sync (optional)")
	return cmd
}

// runInitGit runs git init and sets (or updates) remote origin in dir.
func runInitGit(cmd *cobra.Command, dir, remoteURL string) error {
	_, stderr, err := gitops.RunGit(".", "init", dir)
	if err != nil {
		return fmt.Errorf("git init failed: %s", stderr)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "✓ git init %s\n", dir)

	_, _, checkErr := gitops.RunGit(dir, "remote", "get-url", "origin")
	if checkErr == nil {
		if _, stderr, err := gitops.RunGit(dir, "remote", "set-url", "origin", remoteURL); err != nil {
			return fmt.Errorf("git remote set-url failed: %s", stderr)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ remote origin updated to %s\n", remoteURL)
	} else {
		if _, stderr, err := gitops.RunGit(dir, "remote", "add", "origin", remoteURL); err != nil {
			return fmt.Errorf("git remote add failed: %s", stderr)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ remote origin set to %s\n", remoteURL)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(newInitCommand())
}
