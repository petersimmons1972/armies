"""Tests for git sync safety: dirty-state check (#36) and --ff-only pull (#25)."""

from __future__ import annotations

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
# #36 — Dirty state must block sync
# ---------------------------------------------------------------------------


def test_dirty_state_blocks_sync(tmp_path):
    """sync_armies must return error if working tree has uncommitted changes."""
    config = _base_config()

    def fake_run_git(args, cwd):
        if args == ["rev-parse", "--is-inside-work-tree"]:
            return 0, "true", ""
        if args == ["status", "--porcelain"]:
            # Simulate a dirty working tree
            return 0, "M  profiles/grace.md", ""
        raise AssertionError(f"Unexpected git call: {args}")

    with patch("armies.sync.ARMIES_DIR", tmp_path), \
         patch("armies.sync._run_git", side_effect=fake_run_git):
        tmp_path.mkdir(exist_ok=True)
        result = sync_armies(config)

    assert result.get("error"), "Expected an error for dirty state"
    assert "uncommitted" in result["error"].lower() or "dirty" in result["error"].lower()
    assert result["pull_ok"] is False
    assert result["push_ok"] is False


def test_clean_state_allows_sync(tmp_path):
    """sync_armies must proceed with pull if working tree is clean."""
    config = _base_config()
    calls_seen = []

    def fake_run_git(args, cwd):
        calls_seen.append(args)
        if args == ["rev-parse", "--is-inside-work-tree"]:
            return 0, "true", ""
        if args == ["status", "--porcelain"]:
            return 0, "", ""  # Clean
        if args[0] == "pull":
            return 0, "Already up to date.", ""
        if args[0] == "push":
            return 0, "", ""
        raise AssertionError(f"Unexpected git call: {args}")

    with patch("armies.sync.ARMIES_DIR", tmp_path), \
         patch("armies.sync._run_git", side_effect=fake_run_git):
        tmp_path.mkdir(exist_ok=True)
        result = sync_armies(config)

    assert result.get("error") is None
    assert result["pull_ok"] is True


# ---------------------------------------------------------------------------
# #25 — Pull must use --ff-only
# ---------------------------------------------------------------------------


def test_pull_uses_ff_only(tmp_path):
    """git pull must be called with --ff-only to prevent silent merge commits."""
    config = _base_config()
    pull_args_seen = []

    def fake_run_git(args, cwd):
        if args == ["rev-parse", "--is-inside-work-tree"]:
            return 0, "true", ""
        if args == ["status", "--porcelain"]:
            return 0, "", ""
        if args[0] == "pull":
            pull_args_seen.append(args)
            return 0, "Already up to date.", ""
        if args[0] == "push":
            return 0, "", ""
        raise AssertionError(f"Unexpected git call: {args}")

    with patch("armies.sync.ARMIES_DIR", tmp_path), \
         patch("armies.sync._run_git", side_effect=fake_run_git):
        tmp_path.mkdir(exist_ok=True)
        sync_armies(config)

    assert pull_args_seen, "pull was never called"
    assert "--ff-only" in pull_args_seen[0], (
        f"pull was not called with --ff-only. Got: {pull_args_seen[0]}"
    )


def test_push_uses_explicit_branch(tmp_path):
    """git push must be called with an explicit remote and branch name.

    A bare `git push` can silently push to the wrong tracking branch if the
    local checkout is in an unexpected state (issue #25).
    """
    config = _base_config()
    push_args_seen = []

    def fake_run_git(args, cwd):
        if args == ["rev-parse", "--is-inside-work-tree"]:
            return 0, "true", ""
        if args == ["status", "--porcelain"]:
            return 0, "", ""
        if args[0] == "pull":
            return 0, "Already up to date.", ""
        if args[0] == "push":
            push_args_seen.append(args)
            return 0, "", ""
        raise AssertionError(f"Unexpected git call: {args}")

    with patch("armies.sync.ARMIES_DIR", tmp_path), \
         patch("armies.sync._run_git", side_effect=fake_run_git):
        tmp_path.mkdir(exist_ok=True)
        sync_armies(config)

    assert push_args_seen, "push was never called"
    push_cmd = push_args_seen[0]
    # Must include both a remote name and a branch name — bare `push` is not acceptable
    assert len(push_cmd) >= 3, (
        f"push must specify remote and branch. Got: {push_cmd}"
    )
    assert "origin" in push_cmd, (
        f"push must explicitly name the 'origin' remote. Got: {push_cmd}"
    )
