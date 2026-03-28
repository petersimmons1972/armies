"""Click CLI entry point for the `armies` command."""

from __future__ import annotations

import importlib.resources
import subprocess
import sys
from datetime import date
from pathlib import Path
from textwrap import dedent

import click
import yaml
from rich.console import Console
from rich.table import Table

from .config import ARMIES_DIR, CONFIG_PATH, load_config, malus_ledger_path, profiles_dir
from .eligibility import KNOWN_ROLES, compute_effective_malus, eligibility_status, tier_for_malus
from .profiles import iter_profile_paths, read_frontmatter, read_frontmatter_and_sections
from .sync import sync_armies

console = Console()


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

STATUS_COLOUR = {
    "CLEAR": "[green]CLEAR[/green]",
    "BLOCKED": "[red]BLOCKED[/red]",
    "FOUNDER": "[yellow]FOUNDER[/yellow]",
    "REVIEW": "[yellow]REVIEW[/yellow]",
    "ESCALATE": "[yellow]ESCALATE[/yellow]",
}

OVERALL_COLOUR = {
    "eligible": "[green]eligible[/green]",
    "restricted": "[yellow]restricted[/yellow]",
    "blocked": "[red]BLOCKED[/red]",
}


def _status_display(status: str) -> str:
    return STATUS_COLOUR.get(status, status)


def _overall_display(overall: str) -> str:
    return OVERALL_COLOUR.get(overall, overall)


# ---------------------------------------------------------------------------
# CLI group
# ---------------------------------------------------------------------------


@click.group()
def cli() -> None:
    """Armies — multi-agent coordination engine for Claude Code."""


# ---------------------------------------------------------------------------
# armies roster
# ---------------------------------------------------------------------------


@cli.command("roster")
def cmd_roster() -> None:
    """Scan profiles and display the agent roster table."""
    config = load_config()
    pdir = profiles_dir(config)
    ledger = malus_ledger_path()

    table = Table(title="Agent Roster", show_lines=True)
    table.add_column("name", style="bold cyan", no_wrap=True)
    table.add_column("display_name")
    table.add_column("primary_role")
    table.add_column("xp", justify="right")
    table.add_column("rank")
    table.add_column("eligibility_status")

    found = False
    for path in iter_profile_paths(pdir):
        found = True
        fm = read_frontmatter(path)

        name = fm.get("name", path.stem)
        display_name = fm.get("display_name", fm.get("name", path.stem))
        roles = fm.get("roles", {})
        primary_role = fm.get("primary_role") or (roles.get("primary") if isinstance(roles, dict) else None) or "—"
        xp = str(fm.get("xp", "—"))
        rank = fm.get("rank", "—")

        # Eligibility
        if ledger.exists():
            info = eligibility_status(str(name), ledger)
            elig = _overall_display(info["overall"])
        else:
            elig = "[green]eligible[/green]"

        table.add_row(str(name), str(display_name), str(primary_role), xp, str(rank), elig)

    if not found:
        console.print(f"[yellow]No profiles found in {pdir}[/yellow]")
        console.print("Run [bold]armies init[/bold] to create the directory structure.")
        return

    console.print(table)


# ---------------------------------------------------------------------------
# armies spawn
# ---------------------------------------------------------------------------


