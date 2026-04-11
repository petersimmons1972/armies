package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/petersimmons1972/armies/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// setupRecordFixture creates a temp profiles dir with a copy of valid.md
// renamed to agent.md, and a temp armies dir. Returns (profilesDir, armiesDir).
func setupRecordFixture(t *testing.T, agentName string) (string, string) {
	t.Helper()

	profilesDir := t.TempDir()
	armiesDir := t.TempDir()

	// Copy valid.md as <agentName>.md
	src, err := os.ReadFile("../internal/profiles/testdata/profiles/valid.md")
	require.NoError(t, err, "read test fixture")

	dst := filepath.Join(profilesDir, agentName+".md")
	require.NoError(t, os.WriteFile(dst, src, 0644))

	return profilesDir, armiesDir
}

func TestRecord_WritesServiceRecord(t *testing.T) {
	profilesDir, armiesDir := setupRecordFixture(t, "test-agent")

	recCmd := cmd.NewRecordCommand()
	var buf bytes.Buffer
	recCmd.SetOut(&buf)
	recCmd.SetErr(&buf)
	recCmd.SetArgs([]string{
		"test-agent",
		"completed the training mission",
		"--xp", "25",
		"--outcome", "success",
		"--profiles-dir", profilesDir,
		"--armies-dir", armiesDir,
	})

	err := recCmd.Execute()
	require.NoError(t, err)

	// Verify service records file was created
	srFile := filepath.Join(armiesDir, "service-records", "test-agent.yaml")
	require.FileExists(t, srFile)

	// Verify YAML content
	data, err := os.ReadFile(srFile)
	require.NoError(t, err)

	var entries []map[string]any
	require.NoError(t, yaml.Unmarshal(data, &entries))
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "completed the training mission", entry["task"])
	assert.Equal(t, "success", entry["outcome"])
	assert.Equal(t, 25, entry["xp_earned"])
	assert.Equal(t, 125, entry["xp_total"])

	// Verify stdout output
	out := buf.String()
	assert.Contains(t, out, "Service record written")
	assert.Contains(t, out, "XP updated")
	assert.Contains(t, out, "100")
	assert.Contains(t, out, "125")

	// Verify profile XP was updated
	profilePath := filepath.Join(profilesDir, "test-agent.md")
	profData, err := os.ReadFile(profilePath)
	require.NoError(t, err)
	assert.Contains(t, string(profData), "xp: 125")
}

func TestRecord_AgentNotFound_Error(t *testing.T) {
	profilesDir := t.TempDir()
	armiesDir := t.TempDir()

	recCmd := cmd.NewRecordCommand()
	var buf bytes.Buffer
	recCmd.SetOut(&buf)
	recCmd.SetErr(&buf)
	recCmd.SetArgs([]string{
		"ghost-agent",
		"did something",
		"--profiles-dir", profilesDir,
		"--armies-dir", armiesDir,
	})

	err := recCmd.Execute()
	require.Error(t, err)
}

func TestRecord_InvalidOutcome_Error(t *testing.T) {
	profilesDir, armiesDir := setupRecordFixture(t, "test-agent")

	recCmd := cmd.NewRecordCommand()
	var buf bytes.Buffer
	recCmd.SetOut(&buf)
	recCmd.SetErr(&buf)
	recCmd.SetArgs([]string{
		"test-agent",
		"some note",
		"--outcome", "invalid",
		"--profiles-dir", profilesDir,
		"--armies-dir", armiesDir,
	})

	err := recCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestRecord_ZeroXP_AcceptedSilently(t *testing.T) {
	profilesDir, armiesDir := setupRecordFixture(t, "test-agent")

	recCmd := cmd.NewRecordCommand()
	var buf bytes.Buffer
	recCmd.SetOut(&buf)
	recCmd.SetErr(&buf)
	recCmd.SetArgs([]string{
		"test-agent",
		"documented outcome without XP change",
		// --xp flag intentionally omitted; defaults to 0
		"--outcome", "success",
		"--profiles-dir", profilesDir,
		"--armies-dir", armiesDir,
	})

	err := recCmd.Execute()
	require.NoError(t, err)

	// Verify service record was created with xp_earned: 0
	srFile := filepath.Join(armiesDir, "service-records", "test-agent.yaml")
	require.FileExists(t, srFile)

	data, readErr := os.ReadFile(srFile)
	require.NoError(t, readErr)

	var entries []map[string]any
	require.NoError(t, yaml.Unmarshal(data, &entries))
	require.Len(t, entries, 1)
	assert.Equal(t, 0, entries[0]["xp_earned"])

	// XP in profile must remain unchanged (100 → 100)
	profilePath := filepath.Join(profilesDir, "test-agent.md")
	profData, err := os.ReadFile(profilePath)
	require.NoError(t, err)
	assert.Contains(t, string(profData), "xp: 100")
}

func TestRecord_AppendsTwoEntries(t *testing.T) {
	profilesDir, armiesDir := setupRecordFixture(t, "test-agent")

	// First record
	recCmd1 := cmd.NewRecordCommand()
	recCmd1.SetOut(&bytes.Buffer{})
	recCmd1.SetErr(&bytes.Buffer{})
	recCmd1.SetArgs([]string{
		"test-agent",
		"first mission",
		"--xp", "10",
		"--outcome", "success",
		"--profiles-dir", profilesDir,
		"--armies-dir", armiesDir,
	})
	require.NoError(t, recCmd1.Execute())

	// Second record
	recCmd2 := cmd.NewRecordCommand()
	recCmd2.SetOut(&bytes.Buffer{})
	recCmd2.SetErr(&bytes.Buffer{})
	recCmd2.SetArgs([]string{
		"test-agent",
		"second mission",
		"--xp", "15",
		"--outcome", "partial",
		"--profiles-dir", profilesDir,
		"--armies-dir", armiesDir,
	})
	require.NoError(t, recCmd2.Execute())

	// Verify YAML list has 2 entries
	srFile := filepath.Join(armiesDir, "service-records", "test-agent.yaml")
	data, err := os.ReadFile(srFile)
	require.NoError(t, err)

	var entries []map[string]any
	require.NoError(t, yaml.Unmarshal(data, &entries))
	require.Len(t, entries, 2)

	assert.Equal(t, "first mission", entries[0]["task"])
	assert.Equal(t, "second mission", entries[1]["task"])
	assert.Equal(t, 10, entries[0]["xp_earned"])
	assert.Equal(t, 110, entries[0]["xp_total"])
	assert.Equal(t, 15, entries[1]["xp_earned"])
	assert.Equal(t, 125, entries[1]["xp_total"])
}
