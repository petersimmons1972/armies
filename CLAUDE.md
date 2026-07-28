---
project: armies
purpose: CLI tool giving AI agents persistent identity, XP, and role constraints via historical-figure profiles.
stack: [go]
status: active
entrypoints:
  - main.go
  - cmd/
  - teams/
related: [generals, aifleet]
notes: Experimental status per README. XP values are write-once by session-close routine — never edit directly.
---

# Armies — Agent Session Guardrails

These rules apply to ALL agent sessions operating within the Armies project.
They are NON-NEGOTIABLE and override any agent's default behavior.

**Founder profile**: `profiles/examples/founder.md` — read before writing any content as Peter or making decisions on his behalf.

---

## XP Integrity

**NEVER modify XP values directly.**
XP is calculated and written only by the designated session-close routine. If you believe XP is wrong, file a GitHub Issue. Do not edit the profile. [AP.12]

---

## Git Discipline — NON-NEGOTIABLE [AP.5]

- Every profile change (XP update, malus event, rule amendment) MUST be committed.
- Commit message format: `profile(<agent-name>): <what changed and why>`
- Do NOT batch unrelated profile changes into a single commit.
- Do NOT amend existing commits. Always create a new commit.
- `git diff --staged` before every commit — no surprises.
- Profiles that are not committed are not real. If it's not in git, it didn't happen.

---

## Service Records Are Mandatory [AP.12]

After every deployment or significant task:
1. Update the agent's service record block in their profile.
2. Include: date, task description, outcome (success/partial/failure), XP delta.
3. Commit the updated profile immediately.

A profile without a service record is an unaccountable agent.

---

## Coordinator Doctrine [AP.1]

Coordinators ONLY coordinate. This means:

- **Allowed:** Task, Read, Glob, Grep (for situational awareness)
- **Forbidden:** Write, Edit, Bash, any tool that modifies files or runs commands

If a coordinator needs to make a change, it spawns an implementer sub-agent.
A coordinator that writes code is a scope violation. Log it as a malus event.

---

## Profile Changes

Profile changes (to behavioral rules, tool restrictions, or role class) require:
1. A written rationale in the commit message.
2. Commit and push before the change takes effect.
3. No profile change is valid until it exists on the remote.

---

## Malus Events [AP.12]

Malus events are logged to `accountability/malus-ledger.yaml`.

Record format: `{agent, date, event, severity: minor|major|critical, malus_points: int, resolved: false}`

Unresolved malus events reduce an agent's spawn eligibility tier.
Only the agent's owner (the human operator) can mark a malus event resolved.

---

## Observer Rules

Observer agents receive ONLY the raw task input and source materials (code, docs, data).
Observers NEVER receive prior agent findings, synthesis outputs, or debug logs from other agents.
An observer that is contaminated with prior findings provides no independent signal; discard their output and re-run clean.

---

## Emergency Override

If you are blocked and none of the above rules can be followed without causing greater harm,
halt and surface the conflict to the human operator immediately. Do not self-authorize exceptions.
