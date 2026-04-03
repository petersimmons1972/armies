"""Tests for sync.py URL validation — closes GitHub Issues #19, #24, #38."""

from __future__ import annotations

import pytest

from armies.sync import _validate_remote_url


# ---------------------------------------------------------------------------
# Protocol rejection
# ---------------------------------------------------------------------------


def test_rejects_http_url():
    """Unencrypted http:// must be rejected (issue #24)."""
    with pytest.raises(ValueError, match="http://"):
        _validate_remote_url("http://github.com/user/repo.git")


def test_rejects_file_url():
    """Local file:// paths must be rejected — local path injection vector (issue #19)."""
    with pytest.raises(ValueError, match="file://"):
        _validate_remote_url("file:///home/attacker/evil-repo")


def test_rejects_empty_string():
    """Empty URL must be rejected."""
    with pytest.raises(ValueError, match="empty"):
        _validate_remote_url("")


def test_rejects_blank_string():
    """Whitespace-only URL must be rejected."""
    with pytest.raises(ValueError, match="empty"):
        _validate_remote_url("   ")


def test_rejects_ftp_url():
    """Any other non-allowed protocol must be rejected."""
    with pytest.raises(ValueError):
        _validate_remote_url("ftp://example.com/repo.git")


# ---------------------------------------------------------------------------
# Protocol acceptance
# ---------------------------------------------------------------------------


def test_accepts_https_url():
    """https:// GitHub URL must pass validation."""
    # Should not raise
    _validate_remote_url("https://github.com/user/repo.git")


def test_accepts_git_ssh_url():
    """git@ SCP-style SSH URL must pass validation."""
    # Should not raise
    _validate_remote_url("git@github.com:user/repo.git")


def test_accepts_ssh_scheme_url():
    """ssh:// URL must pass validation."""
    # Should not raise
    _validate_remote_url("ssh://git@github.com/user/repo.git")
