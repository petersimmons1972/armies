package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/petersimmons1972/armies/internal/config"
	"github.com/petersimmons1972/armies/internal/eligibility"
	"github.com/petersimmons1972/armies/internal/profiles"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewRosterCommand returns the roster cobra.Command.
func NewRosterCommand() *cobra.Command {
	var profilesDir string
	var malusLedger string

	c := &cobra.Command{
		Use:   "roster",
		Short: "Scan profiles and display the agent roster table",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Resolve malus ledger
			ledgerPath := malusLedger
			if ledgerPath == "" {
				lp, err := config.MalusLedgerPath()
				if err != nil {
					return err
				}
				ledgerPath = lp
			}

			paths, err := profiles.StreamProfiles(pdir)
			if err != nil {
				return fmt.Errorf("cannot read profiles directory: %w", err)
			}

			if len(paths) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No profiles found in %s\n", pdir)
				fmt.Fprintf(cmd.OutOrStdout(), "Run armies init to create the directory structure.\n")
				return nil
			}

			// Check if ledger exists
			ledgerExists := false
			if _, statErr := os.Stat(ledgerPath); statErr == nil {
				ledgerExists = true
			}

			t := table.NewWriter()
			t.SetOutputMirror(cmd.OutOrStdout())
			t.AppendHeader(table.Row{"name", "display_name", "primary_role", "model", "effort", "xp", "rank", "eligibility"})

			for _, path := range paths {
				fm, _, parseErr := profiles.ParseProfile(path, nil)
				if parseErr != nil {
					// Skip malformed profiles silently — log warning to stderr
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: skipping %s: %v\n", path, parseErr)
					continue
				}

				name := stringOrDash(fm, "name")
				// Fall back to name when display_name is absent.
				displayName := stringOrDash(fm, "display_name")
				if displayName == "—" {
					displayName = name
				}
				primaryRole := resolvePrimaryRole(fm)
				model := stringOrDash(fm, "model")
				effort := stringOrDash(fm, "effort_level")
				if effort == "—" {
					effort = "medium"
				}
				xp := stringOrDash(fm, "xp")
				rank := stringOrDash(fm, "rank")

				elig := "eligible"
				if ledgerExists {
					entries := loadMalusEntries(ledgerPath, name)
					effectiveMalus := eligibility.EffectiveMalus(entries)
					status := eligibility.EligibilityStatus(name, effectiveMalus)
					elig = status.Overall
				}

				t.AppendRow(table.Row{name, displayName, primaryRole, model, effort, xp, rank, elig})
			}

			t.Render()
			return nil
		},
	}

	c.Flags().StringVar(&profilesDir, "profiles-dir", "", "Profiles directory")
	c.Flags().StringVar(&malusLedger, "malus-ledger", "", "Path to malus-ledger.yaml")
	return c
}

// stringOrDash returns the string representation of fm[key], or "—" if missing/nil/empty.
func stringOrDash(fm map[string]any, key string) string {
	v, ok := fm[key]
	if !ok || v == nil {
		return "—"
	}
	s := fmt.Sprintf("%v", v)
	if s == "" {
		return "—"
	}
	return s
}

// resolvePrimaryRole resolves primary_role using the Python precedence:
// 1. fm["primary_role"] if present and non-empty
// 2. fm["roles"]["primary"] if roles is a map and primary is non-empty
// 3. "—"
func resolvePrimaryRole(fm map[string]any) string {
	if v, ok := fm["primary_role"]; ok && v != nil {
		if s := fmt.Sprintf("%v", v); s != "" {
			return s
		}
	}
	if rolesRaw, ok := fm["roles"]; ok {
		if rolesMap, ok := rolesRaw.(map[string]any); ok {
			if p, ok := rolesMap["primary"]; ok && p != nil {
				if s := fmt.Sprintf("%v", p); s != "" {
					return s
				}
			}
		}
	}
	return "—"
}

// loadMalusEntries reads malus-ledger.yaml and returns entries attributed to agentName.
// The YAML ledger supports two formats:
//   - flat entry: {agent: name, share: N, raw_malus: N, date: ..., decays: bool}
//   - allocation list: {allocation: [{agent: name, share: N}, ...], raw_malus: N, date: ..., decays: bool}
func loadMalusEntries(ledgerPath, agentName string) []eligibility.MalusEntry {
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		return nil
	}

	var raw []map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil
	}

	target := strings.ToLower(strings.TrimSpace(agentName))
	var entries []eligibility.MalusEntry

	for _, item := range raw {
		rawMalus := toFloat(item["raw_malus"])
		decays := toBool(item["decays"], true)

		date := ""
		if d, ok := item["date"]; ok {
			date = fmt.Sprintf("%v", d)
		}

		// Build a unified list of {agent, share} allocations
		var allocs []map[string]any
		if allocList, ok := item["allocation"].([]any); ok {
			for _, a := range allocList {
				if am, ok := a.(map[string]any); ok {
					allocs = append(allocs, am)
				}
			}
		} else if agentVal, ok := item["agent"]; ok {
			allocs = []map[string]any{
				{"agent": agentVal, "share": item["share"]},
			}
		}

		for _, alloc := range allocs {
			if strings.ToLower(fmt.Sprintf("%v", alloc["agent"])) != target {
				continue
			}
			share := toFloat(alloc["share"])
			if share == 0 {
				share = 100
			}
			entries = append(entries, eligibility.MalusEntry{
				RawMalus: rawMalus,
				Share:    share,
				Date:     date,
				Decays:   decays,
			})
		}
	}
	return entries
}

// toFloat converts YAML numeric types (int, float32, float64) to float64.
// yaml.v3 unmarshals integers as int, not float64.
func toFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return float64(val)
	case float64:
		return val
	case float32:
		return float64(val)
	default:
		return 0
	}
}

// toBool extracts a bool from an any value, returning defaultVal if nil or wrong type.
func toBool(v any, defaultVal bool) bool {
	if v == nil {
		return defaultVal
	}
	b, ok := v.(bool)
	if !ok {
		return defaultVal
	}
	return b
}

func init() {
	rootCmd.AddCommand(NewRosterCommand())
}
