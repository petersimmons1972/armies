package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/petersimmons1972/armies/internal/config"
	"github.com/petersimmons1972/armies/internal/eligibility"
	"github.com/spf13/cobra"
)

// NewEligibleCommand returns the eligible cobra.Command.
func NewEligibleCommand() *cobra.Command {
	var malusLedger string

	cmd := &cobra.Command{
		Use:   "eligible <agent>",
		Short: "Compute and display spawn eligibility for an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

			ledgerPath := malusLedger
			if ledgerPath == "" {
				lp, err := config.MalusLedgerPath()
				if err != nil {
					return err
				}
				ledgerPath = lp
			}

			// Load entries; if ledger missing, entries is nil (zero malus)
			ledgerExists := false
			if _, err := os.Stat(ledgerPath); err == nil {
				ledgerExists = true
			}

			var entries []eligibility.MalusEntry
			if ledgerExists {
				entries = loadMalusEntries(ledgerPath, agentName)
			}

			effectiveMalus := eligibility.EffectiveMalus(entries)
			status := eligibility.EligibilityStatus(agentName, effectiveMalus)

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "\nAgent: %s\n", agentName)
			fmt.Fprintf(out, "Effective malus: %.1f\n", effectiveMalus)
			fmt.Fprintf(out, "Tier: %s\n", status.Tier)

			if !ledgerExists {
				fmt.Fprintf(out, "\nNote: Ledger not found at %s. Showing gates for zero malus.\n", ledgerPath)
			}

			t := table.NewWriter()
			t.SetOutputMirror(out)
			t.AppendHeader(table.Row{"Role", "Status"})
			for _, role := range eligibility.KnownRoles() {
				t.AppendRow(table.Row{
					strings.ReplaceAll(role, "_", " "),
					status.Roles[role],
				})
			}
			t.Render()

			return nil
		},
	}

	cmd.Flags().StringVar(&malusLedger, "malus-ledger", "", "Path to malus-ledger.yaml")
	return cmd
}

func init() {
	rootCmd.AddCommand(NewEligibleCommand())
}
