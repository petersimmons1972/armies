"""Tests for frontmatter-safe XP update (#20, #32, #37).

Root cause: re.sub matches the first `xp:` token anywhere in the file,
including inside role body sections, corrupting profile content.

Fix: _update_frontmatter_field() isolates the frontmatter, parses it with
yaml.safe_load, updates the key, and re-serializes — leaving the body intact.
"""

from __future__ import annotations

from pathlib import Path

import pytest
import yaml


# ---------------------------------------------------------------------------
# Helper to write a profile file with xp: in the body too
# ---------------------------------------------------------------------------

_PROFILE_WITH_BODY_XP = """\
---
name: grace
xp: 10
rank: Admiral
---

## Base Persona

Grace is relentless.

## Role: implementer

This role gains xp: 5 per successful mission. The more missions, the more xp: awarded.
"""

_PROFILE_NO_BODY_XP = """\
---
name: grace
xp: 0
rank: Lieutenant
---

## Base Persona

Grace is relentless.
"""

_PROFILE_XP_MISSING = """\
---
name: grace
rank: Ensign
---

## Base Persona

Grace is relentless.
"""


# ---------------------------------------------------------------------------
# #20 / #32 / #37 — Body xp: must not be corrupted
# ---------------------------------------------------------------------------


def test_body_xp_not_corrupted(tmp_path):
    """_update_frontmatter_field must not modify xp: occurrences in the body."""
    from armies.cli import _update_frontmatter_field

    profile = tmp_path / "grace.md"
    profile.write_text(_PROFILE_WITH_BODY_XP, encoding="utf-8")

    _update_frontmatter_field(profile, "xp", 25)

    content = profile.read_text(encoding="utf-8")

    # Body text must be completely unchanged
    assert "This role gains xp: 5 per successful mission." in content
    assert "the more xp: awarded." in content

    # Frontmatter xp must be updated
    # Parse frontmatter to verify
    parts = content.split("---", 2)
    assert len(parts) >= 3, "Profile must have opening and closing ---"
    fm = yaml.safe_load(parts[1])
    assert fm["xp"] == 25, f"Expected xp=25 in frontmatter, got {fm['xp']}"


def test_frontmatter_xp_updated_correctly(tmp_path):
    """_update_frontmatter_field increments xp in frontmatter correctly."""
    from armies.cli import _update_frontmatter_field

    profile = tmp_path / "grace.md"
    profile.write_text(_PROFILE_NO_BODY_XP, encoding="utf-8")

    _update_frontmatter_field(profile, "xp", 42)

    content = profile.read_text(encoding="utf-8")
    parts = content.split("---", 2)
    fm = yaml.safe_load(parts[1])
    assert fm["xp"] == 42


def test_body_preserved_verbatim(tmp_path):
    """The body section must be byte-identical after frontmatter update."""
    from armies.cli import _update_frontmatter_field

    profile = tmp_path / "grace.md"
    profile.write_text(_PROFILE_WITH_BODY_XP, encoding="utf-8")

    _update_frontmatter_field(profile, "xp", 99)

    content = profile.read_text(encoding="utf-8")
    # Split at closing --- to get body
    parts = content.split("---", 2)
    body = parts[2]

    original_parts = _PROFILE_WITH_BODY_XP.split("---", 2)
    original_body = original_parts[2]

    assert body == original_body, "Body must be byte-identical after frontmatter update"


def test_other_frontmatter_keys_preserved(tmp_path):
    """Non-xp frontmatter keys must survive the update unchanged."""
    from armies.cli import _update_frontmatter_field

    profile = tmp_path / "grace.md"
    profile.write_text(_PROFILE_WITH_BODY_XP, encoding="utf-8")

    _update_frontmatter_field(profile, "xp", 7)

    content = profile.read_text(encoding="utf-8")
    parts = content.split("---", 2)
    fm = yaml.safe_load(parts[1])

    assert fm["name"] == "grace"
    assert fm["rank"] == "Admiral"


def test_missing_xp_key_creates_it(tmp_path):
    """If xp is absent from frontmatter, it should be created."""
    from armies.cli import _update_frontmatter_field

    profile = tmp_path / "grace.md"
    profile.write_text(_PROFILE_XP_MISSING, encoding="utf-8")

    _update_frontmatter_field(profile, "xp", 5)

    content = profile.read_text(encoding="utf-8")
    parts = content.split("---", 2)
    fm = yaml.safe_load(parts[1])
    assert fm["xp"] == 5


# ---------------------------------------------------------------------------
# #43 — '---' substring in YAML values must not break the parser
# ---------------------------------------------------------------------------


_PROFILE_DASH_IN_YAML_VALUE = """\
---
name: grace
xp: 10
description: 'range ----> max'
rank: Admiral
---

## Base Persona

Grace is relentless.
"""

