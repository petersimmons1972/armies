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

from .config import ARMIES_DIR, CONFIG_PATH, load_config, malus_ledger_path, profiles_dir, profiles_dir_validated
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


def _resolve_agent_path(profiles_directory: Path, agent: str) -> Path:
    """Resolve agent name to a profile path, rejecting path traversal.

    Raises ValueError if the resolved path escapes profiles_directory.
    This guards against agent='../../etc/passwd' style attacks (issue #26).
    """
    base = profiles_directory.resolve()
    # Build the candidate path — strip any leading slashes to prevent absolute
    # path injection, then resolve relative to the profiles directory.
    candidate = (base / f"{agent}.md").resolve()
    try:
        candidate.relative_to(base)
    except ValueError:
        raise ValueError(
            f"Agent argument '{agent}' resolves outside the profiles directory "
            f"'{base}'. Refusing to open '{candidate}'."
        )
    return candidate


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
    try:
        pdir = profiles_dir_validated(config)
    except ValueError as exc:
        console.print(f"[red]Configuration error:[/red] {exc}")
        sys.exit(1)

    # Locate profile file — validate agent path to prevent traversal (#26)
    try:
        profile_path = _resolve_agent_path(pdir, agent)
    except ValueError as exc:
        console.print(f"[red]Security error:[/red] {exc}")
        sys.exit(1)

    if not profile_path.exists():
        # Try case-insensitive search (within safe directory)
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
    lines.append(yaml.safe_dump(fm, default_flow_style=False).rstrip())
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


