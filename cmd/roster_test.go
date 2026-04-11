package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/petersimmons1972/armies/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeProfile(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
}

const rosterProfile = `---
name: test-agent
display_name: Test Agent
xp: 150
rank: captain
primary_role: specialist
role: specialist
---
## Base Persona
body
`

const rosterProfileMinimal = `---
name: minimal-agent
xp: 10
role: validator
---
## Base Persona
body
`

func TestRoster_ShowsProfileTable(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "test-agent.md", rosterProfile)

	rosterCmd := cmd.NewRosterCommand()
	var buf bytes.Buffer
	rosterCmd.SetOut(&buf)
	rosterCmd.SetArgs([]string{"--profiles-dir", dir})
	err := rosterCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "test-agent")
	assert.Contains(t, out, "150")
	assert.Contains(t, out, "captain")
	assert.Contains(t, out, "specialist")
}

func TestRoster_NoProfilesDir_PrintsMessage(t *testing.T) {
	rosterCmd := cmd.NewRosterCommand()
	var buf bytes.Buffer
	rosterCmd.SetOut(&buf)
	rosterCmd.SetArgs([]string{"--profiles-dir", "/nonexistent/dir/12345"})
	err := rosterCmd.Execute()
	require.NoError(t, err) // not an error exit — just prints message
	out := buf.String()
	assert.Contains(t, out, "No profiles found in")
}

func TestRoster_MissingFieldsDefaultToDash(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "minimal-agent.md", rosterProfileMinimal)

	rosterCmd := cmd.NewRosterCommand()
	var buf bytes.Buffer
	rosterCmd.SetOut(&buf)
	rosterCmd.SetArgs([]string{"--profiles-dir", dir})
	err := rosterCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "minimal-agent")
	assert.Contains(t, out, "—", "missing fields must show em-dash")
}

const rosterProfileNoDisplayName = `---
name: nodisplay-agent
xp: 50
role: specialist
---
## Base Persona
body
`

func TestRoster_DisplayNameFallsBackToName(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "nodisplay-agent.md", rosterProfileNoDisplayName)

	rosterCmd := cmd.NewRosterCommand()
	var buf bytes.Buffer
	rosterCmd.SetOut(&buf)
	rosterCmd.SetArgs([]string{"--profiles-dir", dir})
	err := rosterCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	// name should appear twice in output — once in name column, once in display_name column
	assert.Equal(t, 2, strings.Count(out, "nodisplay-agent"), "name must appear as display_name fallback")
}

func TestRoster_EligibilityDefaultsToEligibleWhenNoLedger(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "test-agent.md", rosterProfile)

	rosterCmd := cmd.NewRosterCommand()
	var buf bytes.Buffer
	rosterCmd.SetOut(&buf)
	rosterCmd.SetArgs([]string{"--profiles-dir", dir, "--malus-ledger", "/nonexistent/ledger.yaml"})
	err := rosterCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "eligible")
}
