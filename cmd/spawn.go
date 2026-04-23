package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/petersimmons1972/armies/internal/config"
	"github.com/petersimmons1972/armies/internal/profiles"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var roleHeadingRe = regexp.MustCompile(`(?i)^## (Role: .+)$`)

// NewSpawnCommand returns the spawn cobra.Command.
func NewSpawnCommand() *cobra.Command {
	var profilesDir string

	cmd := &cobra.Command{
		Use:   "spawn <agent>",
		Short: "Output frontmatter + Base Persona + one role block for an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

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

			// Resolve path — traversal guard
			profilePath, err := profiles.ResolveAgentPath(pdir, agentName)
			if err != nil {
				return err
			}

			// Check existence; fall back to case-insensitive search.
			// The fallback result MUST be re-validated with EnsureContained —
			// ResolveAgentPath's guard ran only against the first candidate.
			if _, statErr := os.Stat(profilePath); statErr != nil {
				found := caseInsensitiveSearch(pdir, agentName)
				if found == "" {
					return fmt.Errorf("profile not found: %s (searched in %s)", agentName, pdir)
				}
				resolved, err := profiles.EnsureContained(pdir, found)
				if err != nil {
					return fmt.Errorf("case-insensitive match %s failed containment check: %w", found, err)
				}
				profilePath = resolved
			}

			// Determine role from --role flag
			roleName, _ := cmd.Flags().GetString("role")
			roleHeading := "Role: " + roleName
			sectionsWanted := []string{"Base Persona", roleHeading}

			fm, sections, err := profiles.ParseProfile(profilePath, sectionsWanted)
			if err != nil {
				return fmt.Errorf("cannot parse profile %s: %w", profilePath, err)
			}

			// Check role exists
			if _, ok := sections[roleHeading]; !ok {
				available := listRoleHeadings(profilePath)
				msg := fmt.Sprintf("role block '## %s' not found in %s", roleHeading, profilePath)
				if len(available) > 0 {
					msg += "\nAvailable role blocks:\n  " + strings.Join(available, "\n  ")
				}
				return fmt.Errorf("%s", msg)
			}

			// Marshal frontmatter
			fmYAML, err := yaml.Marshal(fm)
			if err != nil {
				return fmt.Errorf("cannot marshal frontmatter: %w", err)
			}

			// Build output
			var sb strings.Builder
			sb.WriteString("---\n")
			sb.WriteString(string(fmYAML))
			sb.WriteString("---\n\n")

			if body, ok := sections["Base Persona"]; ok && body != "" {
				sb.WriteString("## Base Persona\n\n")
				sb.WriteString(body)
				sb.WriteString("\n\n")
			}

			sb.WriteString("## " + roleHeading + "\n\n")
			sb.WriteString(sections[roleHeading])
			sb.WriteString("\n\n")

			fmt.Fprint(cmd.OutOrStdout(), sb.String())
			return nil
		},
	}

	cmd.Flags().StringVar(&profilesDir, "profiles-dir", "", "Profiles directory")
	cmd.Flags().String("role", "", "Role block to extract (e.g. 'implementer')")
	if err := cmd.MarkFlagRequired("role"); err != nil {
		panic(fmt.Sprintf("spawn: MarkFlagRequired: %v", err))
	}
	return cmd
}

// caseInsensitiveSearch finds the first .md file in dir whose stem matches name
// case-insensitively. Returns empty string if not found.
func caseInsensitiveSearch(dir, name string) string {
	target := strings.ToLower(name)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if !strings.HasSuffix(n, ".md") {
			continue
		}
		stem := strings.TrimSuffix(n, ".md")
		if strings.ToLower(stem) == target {
			// Safe: os.ReadDir only yields direct children (no subdirectory traversal),
			// and the loop skips IsDir entries, so the path cannot escape dir.
			return filepath.Join(dir, n)
		}
	}
	return ""
}

// listRoleHeadings scans a profile file for all "## Role: ..." headings.
func listRoleHeadings(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var result []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if m := roleHeadingRe.FindStringSubmatch(line); m != nil {
			result = append(result, m[1])
		}
	}
	return result
}

func init() {
	rootCmd.AddCommand(NewSpawnCommand())
}