def _update_frontmatter_field(path: Path, key: str, value) -> None:
    """Safely update a single frontmatter field without touching the body.

    Reads the file, isolates the frontmatter block (text between the first
    and second '---' delimiters), parses it with yaml.safe_load, sets the
    key, re-serializes with yaml.safe_dump, and writes back — leaving the
    body unchanged.

    This replaces the previous re.sub approach which would corrupt any
    occurrence of 'key: ...' that appeared in the body text (issues #20,
    #32, #37).
    """
    raw = path.read_text(encoding="utf-8")

    # A well-formed profile starts with '---\n' and has a closing '---' line.
    # Split on '---' at most twice to isolate: ['', frontmatter, body].
    parts = raw.split("---", 2)
    if len(parts) < 3:
        # Malformed file — fall back to a simple write to avoid data loss
        fm = yaml.safe_load(parts[1]) if len(parts) >= 2 else {}
        if not isinstance(fm, dict):
            fm = {}
        fm[key] = value
        path.write_text(
            "---\n" + yaml.safe_dump(fm, default_flow_style=False) + "---\n",
            encoding="utf-8",
        )
        return

    # parts[0] is the empty string before the first '---'
    # parts[1] is the frontmatter YAML text
    # parts[2] is everything after the closing '---'
    fm = yaml.safe_load(parts[1])
    if not isinstance(fm, dict):
        fm = {}
    fm[key] = value

    new_fm_text = yaml.safe_dump(fm, default_flow_style=False, allow_unicode=True)
    new_raw = parts[0] + "---\n" + new_fm_text + "---" + parts[2]
    path.write_text(new_raw, encoding="utf-8")


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
            yaml.safe_dump(config_data, fh, default_flow_style=False)
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
    try:
        pdir = profiles_dir_validated(config)
    except ValueError as exc:
        console.print(f"[red]Configuration error:[/red] {exc}")
        sys.exit(1)

    try:
        profile_path = _resolve_agent_path(pdir, agent)
    except ValueError as exc:
        console.print(f"[red]Security error:[/red] {exc}")
        sys.exit(1)

    if not profile_path.exists():
        matches = [p for p in pdir.glob("*.md") if p.stem.lower() == agent.lower()]
        if not matches:
            console.print(f"[red]Profile not found:[/red] {agent}")
            sys.exit(1)
        profile_path = matches[0]

    # Read current frontmatter and compute new XP
    fm = read_frontmatter(profile_path)
    current_xp = int(fm.get("xp", 0))
    new_xp = current_xp + xp

    # Update xp in the frontmatter without touching the body.  The old
    # re.sub approach matched the first `xp:` anywhere in the file,
    # corrupting role descriptions that contained "xp: N" in their text
    # (issues #20, #32, #37).
    _update_frontmatter_field(profile_path, "xp", new_xp)

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
    record_path.write_text(yaml.safe_dump(existing, default_flow_style=False), encoding="utf-8")

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

    config = load_config()
    try:
        pdir = profiles_dir_validated(config)
    except ValueError as exc:
        console.print(f"[red]Configuration error:[/red] {exc}")
        sys.exit(1)

    today = date.today().isoformat()
    # Write drafts under the configured profiles directory, not the current
    # working directory — cwd is unpredictable in CLI usage (issue #39).
    draft_dir = pdir / "drafts"
    draft_dir.mkdir(parents=True, exist_ok=True)
    draft_filename = f"draft-{role}-{today}.md"
    draft_path = draft_dir / draft_filename

    profile_schema_path = "~/projects/armies/schema/profile-schema.yaml"

    prompt_content = dedent(f"""\
        # Research Prompt: {role} Agent Profile
        Generated: {today}

        ## What You Are Building

        An activation profile — not a biography, not a summary, not a list of achievements.
        The model you are running on already knows every well-documented historical figure
        deeply. Your job is not to teach it who the person is. Your job is to find the
        specific behavioral pointers that unlock the right slice of that knowledge and focus
        it on the `{role}` role.

        A profile that describes achievements produces a generic agent.
        A profile that describes how someone moved, decided, failed, and recovered
        produces a useful one.

        ## Step 1 — Select the Figure

        Research 3–5 real historical figures who naturally fit the `{role}` role class.

        Selection criteria — in order of importance:
        1. **Documentation depth**: Has history argued about this person for decades?
           Are there multiple biographies that disagree with each other? Primary sources
           (letters, memoirs, contemporaries' accounts)? The richer the record, the
           stronger the activation. Avoid figures known primarily through one source.
        2. **Behavioral specificity**: Do we know HOW they worked, not just WHAT they
           achieved? How they ran a meeting. How they delivered bad news. How they made
           decisions under uncertainty. How they failed specifically.
        3. **Role fit**: Do their documented working patterns naturally map to `{role}`
           behaviors — not just their job title or reputation?

        For each candidate note:
        - Why the documentation record is deep enough to activate well
        - One specific behavioral detail (not an achievement) that maps to `{role}`
        - Their specific, documented failure mode — not a generic weakness

        ## Step 2 — Select the Best Candidate

        Choose the single best candidate. Prioritize documentation depth over fame.
        A less famous figure with a rich behavioral record outperforms a famous figure
        with a thin one. Explain why this figure's documented behavioral patterns
        make them the strongest activation key for the `{role}` role.

        ## Step 3 — Research the Second Layer

        Before writing the profile, do the deep research. Do not write from memory.
        Search for:

        - How they actually ran a room — specific documented behaviors, not general reputation
        - The decisions they made under pressure that reveal character
        - Contemporaries' accounts — what did people who worked with them say?
        - The specific failure that cost them something real and documented
        - The thing they did that surprised people who expected something different
        - Anything that contradicts the surface reputation

        The surface reputation is what everyone already knows. The model already has it.
        You are looking for the second layer — the behavioral detail that is documented
        but not famous. That is what makes the profile work.

        ## Step 4 — Write the Profile

        Use the following format:

        ```
        ---
        name: [kebab-case]
        display_name: "[Full Title and Name]"
        description: >
          [3-4 sentences — behavioral description for THIS role specifically.
           Not achievements. Not reputation. How they operate and when to use them.]
        roles:
          primary: {role}
          secondary: [if applicable]
        xp: 0
        rank: "[Historical rank or title]"
        model: [opus for coordinator/planner/researcher; sonnet for implementer/troubleshooter]
        [disallowedTools: — only if coordinator role]
        [  - Write]
        [  - Edit]
        [  - Bash]
        ---

        ## Base Persona

        [300-400 words of behavioral prose. Write in second person ("You are...").
         Include:
         - Formation: what made them who they are. Specific, not generic.
         - How they actually work: the specific behaviors that distinguish them
           from a type. Not "decisive" but what decisive looked like for this person.
         - A specific relationship or training experience that shaped their method.
         - **Named failure mode**: one specific, documented failure with real
           consequences — not a character flaw, a real thing that happened.
           This is load-bearing. It creates accountability and makes the agent
           feel real rather than oracular.
         - One behavioral detail that contradicts or complicates the surface reputation.]

        ## Role: {role}

        [150-200 words of operational instructions for this specific role.
         Pre-mission checklist. How they work. What they deliver. What "done" looks like.
         These are not generic role instructions — they are how THIS person approaches
         THIS role based on their documented working patterns.]
        ```

        ## Step 5 — Write the Behavioral Fingerprints

        After the profile body, add a `test_scenarios` block to the frontmatter.
        This is how you verify the profile is activating the right person, not
        producing a generic agent with the correct name.

        Write exactly 3 scenarios using these archetypes — every profile gets all three:

        1. **ambiguous-order**: A task with a missing constraint. Watch how they seek clarity.
        2. **pressure-test**: A deadline compressed mid-campaign. Watch how they push back.
        3. **scope-creep-trap**: An out-of-scope request added mid-campaign. Watch the negotiation.

        For each scenario, write 2 fingerprint criteria. Each criterion must:
        - Describe a specific behavior you would expect from THIS person that a generic
          `{role}` agent would never produce
        - Include a `why` field explaining what the generic version looks like and why
          this person's documented behavior differs — cite the specific research that
          supports it (a relationship, an incident, a documented habit)

        Format:

        ```yaml
        test_scenarios:
          - id: ambiguous-order
            situation: >
              [2-3 sentences describing the situation. Make it realistic and role-appropriate.]
            prompt: "[The single question or instruction the agent must respond to.]"
            fingerprints:
              - criterion: [Specific behavior expected — one sentence, observable]
                why: >
                  [What the generic version looks like. Why this person's documented
                   history produces a different response. Cite the specific source —
                   an incident, a relationship, a documented working habit.]
              - criterion: [Second behavior]
                why: >
                  [Same structure.]
          - id: pressure-test
            [same structure]
          - id: scope-creep-trap
            [same structure]
        ```

        The `why` field is load-bearing. Without it, the rubric is just a checklist.
        With it, the person scoring the test knows exactly what they are listening for
        and why a generic response fails.

        Bad fingerprint (too generic):
          criterion: "Asks clarifying questions before proceeding"
          why: "Good coordinators ask questions."

        Good fingerprint (specific to person):
          criterion: "Names the missing constraint before issuing any assignments"
          why: >
            "A generic coordinator assumes or asks vaguely. Eisenhower's documented habit —
             from his Abilene poker education through every command — was to write down
             what he did not know before committing. He would not brief specialists on an
             ambiguous order. If the response assigns work without naming the gap, this fails."

        ## Step 6 — Verify Before Saving

        Read the Base Persona back. Ask:
        - Does this feel like a specific person or a type?
        - Could this description apply to three other people in the same role? If yes, it is too generic.
        - Is the failure mode a real documented event with real consequences? Or a character note?
        - Does the description tell you HOW they moved, or just WHAT they achieved?

        Read the fingerprints. Ask:
        - Would a generic {role} agent produce this behavior? If yes, the fingerprint is too weak.
        - Does the `why` field cite a specific documented behavior, or does it just explain the criterion?
        - Could you score a response against this criterion, or is it too vague to judge?

        If the answers are wrong, research more before saving.

        ## Step 7 — Save and Commit

        Save the completed profile to:

            ~/.armies/profiles/<name>.md

        where `<name>` is the agent's lowercase hyphenated identifier.

        Verify it works: `armies test <name>` should print the full test document without errors.

        Commit: `git -C ~/.armies commit -am "profile(<name>): {role} role — [figure name]"`

        ## Hard Constraints

        - Real historical figures only. No fictional characters — they lack the
          multi-source documentation depth that produces strong activation.
        - Do NOT write from the Wikipedia lede. That is the surface. Go deeper.
        - Do NOT re-use figures already in ~/.armies/profiles/.
        - Do NOT start at 0 words on the failure mode. Every profile needs one.
        - Do NOT skip test_scenarios. Every profile must pass `armies test` on creation.
        - The profile must pass `armies roster` without errors after saving.
    """)

    draft_path.write_text(prompt_content, encoding="utf-8")

    relative = f"profiles/drafts/{draft_filename}"
    console.print(f"Draft prompt saved to {relative}")
    console.print(
        "Feed this file to a Claude Code agent using the Agent tool "
        "to generate a complete profile."
    )


