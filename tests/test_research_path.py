"""Test that cmd_research writes drafts to configured profiles_dir, not cwd — closes #39."""

from __future__ import annotations

from pathlib import Path
from unittest.mock import patch

from click.testing import CliRunner

from armies.cli import cli


def test_cmd_research_uses_configured_dir(tmp_path):
    """Draft must be written to <profiles_dir>/drafts/, not cwd/profiles/drafts/."""
    profiles = tmp_path / "profiles"
    profiles.mkdir()

    config = {
        "profiles_dir": str(profiles),
        "remote_url": "",
        "default_model": "sonnet",
    }

    runner = CliRunner()
    with patch("armies.cli.load_config", return_value=config), \
         patch("armies.cli.profiles_dir_validated", return_value=profiles.resolve()):
        result = runner.invoke(cli, ["research", "implementer"])

    assert result.exit_code == 0, result.output
    drafts_dir = profiles / "drafts"
    assert drafts_dir.exists(), f"Expected {drafts_dir} to exist"
    draft_files = list(drafts_dir.glob("draft-implementer-*.md"))
    assert len(draft_files) == 1, f"Expected one draft file, found: {draft_files}"


def test_cmd_research_does_not_write_to_cwd(tmp_path):
    """Draft must NOT appear in cwd/profiles/drafts/."""
    profiles = tmp_path / "profiles"
    profiles.mkdir()
    cwd_profiles = tmp_path / "cwd" / "profiles" / "drafts"

    config = {
        "profiles_dir": str(profiles),
        "remote_url": "",
        "default_model": "sonnet",
    }

    runner = CliRunner(mix_stderr=False)
    with runner.isolated_filesystem(temp_dir=tmp_path):
        with patch("armies.cli.load_config", return_value=config), \
             patch("armies.cli.profiles_dir_validated", return_value=profiles.resolve()):
            result = runner.invoke(cli, ["research", "coordinator"])

    assert result.exit_code == 0, result.output
    # The cwd-relative path must NOT have been created
    assert not cwd_profiles.exists(), (
        "cmd_research wrote to cwd/profiles/drafts/ — it should use configured profiles_dir"
    )
