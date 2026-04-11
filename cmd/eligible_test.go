package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/petersimmons1972/armies/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEligible_CleanAgentNoLedger(t *testing.T) {
	eligCmd := cmd.NewEligibleCommand()
	var buf bytes.Buffer
	eligCmd.SetOut(&buf)
	eligCmd.SetArgs([]string{"clean-agent", "--malus-ledger", "/nonexistent/ledger.yaml"})
	err := eligCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "clean-agent")
	assert.Contains(t, out, "0.0")       // zero malus
	assert.Contains(t, out, "Clean")     // tier name
	assert.Contains(t, out, "coordinator")
	assert.Contains(t, out, "CLEAR")
	assert.Contains(t, out, "Ledger not found")
}

func TestEligible_WithLedger(t *testing.T) {
	// Create a temp ledger with 200 malus for test-agent
	dir := t.TempDir()
	ledger := filepath.Join(dir, "malus-ledger.yaml")
	content := `
- agent: test-agent
  raw_malus: 200
  decays: false
  share: 100
`
	require.NoError(t, os.WriteFile(ledger, []byte(content), 0644))

	eligCmd := cmd.NewEligibleCommand()
	var buf bytes.Buffer
	eligCmd.SetOut(&buf)
	eligCmd.SetArgs([]string{"test-agent", "--malus-ledger", ledger})
	err := eligCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "200.0")     // effective malus
	assert.Contains(t, out, "Probation") // tier
	assert.Contains(t, out, "BLOCKED")   // coordinator is blocked in Probation
	assert.Contains(t, out, "REVIEW")    // specialist is REVIEW in Probation
}

func TestEligible_RoleNamesHaveSpaces(t *testing.T) {
	eligCmd := cmd.NewEligibleCommand()
	var buf bytes.Buffer
	eligCmd.SetOut(&buf)
	eligCmd.SetArgs([]string{"any-agent", "--malus-ledger", "/nonexistent/ledger.yaml"})
	err := eligCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	// "emergency_reserve" should appear as "emergency reserve" in the table
	assert.Contains(t, out, "emergency reserve")
}
