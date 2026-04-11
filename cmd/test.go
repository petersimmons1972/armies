package cmd

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/petersimmons1972/armies/internal/config"
	"github.com/petersimmons1972/armies/internal/profiles"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewTestCommand returns the test cobra.Command.
// It generates a behavioral fingerprint test document for an agent profile.
func NewTestCommand() *cobra.Command {
	var profilesDir string

	cmd := &cobra.Command{
		Use:   "test <agent>",
		Short: "Generate a behavioral fingerprint test document for an agent profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

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

			// Read frontmatter only (no sections needed yet)
			fm, _, err := profiles.ParseProfile(profilePath, nil)
			if err != nil {
				return fmt.Errorf("cannot parse profile %s: %w", profilePath, err)
			}

			// Extract test_scenarios from frontmatter
			scenarios, err := extractTestScenarios(fm)
			if err != nil || len(scenarios) == 0 {
				filename := filepath.Base(profilePath)
				name := agentName
				if n, ok := fm["name"].(string); ok && n != "" {
					name = n
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "No test scenarios defined in %s\nAdd a test_scenarios block to the frontmatter of %s.\nSee the armies research command output for the schema.\n", filename, name)
				return fmt.Errorf("No test scenarios defined in %s", filename)
			}

			// Determine primaryRole
			primaryRole := "coordinator"
			if pr, ok := fm["primary_role"].(string); ok && pr != "" {
				primaryRole = pr
			} else if rolesMap, ok := fm["roles"].(map[string]interface{}); ok {
				if p, ok := rolesMap["primary"].(string); ok && p != "" {
					primaryRole = p
				}
			}

			// Read profile body for Base Persona and the primary role section
			roleHeading := "Role: " + primaryRole
			sectionsWanted := []string{"Base Persona", roleHeading}
			_, sections, err := profiles.ParseProfile(profilePath, sectionsWanted)
			if err != nil {
				return fmt.Errorf("cannot parse profile body %s: %w", profilePath, err)
			}

			// Determine display name
			displayName := agentName
			if dn, ok := fm["display_name"].(string); ok && dn != "" {
				displayName = dn
			} else if n, ok := fm["name"].(string); ok && n != "" {
				displayName = n
			}

			// Build spawn YAML block — exclude test_scenarios from the output
			spawnFM := make(map[string]interface{})
			for k, v := range fm {
				if k == "test_scenarios" {
					continue
				}
				spawnFM[k] = v
			}
			fmYAML, err := yaml.Marshal(spawnFM)
			if err != nil {
				return fmt.Errorf("cannot marshal frontmatter: %w", err)
			}

			// Count total criteria
			totalCriteria := 0
			for _, sc := range scenarios {
				totalCriteria += len(sc.Fingerprints)
			}

			// Build the document
			doc := buildTestDocument(displayName, primaryRole, roleHeading, string(fmYAML), sections, scenarios, totalCriteria)
			fmt.Fprint(cmd.OutOrStdout(), doc)
			return nil
		},
	}

	cmd.Flags().StringVar(&profilesDir, "profiles-dir", "", "Profiles directory")
	return cmd
}

// testScenario represents a single test scenario from the profile frontmatter.
type testScenario struct {
	ID          string
	Situation   string
	Prompt      string
	Fingerprints []fingerprint
}

// fingerprint is a single scoring criterion within a scenario.
type fingerprint struct {
	Criterion string
	Why       string
}

// extractTestScenarios pulls test_scenarios out of the frontmatter map.
// Returns an error if the key is missing or malformed.
func extractTestScenarios(fm map[string]any) ([]testScenario, error) {
	raw, ok := fm["test_scenarios"]
	if !ok || raw == nil {
		return nil, fmt.Errorf("test_scenarios key absent")
	}

	rawSlice, ok := raw.([]interface{})
	if !ok || len(rawSlice) == 0 {
		return nil, fmt.Errorf("test_scenarios is empty or not a list")
	}

	var result []testScenario
	for _, item := range rawSlice {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		sc := testScenario{
			ID:        stringField(m, "id"),
			Situation: stringField(m, "situation"),
			Prompt:    stringField(m, "prompt"),
		}

		if fps, ok := m["fingerprints"].([]interface{}); ok {
			for _, fpItem := range fps {
				fpMap, ok := fpItem.(map[string]interface{})
				if !ok {
					continue
				}
				sc.Fingerprints = append(sc.Fingerprints, fingerprint{
					Criterion: stringField(fpMap, "criterion"),
					Why:       stringField(fpMap, "why"),
				})
			}
		}

		result = append(result, sc)
	}

	return result, nil
}

