package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/petersimmons1972/armies/internal/config"
	"github.com/petersimmons1972/armies/internal/profiles"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewRecordCommand returns the record cobra.Command.
func NewRecordCommand() *cobra.Command {
	var xp int
	var outcome string
	var profilesDir string
	var armiesDir string

	cmd := &cobra.Command{
		Use:   "record <agent> <note>",
		Short: "Write a service record entry for an agent and update XP",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]
			note := args[1]

			// Validate outcome
			validOutcomes := map[string]bool{"success": true, "partial": true, "failure": true}
			if !validOutcomes[outcome] {
				return fmt.Errorf("invalid outcome %q: must be one of success, partial, failure", outcome)
			}

			// Resolve profiles dir
			pdir := profilesDir
			if pdir == "" {
				cfg, err := config.Load()
				if err != nil {
					return err
				}
				pdir, err = cfg.ProfilesDirLexicallyValidated()
				if err != nil {
					return err
				}
			}

			// Resolve armies dir
			adir := armiesDir
			if adir == "" {
				var err error
				adir, err = config.ArmiesDir()
				if err != nil {
					return err
				}
			}

			// Resolve agent profile path with traversal guard
			profilePath, err := profiles.ResolveAgentPath(pdir, agentName)
			if err != nil {
				return err
			}

			// Check existence; fall back to case-insensitive search
			if _, statErr := os.Stat(profilePath); statErr != nil {
				found := caseInsensitiveSearch(pdir, agentName)
				if found == "" {
					return fmt.Errorf("profile not found: %s (searched in %s)", agentName, pdir)
				}
				profilePath = found
			}

			// Derive the canonical filename base from the resolved path, not the raw arg.
			// This ensures case normalization and avoids path traversal in the filename.
			agentBase := strings.TrimSuffix(filepath.Base(profilePath), ".md")

			// Read current XP from frontmatter
			fm, _, err := profiles.ParseProfile(profilePath, nil)
			if err != nil {
				return fmt.Errorf("cannot parse profile %s: %w", profilePath, err)
			}
			currentXP := int(toFloat(fm["xp"]))
			newXP := currentXP + xp

			// Update profile XP
			if err := profiles.UpdateFrontmatterField(profilePath, "xp", newXP); err != nil {
				return fmt.Errorf("cannot update XP in profile: %w", err)
			}

			// Create service-records dir if needed
			srDir := filepath.Join(adir, "service-records")
			if err := os.MkdirAll(srDir, 0755); err != nil {
				return fmt.Errorf("cannot create service-records directory: %w", err)
			}

			// Build the record entry
			today := time.Now().Format("2006-01-02")
			newEntry := map[string]any{
				"date":     today,
				"task":     note,
				"outcome":  outcome,
				"xp_earned": xp,
				"xp_total": newXP,
			}

			// Read existing entries (empty list if file does not exist)
			srFile := filepath.Join(srDir, agentBase+".yaml")
			var entries []map[string]any
			if data, readErr := os.ReadFile(srFile); readErr == nil {
				if unmErr := yaml.Unmarshal(data, &entries); unmErr != nil {
					return fmt.Errorf("cannot parse existing service record %s: %w", srFile, unmErr)
				}
			}
			if entries == nil {
				entries = []map[string]any{}
			}
			entries = append(entries, newEntry)

			// Write back
			out, err := yaml.Marshal(entries)
			if err != nil {
				return fmt.Errorf("cannot marshal service record: %w", err)
			}
			if err := os.WriteFile(srFile, out, 0644); err != nil {
				return fmt.Errorf("cannot write service record: %w", err)
			}

			// Build relative path for display
			relSR := filepath.Join("service-records", agentBase+".yaml")

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Service record written: %s\n", relSR)
			fmt.Fprintf(cmd.OutOrStdout(), "✓ XP updated: %d → %d\n", currentXP, newXP)
			fmt.Fprintf(cmd.OutOrStdout(), "Commit hint: git -C %s commit -am \"record: %s — %s\"\n",
				adir, agentBase, note)

			return nil
		},
	}

	// --xp 0 (the default) is accepted: zero-XP records document outcomes without XP changes.
	cmd.Flags().IntVar(&xp, "xp", 0, "XP to award")
	cmd.Flags().StringVar(&outcome, "outcome", "success", "Outcome: success, partial, or failure")
	cmd.Flags().StringVar(&profilesDir, "profiles-dir", "", "Profiles directory")
	cmd.Flags().StringVar(&armiesDir, "armies-dir", "", "Armies directory (default: ~/.armies)")

	return cmd
}

func init() {
	rootCmd.AddCommand(NewRecordCommand())
}
