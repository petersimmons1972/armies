package sync

import (
	"fmt"
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

// ValidateRemoteURL delegates to gitops.ValidateRemoteURL, which enforces
// protocol allow-list plus flag-injection and git-config-substring defenses.
// Preserved here as a thin wrapper for historical call sites and to keep the
// "remote_url is empty" phrasing that references the config file.
func ValidateRemoteURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("remote_url is empty; set https:// or git@ URL in ~/.armies/config.yaml")
	}
	return gitops.ValidateRemoteURL(rawURL)
}

// Sync pulls then pushes the armies directory.
func Sync(opts SyncOptions) SyncResult {
	if err := ValidateRemoteURL(opts.RemoteURL); err != nil {
		return errResult(err.Error())
	}

	lf, err := acquireLock(opts.ArmiesDir, ".sync.lock")
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
