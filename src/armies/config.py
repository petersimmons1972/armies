"""Config loading from ~/.armies/config.yaml."""

from __future__ import annotations

import os
from pathlib import Path
from typing import Any

import yaml

ARMIES_DIR = Path.home() / ".armies"
CONFIG_PATH = ARMIES_DIR / "config.yaml"

DEFAULT_CONFIG: dict[str, Any] = {
    "remote_url": "",
    "default_model": "sonnet",
    "profiles_dir": str(ARMIES_DIR / "profiles"),
}


def load_config() -> dict[str, Any]:
    """Load config from ~/.armies/config.yaml, merging with defaults."""
    if not CONFIG_PATH.exists():
        return dict(DEFAULT_CONFIG)

    with CONFIG_PATH.open() as f:
        data = yaml.safe_load(f) or {}

    config = dict(DEFAULT_CONFIG)
    config.update(data)
    return config


def profiles_dir(config: dict[str, Any] | None = None) -> Path:
    """Return the resolved profiles directory path."""
    if config is None:
        config = load_config()
    raw = config.get("profiles_dir", str(ARMIES_DIR / "profiles"))
    return Path(os.path.expanduser(raw))


def malus_ledger_path() -> Path:
    """Return the path to the malus ledger YAML."""
    return ARMIES_DIR / "accountability" / "malus-ledger.yaml"
