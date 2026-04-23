//go:build linux || darwin

package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireLock_RejectsSeparatorInName(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"../evil", "sub/evil", "", ".", ".."} {
		if _, err := acquireLock(dir, name); err == nil {
			t.Errorf("expected error for name %q, got nil", name)
		}
	}
}

func TestAcquireLock_CreatesInsideDir(t *testing.T) {
	dir := t.TempDir()
	f, err := acquireLock(dir, ".sync.lock")
	if err != nil {
		t.Fatalf("acquireLock: %v", err)
	}
	defer releaseLock(f)

	want := filepath.Join(dir, ".sync.lock")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("lock file not at %s: %v", want, err)
	}
	info, _ := os.Stat(want)
	if info.Mode().Perm() != 0o600 {
		t.Errorf("lock mode = %o, want 0600", info.Mode().Perm())
	}
}

func TestAcquireLock_ResolvesSymlinkedDir(t *testing.T) {
	realDir := t.TempDir()
	linkDir := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	f, err := acquireLock(linkDir, ".sync.lock")
	if err != nil {
		t.Fatalf("acquireLock via symlinked dir: %v", err)
	}
	defer releaseLock(f)

	if _, err := os.Stat(filepath.Join(realDir, ".sync.lock")); err != nil {
		t.Errorf("lock file not materialized in real dir: %v", err)
	}
}

func TestAcquireLock_RejectsMissingDir(t *testing.T) {
	if _, err := acquireLock(filepath.Join(t.TempDir(), "does-not-exist"), ".sync.lock"); err == nil {
		t.Error("expected error for missing dir, got nil")
	}
}
