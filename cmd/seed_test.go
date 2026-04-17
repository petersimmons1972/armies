package cmd_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/petersimmons1972/armies/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTestFS creates an fs.FS with general profile files under examples/generals/.
func makeTestFS(t *testing.T, filenames ...string) fs.FS {
	t.Helper()
	m := fstest.MapFS{}
	for _, name := range filenames {
		m["examples/generals/"+name] = &fstest.MapFile{
			Data: []byte("---\nname: " + name[:len(name)-3] + "\nxp: 0\nrole: specialist\n---\nBody\n"),
		}
	}
	return m
}

func TestSeed_InstallsProfiles(t *testing.T) {
	dir := t.TempDir()
	testFS := makeTestFS(t, "test-general.md")

	seedCmd := cmd.NewSeedCommandFS(testFS)
	seedCmd.SetArgs([]string{"--profiles-dir", dir})
	var buf strings.Builder
	seedCmd.SetOut(&buf)
	err := seedCmd.Execute()
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(dir, "test-general.md"))
	assert.Contains(t, buf.String(), "Installed 1")
}

func TestSeed_SkipsExistingWithoutForce(t *testing.T) {
	dir := t.TempDir()
	// Pre-create file with different content
	existing := filepath.Join(dir, "test-general.md")
	os.WriteFile(existing, []byte("original content"), 0644)

	testFS := makeTestFS(t, "test-general.md")
	seedCmd := cmd.NewSeedCommandFS(testFS)
	seedCmd.SetArgs([]string{"--profiles-dir", dir})
	var buf strings.Builder
	seedCmd.SetOut(&buf)
	err := seedCmd.Execute()
	require.NoError(t, err)

	data, _ := os.ReadFile(existing)
	assert.Equal(t, "original content", string(data), "existing file must not be overwritten without --force")
	assert.Contains(t, buf.String(), "skipped 1")
}

func TestSeed_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "test-general.md")
	os.WriteFile(existing, []byte("old"), 0644)

	testFS := makeTestFS(t, "test-general.md")
	seedCmd := cmd.NewSeedCommandFS(testFS)
	seedCmd.SetArgs([]string{"--profiles-dir", dir, "--force"})
	var buf strings.Builder
	seedCmd.SetOut(&buf)
	err := seedCmd.Execute()
	require.NoError(t, err)

	data, _ := os.ReadFile(existing)
	assert.NotEqual(t, "old", string(data), "force must overwrite existing file")
	assert.Contains(t, buf.String(), "Installed 1")
}

func TestSeed_MissingProfilesDirFlag(t *testing.T) {
	testFS := makeTestFS(t, "test-general.md")
	seedCmd := cmd.NewSeedCommandFS(testFS)
	seedCmd.SetArgs([]string{})
	err := seedCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--profiles-dir")
}
