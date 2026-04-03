"""Test that cli.py does not use yaml.dump (only yaml.safe_dump) — closes #35."""

from __future__ import annotations

import ast
import sys
from pathlib import Path


def test_no_yaml_dump_in_cli():
    """cli.py must not contain calls to yaml.dump — only yaml.safe_dump."""
    cli_path = Path(__file__).parent.parent / "src" / "armies" / "cli.py"
    source = cli_path.read_text(encoding="utf-8")
    tree = ast.parse(source)

    violations = []
    for node in ast.walk(tree):
        if (
            isinstance(node, ast.Call)
            and isinstance(node.func, ast.Attribute)
            and node.func.attr == "dump"
            and isinstance(node.func.value, ast.Name)
            and node.func.value.id == "yaml"
        ):
            violations.append(node.lineno)

    assert violations == [], (
        f"cli.py still calls yaml.dump (not yaml.safe_dump) at lines: {violations}"
    )
