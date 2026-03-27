# Example Profiles — Armies 2.0

These profiles are public demonstrations that the Armies engine works with
any persona — not just military figures. A profile is a behavioral contract,
not a costume. What matters is the alignment between a historical figure's
actual decision-making style and the role class they are assigned to.

---

## The Methodology: Personality → Role Class

Every profile here was chosen because the historical figure's *actual working
style* maps cleanly onto one of the Armies role classes. This is not about
who was "important" or who sounds impressive in a system prompt. It is about
behavioral fit.

| Profile | Role Class | Why the Mapping Works |
|---|---|---|
| Roy O. Disney | coordinator | Spent his career making others' visions real. Raised financing, sequenced work, protected the team from outside interference. Never in the spotlight, never touched the brushes. |
| Walt Disney | artist | Communicated through sketches and stories. Pushed past every constraint on quality. Every deliverable was something you could see and feel. |
| Theodor Geisel | planner + artist | Designed around constraints before drawing a line. The Cat in the Hat was an engineering problem first. Green Eggs and Ham was a bet that generated its own architecture. |

The Scientists set (ada-lovelace, grace-hopper, alan-turing, nikola-tesla,
hedy-lamarr) covers the implementer, qa-validator, troubleshooter, researcher,
and observer role classes. Together, the two sets cover all 8 role classes in
the taxonomy.

---

## The Walt/Roy Pairing Concept

Walt Disney's profile includes a special frontmatter field:

```yaml
affinity: roy-disney
```

This field is **advisory, not enforced**. The engine does not require Roy to
be present before Walt can be deployed. But it is guidance for the coordinator
doing roster selection: Walt works best when Roy is coordinating the mission.

The historical reason is obvious. Walt's obsessiveness made him an extraordinary
creative force and a budget catastrophe. Roy's financial discipline and
operational sequencing was the counterweight that turned Walt's visions from
expensive dreams into finished projects. Disneyland exists because Roy kept
the lights on while Walt spent everything.

In Armies terms: a coordinator who knows their artist is Walt should brief
more tightly, set explicit scope boundaries, and check deliverables against
the brief before accepting them. Walt will expand scope in pursuit of quality.
That is a feature, not a bug — but it needs a Roy to manage it.

**Other pairings may emerge as this profile set grows.** If you build a
creative partner pair, use the `affinity` field to document the relationship.
The field takes a profile `name` value (kebab-case) pointing to the natural
partner.

---

## How to Build Your Own Profile Pack

A profile is a Markdown file with YAML frontmatter. The full specification
lives at:

```
~/projects/armies/schema/profile-schema.yaml
```

The short version:

1. **Choose a historical figure** whose actual working style maps to a role class.
   Read the role taxonomy (`~/projects/armies/rules/role-taxonomy.yaml`) and
   ask: what did this person *actually do* — not what are they famous for?

2. **Write the frontmatter**: `name`, `description`, `roles`, `xp: 0`,
   `rank: Colonel`, `model`, and tool restrictions appropriate for the role.

3. **Write the Base Persona**: 200-400 words. Historical context, personality,
   decision-making philosophy, known failure modes. Remove the name — it should
   still be recognizably this person.

4. **Write one Role block per declared role**: pre-mission checklist, how to
   work, post-mission requirements. This is the agent's methodology, not a
   generic role description.

5. **Validate against the checklist** in `schema/profile-schema.yaml` Section 7
   before committing.

Profiles live in `.claude/agents/` when deployed. The `profiles/examples/`
directory is for reference implementations and onboarding — copy a profile to
`.claude/agents/` in your project to use it.

---

## Have a Historical Figure Who Belongs Here?

Submit a PR. The only requirement is that the behavioral mapping is honest:
the figure's actual working style, including their known failure modes, must
fit the role class. A coordinator who never delegated, a planner who never
planned — these are not good fits regardless of historical fame.

The `affinity` field is optional but encouraged for figures who are natural
pairs. Document the historical relationship in a comment in the frontmatter
so future readers understand why the pairing exists.

```
schema/profile-schema.yaml   ← full format specification
rules/role-taxonomy.yaml     ← role class definitions and tool policies
rules/rank-schema.yaml       ← rank ladder and advancement requirements
```