@cli.command("spawn")
@click.argument("agent")
@click.option("--role", required=True, help="Role block to extract (e.g. 'implementer')")
def cmd_spawn(agent: str, role: str) -> None:
    """Read a profile and output frontmatter + Base Persona + one role block."""
    config = load_config()
    pdir = profiles_dir(config)

    # Locate profile file
    profile_path = pdir / f"{agent}.md"
    if not profile_path.exists():
        # Try case-insensitive search
        matches = [p for p in pdir.glob("*.md") if p.stem.lower() == agent.lower()]
        if not matches:
            console.print(f"[red]Profile not found:[/red] {agent}")
            console.print(f"Searched in: {pdir}")
            sys.exit(1)
        profile_path = matches[0]

    role_heading = f"Role: {role}"
    sections_wanted = ["Base Persona", role_heading]

    fm, sections = read_frontmatter_and_sections(profile_path, sections_wanted)

    if role_heading not in sections:
        console.print(
            f"[red]Role block '## {role_heading}' not found in {profile_path.name}[/red]"
        )
        available = _list_role_headings(profile_path)
        if available:
            console.print("Available role blocks:")
            for h in available:
                console.print(f"  {h}")
        sys.exit(1)

    # Build output: frontmatter + Base Persona + role block
    lines: list[str] = ["---"]
    lines.append(yaml.dump(fm, default_flow_style=False).rstrip())
    lines.append("---")
    lines.append("")

    if "Base Persona" in sections:
        lines.append("## Base Persona")
        lines.append("")
        lines.append(sections["Base Persona"])
        lines.append("")

    lines.append(f"## {role_heading}")
    lines.append("")
    lines.append(sections[role_heading])
    lines.append("")

    click.echo("\n".join(lines))


def _list_role_headings(path: Path) -> list[str]:
    """Return all '## Role: ...' headings found in a profile file."""
    import re

    results = []
    with path.open(encoding="utf-8") as fh:
        for line in fh:
            m = re.match(r"^## (Role: .+)$", line.strip())
            if m:
                results.append(m.group(1))
    return results


# ---------------------------------------------------------------------------
# armies eligible
# ---------------------------------------------------------------------------


@cli.command("eligible")
@click.argument("agent")
def cmd_eligible(agent: str) -> None:
    """Compute and display spawn eligibility for an agent."""
    ledger = malus_ledger_path()

    effective = compute_effective_malus(agent, ledger)
    tier = tier_for_malus(effective)

    console.print(f"\n[bold]Agent:[/bold] {agent}")
    console.print(f"[bold]Effective malus:[/bold] {effective:.1f}")
    console.print(f"[bold]Tier:[/bold] {tier['name']}")

    if not ledger.exists():
        console.print(
            f"\n[yellow]Note:[/yellow] Ledger not found at {ledger}. "
            "Showing gates for zero malus."
        )

    table = Table(title="Role Eligibility", show_lines=False)
    table.add_column("Role", style="bold")
    table.add_column("Status")

    for role in KNOWN_ROLES:
        status = tier[role]
        table.add_row(role.replace("_", " "), _status_display(status))

    console.print(table)


# ---------------------------------------------------------------------------
# armies sync
# ---------------------------------------------------------------------------


@cli.command("sync")
def cmd_sync() -> None:
    """Pull then push the ~/.armies git repository."""
    config = load_config()
    result = sync_armies(config)

    if result.get("error"):
        console.print(f"[red]Error:[/red] {result['error']}")
        sys.exit(1)

    # Pull
    pull_icon = "[green]✓[/green]" if result["pull_ok"] else "[red]✗[/red]"
    console.print(f"{pull_icon} pull: {result['pull_msg'] or '(no output)'}")

    # Push
    push_icon = "[green]✓[/green]" if result["push_ok"] else "[red]✗[/red]"
    console.print(f"{push_icon} push: {result['push_msg'] or '(no output)'}")

    if not result["pull_ok"] or not result["push_ok"]:
        sys.exit(1)


# ---------------------------------------------------------------------------
# armies init
# ---------------------------------------------------------------------------


@cli.command("init")
def cmd_init() -> None:
    """Create the ~/.armies/ directory structure."""
    subdirs = ["profiles", "accountability", "service-records", "teams"]

    # Create dirs
    for sub in subdirs:
        d = ARMIES_DIR / sub
        d.mkdir(parents=True, exist_ok=True)
        console.print(f"[green]✓[/green] {d}")

    # Write config.yaml if missing
    if not CONFIG_PATH.exists():
        remote_url = click.prompt(
            "GitHub remote URL for your private profiles repo (leave blank to skip sync setup)",
            default="",
            show_default=False,
        ).strip()

        config_data = {
            "remote_url": remote_url,
            "default_model": "sonnet",
            "profiles_dir": str(ARMIES_DIR / "profiles"),
        }
        with CONFIG_PATH.open("w", encoding="utf-8") as fh:
            yaml.dump(config_data, fh, default_flow_style=False)
        console.print(f"[green]✓[/green] {CONFIG_PATH}")

        if remote_url:
            _init_git(remote_url)
    else:
        console.print(f"[dim]config.yaml already exists — skipping prompt[/dim]")
        config = load_config()
        remote_url = config.get("remote_url", "").strip()
        if remote_url:
            _init_git(remote_url)

    console.print("\n[bold green]Done.[/bold green] ~/.armies/ is ready.")


