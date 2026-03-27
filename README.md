# Armies — Multi-Agent Coordination System

A self-learning multi-agent coordination engine for Claude Code. Personalities persist. Experience compounds. The right specialist for every mission.

**STATUS: experimental**

> Full documentation in progress — see /docs/

---

## Overview

Armies is the next generation of the "generals" multi-agent coordination system. It defines a portable, schema-driven profile format for Claude Code sub-agents — giving each agent a persistent identity, role class, behavioral constraints, and an XP-based progression system.

Armies profiles are:
- **Portable** — drop a profile into any Claude Code project
- **Composable** — assemble pre-built team templates for any mission type
- **Accountable** — XP and malus ledgers track performance across sessions
- **Open** — fully open-source, community-contributed profiles welcome

---

## Quick Start

```bash
# Clone into your project's .claude/agents/ directory
git clone https://github.com/your-org/armies .claude/armies

# Pick a team template
cp armies/teams/standard.yaml .claude/team.yaml

# Spawn your coordinator
# (see docs/spawning.md for full instructions)
```

---

## How It Works

1. Each agent has a **profile** — a YAML/Markdown file defining role, personality, tools allowed, XP, and behavioral rules.
2. A **coordinator** agent reads the team template and spawns sub-agents using the `Task` tool.
3. Sub-agents execute their mission and report back. Coordinators synthesize results — they never implement directly.
4. After each session, XP is updated in the profile. Malus events are logged to the accountability ledger.
5. Profiles commit to git — experience persists across sessions, machines, and projects.

---

## Role Classes

| Role            | Responsibility                                      | Allowed Tools                   |
| --------------- | --------------------------------------------------- | ------------------------------- |
| `coordinator`   | Delegates tasks, synthesizes results                | Task, Read (no Write/Edit/Bash) |
| `planner`       | Architecture, specs, design documents               | Read, Write, WebSearch          |
| `implementer`   | Code, config, and file changes                      | Full toolset                    |
| `qa-validator`  | Tests, audits, and verification                     | Read, Bash (read-only)          |
| `researcher`    | Intelligence gathering and prior art                | Read, WebSearch, WebFetch       |
| `troubleshooter`| Root cause analysis and emergency fixes             | Full toolset                    |
| `observer`      | Zero-context review — receives no prior findings    | Read only                       |

---

## Profile System

Profiles live in `.claude/agents/` and follow the schema defined in `schema/profile.yaml`.

Each profile contains:
- **Frontmatter** — role class, model, XP, malus balance, spawn eligibility
- **Personality block** — name, backstory, communication style
- **Behavioral rules** — what the agent will and will not do
- **Tool restrictions** — explicit allowlist per role class
- **Service record** — recent deployments and outcomes

See `profiles/examples/` for reference implementations.

---

## Progression System

Agents earn XP for:
- Successful task completions
- Catching bugs before deployment
- Accurate root-cause diagnoses
- Clean first-pass implementations (no rework)

Malus events are recorded for:
- Scope creep (modifying files outside the task boundary)
- Skipping tests
- Coordinator implementing code directly
- Observer anchoring on prior findings

XP thresholds unlock higher `spawn_eligibility` tiers — which coordinators use to assign harder tasks.

---

## Team Templates

Pre-built team compositions live in `teams/`. Available templates:

| Template                | Use Case                                      |
| ----------------------- | --------------------------------------------- |
| `standard.yaml`         | General feature work and implementation       |
| `quality-sprint.yaml`   | Security reviews, critical releases           |
| `research-spike.yaml`   | Deep-dive R&D and intelligence gathering      |
| `firefighting.yaml`     | Production incidents and emergency response   |
| `planning-session.yaml` | Architecture and design before implementation |

---

## Installation

```bash
# As a git submodule (recommended)
git submodule add https://github.com/your-org/armies .claude/armies

# Or clone directly
git clone https://github.com/your-org/armies
```

Copy agent profiles from `profiles/examples/` into your project's `.claude/agents/` directory and customize as needed.

---

## Contributing

Contributions welcome — especially:
- New example agent profiles
- Additional team templates
- Schema improvements
- Documentation

Please read `docs/contributing.md` before submitting a PR.