// stringField safely extracts a string from a map[string]interface{}.
func stringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// titleCase converts a hyphen-separated id into Title Case.
// e.g. "ambiguous-order" → "Ambiguous Order"
func titleCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

// buildTestDocument assembles the final output document.
func buildTestDocument(
	displayName, primaryRole, roleHeading, fmYAML string,
	sections map[string]string,
	scenarios []testScenario,
	totalCriteria int,
) string {
	tick := "```"
	var sb strings.Builder

	// Header
	sb.WriteString("# Behavioral Fingerprint Test — ")
	sb.WriteString(displayName)
	sb.WriteString("\n\n")
	sb.WriteString("Paste this entire document into a new Claude Code conversation. Read the agent's response to each scenario, then score it against the rubric below each one.\n\n")
	sb.WriteString("---\n\n")

	// Agent Context
	sb.WriteString("## Agent Context\n\n")
	sb.WriteString("The following profile is loaded for this session. You are this person for the duration of this conversation.\n\n")
	sb.WriteString(tick + "\n")
	sb.WriteString("---\n")
	sb.WriteString(fmYAML)
	sb.WriteString("---\n\n")
	if body, ok := sections["Base Persona"]; ok && body != "" {
		sb.WriteString("## Base Persona\n\n")
		sb.WriteString(body)
		sb.WriteString("\n\n")
	}
	if body, ok := sections[roleHeading]; ok && body != "" {
		sb.WriteString("## ")
		sb.WriteString(roleHeading)
		sb.WriteString("\n\n")
		sb.WriteString(body)
		sb.WriteString("\n")
	}
	sb.WriteString(tick + "\n\n")
	sb.WriteString("---\n\n")

	// Scenarios
	sb.WriteString("## Scenarios\n\n")
	sb.WriteString("Read each scenario, respond in character, then use the rubric below to score your own response.\n\n")

	for i, sc := range scenarios {
		criterionIndex := 1
		scenarioNum := i + 1
		scenarioTitle := titleCase(sc.ID)

		fmt.Fprintf(&sb, "### Scenario %d: %s\n\n", scenarioNum, scenarioTitle)
		sb.WriteString(sc.Situation)
		sb.WriteString("\n\n")
		fmt.Fprintf(&sb, "**%s**\n\n", sc.Prompt)
		sb.WriteString("*Respond in character before reading the rubric below.*\n\n")
		sb.WriteString("---\n\n")
		fmt.Fprintf(&sb, "#### Scoring Rubric — Scenario %d\n\n", scenarioNum)

		for _, fp := range sc.Fingerprints {
			fmt.Fprintf(&sb, "**Criterion %d.%d:** %s\n\n", scenarioNum, criterionIndex, fp.Criterion)
			fmt.Fprintf(&sb, "*Why this is specific to %s:* %s\n\n", displayName, fp.Why)
			sb.WriteString(tick + "\n")
			sb.WriteString("[ ] PASS   [ ] FAIL\n\n")
			sb.WriteString("Notes: \n")
			sb.WriteString(tick + "\n\n")
			criterionIndex++
		}

		sb.WriteString("---\n\n")
	}

	// Summary Scorecard
	high := int(math.Round(float64(totalCriteria) * 0.75))
	if high < 1 {
		high = 1
	}
	mid := int(math.Round(float64(totalCriteria) * 0.50))
	if mid < 1 {
		mid = 1
	}

	sb.WriteString("## Summary Scorecard\n\n")
	fmt.Fprintf(&sb, "Total criteria: **%d**\n\n", totalCriteria)
	sb.WriteString("| Score | Interpretation |\n")
	sb.WriteString("|-------|---------------|\n")
	fmt.Fprintf(&sb, "| %d/%d | Strong activation — profile is working |\n", totalCriteria, totalCriteria)
	fmt.Fprintf(&sb, "| %d/%d | Good activation — minor gaps |\n", high, totalCriteria)
	fmt.Fprintf(&sb, "| %d/%d | Partial activation — Base Persona needs more behavioral specifics |\n", mid, totalCriteria)
	fmt.Fprintf(&sb, "| <%d/%d | Weak activation — profile is producing a generic agent |\n", mid, totalCriteria)
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "If score is below 50%%: review the Base Persona for generic sentences (sentences that could describe any %s). Replace them with documented behavioral specifics from %s's record.\n", primaryRole, displayName)

	return sb.String()
}

func init() {
	rootCmd.AddCommand(NewTestCommand())
}
