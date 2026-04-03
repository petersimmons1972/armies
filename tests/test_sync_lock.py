"""Tests for sync concurrency guard via file locking (#42).

armies sync must acquire an exclusive lock on ~/.armies/.sync.lock before
running git operations, and release it when done. This is a best-effort
guard against simultaneous syncs from the same machine.
"""

from __future__ import annotations

import fcntl
import os
import threading
from pathlib import Path
from unittest.mock import patch, call

import pytest

from armies.sync import sync_armies


def _base_config():
    return {
        "remote_url": "https://github.com/user/repo.git",
        "default_model": "sonnet",
    }


# ---------------------------------------------------------------------------
# Lock file is created and used
# ---------------------------------------------------------------------------


def test_sync_creates_lock_file(tmp_path):
    """sync_armies must create a .sync.lock file in the armies dir."""
    config = _base_config()
    lock_path = tmp_path / ".sync.lock"

    def fake_run_git(args, cwd):
        if args == ["rev-parse", "--is-inside-work-tree"]:
            return 0, "true", ""
        if args == ["status", "--porcelain"]:
            return 0, "", ""
        if args[0] == "pull":
            return 0, "Already up to date.", ""
        if args[0] == "push":
            return 0, "", ""
        raise AssertionError(f"Unexpected git call: {args}")

    with patch("armies.sync.ARMIES_DIR", tmp_path), \
         patch("armies.sync._run_git", side_effect=fake_run_git):
        sync_armies(config)

    assert lock_path.exists(), f"Expected .sync.lock at {lock_path}"


def test_sync_releases_lock_after_completion(tmp_path):
    """After sync_armies returns, another process must be able to acquire the lock."""
    config = _base_config()
    lock_path = tmp_path / ".sync.lock"

    def fake_run_git(args, cwd):
        if args == ["rev-parse", "--is-inside-work-tree"]:
            return 0, "true", ""
        if args == ["status", "--porcelain"]:
            return 0, "", ""
        if args[0] == "pull":
            return 0, "Already up to date.", ""
        if args[0] == "push":
            return 0, "", ""
        raise AssertionError(f"Unexpected git call: {args}")

    with patch("armies.sync.ARMIES_DIR", tmp_path), \
         patch("armies.sync._run_git", side_effect=fake_run_git):
        sync_armies(config)

    # After sync returns, we must be able to acquire the lock non-blocking
    # (if it were still held, LOCK_NB would raise BlockingIOError)
    with open(lock_path, "w") as fh:
        try:
            fcntl.flock(fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            fcntl.flock(fh, fcntl.LOCK_UN)
        except BlockingIOError:
            pytest.fail("Lock was not released after sync_armies returned")


def test_sync_releases_lock_on_error(tmp_path):
    """Lock must be released even if a git operation fails."""
    config = _base_config()
    lock_path = tmp_path / ".sync.lock"

    def fake_run_git(args, cwd):
        if args == ["rev-parse", "--is-inside-work-tree"]:
            return 0, "true", ""
        if args == ["status", "--porcelain"]:
            return 0, "", ""
        if args[0] == "pull":
            return 1, "", "fatal: could not read from remote"
        raise AssertionError(f"Unexpected git call: {args}")

    with patch("armies.sync.ARMIES_DIR", tmp_path), \
         patch("armies.sync._run_git", side_effect=fake_run_git):
        result = sync_armies(config)

    assert result["pull_ok"] is False  # Pull failed

    # Lock must still be released
    with open(lock_path, "w") as fh:
        try:
            fcntl.flock(fh, fcntl.LOCK_EX | fcntl.LOCK_NB)
            fcntl.flock(fh, fcntl.LOCK_UN)
        except BlockingIOError:
            pytest.fail("Lock was not released after failed sync")
