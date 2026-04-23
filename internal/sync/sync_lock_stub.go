//go:build !linux && !darwin

package sync

import "os"

func acquireLock(dir, name string) (*os.File, error) { return nil, nil }
func releaseLock(f *os.File)                         {}
