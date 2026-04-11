package sync

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/petersimmons1972/armies/internal/gitops"
)

// SyncOptions configures a Sync call.
type SyncOptions struct {
	ArmiesDir string
	RemoteURL string
}

// SyncResult reports the outcome of a Sync call.
// Error is non-nil only for hard failures that prevented the operation.
type SyncResult struct {
	PullOK  bool
	PushOK  bool
	PullMsg string
	PushMsg string
	Error   *string // nil if no hard error
}

func errResult(msg string) SyncResult {
	s := msg
	return SyncResult{Error: &s}
}

// ValidateRemoteURL validates that the URL is non-empty and uses an allowed
// protocol (https, ssh, or git@ SCP). Rejects empty, http://, file://, etc.
func ValidateRemoteURL(rawURL string) error {
	s := strings.TrimSpace(rawURL)
	if s == "" {
		return fmt.Errorf("remote_url is empty; set https:// or git@ URL in ~/.armies/config.yaml")
	}
	if strings.HasPrefix(s, "git@") {
		return nil // SCP-style SSH is valid
	}
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("malformed remote_url: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" && scheme != "ssh" {
		return fmt.Errorf("remote_url uses disallowed protocol %q; only https://, ssh://, git@ allowed", scheme)
	}
	return nil
}

// Sync pulls then pushes the armies directory.
func Sync(opts SyncOptions) SyncResult {
	if err := ValidateRemoteURL(opts.RemoteURL); err != nil {
		return errResult(err.Error())
	}

	lockPath := filepath.Join(opts.ArmiesDir, ".sync.lock")
	lf, err := acquireLock(lockPath)
	if err != nil {
		return errResult(err.Error())
	}
	defer releaseLock(lf)

	// Check for uncommitted changes before any git network operation.
	stdout, _, _ := gitops.RunGit(opts.ArmiesDir, "status", "--porcelain")
	if strings.TrimSpace(stdout) != "" {
		return errResult("Cannot sync: uncommitted changes in ~/.armies. Commit or stash first.")
	}

	// Pull --ff-only to prevent silent merge commits.
	pullOut, pullErr, pullCmdErr := gitops.RunGit(opts.ArmiesDir, "pull", "--ff-only", "origin", "master")
	pullMsg := pullOut
	if pullMsg == "" {
		pullMsg = pullErr
	}
	if pullCmdErr != nil {
		return SyncResult{PullOK: false, PullMsg: pullMsg}
	}

	// Push only if pull succeeded.
	pushOut, pushErr, pushCmdErr := gitops.RunGit(opts.ArmiesDir, "push", "origin", "master")
	pushMsg := pushOut
	if pushMsg == "" {
		pushMsg = pushErr
	}

	return SyncResult{
		PullOK:  true,
		PushOK:  pushCmdErr == nil,
		PullMsg: pullMsg,
		PushMsg: pushMsg,
	}
}
