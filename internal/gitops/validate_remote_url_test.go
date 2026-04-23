package gitops_test

import (
	"strings"
	"testing"

	"github.com/petersimmons1972/armies/internal/gitops"
)

func TestValidateRemoteURL_AllowedForms(t *testing.T) {
	allowed := []string{
		"https://github.com/owner/repo.git",
		"https://github.com/owner/repo",
		"ssh://git@github.com/owner/repo.git",
		"git@github.com:owner/repo.git",
		"git@example.org:group/sub/repo",
	}
	for _, u := range allowed {
		if err := gitops.ValidateRemoteURL(u); err != nil {
			t.Errorf("expected %q to be allowed, got error: %v", u, err)
		}
	}
}

func TestValidateRemoteURL_RejectsEmpty(t *testing.T) {
	for _, u := range []string{"", "  ", "\t\n"} {
		if err := gitops.ValidateRemoteURL(u); err == nil {
			t.Errorf("expected empty input %q to be rejected", u)
		}
	}
}

func TestValidateRemoteURL_RejectsFlagInjection(t *testing.T) {
	// A value that starts with '-' would be consumed by git as a flag rather
	// than a positional URL argument — must be refused.
	for _, u := range []string{"-", "--upload-pack=/tmp/pwn", "-c", "-core.pager=curl"} {
		err := gitops.ValidateRemoteURL(u)
		if err == nil {
			t.Errorf("expected %q to be rejected as flag injection", u)
		}
	}
}

func TestValidateRemoteURL_RejectsDangerousSubstrings(t *testing.T) {
	// Even embedded in a URL-ish form, these config keys must be refused.
	cases := []string{
		"https://example.com/repo.git --upload-pack=/tmp/pwn",
		"https://example.com/repo?core.pager=/tmp/pwn",
		"https://example.com/repo#core.gitproxy=/tmp/pwn",
		"git@host:repo core.sshcommand=/tmp/pwn",
		"https://example.com/?core.fsmonitor=/tmp/pwn",
	}
	for _, u := range cases {
		err := gitops.ValidateRemoteURL(u)
		if err == nil {
			t.Errorf("expected %q to be rejected for dangerous substring", u)
			continue
		}
		if !strings.Contains(err.Error(), "disallowed substring") &&
			!strings.Contains(err.Error(), "whitespace") {
			t.Errorf("unexpected error for %q: %v", u, err)
		}
	}
}

func TestValidateRemoteURL_RejectsDisallowedSchemes(t *testing.T) {
	for _, u := range []string{
		"http://example.com/repo.git",
		"file:///etc/passwd",
		"ext::sh -c pwn",
		"ftp://example.com/repo",
	} {
		if err := gitops.ValidateRemoteURL(u); err == nil {
			t.Errorf("expected %q to be rejected by scheme allow-list", u)
		}
	}
}

func TestValidateRemoteURL_RejectsMalformedScp(t *testing.T) {
	// git@host without ':' is not SCP-style and must not be silently accepted.
	for _, u := range []string{"git@host", "git@", "git@ host:repo"} {
		if err := gitops.ValidateRemoteURL(u); err == nil {
			t.Errorf("expected %q to be rejected", u)
		}
	}
}
