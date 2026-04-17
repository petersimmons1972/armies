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

func TestInit_CreatesDirectoryStructure(t *testing.T) {
	dir := t.TempDir()
	initCmd := cmd.NewInitCommand()
	var buf bytes.Buffer
	initCmd.SetOut(&buf)
	initCmd.SetArgs([]string{"--armies-dir", dir})
	err := initCmd.Execute()
	require.NoError(t, err)

	for _, sub := range []string{"profiles", "accountability", "service-records", "teams"} {
		assert.DirExists(t, filepath.Join(dir, sub))
	}
	assert.FileExists(t, filepath.Join(dir, "config.yaml"))
}

func TestInit_WritesConfigYAML(t *testing.T) {
	dir := t.TempDir()
	initCmd := cmd.NewInitCommand()
	initCmd.SetArgs([]string{"--armies-dir", dir, "--remote-url", "https://github.com/user/repo.git"})
	var buf bytes.Buffer
	initCmd.SetOut(&buf)
	err := initCmd.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "remote_url")
	assert.Contains(t, string(data), "sonnet")
}

func TestInit_SkipsConfigIfExists(t *testing.T) {
	dir := t.TempDir()
	// Pre-create config
	cfgPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(cfgPath, []byte("remote_url: \"\"\ndefault_model: opus\nprofiles_dir: /tmp\n"), 0644)

	initCmd := cmd.NewInitCommand()
	var buf bytes.Buffer
	initCmd.SetOut(&buf)
	initCmd.SetArgs([]string{"--armies-dir", dir})
	err := initCmd.Execute()
	require.NoError(t, err)

	// Config must not be overwritten
	data, _ := os.ReadFile(cfgPath)
	assert.Contains(t, string(data), "opus", "existing config must be preserved")
	assert.Contains(t, buf.String(), "already exists")
}

func TestInit_IdempotentSubdirs(t *testing.T) {
	dir := t.TempDir()

	// Run twice — must not fail on second run
	for i := 0; i < 2; i++ {
		initCmd := cmd.NewInitCommand()
		initCmd.SetArgs([]string{"--armies-dir", dir})
		var buf bytes.Buffer
		initCmd.SetOut(&buf)
		err := initCmd.Execute()
		require.NoError(t, err, "run %d failed", i+1)
	}
}
