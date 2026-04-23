//go:build linux || darwin

package sync

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// acquireLock opens and flock()s a lock file named `name` inside `dir`.
// The lock path is canonicalized against `dir` (symlinks resolved) and verified
// to lie inside `dir` before open, so a caller cannot be tricked into creating
// a lock file at an arbitrary filesystem location via a symlinked dir.
func acquireLock(dir, name string) (*os.File, error) {
	if strings.ContainsRune(name, filepath.Separator) || name == "" || name == "." || name == ".." {
		return nil, fmt.Errorf("invalid lock name %q", name)
	}
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve lock dir %q: %w", dir, err)
	}
	absDir, err := filepath.Abs(realDir)
	if err != nil {
		return nil, fmt.Errorf("abs lock dir %q: %w", realDir, err)
	}
	path := filepath.Join(absDir, name)
	rel, err := filepath.Rel(absDir, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("lock path %q escapes dir %q", path, absDir)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open sync lock %q: %w", path, err)
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("could not acquire sync lock: another sync may be running")
	}
	return f, nil
}

func releaseLock(f *os.File) {
	if f == nil {
		return
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_UN); err != nil {
		log.Printf("sync: flock(LOCK_UN) failed on %s: %v", f.Name(), err)
	}
	if err := f.Close(); err != nil {
		log.Printf("sync: close lock file %s failed: %v", f.Name(), err)
	}
}
