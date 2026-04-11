# armies v3.0

AI agents are generic. You get the same assistant whether you're debugging a race condition or writing a post-mortem. No memory between sessions. No accountability when something goes wrong. No personality to anchor behavior. Every prompt starts from zero, and every agent is interchangeable. That is the problem. Armies gives your agents identity. Historical figures with earned expertise, accumulated XP, and structural role constraints. Grace Hopper ships code fast and asks forgiveness later. Jane Goodall observes without contaminating the scene. Roy Disney keeps Walt's impossible vision from burning the budget. The right specialist for every mission -- and they get better every time they're deployed.

**STATUS: experimental**

---

## Installation

### Pre-built binary (recommended)

Download the latest binary for your platform from the [Releases page](https://github.com/petersimmons1972/armies/releases).

```bash
# macOS (arm64)
curl -L https://github.com/petersimmons1972/armies/releases/latest/download/armies-darwin-arm64 -o armies
chmod +x armies
sudo mv armies /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/petersimmons1972/armies/releases/latest/download/armies-linux-amd64 -o armies
chmod +x armies
sudo mv armies /usr/local/bin/
```

### go install

If you have Go 1.22 or later installed:

```bash
go install github.com/petersimmons1972/armies@latest
```

This compiles and installs the binary in `$GOPATH/bin`. Make sure that directory is in your `PATH`. If you are not sure where that is, run `go env GOPATH` and add `$(go env GOPATH)/bin` to your `PATH` in `~/.bashrc` or `~/.zshrc`.

### Build from source

```bash
git clone https://github.com/petersimmons1972/armies
cd armies
go build -o armies .
sudo mv armies /usr/local/bin/
```

That is the complete installation. No Python. No virtual environment. No Docker. One binary, no runtime dependencies.

---

## Quick Start

```bash
# 1. Initialize your private profile store
armies init

# 2. Install the bundled example profiles
armies seed

# 3. Spawn Grace Hopper as an implementer
armies spawn grace-hopper --role implementer

# 4. After the mission, record it
armies record grace-hopper "implemented user auth" --xp 100
```

`armies init` creates `~/.armies/` -- your private profile store, separate from the armies repo. `armies seed` installs all bundled profiles from `examples/generals/` into `~/.armies/profiles/`, so you have a working roster immediately. `spawn` reads the profile, merges the personality and role blocks, and outputs a prompt you paste into Claude Code. `record` writes the service record and updates XP -- next spawn, she's smarter.

See [Getting Started](docs/getting-started.md) for the full narrative walkthrough.

---

## The Idea

Most agent frameworks treat personality as decoration -- a system prompt seasoning sprinkled over the same underlying behavior. Armies treats personality as the *anchor*. When you spawn Grace Hopper, you are not getting "an agent with a pirate-themed prompt." You are getting a mathematician who invented the compiler, who believes it is easier to ask forgiveness than permission, and whose known failure mode is cutting corners on tests when she moves too fast. That failure mode is documented in her profile. It constrains her behavior. It makes the agent self-aware of its own weaknesses in a way that generic assistants never are.

The personality is who they *are*. The role is what they *do this time*. The same historical figure can play different roles depending on the mission. Walt Disney as an `artist` produces visual output -- SVGs, layouts, design systems. But Walt Disney as a `planner` produces creative briefs and architectural visions. The personality stays constant (ambitious, visual, allergic to compromise), but the behavioral constraints change with the role.

This matters because an agent that *is* someone behaves coherently. It makes consistent decisions under pressure. It has predictable failure modes you can plan around. An agent that has tags is just a prompt. An agent with identity is a team member.

---

## How It Works

```mermaid
sequenceDiagram
    participant You
    participant armies CLI
    participant Profile
    participant Claude

    You->>armies CLI: armies spawn grace-hopper --role implementer
    armies CLI->>Profile: Read frontmatter + Base Persona + Role:implementer block
    Note over Profile: Role:researcher block stays on disk
    armies CLI->>You: Merged spawn prompt (focused, personality-laden)
    You->>Claude: Paste prompt → spawn agent
    Claude->>You: Grace Hopper executes mission
    You->>armies CLI: armies record grace-hopper "fixed the race condition"
    armies CLI->>Profile: +100 XP, write service record
    Note over Profile: Next spawn: Grace Hopper is smarter
```

You ask the CLI to spawn a profile in a specific role. The CLI reads the profile's frontmatter (name, XP, rank, model preference, tool restrictions), the Base Persona (always loaded -- this is the personality anchor), and the single Role block you selected. Everything else stays on disk. The output is a merged prompt that you paste into Claude Code to spawn the agent.

After the mission, you record what happened. The CLI writes a service record entry and updates XP. Next time you spawn Grace Hopper, her XP is higher, her service record is longer, and the spawn prompt includes her deployment history. She is literally more experienced.

---

## Anatomy of a Profile

```mermaid
graph TD
    A["profile.md"] --> B["Frontmatter\n(name, role, XP, rank, model)"]
    A --> C["## Base Persona\n(always loaded — who they ARE)"]
    A --> D["## Role: implementer\n(loaded when role=implementer)"]
    A --> E["## Role: troubleshooter\n(stays on disk unless selected)"]

    style C fill:#2d5a27,color:#fff
    style D fill:#1a3a6b,color:#fff
    style E fill:#3a3a3a,color:#888
```

The **Base Persona** is the personality anchor. It loads every time, regardless of role. The **Role blocks** define how the agent operates this time -- behavioral instructions for one specific mission type. Only one loads per spawn. The **frontmatter** carries tool restrictions (enforced by Claude Code natively), model preference, XP, and rank.

---

## Role Classes

| Role              | Archetype                                                                  | Responsibility                                      | Allowed Tools                   |
| ----------------- | -------------------------------------------------------------------------- | --------------------------------------------------- | ------------------------------- |
| `coordinator`     | Orchestrates without touching the work -- tool restriction enforced        | Delegates tasks, synthesizes results                | Agent, Read, Grep, Glob        |
| `implementer`     | Ships first, documents after -- asks forgiveness not permission            | Code, config, and file changes                      | Full toolset (no Agent)         |
| `qa-validator`    | Finds the assumption everyone else missed -- read-only by design           | Tests, audits, and verification                     | Read, Bash (read-only), Grep   |
| `planner`         | Refuses to fight until certain of winning -- preparation over improvisation | Architecture, specs, design documents               | Read, Write, Edit, Grep        |
| `researcher`      | Raw signal collection -- feeds the coordinator's synthesis                 | Intelligence gathering and prior art                | Read, Write, Bash, Grep        |
| `troubleshooter`  | Pivots under pressure -- high autonomy, documents after                    | Root cause analysis and emergency fixes             | Full toolset                    |
| `artist`          | Named aesthetic, not generic output -- visual deliverables                 | SVG, layout, design systems, brand assets           | Read, Write, Edit, Bash        |
| `observer`        | Zero-context review -- receives no prior findings, absolute malus immunity | Independent cross-check of completed work           | Read only                       |

Tool restrictions are a **structural guarantee**, not a suggestion. The Eisenhower Precedent taught a hard lesson: a coordinator with Write/Edit/Bash tools *will* use them under pressure, creating unreviewed changes with no accountability trail. These restrictions exist because violations happened.

---

## Example Profiles

The bundled profiles in `examples/generals/` demonstrate obvious historical matches between personality and role.

**Grace Hopper / implementer** -- She invented COBOL, coined the term "debugging," and retired from the Navy at 79 after they kept recalling her because nobody else could do what she did. Her motto was "It's easier to ask forgiveness than permission." If you need someone to take a specification and make it real without getting stuck in committee, this is the profile.

**Jane Goodall / observer** -- Sixty years of observation without interference at Gombe Stream. Her entire scientific method was *watch and document*. She receives only raw inputs, no prior findings, and her structural tool restriction means she literally cannot modify what she's reviewing. Her malus immunity is absolute -- you don't court-martial a scout for reporting what they saw, even if the report turns out to be a false alarm.

**Vannevar Bush / coordinator** -- As Director of the Office of Scientific Research and Development, he coordinated the Manhattan Project, radar, the proximity fuse, and penicillin mass production simultaneously. He never built any of it himself. He found the right people, gave them what they needed, and kept the bureaucracy away from their doors. His tool restriction means he cannot write code or run commands -- every deliverable routes through a specialist he has briefed and dispatched.

---

## Progression

Agents earn XP for successful completions, catching bugs early, accurate diagnoses, and clean first-pass implementations. XP accumulates across sessions and unlocks higher spawn eligibility tiers.

Malus is the other side of the ledger. Scope creep, skipping tests, a coordinator implementing code directly, an observer anchoring on prior findings -- these are logged as malus events with severity levels. The malus ledger is permanent. Only the human operator can resolve a malus event. Unresolved malus reduces spawn eligibility.

Stars are earned at XP thresholds and represent sustained competence across specific categories. A three-star implementer has proven reliability across many deployments, not just one good session.

For the full progression system -- XP schedules, rank ladder, star thresholds, and eligibility gates -- see [docs/progression.md](docs/progression.md).

---

## Contributing

Contributions welcome -- especially new profiles, team templates, and documentation. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a PR.

---

## Documentation

| Document                                                  | What It Covers                                           |
| --------------------------------------------------------- | -------------------------------------------------------- |
| [Getting Started](docs/getting-started.md)                | Zero to first spawn -- narrative walkthrough             |
| [How It Works](docs/how-it-works.md)                      | Architecture, spawn flow, profile resolution             |
| [Creating Profiles](docs/creating-profiles.md)            | Building your own roster from scratch                    |
| [Progression](docs/progression.md)                        | XP, stars, rank, malus, and eligibility gates            |
| [Team Templates](docs/team-templates.md)                  | Coordinated multi-agent mission compositions             |
| [Coordinator Guide](docs/coordinator-guide.md)            | Running campaigns with structural tool restrictions      |
| [Accountability](docs/accountability.md)                  | Malus ledger, service records, and audit trails          |
| [CLI Reference](docs/cli-reference.md)                    | Every command, every flag                                |
| [Security](docs/security.md)                              | Binary distribution, profile integrity, what armies does not do |
| [Troubleshooting](docs/troubleshooting.md)                | Common issues and how to resolve them                    |
