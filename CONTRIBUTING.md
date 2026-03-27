# Contributing to Armies

Thank you for wanting to make this better. This is a short guide — we'd rather you spent your time building profiles than reading contribution docs.

## The Most Valuable Thing You Can Do

Submit a profile pack. A profile pack is a set of historical figures mapped to role classes — people whose personality, working style, and known history make them an obvious fit for a specific kind of mission.

The test for a good profile: **when you read it, the match should feel inevitable**. If you have to explain why Hedy Lamarr is a troubleshooter, the profile isn't working yet. When it's working, the reader thinks "of course she is."

## Profile Packs

A profile pack is a directory of `.md` files following the schema at `schema/profile-schema.yaml`.

You can contribute:
- **A single profile** (one historical figure, one PR)
- **A themed set** (e.g., "Renaissance Artists," "Women of Computing," "Jazz Musicians")
- **A domain pack** (e.g., "Film Crew," "Kitchen Brigade," "Early Aviators")

See `profiles/examples/` for complete working examples. See `profiles/examples/README.md` for the methodology behind profile selection.

### Profile Requirements

1. **The personality-to-role match must be obvious.** Research the person's actual working style, not just their achievements. Achievements tell you what they did; working style tells you how they thought.

2. **Base Persona is the anchor.** 200–400 words on who this person actually was — their voice, their method, their known failure modes. This is what makes the agent distinctly them rather than a generic specialist.

3. **Role blocks are scoped.** Each `## Role: <name>` block contains only behavioral instructions relevant to that specific role. An implementer block is not a general-purpose prompt. It tells the agent exactly how THIS person would approach implementation specifically.

4. **No living people.** Profiles are based on historical figures — people whose work and personality are part of the public record. No living people, no fictional characters.

5. **XP starts at 0.** All contributed profiles start at zero experience. They earn it the same way everyone else does.

### Profile Format

```markdown
---
name: kebab-case-name
display_name: "Full Name"
roles:
  primary: <role-class>
  secondary: <role-class>   # optional
xp: 0
rank: Colonel
model: sonnet               # or opus for complex coordinators/planners
---

## Base Persona
[200–400 words. Core personality. Historical context. Voice. Known failure modes.]

## Role: <primary-role>
[100–200 words. Behavioral instructions for this specific role only.]

## Role: <secondary-role>   # if listed in frontmatter
[100–200 words.]
```

Valid role classes: `coordinator`, `implementer`, `qa-validator`, `planner`, `researcher`, `troubleshooter`, `artist`, `observer`. Additional roles (e.g., `writer`, `security-auditor`) can be defined in private packs — see `rules/role-taxonomy.yaml`.

## Code Contributions

The core engine lives in `src/armies/`. It's a small Python CLI built with Click and PyYAML.

Before contributing code:
- Run `pip install -e ".[dev]"` to install with dev dependencies
- The progressive loading rule is non-negotiable: `armies spawn` must never load more than one role block into memory. Don't break this.
- Streaming profile reader is in `src/armies/profiles.py` — understand it before touching it.

### Running Locally

```bash
git clone https://github.com/petersimmons1972/armies
cd armies
pip install -e .
armies --help
```

For Docker:

```bash
cd docker
docker compose build
docker compose run armies --help
```

## Opening Issues

If you find a bug: open an issue with the exact command that failed and the error output.

If you have a feature idea: open an issue describing the use case, not the implementation. "I want to be able to X" is more useful than "add a Y flag."

## Pull Requests

Keep PRs focused. A profile pack is one PR. A bug fix is one PR. Don't mix profiles and code changes.

Write a clear PR description: what you're adding, why the profile match works, and where you found your primary sources for the Base Persona.

That's it. Go build something.
