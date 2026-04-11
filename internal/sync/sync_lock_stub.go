//go:build !linux && !darwin

package sync

import "os"

func acquireLock(path string) (*os.File, error) { return nil, nil }
func releaseLock(f *os.File)                    {}
