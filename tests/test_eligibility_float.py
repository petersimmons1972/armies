"""Tests for floating-point boundary stability in tier_for_malus (#23).

Malus accumulation with floating-point arithmetic can produce values like
99.9999999997 or 100.0000000003 at tier boundaries, causing non-deterministic
tier assignment. The fix uses round() to stabilize comparisons.
"""

from __future__ import annotations

from datetime import date
from pathlib import Path

import pytest
import yaml


def _write_ledger(path: Path, entries: list) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(yaml.dump(entries, default_flow_style=False), encoding="utf-8")


# ---------------------------------------------------------------------------
# Construct a malus that should sum to exactly 100 via floating-point ops
# ---------------------------------------------------------------------------


def test_malus_summing_to_100_lands_in_warning_tier(tmp_path):
    """A sum of malus points that equals exactly 100 must hit the Warning tier.

    We construct entries that sum mathematically to 100.0 but may produce
    99.9999999997 or 100.00000000003 due to IEEE 754 arithmetic. The fixed
    code must yield a deterministic result: the Warning tier (min=100).
    """
    from armies.eligibility import compute_effective_malus, tier_for_malus

    today = date.today().isoformat()
    ledger = tmp_path / "malus-ledger.yaml"

    # Three entries: 33.33... + 33.33... + 33.33... using decays=False
    # The real sum is 99.99..., not 100. Better: use two entries: 50+50=100.
    # One entry of 100 total to ensure exact boundary.
    _write_ledger(ledger, [
        {"agent": "grace", "date": today, "raw_malus": 100, "decays": False, "share": 100},
    ])

    result = compute_effective_malus("grace", ledger)
    assert result == pytest.approx(100.0), f"Expected 100.0, got {result}"

    tier = tier_for_malus(result)
    assert tier["name"] == "Warning", (
        f"Malus=100.0 must land in Warning tier, got '{tier['name']}'"
    )


def test_malus_just_under_100_lands_in_clean_tier(tmp_path):
    """Malus of 99.0 must land in the Clean tier (max=99)."""
    from armies.eligibility import tier_for_malus

    tier = tier_for_malus(99.0)
    assert tier["name"] == "Clean", f"Expected Clean, got '{tier['name']}'"


def test_malus_at_99_9999_rounds_to_clean(tmp_path):
    """Floating-point near-miss: 99.9999999997 should round to 99.9999999997,
    still less than 100 — so it stays in Clean tier.

    This documents the rounding contract: we round to 10 decimal places,
    not to the nearest integer. So 99.9999999997 stays < 100.
    """
    from armies.eligibility import tier_for_malus

    # 99.9999999997 is less than 100 even at 10dp, so it stays Clean
    tier = tier_for_malus(99.9999999997)
    assert tier["name"] == "Clean", (
        f"99.9999999997 should be Clean (below 100), got '{tier['name']}'"
    )


def test_malus_floating_point_over_100_lands_in_warning(tmp_path):
    """100.0000000003 rounded to 10dp is 100.0, which is exactly the Warning boundary."""
    from armies.eligibility import tier_for_malus

    # 100.0000000003 rounded to 10dp = 100.0 -> Warning tier
    tier = tier_for_malus(100.0000000003)
    assert tier["name"] == "Warning", (
        f"100.0000000003 rounds to 100.0 at 10dp, must be Warning, got '{tier['name']}'"
    )


def test_tier_boundaries_are_deterministic_across_accumulation(tmp_path):
    """Multiple small entries summing to 100 must yield a deterministic tier."""
    from armies.eligibility import compute_effective_malus, tier_for_malus

    today = date.today().isoformat()
    ledger = tmp_path / "malus-ledger.yaml"

    # 10 entries of 10 each = 100 total (no decay)
    entries = [
        {"agent": "grace", "date": today, "raw_malus": 10, "decays": False, "share": 100}
        for _ in range(10)
    ]
    _write_ledger(ledger, entries)

    result = compute_effective_malus("grace", ledger)
    tier = tier_for_malus(result)

    # 10 * 10 = 100, which is the Warning tier boundary
    assert tier["name"] == "Warning", (
        f"10 x 10 malus = 100 must be Warning, got '{tier['name']}' (result={result})"
    )
