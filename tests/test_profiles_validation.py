"""Tests for profile YAML schema validation and size cap — closes #28, #29."""

from __future__ import annotations

from pathlib import Path

import pytest


def _write_profile(path: Path, frontmatter: dict, body: str = "") -> Path:
    """Write a minimal profile file for testing."""
    import yaml
    fm_str = yaml.dump(frontmatter, default_flow_style=False)
    content = f"---\n{fm_str}---\n{body}\n"
    path.write_text(content, encoding="utf-8")
    return path


# ---------------------------------------------------------------------------
# #28 — Schema validation: required fields
# ---------------------------------------------------------------------------


def test_read_frontmatter_rejects_missing_name(tmp_path):
    """Profile without 'name' field must raise ValueError."""
    from armies.profiles import read_frontmatter_validated
    p = _write_profile(tmp_path / "bad.md", {"role": "implementer", "xp": 0})
    with pytest.raises(ValueError, match="name"):
        read_frontmatter_validated(p)


def test_read_frontmatter_rejects_missing_role(tmp_path):
    """Profile without 'role' field must raise ValueError."""
    from armies.profiles import read_frontmatter_validated
    p = _write_profile(tmp_path / "bad.md", {"name": "grace", "xp": 0})
    with pytest.raises(ValueError, match="role"):
        read_frontmatter_validated(p)


def test_read_frontmatter_rejects_missing_xp(tmp_path):
    """Profile without 'xp' field must raise ValueError."""
    from armies.profiles import read_frontmatter_validated
    # 'role' here is supplied via a roles dict — only checking xp is missing
    p = _write_profile(tmp_path / "bad.md", {"name": "grace", "roles": {"primary": "implementer"}})
    with pytest.raises(ValueError, match="xp"):
        read_frontmatter_validated(p)


def test_read_frontmatter_accepts_valid_profile(tmp_path):
    """Profile with all required fields must pass validation."""
    from armies.profiles import read_frontmatter_validated
    p = _write_profile(tmp_path / "good.md", {"name": "grace", "role": "implementer", "xp": 0})
    fm = read_frontmatter_validated(p)
    assert fm["name"] == "grace"


# ---------------------------------------------------------------------------
# #29 — Size cap: files > 1MB must be rejected
# ---------------------------------------------------------------------------


def test_read_frontmatter_rejects_large_file(tmp_path):
    """Profile file over 1MB must raise ValueError before reading."""
    from armies.profiles import read_frontmatter_validated
    p = tmp_path / "huge.md"
    # Write just over 1MB of data
    p.write_bytes(b"x" * (1024 * 1024 + 1))
    with pytest.raises(ValueError, match="too large"):
        read_frontmatter_validated(p)


def test_read_frontmatter_and_sections_rejects_large_file(tmp_path):
    """read_frontmatter_and_sections must also enforce the size cap."""
    from armies.profiles import read_frontmatter_and_sections
    p = tmp_path / "huge.md"
    p.write_bytes(b"x" * (1024 * 1024 + 1))
    with pytest.raises(ValueError, match="too large"):
        read_frontmatter_and_sections(p, ["Base Persona"])
