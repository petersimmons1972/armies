package profiles_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/petersimmons1972/armies/internal/profiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixtureProfile = `---
name: test-agent
xp: 100
role: specialist
---
## Base Persona
This body must be preserved exactly.

Line two of body.
`

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestUpdateFrontmatterField_UpdatesXP(t *testing.T) {
	path := writeTemp(t, fixtureProfile)
	err := profiles.UpdateFrontmatterField(path, "xp", 250)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "xp: 250")
}

func TestUpdateFrontmatterField_BodyPreserved(t *testing.T) {
	path := writeTemp(t, fixtureProfile)
	err := profiles.UpdateFrontmatterField(path, "xp", 999)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "## Base Persona\nThis body must be preserved exactly.")
	assert.Contains(t, body, "Line two of body.")
}

func TestUpdateFrontmatterField_DashValueNotCorrupted(t *testing.T) {
	const dashProfile = `---
name: test-agent
xp: 5
role: specialist
desc: "has --- in it"
---
Body
`
	path := writeTemp(t, dashProfile)
	err := profiles.UpdateFrontmatterField(path, "xp", 20)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "xp: 20")
	// The desc value with --- must survive round-trip intact.
	assert.Contains(t, body, "has --- in it")
	assert.Contains(t, body, "Body\n")
}

func TestUpdateFrontmatterField_NoFrontmatter_Error(t *testing.T) {
	path := writeTemp(t, "just body text\n")
	err := profiles.UpdateFrontmatterField(path, "xp", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no frontmatter")
}

func TestUpdateFrontmatterField_AddsNewKey(t *testing.T) {
	path := writeTemp(t, fixtureProfile)
	err := profiles.UpdateFrontmatterField(path, "new_field", "hello")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "new_field: hello")
	assert.Contains(t, body, "xp: 100")
}