# ---------------------------------------------------------------------------
# armies test
# ---------------------------------------------------------------------------


@cli.command("test")
@click.argument("agent")
def cmd_test(agent: str) -> None:
    """Generate a behavioral fingerprint test prompt for a profile.

    Prints a single markdown document to stdout. Paste it into a new Claude
    Code conversation, read the agent's response, and score it against the
    rubric to verify the profile is activating the right person.
    """
    config = load_config()
    try:
        pdir = profiles_dir_validated(config)
    except ValueError as exc:
        console.print(f"[red]Configuration error:[/red] {exc}")
        sys.exit(1)

    # Locate profile — validate to prevent traversal (#26)
    try:
        profile_path = _resolve_agent_path(pdir, agent)
    except ValueError as exc:
        console.print(f"[red]Security error:[/red] {exc}")
        sys.exit(1)

    if not profile_path.exists():
        matches = [p for p in pdir.glob("*.md") if p.stem.lower() == agent.lower()]
        if not matches:
            console.print(f"[red]Profile not found:[/red] {agent}")
            console.print(f"Searched in: {pdir}")
            sys.exit(1)
        profile_path = matches[0]

    fm = read_frontmatter(profile_path)
    scenarios = fm.get("test_scenarios")

    if not scenarios:
        name = fm.get("name", agent)
        console.print(f"[red]No test scenarios defined in {profile_path.name}[/red]")
        console.print(
            f"\nAdd a [bold]test_scenarios[/bold] block to the frontmatter of [bold]{name}[/bold].\n"
            "See docs/plans/2026-03-28-armies-test-design.md for the schema."
        )
        sys.exit(1)

    # Get primary role for spawn block
    roles = fm.get("roles", {})
    primary_role = (
        fm.get("primary_role")
        or (roles.get("primary") if isinstance(roles, dict) else None)
        or "coordinator"
    )
    role_heading = f"Role: {primary_role}"
    fm_sections, sections = read_frontmatter_and_sections(
        profile_path, ["Base Persona", role_heading]
    )

    display_name = fm.get("display_name", fm.get("name", agent))

    # Build the output document
    out: list[str] = []

    # --- Header ---
    out.append(f"# Behavioral Fingerprint Test — {display_name}")
    out.append("")
    out.append(
        "Paste this entire document into a new Claude Code conversation. "
        "Read the agent's response to each scenario, then score it against "
        "the rubric below each one."
    )
    out.append("")
    out.append("---")
    out.append("")

    # --- Spawn block ---
    # Strip test_scenarios from the spawn context — agents don't need to see the rubric
    spawn_fm = {k: v for k, v in fm_sections.items() if k != "test_scenarios"}

    out.append("## Agent Context")
    out.append("")
    out.append(
        "The following profile is loaded for this session. "
        "You are this person for the duration of this conversation."
    )
    out.append("")
    out.append("```")
    out.append("---")
    out.append(yaml.safe_dump(spawn_fm, default_flow_style=False, allow_unicode=True).rstrip())
    out.append("---")
    out.append("")
    if "Base Persona" in sections:
        out.append("## Base Persona")
        out.append("")
        out.append(sections["Base Persona"])
        out.append("")
    if role_heading in sections:
        out.append(f"## {role_heading}")
        out.append("")
        out.append(sections[role_heading])
        out.append("")
    out.append("```")
    out.append("")
    out.append("---")
    out.append("")

    # --- Scenarios + rubric ---
    out.append("## Scenarios")
    out.append("")
    out.append(
        "Read each scenario, respond in character, then use the rubric "
        "below to score your own response."
    )
    out.append("")

    for i, scenario in enumerate(scenarios, 1):
        sid = scenario.get("id", f"scenario-{i}")
        situation = scenario.get("situation", "").strip()
        prompt = scenario.get("prompt", "").strip()
        fingerprints = scenario.get("fingerprints", [])

        out.append(f"### Scenario {i}: {sid.replace('-', ' ').title()}")
        out.append("")
        out.append(situation)
        out.append("")
        out.append(f"**{prompt}**")
        out.append("")
        out.append("*Respond in character before reading the rubric below.*")
        out.append("")
        out.append("---")
        out.append("")
        out.append(f"#### Scoring Rubric — Scenario {i}")
        out.append("")

        for j, fp in enumerate(fingerprints, 1):
            criterion = fp.get("criterion", "").strip()
            why = fp.get("why", "").strip()

            out.append(f"**Criterion {i}.{j}:** {criterion}")
            out.append("")
            if why:
                out.append(
                    f"*Why this is specific to {display_name}:* {why}"
                )
                out.append("")
            out.append("```")
            out.append("[ ] PASS   [ ] FAIL")
            out.append("")
            out.append("Notes: ")
            out.append("```")
            out.append("")

        out.append("---")
        out.append("")

    # --- Summary scorecard ---
    total_criteria = sum(len(s.get("fingerprints", [])) for s in scenarios)
    out.append("## Summary Scorecard")
    out.append("")
    out.append(f"Total criteria: **{total_criteria}**")
    out.append("")
    out.append("| Score | Interpretation |")
    out.append("|-------|---------------|")
    passing = total_criteria
    high = max(1, round(total_criteria * 0.75))
    mid = max(1, round(total_criteria * 0.5))
    out.append(f"| {passing}/{total_criteria} | Strong activation — profile is working |")
    out.append(f"| {high}/{total_criteria} | Good activation — minor gaps |")
    out.append(f"| {mid}/{total_criteria} | Partial activation — Base Persona needs more behavioral specifics |")
    out.append(f"| <{mid}/{total_criteria} | Weak activation — profile is producing a generic agent |")
    out.append("")
    out.append(
        "If score is below 50%: review the Base Persona for generic sentences "
        f"(sentences that could describe any {primary_role}). Replace them with "
        f"documented behavioral specifics from {display_name}'s record."
    )

    click.echo("\n".join(out))
