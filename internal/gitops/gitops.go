package gitops

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// RunGit runs git with args in dir. Returns trimmed stdout, stderr, and error.
// An error is returned if the command exits non-zero.
//
// IMPORTANT — caller contract: args MUST NOT include any element derived from
// untrusted input unless it has been validated. Git has several flags that
// invoke attacker-controlled binaries (`--upload-pack=`, `--receive-pack=`,
// `-c core.gitProxy=…`, `-c core.pager=…`, `-c core.sshCommand=…`) and will
// happily interpret a value that starts with `-` as a flag. When passing
// user-supplied URLs, branches, or paths, use ValidateRemoteURL / ValidateRef
// / a path containment check first, and where the arg is a positional that
// follows subcommand-specific options, pass `--` before it.
func RunGit(dir string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}

// dangerousArgSubstrings rejects any occurrence of a git flag whose value is
// an attacker-controlled binary. These are not reachable through documented
// remote-URL parsing but appear as literal substrings when an attacker injects
// a crafted URL into argv.
var dangerousArgSubstrings = []string{
	"--upload-pack=",
	"--receive-pack=",
	"core.gitproxy=",
	"core.pager=",
	"core.sshcommand=",
	"core.fsmonitor=",
}

// ValidateRemoteURL validates a git remote URL as safe to pass to RunGit.
// Allowed forms: `https://…`, `ssh://…`, SCP-style `git@host:path`.
// Rejects:
//   - empty input
//   - values starting with `-` (would be interpreted as a flag by git)
//   - any value containing a dangerous git-config substring that could
//     invoke an attacker-controlled binary (--upload-pack=, -c core.pager=, …)
//   - http://, file://, ext:: and other non-network schemes that could be
//     abused as SSRF or ext-cmd execution vectors
func ValidateRemoteURL(rawURL string) error {
	s := strings.TrimSpace(rawURL)
	if s == "" {
		return fmt.Errorf("remote url is empty; set https:// or git@ URL")
	}
	if strings.HasPrefix(s, "-") {
		return fmt.Errorf("remote url %q starts with '-'; would be interpreted as a git flag", s)
	}
	lower := strings.ToLower(s)
	for _, bad := range dangerousArgSubstrings {
		if strings.Contains(lower, bad) {
			return fmt.Errorf("remote url contains disallowed substring %q", bad)
		}
	}

	if strings.HasPrefix(s, "git@") {
		// SCP-style SSH: "git@host:path". Validate it has the colon separator
		// and no whitespace that could enable argv splitting downstream.
		if !strings.Contains(s, ":") {
			return fmt.Errorf("scp-style remote url %q missing ':' separator", s)
		}
		if strings.ContainsAny(s, " \t\n\r") {
			return fmt.Errorf("remote url %q contains whitespace", s)
		}
		return nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("malformed remote url: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" && scheme != "ssh" {
		return fmt.Errorf("remote url uses disallowed protocol %q; only https://, ssh://, git@ allowed", scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("remote url %q has no host", s)
	}
	return nil
}
