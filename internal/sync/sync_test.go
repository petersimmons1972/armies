package sync_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	armiesSync "github.com/petersimmons1972/armies/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRemoteURL_Allowed(t *testing.T) {
	for _, url := range []string{
		"https://github.com/user/repo.git",
		"ssh://git@github.com/user/repo.git",
		"git@github.com:user/repo.git",
	} {
		assert.NoError(t, armiesSync.ValidateRemoteURL(url), "should allow: %s", url)
	}
}

func TestValidateRemoteURL_Rejected(t *testing.T) {
	for _, url := range []string{
		"",
		"http://github.com/user/repo.git",
		"file:///home/user/repo",
		"ftp://example.com/repo",
	} {
		assert.Error(t, armiesSync.ValidateRemoteURL(url), "should reject: %s", url)
	}
}

func TestSync_DirtyTreeBlocked(t *testing.T) {
	dir := t.TempDir()
	// Init git repo
	exec.Command("git", "-C", dir, "init").Run()
	exec.Command("git", "-C", dir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", dir, "config", "user.name", "Test").Run()
	exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "init").Run()
	// Stage a file to make the tree dirty
	os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("untracked"), 0644)
	exec.Command("git", "-C", dir, "add", "dirty.txt").Run()

	result := armiesSync.Sync(armiesSync.SyncOptions{
		ArmiesDir: dir,
		RemoteURL: "https://github.com/user/repo.git",
	})
	require.NotNil(t, result.Error)
	assert.Contains(t, *result.Error, "uncommitted")
}

func TestSync_InvalidURL_Blocked(t *testing.T) {
	result := armiesSync.Sync(armiesSync.SyncOptions{
		ArmiesDir: t.TempDir(),
		RemoteURL: "http://insecure.example.com/repo.git",
	})
	require.NotNil(t, result.Error)
	assert.Contains(t, *result.Error, "disallowed")
}