def _init_git(remote_url: str) -> None:
    """Run git init and set remote origin in ~/.armies/."""
    result = subprocess.run(
        ["git", "init", str(ARMIES_DIR)],
        capture_output=True,
        text=True,
    )
    if result.returncode == 0:
        console.print(f"[green]✓[/green] git init {ARMIES_DIR}")
    else:
        console.print(f"[yellow]git init:[/yellow] {result.stderr.strip()}")

    # Set or update remote
    check = subprocess.run(
        ["git", "-C", str(ARMIES_DIR), "remote", "get-url", "origin"],
        capture_output=True,
        text=True,
    )
    if check.returncode == 0:
        subprocess.run(
            ["git", "-C", str(ARMIES_DIR), "remote", "set-url", "origin", remote_url],
            capture_output=True,
        )
        console.print(f"[green]✓[/green] remote origin updated to {remote_url}")
    else:
        subprocess.run(
            ["git", "-C", str(ARMIES_DIR), "remote", "add", "origin", remote_url],
            capture_output=True,
        )
        console.print(f"[green]✓[/green] remote origin set to {remote_url}")


# ---------------------------------------------------------------------------
# armies record
# ---------------------------------------------------------------------------


@cli.command("record")
@click.argument("agent")
@click.argument("note")
@click.option("--xp", default=0, type=int, show_default=True, help="XP earned this deployment.")
@click.option("--outcome", default="success", type=click.Choice(["success", "partial", "failure"]), show_default=True)
def cmd_record(agent: str, note: str, xp: int, outcome: str) -> None:
    """Write a service record entry and update XP in the profile."""
    config = load_config()
    pdir = profiles_dir(config)
    profile_path = pdir / f"{agent}.md"

    if not profile_path.exists():
        matches = [p for p in pdir.glob("*.md") if p.stem.lower() == agent.lower()]
        if not matches:
            console.print(f"[red]Profile not found:[/red] {agent}")
            sys.exit(1)
        profile_path = matches[0]

    # Read current profile text and frontmatter
    raw = profile_path.read_text(encoding="utf-8")
    fm = read_frontmatter(profile_path)
    current_xp = int(fm.get("xp", 0))
    new_xp = current_xp + xp

    # Rewrite xp in the profile frontmatter
    import re
    raw = re.sub(r"^xp:\s*\d+", f"xp: {new_xp}", raw, flags=re.MULTILINE)
    profile_path.write_text(raw, encoding="utf-8")

    # Append service record entry
    service_records_dir = ARMIES_DIR / "service-records"
    service_records_dir.mkdir(parents=True, exist_ok=True)
    record_path = service_records_dir / f"{agent}.yaml"

    existing: list = []
    if record_path.exists():
        existing = yaml.safe_load(record_path.read_text(encoding="utf-8")) or []

    entry = {
        "date": date.today().isoformat(),
        "task": note,
        "outcome": outcome,
        "xp_earned": xp,
        "xp_total": new_xp,
    }
    existing.append(entry)
    record_path.write_text(yaml.dump(existing, default_flow_style=False), encoding="utf-8")

    console.print(f"[green]✓[/green] Service record written: {record_path.name}")
    console.print(f"[green]✓[/green] XP updated: {current_xp} → {new_xp}")
    console.print(f"[dim]Commit the profile to make it permanent:[/dim] git -C ~/.armies commit -am 'record({agent}): {note}'")


# ---------------------------------------------------------------------------
# armies seed
# ---------------------------------------------------------------------------


