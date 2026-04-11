package gitops

import (
	"bytes"
	"os/exec"
	"strings"
)

// RunGit runs git with args in dir. Returns trimmed stdout, stderr, and error.
// An error is returned if the command exits non-zero.
func RunGit(dir string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}
