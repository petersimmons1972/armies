"""Tests for eligibility edge cases — closes #30, #31, #21, #40."""

from __future__ import annotations

import logging
from datetime import date, timedelta
from pathlib import Path

import pytest
import yaml


def _write_ledger(path: Path, entries: list) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(yaml.dump(entries, default_flow_style=False), encoding="utf-8")


# ---------------------------------------------------------------------------
# #30 — Future-dated entries must produce days_since = 0, not negative
# ---------------------------------------------------------------------------


def test_future_date_days_since_not_negative(tmp_path):
    """Entry dated in the future must yield effective malus >= 0, not blow up."""
    from armies.eligibility import compute_effective_malus

    future = (date.today() + timedelta(days=30)).isoformat()
    ledger = tmp_path / "malus-ledger.yaml"
    _write_ledger(ledger, [
        {"agent": "grace", "date": future, "raw_malus": 50, "decays": True, "share": 100}
    ])
    result = compute_effective_malus("grace", ledger)
    # With days_since clamped to 0: 50 * 1.0 * (0.5**0) = 50.0
    assert result >= 0.0
    # Specifically: days_since=0 → contribution = 50.0
    assert result == pytest.approx(50.0)


# ---------------------------------------------------------------------------
# #31 — share > 100 must be clamped
# ---------------------------------------------------------------------------


def test_share_over_100_clamped(tmp_path):
    """share=150 must be treated as 100, not produce a malus over raw_malus."""
    from armies.eligibility import compute_effective_malus

    today = date.today().isoformat()
    ledger = tmp_path / "malus-ledger.yaml"
    _write_ledger(ledger, [
        {"agent": "grace", "date": today, "raw_malus": 100, "decays": False, "share": 150}
    ])
    result = compute_effective_malus("grace", ledger)
    # share clamped to 100 → contribution = 100 * (100/100) = 100.0
    assert result == pytest.approx(100.0)


def test_share_negative_clamped(tmp_path):
    """share=-10 must be clamped to 0, yielding zero contribution."""
    from armies.eligibility import compute_effective_malus

    today = date.today().isoformat()
    ledger = tmp_path / "malus-ledger.yaml"
    _write_ledger(ledger, [
        {"agent": "grace", "date": today, "raw_malus": 100, "decays": False, "share": -10}
    ])
    result = compute_effective_malus("grace", ledger)
    assert result == pytest.approx(0.0)


# ---------------------------------------------------------------------------
# #21 / #40 — Missing or unparseable date must log warning, not silently use 0
# ---------------------------------------------------------------------------


def test_missing_date_logs_warning(tmp_path, caplog):
    """Entry with no date field must log a warning and fall back to today."""
    from armies.eligibility import compute_effective_malus

    ledger = tmp_path / "malus-ledger.yaml"
    _write_ledger(ledger, [
        {"agent": "grace", "raw_malus": 50, "decays": True, "share": 100}
        # 'date' key intentionally absent
    ])
    with caplog.at_level(logging.WARNING, logger="armies.eligibility"):
        result = compute_effective_malus("grace", ledger)

    assert result >= 0.0
    assert any("unparseable" in r.message.lower() or "date" in r.message.lower()
               for r in caplog.records), (
        f"Expected a warning about the missing date, got: {[r.message for r in caplog.records]}"
    )


def test_bad_date_string_logs_warning(tmp_path, caplog):
    """Entry with an invalid date string must log a warning and not crash."""
    from armies.eligibility import compute_effective_malus

    ledger = tmp_path / "malus-ledger.yaml"
    _write_ledger(ledger, [
        {"agent": "grace", "date": "not-a-date", "raw_malus": 50, "decays": True, "share": 100}
    ])
    with caplog.at_level(logging.WARNING, logger="armies.eligibility"):
        result = compute_effective_malus("grace", ledger)

    assert result >= 0.0
    assert any("unparseable" in r.message.lower() or "date" in r.message.lower()
               for r in caplog.records)
