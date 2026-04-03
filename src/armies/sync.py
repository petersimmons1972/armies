"""GitHub sync via git CLI — operates on the ~/.armies directory."""

from __future__ import annotations

import fcntl
import subprocess
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

from .config import ARMIES_DIR

# Protocols that are allowed for remote_url.  We require encryption (https or
# SSH) and explicitly forbid file:// and http:// which are either local-path
# injection vectors or unencrypted channels.
_ALLOWED_SCHEMES = {"https", "ssh"}
_ALLOWED_SCP_PREFIX = "git@"


def _validate_remote_url(url: str) -> None:
    """Validate that *url* is a non-empty, secure remote URL.

    Raises ValueError with a human-readable message if the URL is empty,
    uses a disallowed protocol, or is otherwise malformed.

    Allowed forms:
        https://github.com/user/repo.git
        git@github.com:user/repo.git   (SCP-style SSH)
        ssh://git@github.com/user/repo.git
    """
    stripped = url.strip()
    if not stripped:
        raise ValueError(
            "remote_url is empty. "
            "Set a valid https:// or git@/ssh:// URL in ~/.armies/config.yaml."
        )

    # SCP-style git@ URLs (git@host:path) are not parseable by urlparse as a
    # scheme URL, so we handle them separately first.
    if stripped.startswith(_ALLOWED_SCP_PREFIX):
        return  # git@... is valid SSH

    parsed = urlparse(stripped)
    scheme = parsed.scheme.lower()

    if scheme not in _ALLOWED_SCHEMES:
        raise ValueError(
            f"remote_url uses disallowed protocol '{scheme}://'. "
            f"Only https://, ssh://, and git@ URLs are permitted. "
            f"Got: {stripped}"
        )


def _run_git(args: list[str], cwd: Path) -> tuple[int, str, str]:
    """Run a git command and return (returncode, stdout, stderr)."""
    result = subprocess.run(
        ["git", "-C", str(cwd)] + args,
        capture_output=True,
        text=True,
    )
    return result.returncode, result.stdout.strip(), result.stderr.strip()


def sync_armies(config: dict[str, Any]) -> dict[str, Any]:
    """Pull then push the ~/.armies git repository.

    Returns a dict with keys:
        pull_ok    — bool
        push_ok    — bool
        pull_msg   — str
        push_msg   — str
        error      — str | None  (set if a hard error occurred before any git op)
    """
    remote_url: str = config.get("remote_url", "").strip()
    armies_dir = ARMIES_DIR

    if not armies_dir.is_dir():
        return {
            "pull_ok": False,
            "push_ok": False,
            "pull_msg": "",
            "push_msg": "",
            "error": f"{armies_dir} does not exist. Run `armies init` first.",
        }

    # Check git repo
    rc, _, _ = _run_git(["rev-parse", "--is-inside-work-tree"], armies_dir)
    if rc != 0:
        return {
            "pull_ok": False,
            "push_ok": False,
            "pull_msg": "",
            "push_msg": "",
            "error": (
                f"{armies_dir} is not a git repository. "
                "Run `armies init` with a remote URL to set one up."
            ),
        }

    # Validate the URL before running any git commands.  A compromised config
    # or DNS-hijacked remote could deliver adversarial content; rejecting bad
    # protocols here is the first line of defence (issues #19, #24).
    try:
        _validate_remote_url(remote_url)
    except ValueError as exc:
        return {
            "pull_ok": False,
            "push_ok": False,
            "pull_msg": "",
            "push_msg": "",
            "error": str(exc),
        }

    # Acquire an exclusive file lock before running any git operations.  This
    # is a best-effort guard against simultaneous `armies sync` invocations on
    # the same machine (e.g. two terminal sessions, a cron job, and an IDE).
    # It does NOT protect against truly concurrent syncs from different machines
    # — git itself handles that case via fast-forward rejection (issue #42).
    lock_path = armies_dir / ".sync.lock"
    lock_fh = open(lock_path, "w")
    try:
        fcntl.flock(lock_fh, fcntl.LOCK_EX)
    except OSError:
        lock_fh.close()
        return {
            "pull_ok": False,
            "push_ok": False,
            "pull_msg": "",
            "push_msg": "",
            "error": "Could not acquire sync lock. Another sync may be running.",
        }

    try:
        return _sync_with_lock(armies_dir, remote_url)
    finally:
        fcntl.flock(lock_fh, fcntl.LOCK_UN)
        lock_fh.close()


def _sync_with_lock(armies_dir: Path, remote_url: str) -> dict[str, Any]:
    """Inner sync logic — called while the .sync.lock is held."""
    # Check for uncommitted changes before any git network operation.  Syncing
    # over a dirty working tree can silently overwrite local edits or produce
    # a confusing merge state (issue #36).
    dirty_rc, dirty_out, _ = _run_git(["status", "--porcelain"], armies_dir)
    if dirty_rc == 0 and dirty_out.strip():
        return {
            "pull_ok": False,
            "push_ok": False,
            "pull_msg": "",
            "push_msg": "",
            "error": (
                "Cannot sync: uncommitted changes in ~/.armies. "
                "Commit or stash first."
            ),
        }

    # Pull with --ff-only to prevent silent merge commits.  If the remote has
    # diverged, the operator must resolve the conflict manually before syncing
    # (issue #25).
    pull_rc, pull_out, pull_err = _run_git(["pull", "--ff-only", "origin", "master"], armies_dir)
    pull_ok = pull_rc == 0
    pull_msg = pull_out or pull_err

    # Do not push if pull failed — pushing after a failed pull can lose work
    # or push to a state the operator has not reviewed (issue #38).
    if not pull_ok:
        return {
            "pull_ok": False,
            "push_ok": False,
            "pull_msg": pull_msg,
            "push_msg": "",
            "error": None,
        }

    # Push with explicit remote and branch to match the pull above.  A bare
    # `git push` can operate on the wrong tracking branch if the local
    # checkout is in an unexpected state (issue #25).
    push_rc, push_out, push_err = _run_git(["push", "origin", "master"], armies_dir)
    push_ok = push_rc == 0
    push_msg = push_out or push_err

    return {
        "pull_ok": pull_ok,
        "push_ok": push_ok,
        "pull_msg": pull_msg,
        "push_msg": push_msg,
        "error": None,
    }
