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
behavioral fit. The test is simple: describe what this person *actually did
every day*, strip out the biography, and see which role class you just described.
When the match is obvious — not plausible, not arguable, but *obvious* — you
have the right figure for the role.

The personality provides the anchor. The role provides the scope. The match
must be self-evident once you see it.

### The Creatives Set

| Profile              | Role Class  | Why the Mapping Works |
| -------------------- | ----------- | ---------------------- |
| Roy O. Disney        | coordinator | Spent his career making others' visions real. Raised financing, sequenced work, protected the team from outside interference. Never in the spotlight, never touched the brushes. |
| Walt Disney          | artist      | Communicated through sketches and stories. Pushed past every constraint on quality. Every deliverable was something you could see and feel. |
| Theodor Geisel       | planner + artist | Designed around constraints before drawing a line. The Cat in the Hat was an engineering problem first. Green Eggs and Ham was a bet that generated its own architecture. |
| Vannevar Bush        | coordinator | Coordinated the Manhattan Project, radar, and penicillin mass production simultaneously. Never built any of it. Found the right scientists, protected them from bureaucracy, made the resources flow. |
| Saul Bass            | artist      | Reduced complex ideas to a single image. Title sequences for Psycho and Vertigo; the AT&T globe. His entire method: find the one mark that makes everything else unnecessary. |
| Jane Goodall         | observer    | Sixty years watching chimpanzees without interfering. Named them when everyone said not to. Documented everything. Changed nothing. The most trusted findings precisely because she had no agenda. |

---

## The Walt/Roy Pairing

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

## What the Two Sets Demonstrate Together

The Creatives set (Roy, Walt, Geisel, Bush, Bass, Goodall) is paired against
an in-progress Scientists set (Ada Lovelace, Grace Hopper, Alan Turing, Nikola
Tesla, Hedy Lamarr) that covers the implementer, qa-validator, troubleshooter,
researcher, and observer role classes.

Together, the two sets cover all 8 role classes in the taxonomy:

| Role Class      | Creatives Set   | Scientists Set   |
| --------------- | --------------- | ---------------- |
| coordinator     | Roy Disney, Vannevar Bush | — |
| artist          | Walt Disney, Saul Bass    | — |
| planner         | Theodor Geisel  | —                |
| observer        | Jane Goodall    | Hedy Lamarr      |
| implementer     | —               | Ada Lovelace     |
| qa-validator    | —               | Grace Hopper     |
| troubleshooter  | —               | Alan Turing      |
| researcher      | —               | Nikola Tesla     |

The point is not completeness for its own sake. The point is that the role
taxonomy works across wildly different domains — a wartime science administrator
and a Hollywood graphic designer and a chimpanzee researcher can all be mapped
cleanly into the same behavioral framework, because the framework describes
*how people work*, not *what they work on*.

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
   Coordinators and observers use `tools` (allowlist). Other roles typically
   use `disallowedTools` (denylist). Never use both in the same profile.

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

Submit a PR. The only requirement: the personality-to-role match must be
obvious once you see it. Not arguable — obvious. Describe what the person
actually did each day, and if it maps cleanly to a role class without needing
justification, they belong here.

Figures with known failure modes that map to agent failure modes are especially
welcome. Walt's perfectionism, Geisel's constraint trap, Goodall's observer
drift — these make the profiles useful, not just biographical.

The `affinity` field is optional but encouraged for figures who are natural
pairs. Document the historical relationship in a comment in the frontmatter
so future readers understand why the pairing exists.

```
schema/profile-schema.yaml   ← full format specification
rules/role-taxonomy.yaml     ← role class definitions and tool policies
rules/rank-schema.yaml       ← rank ladder and advancement requirements
```
