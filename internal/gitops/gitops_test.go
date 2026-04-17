package gitops_test

import (
	"testing"

	"github.com/petersimmons1972/armies/internal/gitops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunGit_Version(t *testing.T) {
	stdout, _, err := gitops.RunGit(".", "version")
	require.NoError(t, err)
	assert.Contains(t, stdout, "git version")
}

func TestRunGit_InvalidCommand_ReturnsError(t *testing.T) {
	_, _, err := gitops.RunGit(".", "not-a-real-git-command-xyz")
	require.Error(t, err)
}
