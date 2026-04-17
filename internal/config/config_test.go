package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/petersimmons1972/armies/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfilesDirLexicallyValidated_EscapeRejected(t *testing.T) {
	cfg := config.Config{ProfilesDir: "/tmp/evil"}
	_, err := cfg.ProfilesDirLexicallyValidated()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside your home directory")
}

func TestProfilesDirLexicallyValidated_InsideHomeSucceeds(t *testing.T) {
	// Use LoadFrom with a nonexistent path — defaults to ~/.armies/profiles which is inside home.
	cfg, err := config.LoadFrom("/nonexistent/path/config.yaml")
	require.NoError(t, err)
	resolved, err := cfg.ProfilesDirLexicallyValidated()
	require.NoError(t, err)
	assert.NotEmpty(t, resolved)
}

func TestProfilesDirLexicallyValidated_TildeExpansion(t *testing.T) {
	cfg := config.Config{ProfilesDir: "~/.armies/custom"}
	resolved, err := cfg.ProfilesDirLexicallyValidated()
	require.NoError(t, err)
	assert.NotContains(t, resolved, "~")
	assert.Contains(t, resolved, ".armies")
}

func TestProfilesDirLexicallyValidated_EmptyRejected(t *testing.T) {
	cfg := config.Config{ProfilesDir: ""}
	_, err := cfg.ProfilesDirLexicallyValidated()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not set")
}

func TestLoad_DefaultsWhenNoFile(t *testing.T) {
	cfg, err := config.LoadFrom("/nonexistent/path/config.yaml")
	require.NoError(t, err)
	assert.Equal(t, "sonnet", cfg.DefaultModel)
	assert.Equal(t, "", cfg.RemoteURL)
}

func TestLoad_MergesOverDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(cfgPath, []byte("remote_url: https://github.com/user/repo.git\n"), 0600)
	require.NoError(t, err)

	cfg, err := config.LoadFrom(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/user/repo.git", cfg.RemoteURL)
	assert.Equal(t, "sonnet", cfg.DefaultModel) // default preserved
}

func TestLoad_ExplicitEmptyProfilesDirUsesDefault(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	// Explicit empty profiles_dir must not clobber the default.
	err := os.WriteFile(cfgPath, []byte("profiles_dir: \"\"\n"), 0600)
	require.NoError(t, err)

	cfg, err := config.LoadFrom(cfgPath)
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.ProfilesDir, "profiles_dir should fall back to default when set to empty in YAML")
}

func TestMalusLedgerPath_ContainsAccountability(t *testing.T) {
	p, err := config.MalusLedgerPath()
	require.NoError(t, err)
	assert.Contains(t, p, "accountability")
	assert.Contains(t, p, "malus-ledger.yaml")
}
