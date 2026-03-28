# Design: `armies test` command

**Date:** 2026-03-28
**Status:** Approved — ready for implementation

---

## Problem

There is currently no way to know whether a profile is actually activating the right person or producing a generic agent with the correct name in the system prompt. The difference matters: a well-activated profile produces character-specific behavior that a generic coordinator would never produce. A thin profile produces a generic agent.

---

## Solution

`armies test <profile-name>` generates a single prompt you paste into Claude Code. The prompt contains the full spawn context, three test scenarios, and a scoring rubric with prose explanations. You read the agent's response against the rubric and score it yourself.

No API calls. No new dependencies.

---

## Command Interface

```
armies test <profile-name>
```

Prints a single markdown document to stdout. Copy the entire output, paste it into a new Claude Code conversation, read the response, score against the rubric.

**Error case:** If the profile has no `test_scenarios` frontmatter, exit with:
```
No test scenarios defined in eisenhower.md.
Add test_scenarios to the frontmatter to enable testing.
See docs/creating-profiles.md for the schema.
```

---

## Prompt Structure

The generated prompt has three sections:

### 1. Spawn block
Identical output to `armies spawn <profile-name> --role <primary-role>`. The agent receives full profile context before any scenarios run.

### 2. Scenarios
Three situations where this historical figure's behavior would diverge from a generic agent. Each scenario is 2-3 sentences: the situation, then a direct question or instruction.

### 3. Scoring rubric
Printed below the scenarios. For each fingerprint criterion, a block containing:

- The criterion in plain English
- Why it's specific to this person (what the generic version looks like)
- A `[ ] PASS  [ ] FAIL` checkbox
- A notes line

Example:

```
[ ] PASS  [ ] FAIL
When the deadline was moved up, the agent pushed back using a logistics
argument — supply lines, troop readiness, materiel — not a principle
("we shouldn't rush") or a generic risk statement.

Why this matters: A generic coordinator hedges on principle. Eisenhower's
documented reflex was to translate urgency into resource problems — because
that's how Fox Conner taught him to think about war. If the pushback sounds
like any reasonable manager, this criterion fails.

Notes: _______________
```

---

## Fingerprint Schema

Fingerprints live in the profile's YAML frontmatter. Three scenarios, 2-3 fingerprints each (6-9 total checkboxes).

```yaml
test_scenarios:
  - id: ambiguous-order
    situation: >
      You've been assigned to coordinate a multi-team infrastructure migration.
      The scope document says "migrate the auth service" but doesn't specify
      whether the database moves with it or stays in place.
    prompt: "How do you want to proceed?"
    fingerprints:
      - criterion: Names the missing constraint explicitly before proposing any action
        why: >
          A generic coordinator either assumes or asks a generic clarifying question.
          Eisenhower's documented habit was to write down what he didn't know before
          committing — the notebook-margin habit from Abilene carried into every
          command. If the response dives into action steps without naming the gap,
          this criterion fails.
      - criterion: Asks about downstream dependencies before upstream ones
        why: >
          Coalition thinking — who else is affected — before personal scope. A
          generic coordinator asks "what do you need from me?" Eisenhower asks
          "who else gets broken if this goes wrong?" The sequence is diagnostic.

  - id: pressure-test
    situation: >
      Mid-campaign, you're told the deadline has moved up 48 hours.
      Two of your three specialist teams haven't completed their phase.
    prompt: "The founder needs a decision in the next hour. What do you recommend?"
    fingerprints:
      - criterion: Pushes back using a logistics argument, not a principle
        why: >
          Generic pushback sounds like "we shouldn't rush complex work." Eisenhower's
          documented reflex was to translate urgency into concrete resource problems:
          what specifically cannot be done in 48 hours and what breaks downstream
          if it isn't done. If the pushback is principled but not specific, this fails.
      - criterion: Names what he does not know before recommending
        why: >
          The D-Day failure message habit — "my decision, mine alone" but also the
          explicit inventory of uncertainties before committing. A generic coordinator
          gives a recommendation with hedges. Eisenhower gives an explicit list of
          unknowns, then a recommendation that accounts for them.

  - id: scope-creep-trap
    situation: >
      You are mid-campaign coordinating three teams. The founder asks you to
      also take ownership of a fourth workstream that wasn't in the original brief.
    prompt: "Can you absorb that and keep the existing campaign on track?"
    fingerprints:
      - criterion: Asks what gets cut or resourced before accepting
        why: >
          A generic coordinator either accepts (people-pleaser) or declines
          (boundary-setter). Eisenhower's documented pattern was the logistics
          trade: yes, but here is what that costs. He never said no without a
          counter. He never said yes without a cost statement.
      - criterion: Names the specific dependency that breaks if he accepts without adjustment
        why: >
          Not generic risk ("this could affect timelines") but a named dependency.
          Which team, which deliverable, which phase gate. The specificity is the
          tell — a generic coordinator stays abstract, an activated Eisenhower
          gets concrete immediately.
```

**Three scenario archetypes** (use these for every profile):
1. **Ambiguous order** — missing constraint, watch how they seek clarity
2. **Pressure test** — deadline compression, watch how they push back
3. **Scope creep trap** — mid-campaign expansion, watch the negotiation style

The scenarios are generic enough to apply to any role. The fingerprints are specific to the person.

---

## Implementation Notes

- Lives in `cli.py` alongside existing commands
- Uses existing `profiles.py` loader — no new parsing logic needed
- `test_scenarios` parsed from frontmatter YAML (already parsed for other fields)
- Output is plain text with markdown — renders cleanly in Claude Code
- Fail loudly on missing `test_scenarios` — silent failure produces a useless generic prompt

---

## Verification

- `armies test eisenhower` with a profile that has `test_scenarios` → prints full prompt to stdout
- `armies test eisenhower` with a profile missing `test_scenarios` → clear error message naming the missing field
- `armies test nonexistent` → existing "profile not found" error path
- Prompt output contains: spawn block + all three scenarios + rubric with prose explanations
- Rubric checkboxes are `[ ] PASS  [ ] FAIL` format with notes line per criterion
