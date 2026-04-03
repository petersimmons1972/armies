"""Profile loading/parsing — progressive (load only what's needed).

Design constraints:
- NEVER load the entire profiles/ directory into memory at once.
- For `roster`: read frontmatter only (stop after closing ---).
- For `spawn`: read frontmatter + Base Persona + ONE role block only.
- Use a streaming/line-by-line reader, not full file reads.
"""

from __future__ import annotations

import re
from pathlib import Path
from typing import Any, Generator

import yaml

# Maximum allowed profile file size (1 MB).  Files larger than this are
# rejected before reading to prevent memory exhaustion from crafted profiles
# (issue #29).
_MAX_PROFILE_BYTES = 1024 * 1024

# Required frontmatter fields.  A profile missing any of these is malformed
# and must not be used (issue #28).
_REQUIRED_FIELDS = ("name", "xp")


# ---------------------------------------------------------------------------
# Low-level streaming helpers
# ---------------------------------------------------------------------------


def _iter_lines(path: Path) -> Generator[str, None, None]:
    """Yield lines one at a time from a file."""
    with path.open(encoding="utf-8") as fh:
        for line in fh:
            yield line.rstrip("\n")


def read_frontmatter(path: Path) -> dict[str, Any]:
    """Read ONLY the YAML frontmatter block from a profile .md file.

    Stops reading as soon as the closing '---' delimiter is found.
    Returns an empty dict if no valid frontmatter is present.
    """
    lines = _iter_lines(path)

    # First non-empty line must be '---'
    first = next(lines, None)
    if first is None or first.strip() != "---":
        return {}

    fm_lines: list[str] = []
    for line in lines:
        if line.strip() == "---":
            break
        fm_lines.append(line)

    raw = "\n".join(fm_lines)
    try:
        data = yaml.safe_load(raw)
        return data if isinstance(data, dict) else {}
    except yaml.YAMLError:
        return {}


def _check_file_size(path: Path) -> None:
    """Raise ValueError if path exceeds _MAX_PROFILE_BYTES (issue #29)."""
    size = path.stat().st_size
    if size > _MAX_PROFILE_BYTES:
        raise ValueError(
            f"Profile too large: '{path.name}' is {size} bytes "
            f"(limit {_MAX_PROFILE_BYTES} bytes / 1 MB)."
        )


def _validate_schema(fm: dict[str, Any], path: Path) -> None:
    """Raise ValueError if required frontmatter fields are missing (issue #28).

    Required fields: name, xp, and at least one of role/roles.
    """
    missing = []
    for field in _REQUIRED_FIELDS:
        if field not in fm:
            missing.append(field)
    # Accept either 'role' or 'roles' key — profiles use both conventions
    if "role" not in fm and "roles" not in fm:
        missing.append("role")
    if missing:
        raise ValueError(
            f"Profile '{path.name}' is missing required frontmatter field(s): "
            + ", ".join(repr(f) for f in missing)
        )


def read_frontmatter_validated(path: Path) -> dict[str, Any]:
    """Read and validate frontmatter from a profile.

    Enforces size cap (#29) and required field schema (#28).
    Raises ValueError on violations.
    """
    _check_file_size(path)
    fm = read_frontmatter(path)
    _validate_schema(fm, path)
    return fm


def read_frontmatter_and_sections(
    path: Path,
    sections: list[str],
) -> tuple[dict[str, Any], dict[str, str]]:
    """Read frontmatter plus specific named ## sections from a profile.

    ``sections`` is a list of section headings to capture (without the
    leading ``##``), e.g. ``["Base Persona", "Role: implementer"]``.

    Returns:
        (frontmatter_dict, {section_name: section_body, ...})

    Only the requested sections are collected; the rest of the file is
    discarded as soon as it has been scanned past.
    """
    _check_file_size(path)
    lines = _iter_lines(path)

    # --- parse frontmatter ---
    first = next(lines, None)
    if first is None or first.strip() != "---":
        return {}, {}

    fm_lines: list[str] = []
    for line in lines:
        if line.strip() == "---":
            break
        fm_lines.append(line)

    raw = "\n".join(fm_lines)
    try:
        frontmatter: dict[str, Any] = yaml.safe_load(raw) or {}
    except yaml.YAMLError:
        frontmatter = {}

    # Build a lookup set for fast membership tests
    wanted = {s.strip() for s in sections}

    # --- scan body for requested sections ---
    collected: dict[str, list[str]] = {}
    current_section: str | None = None
    capturing = False

    for line in lines:
        # Detect any ## heading
        m = re.match(r"^## (.+)$", line)
        if m:
            heading = m.group(1).strip()
            current_section = heading
            capturing = heading in wanted
            if capturing:
                collected.setdefault(current_section, [])
            continue

        if capturing and current_section is not None:
            collected[current_section].append(line)

    # Convert lists to stripped strings
    body_sections = {k: "\n".join(v).strip() for k, v in collected.items()}
    return frontmatter, body_sections


def iter_profile_paths(profiles_dir: Path) -> Generator[Path, None, None]:
    """Yield .md file paths from profiles_dir one at a time (no bulk load)."""
    if not profiles_dir.is_dir():
        return
    for p in sorted(profiles_dir.glob("*.md")):
        if p.is_file():
            yield p