_PROFILE_DASH_HORIZONTAL_RULE_IN_BODY = """\
---
name: grace
xp: 10
rank: Admiral
---

## Base Persona

Grace is relentless.

---

This paragraph follows a horizontal rule in the body.
"""


def test_dash_in_yaml_value_preserved(tmp_path):
    """A YAML value containing '---' must survive the update without corruption.

    The old split('---', 2) approach would tear the frontmatter at the first
    '---' substring inside a value, producing a broken YAML parse.  The
    line-by-line approach matches only lines whose stripped content is exactly
    '---', so embedded substrings are ignored.
    """
    from armies.cli import _update_frontmatter_field

    profile = tmp_path / "grace.md"
    profile.write_text(_PROFILE_DASH_IN_YAML_VALUE, encoding="utf-8")

    _update_frontmatter_field(profile, "xp", 99)

    content = profile.read_text(encoding="utf-8")

    # The description value must be byte-identical — no corruption
    assert "description: 'range ----> max'" in content or \
        "description: range ----> max" in content, (
        f"description value was corrupted. Content:\n{content}"
    )

    # Frontmatter xp must be updated
    import yaml as _yaml
    lines = content.splitlines(keepends=True)
    fm_lines = []
    in_fm = False
    delimiters_seen = 0
    for line in lines:
        if line.strip() == "---":
            delimiters_seen += 1
            if delimiters_seen == 1:
                in_fm = True
                continue
            if delimiters_seen == 2:
                break
        if in_fm:
            fm_lines.append(line)
    fm = _yaml.safe_load("".join(fm_lines))
    assert fm["xp"] == 99, f"Expected xp=99, got {fm.get('xp')}"
    assert fm["description"] == "range ----> max", (
        f"description was corrupted: {fm.get('description')!r}"
    )


def test_horizontal_rule_in_body_preserved(tmp_path):
    """A markdown horizontal rule ('---' on its own line) in the body must not
    confuse the parser.  The line-by-line approach stops at the second exact
    '---' delimiter and treats the remainder as body — so a horizontal rule
    in the body is simply part of the body string, never mistaken for a
    frontmatter delimiter.
    """
    from armies.cli import _update_frontmatter_field

    profile = tmp_path / "grace.md"
    profile.write_text(_PROFILE_DASH_HORIZONTAL_RULE_IN_BODY, encoding="utf-8")

    _update_frontmatter_field(profile, "xp", 55)

    content = profile.read_text(encoding="utf-8")

    # The horizontal rule and the paragraph after it must survive intact
    assert "This paragraph follows a horizontal rule in the body." in content, (
        f"Body content after horizontal rule was lost. Content:\n{content}"
    )

    # Frontmatter xp must be updated
    import yaml as _yaml
    lines = content.splitlines(keepends=True)
    fm_lines = []
    in_fm = False
    delimiters_seen = 0
    for line in lines:
        if line.strip() == "---":
            delimiters_seen += 1
            if delimiters_seen == 1:
                in_fm = True
                continue
            if delimiters_seen == 2:
                break
        if in_fm:
            fm_lines.append(line)
    fm = _yaml.safe_load("".join(fm_lines))
    assert fm["xp"] == 55, f"Expected xp=55, got {fm.get('xp')}"


def test_cmd_record_uses_safe_frontmatter_update(tmp_path):
    """cmd_record must not corrupt xp: in profile body (integration path).

    This test calls _update_frontmatter_field directly (the path that
    cmd_record now uses) rather than invoking the full CLI, which requires
    a real ~/.armies config on disk.
    """
    from armies.cli import _update_frontmatter_field
    import yaml

    # Create a minimal profile with xp in body
    pdir = tmp_path / "profiles"
    pdir.mkdir()
    profile = pdir / "grace.md"
    profile.write_text(_PROFILE_WITH_BODY_XP, encoding="utf-8")

    # Simulate what cmd_record now does: read xp, add delta, write back
    fm_before = yaml.safe_load(
        _PROFILE_WITH_BODY_XP.split("---", 2)[1]
    )
    current_xp = int(fm_before.get("xp", 0))  # 10
    new_xp = current_xp + 15  # 25
    _update_frontmatter_field(profile, "xp", new_xp)

    # Body xp text must remain untouched
    content = profile.read_text(encoding="utf-8")
    assert "This role gains xp: 5 per successful mission." in content, (
        f"Body was corrupted. Profile content:\n{content}"
    )
    assert "the more xp: awarded." in content, (
        f"Body was corrupted (second occurrence). Profile content:\n{content}"
    )
    # Frontmatter xp must have advanced
    parts = content.split("---", 2)
    fm = yaml.safe_load(parts[1])
    assert fm["xp"] == 25, f"Expected xp=25 (10+15), got {fm.get('xp')}"
