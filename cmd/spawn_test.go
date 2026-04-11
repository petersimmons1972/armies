package cmd_test

import (
	"bytes"
	"testing"

	"github.com/petersimmons1972/armies/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpawn_OutputsFrontmatterAndSections(t *testing.T) {
	// Use the fixture from testdata (profiles/valid.md has Base Persona + Role: specialist)
	spawnCmd := cmd.NewSpawnCommand()
	var buf bytes.Buffer
	spawnCmd.SetOut(&buf)
	spawnCmd.SetArgs([]string{
		"valid",
		"--role", "specialist",
		"--profiles-dir", "../internal/profiles/testdata/profiles",
	})
	err := spawnCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "---")                 // frontmatter block
	assert.Contains(t, out, "name: test-agent")    // frontmatter content
	assert.Contains(t, out, "## Base Persona")     // base persona heading
	assert.Contains(t, out, "## Role: specialist") // role heading
}

func TestSpawn_MissingRole_ExitsError(t *testing.T) {
	spawnCmd := cmd.NewSpawnCommand()
	var buf bytes.Buffer
	spawnCmd.SetOut(&buf)
	spawnCmd.SetErr(&buf)
	spawnCmd.SetArgs([]string{
		"valid",
		"--role", "nonexistent-role",
		"--profiles-dir", "../internal/profiles/testdata/profiles",
	})
	err := spawnCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, buf.String()+err.Error(), "Role: nonexistent-role")
}

func TestSpawn_ProfileNotFound_ExitsError(t *testing.T) {
	spawnCmd := cmd.NewSpawnCommand()
	var buf bytes.Buffer
	spawnCmd.SetOut(&buf)
	spawnCmd.SetErr(&buf)
	spawnCmd.SetArgs([]string{
		"ghost-agent",
		"--role", "specialist",
		"--profiles-dir", "../internal/profiles/testdata/profiles",
	})
	err := spawnCmd.Execute()
	require.Error(t, err)
}

func TestSpawn_TraversalRejected(t *testing.T) {
	spawnCmd := cmd.NewSpawnCommand()
	var buf bytes.Buffer
	spawnCmd.SetOut(&buf)
	spawnCmd.SetErr(&buf)
	spawnCmd.SetArgs([]string{
		"../../etc/passwd",
		"--role", "specialist",
		"--profiles-dir", "../internal/profiles/testdata/profiles",
	})
	err := spawnCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside profiles directory")
}
