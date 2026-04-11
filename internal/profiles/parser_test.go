package profiles_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/petersimmons1972/armies/internal/profiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAgentPath_TraversalRejected(t *testing.T) {
	dir := t.TempDir()
	_, err := profiles.ResolveAgentPath(dir, "../../../etc/passwd")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside profiles directory")
}

func TestResolveAgentPath_ValidName(t *testing.T) {
	dir := t.TempDir()
	// Write the file so it actually exists (not required for path resolution, but realistic)
	err := os.WriteFile(filepath.Join(dir, "grace-hopper.md"), []byte("---\n"), 0644)
	require.NoError(t, err)

	path, err := profiles.ResolveAgentPath(dir, "grace-hopper")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "grace-hopper.md"), path)
}

func TestParseProfile_FileTooLarge(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "large-*.md")
	require.NoError(t, err)
	// Write 1MB + 1 byte
	data := make([]byte, 1024*1024+1)
	_, err = f.Write(data)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	_, _, err = profiles.ParseProfile(f.Name(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too large")
}

func TestParseProfile_MissingRequiredFields(t *testing.T) {
	// Missing xp field
	content := "---\nname: test\nrole: specialist\n---\n"
	f, err := os.CreateTemp(t.TempDir(), "missing-*.md")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	_, _, err = profiles.ParseProfile(f.Name(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "xp")
}

func TestParseProfile_DashesInFrontmatterValue(t *testing.T) {
	fm, _, err := profiles.ParseProfile("testdata/profiles/dashes-in-value.md", nil)
	require.NoError(t, err)
	assert.Equal(t, "dash-agent", fm["name"])
	desc, ok := fm["description"]
	require.True(t, ok, "expected description key in frontmatter")
	assert.Contains(t, desc, "---")
}

func TestParseProfile_HeadingInCodeFenceIgnored(t *testing.T) {
	fm, sections, err := profiles.ParseProfile(
		"testdata/profiles/code-fence.md",
		[]string{"Base Persona", "Role: specialist"},
	)
	require.NoError(t, err)
	assert.NotNil(t, fm)

	roleContent, ok := sections["Role: specialist"]
	require.True(t, ok, "expected 'Role: specialist' section")
	assert.Contains(t, roleContent, "Actual role content")

	_, fenceHeadingPresent := sections["This heading inside fence must be ignored"]
	assert.False(t, fenceHeadingPresent, "heading inside code fence must not appear as a section key")
}

func TestParseProfile_EarlyExit(t *testing.T) {
	fm, sections, err := profiles.ParseProfile(
		"testdata/profiles/valid.md",
		[]string{"Base Persona", "Role: specialist"},
	)
	require.NoError(t, err)
	assert.NotNil(t, fm)

	assert.Contains(t, sections, "Base Persona")
	assert.Contains(t, sections, "Role: specialist")

	_, coordinatorPresent := sections["Role: coordinator"]
	assert.False(t, coordinatorPresent, "Role: coordinator must not be present due to early exit")
}

func TestStreamProfiles_ReturnsSortedPaths(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.md"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.md"), []byte(""), 0644))

	paths, err := profiles.StreamProfiles(dir)
	require.NoError(t, err)
	require.Len(t, paths, 2)
	assert.True(t, strings.HasSuffix(paths[0], "a.md"), "first path should be a.md, got %s", paths[0])
	assert.True(t, strings.HasSuffix(paths[1], "b.md"), "second path should be b.md, got %s", paths[1])
}

func TestStreamProfiles_NonexistentDir(t *testing.T) {
	paths, err := profiles.StreamProfiles("/nonexistent/dir/that/cannot/exist")
	assert.NoError(t, err)
	assert.Nil(t, paths)
}

func TestParseProfile_EmptyFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "empty-*.md")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	fm, sections, err := profiles.ParseProfile(f.Name(), nil)
	require.NoError(t, err)
	assert.NotNil(t, fm)
	assert.NotNil(t, sections)
	assert.Empty(t, fm)
	assert.Empty(t, sections)
}

func TestParseProfile_FrontmatterOnly(t *testing.T) {
	content := "---\nname: solo\nxp: 1\nrole: tester\n---\n"
	f, err := os.CreateTemp(t.TempDir(), "fm-only-*.md")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	fm, sections, err := profiles.ParseProfile(f.Name(), []string{"Base Persona"})
	require.NoError(t, err)
	assert.Equal(t, "solo", fm["name"])
	assert.Empty(t, sections)
}

func TestParseProfile_ExactlyOneMB(t *testing.T) {
	// A file of exactly 1 MB must pass the size guard (the limit is strictly >1 MB).
	f, err := os.CreateTemp(t.TempDir(), "exact-mb-*.md")
	require.NoError(t, err)
	data := make([]byte, 1024*1024)
	_, err = f.Write(data)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// The file has no frontmatter, so we get empty maps — but no size error.
	_, _, err = profiles.ParseProfile(f.Name(), nil)
	require.NoError(t, err)
}

func TestParseProfile_RequestedSectionAbsent(t *testing.T) {
	fm, sections, err := profiles.ParseProfile(
		"testdata/profiles/valid.md",
		[]string{"Nonexistent Section"},
	)
	require.NoError(t, err)
	assert.NotNil(t, fm)
	_, present := sections["Nonexistent Section"]
	assert.False(t, present, "absent section must not appear in returned map")
}

func TestParseProfile_NoFrontmatter(t *testing.T) {
	content := "Just some plain text.\nNo frontmatter here.\n"
	f, err := os.CreateTemp(t.TempDir(), "no-fm-*.md")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	fm, sections, err := profiles.ParseProfile(f.Name(), nil)
	require.NoError(t, err)
	assert.Empty(t, fm)
	assert.Empty(t, sections)
}

func TestParseProfile_UnterminatedFrontmatter(t *testing.T) {
	// Opening "---" but no closing "---".
	content := "---\nname: ghost\nxp: 0\nrole: phantom\n"
	f, err := os.CreateTemp(t.TempDir(), "unterminated-*.md")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	_, _, err = profiles.ParseProfile(f.Name(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unterminated frontmatter")
}

func TestParseProfile_TildeFenceHeadingIgnored(t *testing.T) {
	fm, sections, err := profiles.ParseProfile(
		"testdata/profiles/tilde-fence.md",
		[]string{"Base Persona", "Role: specialist"},
	)
	require.NoError(t, err)
	assert.NotNil(t, fm)

	roleContent, ok := sections["Role: specialist"]
	require.True(t, ok, "expected 'Role: specialist' section")
	assert.Contains(t, roleContent, "Actual tilde role content")

	_, tildeHeadingPresent := sections["This heading inside tilde fence must be ignored"]
	assert.False(t, tildeHeadingPresent, "heading inside tilde fence must not appear as a section key")
}
