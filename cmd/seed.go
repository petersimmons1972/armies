package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// NewSeedCommandFS returns the seed cobra.Command, parameterized on an fs.FS
// so tests can inject fstest.MapFS without needing embed.FS.
func NewSeedCommandFS(generalsFS fs.FS) *cobra.Command {
	var profilesDir string
	var force bool

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Install bundled general profiles to profiles directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if profilesDir == "" {
				return fmt.Errorf("--profiles-dir is required")
			}
			if err := os.MkdirAll(profilesDir, 0o700); err != nil {
				return fmt.Errorf("cannot create profiles directory: %w", err)
			}

			entries, err := fs.ReadDir(generalsFS, "examples/generals")
			if err != nil {
				return fmt.Errorf("cannot read embedded generals: %w", err)
			}

			installed := 0
			skipped := 0
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				if !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				dest := filepath.Join(profilesDir, e.Name())
				if _, err := os.Stat(dest); err == nil && !force {
					skipped++
					continue
				}
				data, err := fs.ReadFile(generalsFS, "examples/generals/"+e.Name())
				if err != nil {
					return err
				}
				if err := os.WriteFile(dest, data, 0o600); err != nil {
					return fmt.Errorf("cannot write %s: %w", dest, err)
				}
				installed++
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed %d profile(s), skipped %d.\n", installed, skipped)
			return nil
		},
	}
	cmd.Flags().StringVar(&profilesDir, "profiles-dir", "", "Target profiles directory (required)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing profiles")
	return cmd
}
