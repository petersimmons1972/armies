//go:build linux || darwin

package sync

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func acquireLock(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
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
	unix.Flock(int(f.Fd()), unix.LOCK_UN)
	f.Close()
}
