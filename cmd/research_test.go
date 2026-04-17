package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/petersimmons1972/armies/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResearch_WritesDraftFile(t *testing.T) {
	tmpDir := t.TempDir()

	researchCmd := cmd.NewResearchCommand()
	var stdout bytes.Buffer
	researchCmd.SetOut(&stdout)
	researchCmd.SetArgs([]string{
		"coordinator",
		"--profiles-dir", tmpDir,
	})

	err := researchCmd.Execute()
	require.NoError(t, err)

	today := time.Now().Format("2006-01-02")
	expectedFile := filepath.Join(tmpDir, "drafts", "draft-coordinator-"+today+".md")

	_, statErr := os.Stat(expectedFile)
	require.NoError(t, statErr, "draft file should exist at %s", expectedFile)

	content, readErr := os.ReadFile(expectedFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "Research Prompt: coordinator")

	// Verify stdout message
	outStr := stdout.String()
	assert.Contains(t, outStr, "drafts/draft-coordinator-"+today+".md")
	assert.Contains(t, outStr, "Feed this file to a Claude Code agent")
}

func TestResearch_ApiModeWarning(t *testing.T) {
	tmpDir := t.TempDir()

	researchCmd := cmd.NewResearchCommand()
	var stdout, stderr bytes.Buffer
	researchCmd.SetOut(&stdout)
	researchCmd.SetErr(&stderr)
	researchCmd.SetArgs([]string{
		"coordinator",
		"--mode", "api",
		"--profiles-dir", tmpDir,
	})

	err := researchCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stderr.String(), "API mode not yet implemented")
}

func TestResearch_InvalidMode_Error(t *testing.T) {
	tmpDir := t.TempDir()

	researchCmd := cmd.NewResearchCommand()
	var buf bytes.Buffer
	researchCmd.SetOut(&buf)
	researchCmd.SetErr(&buf)
	researchCmd.SetArgs([]string{
		"coordinator",
		"--mode", "invalid",
		"--profiles-dir", tmpDir,
	})

	err := researchCmd.Execute()
	require.Error(t, err)
	assert.ErrorContains(t, err, "--mode")
}
