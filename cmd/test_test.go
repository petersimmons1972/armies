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

func TestTest_GeneratesDocument(t *testing.T) {
	testCmd := cmd.NewTestCommand()
	var stdout bytes.Buffer
	testCmd.SetOut(&stdout)
	testCmd.SetArgs([]string{
		"test-agent",
		"--profiles-dir", "../testdata/test",
	})

	err := testCmd.Execute()
	require.NoError(t, err)

	out := stdout.String()
	assert.Contains(t, out, "Behavioral Fingerprint Test")
	assert.Contains(t, out, "Test Agent")
	assert.Contains(t, out, "Scenario 1:")
	assert.Contains(t, out, "Ambiguous Order")
	assert.Contains(t, out, "Criterion 1.1")
	assert.Contains(t, out, "PASS")
	assert.Contains(t, out, "Summary Scorecard")
	assert.Contains(t, out, "Total criteria: **6**")
}

func TestTest_NoScenarios_Error(t *testing.T) {
	// Create a profile without test_scenarios
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "no-scenarios.md")
	content := `---
name: no-scenarios
display_name: "No Scenarios"
xp: 0
role: coordinator
---
## Base Persona
This agent has no scenarios.

## Role: coordinator
Coordinator role body.
`
	err := os.WriteFile(profilePath, []byte(content), 0644)
	require.NoError(t, err)

	testCmd := cmd.NewTestCommand()
	var stderr bytes.Buffer
	testCmd.SetErr(&stderr)
	testCmd.SetArgs([]string{
		"no-scenarios",
		"--profiles-dir", tmpDir,
	})

	err = testCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error()+stderr.String(), "No test scenarios")
}

func TestTest_AgentNotFound_Error(t *testing.T) {
	testCmd := cmd.NewTestCommand()
	var stderr bytes.Buffer
	testCmd.SetErr(&stderr)
	testCmd.SetArgs([]string{
		"nonexistent-agent",
		"--profiles-dir", "../testdata/test",
	})

	err := testCmd.Execute()
	require.Error(t, err)
}