@cli.command("seed")
@click.option("--force", is_flag=True, default=False, help="Overwrite existing profiles.")
def cmd_seed(force: bool) -> None:
    """Install the bundled general profiles into ~/.armies/profiles/."""
    config = load_config()
    pdir = profiles_dir(config)
    pdir.mkdir(parents=True, exist_ok=True)

    pkg = importlib.resources.files("armies") / "examples" / "generals"
    installed = 0
    skipped = 0

    for resource in pkg.iterdir():
        if not resource.name.endswith(".md"):
            continue
        dest = pdir / resource.name
        if dest.exists() and not force:
            console.print(f"[dim]skip[/dim]  {resource.name} (already exists — use --force to overwrite)")
            skipped += 1
            continue
        dest.write_text(resource.read_text(encoding="utf-8"), encoding="utf-8")
        console.print(f"[green]✓[/green]     {resource.name}")
        installed += 1

    console.print(f"\n{installed} installed, {skipped} skipped.")
    if installed:
        console.print("Run [bold]armies roster[/bold] to see them.")


# ---------------------------------------------------------------------------
# armies research
# ---------------------------------------------------------------------------


@cli.command("research")
@click.argument("role")
@click.option(
    "--mode",
    default="prompt",
    type=click.Choice(["prompt", "api"]),
    show_default=True,
    help="Output mode: prompt (default) or api (stub).",
)
def cmd_research(role: str, mode: str) -> None:
    """Generate a structured agent prompt for a given role class."""
    if mode == "api":
        console.print(
            "[yellow]API mode not yet implemented. Default prompt mode used.[/yellow]"
        )

    today = date.today().isoformat()
    draft_dir = Path.cwd() / "profiles" / "drafts"
    draft_dir.mkdir(parents=True, exist_ok=True)
    draft_filename = f"draft-{role}-{today}.md"
    draft_path = draft_dir / draft_filename

    profile_schema_path = "~/projects/armies/schema/profile-schema.yaml"

    prompt_content = dedent(f"""\
        # Research Prompt: {role} Agent Profile
        Generated: {today}

        ## Task

        You are a Claude Code agent. Your task is to research and draft a complete
        agent profile for the role class `{role}`.

        ## Step 1 — Research

        Research 3–5 historical figures whose personality traits, working style,
        and domain expertise naturally fit the `{role}` role class in a software
        engineering context.

        For each candidate, note:
        - Name and brief bio (2–3 sentences)
        - Why their traits map to the `{role}` role
        - Potential weaknesses or malus risk factors
        - Relevance to modern software/AI workflows

        ## Step 2 — Select the Best Candidate

        Choose the single best candidate. Explain your selection rationale in
        2–3 sentences covering: trait fit, uniqueness (not already in the roster),
        and practical value for software coordination tasks.

        ## Step 3 — Draft the Profile

        Using the selected historical figure, draft a complete agent profile
        following the schema at `{profile_schema_path}`.

        The profile MUST include:
        - YAML frontmatter with: name, display_name, primary_role, xp (start at 0),
          rank, archetype, specialties (list), tool_access (list), memory (project field)
        - `## Base Persona` section: voice, decision style, working approach (~200 words)
        - `## Role: {role}` section: specific instructions for this role (~150 words)
        - `## Service Record` section: empty table with headers

        ## Step 4 — Save

        Save the completed profile to:

            ~/.armies/profiles/<name>.md

        where `<name>` is the agent's lowercase hyphenated identifier
        (e.g., `florence-nightingale` → `florence-nightingale.md`).

        ## Constraints

        - Do NOT invent facts about the historical figure — use only well-documented traits.
        - Do NOT re-use figures already in ~/.armies/profiles/.
        - The profile must pass `armies roster` without errors after saving.
        - Commit the new profile: `git -C ~/.armies commit -am "profile(<name>): initial draft for {role} role"`
    """)

    draft_path.write_text(prompt_content, encoding="utf-8")

    relative = f"profiles/drafts/{draft_filename}"
    console.print(f"Draft prompt saved to {relative}")
    console.print(
        "Feed this file to a Claude Code agent using the Agent tool "
        "to generate a complete profile."
    )
