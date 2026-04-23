package profiles_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/petersimmons1972/armies/internal/profiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolveAgentPath_SymlinkOutsideRejected confirms that a symlink placed
// inside profilesDir that points outside it is rejected by ResolveAgentPath.
// This is the fix for GitHub issue #47.
func TestResolveAgentPath_SymlinkOutsideRejected(t *testing.T) {
	profilesDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "passwd.md")
	require.NoError(t, os.WriteFile(outsideFile, []byte("---\n"), 0o600))

	link := filepath.Join(profilesDir, "evil.md")
	require.NoError(t, os.Symlink(outsideFile, link))

	_, err := profiles.ResolveAgentPath(profilesDir, "evil")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside")
}

// TestResolveAgentPath_SymlinkChainOutsideRejected confirms a two-hop symlink
// chain that ultimately resolves outside the profiles directory is rejected.
func TestResolveAgentPath_SymlinkChainOutsideRejected(t *testing.T) {
	profilesDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "passwd.md")
	require.NoError(t, os.WriteFile(outsideFile, []byte("---\n"), 0o600))

	hop1 := filepath.Join(profilesDir, "hop1.md")
	require.NoError(t, os.Symlink(outsideFile, hop1))

	link := filepath.Join(profilesDir, "chain.md")
	require.NoError(t, os.Symlink(hop1, link))

	_, err := profiles.ResolveAgentPath(profilesDir, "chain")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside")
}

// TestResolveAgentPath_SymlinkInsideAccepted confirms a symlink whose final
// target is inside profilesDir is still accepted.
func TestResolveAgentPath_SymlinkInsideAccepted(t *testing.T) {
	profilesDir := t.TempDir()
	real := filepath.Join(profilesDir, "real.md")
	require.NoError(t, os.WriteFile(real, []byte("---\n"), 0o600))

	alias := filepath.Join(profilesDir, "alias.md")
	require.NoError(t, os.Symlink(real, alias))

	got, err := profiles.ResolveAgentPath(profilesDir, "alias")
	require.NoError(t, err)
	// EvalSymlinks on an existing target returns the canonical real path.
	wantReal, err := filepath.EvalSymlinks(real)
	require.NoError(t, err)
	assert.Equal(t, wantReal, got)
}

// TestResolveAgentPath_NonexistentUsesLexicalOnly confirms that when the file
// does not yet exist, the lexical guard is sufficient (no EvalSymlinks call).
// This preserves the existing contract for pre-read validation (e.g., in
// `armies record` before the file is created).
func TestResolveAgentPath_NonexistentUsesLexicalOnly(t *testing.T) {
	profilesDir := t.TempDir()
	got, err := profiles.ResolveAgentPath(profilesDir, "future-agent")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(profilesDir, "future-agent.md"), got)
}

// TestResolveAgentPath_SeparatorInNameRejected confirms names containing a
// path separator are rejected up-front, independent of containment logic.
func TestResolveAgentPath_SeparatorInNameRejected(t *testing.T) {
	profilesDir := t.TempDir()
	for _, name := range []string{"a/b", `a\b`} {
		_, err := profiles.ResolveAgentPath(profilesDir, name)
		require.Error(t, err, "name %q must be rejected", name)
		assert.Contains(t, err.Error(), "outside profiles directory")
	}
	_, err := profiles.ResolveAgentPath(profilesDir, "")
	require.Error(t, err)
}

// TestEnsureContained_RejectsSymlinkEscape confirms the exported helper (used
// by cmd/spawn.go to re-validate case-insensitive fallback results) rejects a
// path that resolves outside base after symlink evaluation. Fix for issue #48.
func TestEnsureContained_RejectsSymlinkEscape(t *testing.T) {
	profilesDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "passwd.md")
	require.NoError(t, os.WriteFile(outsideFile, []byte("---\n"), 0o600))

	link := filepath.Join(profilesDir, "FOO.md")
	require.NoError(t, os.Symlink(outsideFile, link))

	_, err := profiles.EnsureContained(profilesDir, link)
	require.Error(t, err)
}

// TestEnsureContained_AcceptsContainedPath confirms the happy path through
// the exported helper.
func TestEnsureContained_AcceptsContainedPath(t *testing.T) {
	profilesDir := t.TempDir()
	real := filepath.Join(profilesDir, "Agent.md")
	require.NoError(t, os.WriteFile(real, []byte("---\n"), 0o600))

	got, err := profiles.EnsureContained(profilesDir, real)
	require.NoError(t, err)
	wantReal, err := filepath.EvalSymlinks(real)
	require.NoError(t, err)
	assert.Equal(t, wantReal, got)
}
