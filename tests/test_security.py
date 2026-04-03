"""Tests for path traversal security fixes — closes #26, #34."""

from __future__ import annotations

from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from armies.config import profiles_dir


# ---------------------------------------------------------------------------
# #34 — profiles_dir must stay within home
# ---------------------------------------------------------------------------


def test_profiles_dir_rejects_absolute_outside_home(tmp_path):
    """profiles_dir set to /etc/passwd must be rejected."""
    config = {"profiles_dir": "/etc"}
    # Should raise ValueError when outside home
    with pytest.raises(ValueError, match="profiles_dir"):
        _validated_profiles_dir(config)


def test_profiles_dir_accepts_valid_path():
    """profiles_dir inside home must be accepted."""
    home = Path.home()
    config = {"profiles_dir": str(home / ".armies" / "profiles")}
    result = _validated_profiles_dir(config)
    assert result.is_relative_to(home)


# ---------------------------------------------------------------------------
# #26 — agent argument path traversal in CLI commands
# ---------------------------------------------------------------------------


def test_spawn_rejects_path_traversal(tmp_path):
    """agent='../../etc/passwd' must not resolve outside profiles_dir."""
    profiles = tmp_path / "profiles"
    profiles.mkdir()
    with pytest.raises(ValueError, match="outside"):
        _validate_agent_path(profiles, "../../etc/passwd")


def test_spawn_accepts_valid_agent(tmp_path):
    """Normal agent name must resolve correctly."""
    profiles = tmp_path / "profiles"
    profiles.mkdir()
    # Should not raise
    result = _validate_agent_path(profiles, "grace-hopper")
    assert result == (profiles / "grace-hopper.md").resolve()


def test_spawn_rejects_absolute_agent(tmp_path):
    """Absolute path as agent name must be rejected."""
    profiles = tmp_path / "profiles"
    profiles.mkdir()
    with pytest.raises(ValueError, match="outside"):
        _validate_agent_path(profiles, "/etc/passwd")


# ---------------------------------------------------------------------------
# Helpers — these will call the functions we are about to add
# ---------------------------------------------------------------------------


def _validated_profiles_dir(config):
    """Call the validated version of profiles_dir."""
    from armies.config import profiles_dir_validated
    return profiles_dir_validated(config)


def _validate_agent_path(pdir: Path, agent: str) -> Path:
    """Call the path-traversal guard we are about to add."""
    from armies.cli import _resolve_agent_path
    return _resolve_agent_path(pdir, agent)
