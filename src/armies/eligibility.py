"""Malus computation and spawn gate logic.

Formula for each ledger entry allocated to a general:
    if decays:
        effective = raw_malus * (share/100) * (0.5 ** (days_since / 14))
    else:
        effective = raw_malus * (share/100)

Total effective malus = sum of all per-entry values.
"""

from __future__ import annotations

import logging
import math
from datetime import date, datetime
from pathlib import Path
from typing import Any

import yaml

log = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Tier gate table — mirrors rules/eligibility.yaml
# ---------------------------------------------------------------------------

TIERS: list[dict[str, Any]] = [
    {
        "min": 0,
        "max": 99,
        "name": "Clean",
        "coordinator": "CLEAR",
        "emergency_reserve": "CLEAR",
        "specialist": "CLEAR",
        "validator": "CLEAR",
    },
    {
        "min": 100,
        "max": 199,
        "name": "Warning",
        "coordinator": "BLOCKED",
        "emergency_reserve": "FOUNDER",
        "specialist": "CLEAR",
        "validator": "CLEAR",
    },
    {
        "min": 200,
        "max": 299,
        "name": "Probation",
        "coordinator": "BLOCKED",
        "emergency_reserve": "BLOCKED",
        "specialist": "REVIEW",
        "validator": "CLEAR",
    },
    {
        "min": 300,
        "max": 399,
        "name": "Demotion risk",
        "coordinator": "BLOCKED",
        "emergency_reserve": "BLOCKED",
        "specialist": "ESCALATE",
        "validator": "CLEAR",
    },
    {
        "min": 400,
        "max": math.inf,
        "name": "Suspension",
        "coordinator": "BLOCKED",
        "emergency_reserve": "BLOCKED",
        "specialist": "BLOCKED",
        "validator": "BLOCKED",
    },
]

KNOWN_ROLES = ("coordinator", "emergency_reserve", "specialist", "validator")


def _parse_date(value: Any) -> date:
    """Parse a date value that may be a string or datetime.date."""
    if isinstance(value, date):
        return value
    if isinstance(value, datetime):
        return value.date()
    # Try ISO string
    return date.fromisoformat(str(value).strip())


def compute_effective_malus(agent_name: str, ledger_path: Path) -> float:
    """Return the total effective malus for *agent_name* from *ledger_path*.

    If the ledger file does not exist, returns 0.0.
    The agent name comparison is case-insensitive.
    """
    if not ledger_path.exists():
        return 0.0

    with ledger_path.open(encoding="utf-8") as fh:
        raw = yaml.safe_load(fh)

    if not isinstance(raw, list):
        return 0.0

    today = date.today()
    total = 0.0
    target = agent_name.strip().lower()

    for entry in raw:
        if not isinstance(entry, dict):
            continue

        raw_malus = float(entry.get("raw_malus", 0))
        decays = bool(entry.get("decays", True))

        # Support both `allocation` list and flat `agent`/`share` fields
        allocations: list[dict[str, Any]] = []

        if "allocation" in entry and isinstance(entry["allocation"], list):
            allocations = entry["allocation"]
        elif "agent" in entry:
            # Flat format: agent + optional share (default 100)
            allocations = [
                {
                    "agent": entry["agent"],
                    "share": entry.get("share", 100),
                }
            ]

        for alloc in allocations:
            if not isinstance(alloc, dict):
                continue
            alloc_agent = str(alloc.get("agent", "")).strip().lower()
            if alloc_agent != target:
                continue

            # Clamp share to [0, 100] — values outside this range are nonsensical
            # and can produce malus contributions larger than raw_malus (#31).
            share = max(0.0, min(100.0, float(alloc.get("share", 100))))

            if decays:
                entry_date_raw = entry.get("date")
                try:
                    entry_date = _parse_date(entry_date_raw)
                    # Clamp to 0 — future-dated entries decay at full strength, not
                    # negative (which would amplify malus instead of decaying it) (#30).
                    days_since = max(0, (today - entry_date).days)
                except (TypeError, ValueError):
                    # Log the bad entry so operators can find and fix it (#21, #40).
                    log.warning(
                        "Malus entry has unparseable date, treating as today: %r", entry
                    )
                    days_since = 0
                contribution = raw_malus * (share / 100) * (0.5 ** (days_since / 14))
            else:
                contribution = raw_malus * (share / 100)

            total += contribution

    return total


def tier_for_malus(effective_malus: float) -> dict[str, Any]:
    """Return the tier dict from TIERS for the given effective malus value."""
    for tier in TIERS:
        if tier["min"] <= effective_malus <= tier["max"]:
            return tier
    # Fallback — should not happen given the last tier has max=inf
    return TIERS[-1]


def eligibility_status(agent_name: str, ledger_path: Path) -> dict[str, Any]:
    """Compute and return full eligibility info for an agent.

    Returns a dict with keys:
        effective_malus  — float
        tier             — tier name string
        roles            — dict[role_name -> status string]
        overall          — "eligible" | "restricted" | "blocked"
    """
    malus = compute_effective_malus(agent_name, ledger_path)
    tier = tier_for_malus(malus)

    roles = {role: tier[role] for role in KNOWN_ROLES}

    # Overall summary
    statuses = set(roles.values())
    if all(s == "CLEAR" for s in statuses):
        overall = "eligible"
    elif "BLOCKED" in statuses and all(
        s == "BLOCKED" for s in statuses
    ):
        overall = "blocked"
    else:
        overall = "restricted"

    return {
        "effective_malus": malus,
        "tier": tier["name"],
        "roles": roles,
        "overall": overall,
    }
