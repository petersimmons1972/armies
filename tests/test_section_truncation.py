"""Test that ## sub-headings inside a section body don't truncate it — closes #41."""

from __future__ import annotations

from pathlib import Path


PROFILE_WITH_SUBHEADING = """\
---
name: grace
role: implementer
xp: 0
---

## Base Persona

You built the compiler before the committee approved it.

### Sub-heading inside Base Persona

This text must not be lost.

More body text after the sub-heading.

## Role: implementer

Operational instructions here.
"""

PROFILE_WITH_FENCED_HEADING = """\
---
name: grace
role: implementer
xp: 0
---

## Base Persona

Here is some code:

```
## This looks like a heading but is inside a fence
more code here
```

After the fence — must still be captured.

## Role: implementer

Role content.
"""


def _write(path: Path, content: str) -> Path:
    path.write_text(content, encoding="utf-8")
    return path


def test_subheading_in_body_not_truncated(tmp_path):
    """A ### sub-heading inside a section body must not stop capture."""
    from armies.profiles import read_frontmatter_and_sections

    p = _write(tmp_path / "grace.md", PROFILE_WITH_SUBHEADING)
    _, sections = read_frontmatter_and_sections(p, ["Base Persona"])
    body = sections.get("Base Persona", "")
    assert "Sub-heading inside Base Persona" in body, (
        "Sub-heading stripped from Base Persona body"
    )
    assert "This text must not be lost." in body, (
        "Content after sub-heading stripped from Base Persona body"
    )
    assert "More body text after the sub-heading." in body


def test_heading_inside_fenced_block_not_boundary(tmp_path):
    """A ## line inside a fenced code block must not start a new section."""
    from armies.profiles import read_frontmatter_and_sections

    p = _write(tmp_path / "grace.md", PROFILE_WITH_FENCED_HEADING)
    _, sections = read_frontmatter_and_sections(p, ["Base Persona"])
    body = sections.get("Base Persona", "")
    assert "After the fence" in body, (
        "Content after fenced block was stripped because '## inside fence' was treated as a section boundary"
    )


def test_role_section_still_parsed(tmp_path):
    """Role section after a sub-heading-containing Base Persona must still be captured."""
    from armies.profiles import read_frontmatter_and_sections

    p = _write(tmp_path / "grace.md", PROFILE_WITH_SUBHEADING)
    _, sections = read_frontmatter_and_sections(p, ["Base Persona", "Role: implementer"])
    assert "Role: implementer" in sections
    assert "Operational instructions" in sections["Role: implementer"]
