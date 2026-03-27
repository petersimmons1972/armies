"""GitHub sync via git CLI — operates on the ~/.armies directory."""

from __future__ import annotations

import subprocess
from pathlib import Path
from typing import Any

from .config import ARMIES_DIR


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

    if not remote_url:
        return {
            "pull_ok": False,
            "push_ok": False,
            "pull_msg": "",
            "push_msg": "",
            "error": (
                "No remote_url configured. "
                "Set remote_url in ~/.armies/config.yaml and run `armies init` again."
            ),
        }

    # Pull
    pull_rc, pull_out, pull_err = _run_git(["pull"], armies_dir)
    pull_ok = pull_rc == 0
    pull_msg = pull_out or pull_err

    # Push
    push_rc, push_out, push_err = _run_git(["push"], armies_dir)
    push_ok = push_rc == 0
    push_msg = push_out or push_err

    return {
        "pull_ok": pull_ok,
        "push_ok": push_ok,
        "pull_msg": pull_msg,
        "push_msg": push_msg,
        "error": None,
    }
