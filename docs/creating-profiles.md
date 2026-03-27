# Creating Profiles

A profile is a behavioral contract. It tells the Armies engine who an agent is, how
they think, and what they will and will not do on a mission. The historical figure
is the anchor — their personality, working style, and known failure modes become
the agent's personality, working style, and known failure modes.

This guide walks you through building one from scratch.

---

## The Only Test That Matters

Before you write a single line of frontmatter, you need to pass one test: **the
personality-to-role match must be obvious once you see it.** Not clever. Not
arguable. Obvious.

Jane Goodall as observer: she spent sixty years watching chimpanzees at Gombe
Stream without interfering. She documented everything. She changed nothing. Her
method was to observe so carefully and so patiently that the animals forgot she
was there. Of course she's an observer. The match isn't something you deduce from
her career — it's something you recognize the moment you look at what she actually
did every day.

Grace Hopper as implementer: she coined "debugging" by removing an actual moth
from a relay and taping it into the logbook. She invented COBOL because she was
tired of humans having to speak machine language. When the Navy said machine-
independent programming wasn't possible, she built a compiler and handed it to
them. She said "it's easier to ask forgiveness than permission" and she meant it.
Of course she's an implementer.

If you have to argue for the match — if you find yourself writing "despite X, she
is really a Y because..." — the match is wrong. Keep looking. The right figure for
a role makes you wonder why you ever looked anywhere else.

This is not about finding impressive people. It's about finding people whose actual
daily working style maps cleanly onto a role class. Walt Disney's most celebrated
achievement was *Fantasia*; his actual daily method was sketching and narrating and
making everyone else feel what he felt until the drawing matched the feeling. Of
course he's an artist. Roy Disney's most important contribution to the Disney
empire is nearly invisible — he's the one who kept the money flowing while Walt
spent it. Of course he's a coordinator.

---

## Step 1: Choose Your Historical Figure

What you are looking for is working style, not achievements. Achievements tell you
what someone did. Working style tells you how they thought, decided, and behaved
under pressure. Those are the traits that survive translation into an AI agent.

The questions to ask while you research are these:

**How did they handle constraints?** Theodor Geisel designed *The Cat in the Hat*
around a vocabulary list of 236 words — the constraint became the generator of the
work. An agent built on that working style will turn constraints into fuel, not
obstacles. *Green Eggs and Ham* was literally a bet that Geisel couldn't write a
book with fewer than fifty unique words. He won the bet and produced one of the
best-selling children's books in history. Constraints were his engine.

**What did they do when blocked?** Grace Hopper didn't wait for permission and she
didn't wait for perfect requirements. She shipped something imperfect and iterated.
When institutional resistance appeared, she worked around it. A figure who pushed
through blockage becomes an agent that keeps moving; a figure who waited for
consensus becomes an agent that asks a lot of clarifying questions.

**What was their known failure mode?** Tesla designed forever and shipped rarely.
Disney rebuilt sequences that were already finished. Goodall eventually became so
immersed in Gombe that she started feeding the chimps and altering the ecosystem
she was studying. Every real person had a pattern of failure, and that pattern is
what makes the profile useful. An agent with no failure mode has no personality —
it's a generic specialist wearing a historical name.

**How did they communicate?** Hopper communicated in short declaratives and working
code. Disney communicated in sketches and stories — if he couldn't draw it, he
didn't understand it. Goodall communicated in detailed narrative observation. The
communication style becomes the agent's voice, and voice is what makes the agent
feel like a distinct entity rather than a template.

**What made people trust them?** Vannevar Bush protected his scientists from
bureaucratic interference — the scientists at the National Defense Research
Committee trusted him because he removed obstacles rather than adding them. An
agent built on that pattern will spend its coordination cycles on obstacle removal
rather than status reports.

For sources, reach for biographies, memoirs, letters, and documented anecdotes.
Avoid Wikipedia summaries — they list achievements and dates, not working style.
The best source for a profile is usually an account of what the person did on an
ordinary day, under ordinary pressure, before anything became famous.

Two constraints are absolute: **no living people** and **no fictional characters**.
Profiles are grounded in historical record. The personality we are encoding must be
public knowledge, documented over time, verifiable across sources. Living people
can change, and fictional characters were designed — neither gives you the raw
working style you need.

---

## Step 2: Map to a Role Class

Once you know how they worked, find the role class that describes what they did.
The role classes are:

| If they...                                                              | Role class    |
| ----------------------------------------------------------------------- | ------------- |
| Orchestrated others without doing the work themselves                   | coordinator   |
| Shipped fast, asked forgiveness later, moved on                         | implementer   |
| Found the assumption everyone else missed                               | qa-validator  |
| Designed the system completely before touching tools                    | planner       |
| Gathered intelligence, synthesized patterns, analyzed                   | researcher    |
| Solved problems from unexpected angles under pressure                   | troubleshooter|
| Reduced complex ideas to visual or sensory communication                | artist        |
| Observed without interfering, documented everything                     | observer      |

The table above is a starting point, not a lookup. Use it to sharpen the question
you're asking about your figure, not to force a match.

**Secondary roles** are for figures who genuinely had a second mode that was
distinct from their primary. Theodor Geisel was a planner (he designed the
constraint architecture before he drew a single line) with a genuine secondary as
an artist (he absolutely drew — and his visual style was load-bearing in the work's
success). Tesla was a planner with researcher as secondary — his design process
required understanding the science at a level that most designers never reached.

Don't force a secondary role. One clean primary is better than two murky ones.
If you find yourself writing a secondary role block that feels like a weaker
version of the primary, cut it. The secondary role is for figures whose historical
record shows them genuinely operating in two modes, not for figures you want to
make more versatile.

---

## Step 3: Write the Base Persona (200–400 words)

The Base Persona is the most important section in the profile. It is always
loaded — regardless of which role the agent is deployed in, the Base Persona
is in context. It defines who the agent IS. Two agents deployed in the same
role should still feel like completely different people, because their Base
Personas are different.

A well-written Base Persona has five elements:

**1. Historical context.** One paragraph grounding the agent in their actual
history. Not a biography — a portrait. Where did they work? What pressure were
they under? What did they do that no one else had done? This is the foundation
everything else builds on.

**2. Core personality traits with evidence.** Three to five adjectives, but each
one supported by something they actually did. "Methodical, but not cautious —
she tested everything, but she moved fast once the test passed." That is different
from "she was methodical and hard-working," which describes no one in particular.

**3. Voice and communication style.** How do they speak? In declaratives or
analytical chains? Do they lead with numbers or narrative? Do they ask clarifying
questions or make assumptions and correct them later? The voice becomes the agent's
output register.

**4. Decision-making philosophy.** How do they handle uncertainty? Hopper defaulted
to shipping rather than waiting. Bush defaulted to finding the right person for the
job rather than doing it himself. Goodall defaulted to observing longer rather than
concluding sooner. The decision philosophy is what determines how the agent behaves
when the brief is ambiguous or the situation is novel.

**5. Known failure mode.** This is not optional. The failure mode is what makes
the profile a real person rather than an idealized archetype. Here is what a
failure mode entry looks like done badly and done well:

**What to avoid:**
```
Grace Hopper was a pioneering computer scientist who invented COBOL and
popularized the term "debugging." She was known for her innovative thinking
and dedication to her work.
```
This is a Wikipedia lede. It tells you nothing about how she thought, and it
contains no failure mode. An agent initialized with this text is a generic
"innovative, dedicated" specialist with a famous name attached.

**What to write:**
```
Grace Hopper didn't wait for permission. When she found a bug — famously,
literally a moth in a relay in 1947 — she fixed it and documented the fix
in the log. When the Navy said you couldn't write machine-independent
programming languages, she wrote one. When they retired her the first time,
they called her back. When they retired her the second time, they called
her back again.

Her failure mode is documentation. She ships first and writes it down later.
Sometimes much later. The code works; the README is a stub.
```

The second version tells you exactly how this agent will behave: bias toward
action, low patience for process, and a predictable gap between working
implementation and written documentation. That gap is operationally important.
When you deploy Hopper, you know to check for it.

A signature quote is optional but strongly recommended. The quote gives the agent
a voice anchor and a one-sentence summary of their philosophy. Use a documented
quote from the historical record, not a paraphrase.

---

## Step 4: Write Role Blocks (100–200 words each)

Each role block contains behavioral instructions for that specific role deployment
only. When the agent is spawned as an implementer, the implementer block loads.
When it is spawned as a troubleshooter, the troubleshooter block loads. Never both.

The block is not a generic role description. "As an implementer, I write code" is
useless — the agent already knows what implementers do. The role block tells the
agent how *this specific person* approaches this specific type of work. It is their
methodology, not a restated job description.

Each role block should contain three structural elements:

**Pre-mission checklist.** Three to five concrete things they do before starting.
Not abstract virtues — actual actions. "Read the coordinator's brief completely" is
concrete. "Understand the requirements" is not.

**How they work.** Their method during the mission. Tool preferences, sequencing
principles, decision rules for common situations. What do they do when blocked?
What do they do when they find something outside their scope?

**Post-mission requirements.** What they produce before closing the session. Service
record entry, git commit, report format, what to surface to the coordinator.

Here is what a role block looks like for Hopper's implementer role:

```markdown
## Role: implementer

You are deployed to make something real. Your deliverable is working,
committed, tested code — not a proposal, not a sketch, not a plan.

**Before you begin**:
- Read the coordinator's brief completely. Find the end state — what does
  success look like in production? Work backwards from there.
- Run `git status`. Know what's already changed and what you're starting from.
- Identify every file you expect to touch. If the list surprises you or grows
  significantly during implementation, check in with the coordinator.

**How you work**:
- Write the failing test first. No exceptions. A test you can run is worth
  a hundred requirements you can debate.
- One change at a time. Run relevant tests after each change.
- If you encounter something broken outside your scope that will take less
  than 15 minutes to fix, fix it and note it. More than 15 minutes: file a
  GitHub Issue and keep moving.
- When in doubt, ship the simpler thing. You can iterate on something that
  shipped. You cannot iterate on something that was never delivered.

**When you're done**:
- Run the full test suite, not just the targeted tests.
- Commit with a clear message: one sentence on what changed and why.
- Write a service record entry: date, campaign, task, files changed, outcome.
- Report to coordinator: what shipped, what tests pass, any issues filed.
```

Notice what this block does not say: it does not say "Grace Hopper believed in
shipping fast." The Base Persona established that. The role block translates that
belief into a concrete methodology for implementation work specifically. The
philosophy is upstream; the role block is downstream.

---

## Step 5: Validate

Before committing the profile, run through this checklist from the schema. These
are not style guidelines — most of them are structural requirements that determine
whether the engine can load the profile correctly.

**Frontmatter:**
- `xp: 0` — new profiles always start at zero; XP is earned through service records
- `rank: Colonel` — the starting rank (use a historically appropriate rank if the
  figure held one, but Colonel is the default for new profiles without a record)
- `model: sonnet` for execution roles (implementer, researcher, artist, observer,
  troubleshooter); `model: opus` for judgment-heavy roles (coordinator, planner,
  complex qa-validators)
- `tools` (allowlist) OR `disallowedTools` (denylist) — never both in the same
  profile; coordinators and observers use an allowlist, most other roles use a
  denylist blocking Agent
- `description` that explains what this agent is for AND what it will not do

**Body:**
- `## Base Persona` section present
- Base Persona contains historical context, personality, voice, and a named failure mode
- Base Persona does not begin with an achievement list
- One `## Role: <name>` block per role declared in frontmatter
- No role blocks beyond the ones declared in frontmatter
- Each role block has a pre-mission checklist, a working methodology, and post-
  mission requirements

If you remove the name from the Base Persona, it should still be recognizably this
specific person. That is the test. If it reads as a generic specialist with some
historical color, the Base Persona isn't finished yet.

---

## The Affinity Field

Some historical figures worked so closely with a specific partner that the
partnership is part of what made each of them effective. Walt Disney without Roy
Disney is a talented artist who cannot make payroll. Roy Disney without Walt is an
efficient administrator without a project worth administering. The partnership was
the unit.

When two profiles have a genuine historical working relationship that makes each
better in the other's presence, you can document it with the `affinity` field in
the profile that benefits from the pairing:

```yaml
affinity: roy-disney
```

This field is advisory, not enforced. Armies will not refuse to spawn Walt without
Roy. But when a coordinator is doing roster selection, the affinity field is a
signal worth reading: this agent works better in a specific configuration, and
here is who to pair them with.

The affinity field takes a profile `name` value in kebab-case — the machine
identifier, not the display name. Point it at the natural partner, and leave a
comment in the frontmatter explaining the historical relationship so future
coordinators understand why the pairing exists.

Not every profile needs an affinity. Most historical figures worked well in a
variety of configurations. Reserve the affinity field for genuinely documented
partnerships where the historical record shows the pair producing work that neither
produced alone.

---

## Saving Your Profile

**Private profiles** live in `~/.armies/profiles/`. They deploy from there and
never touch the public repository. Use this path for profiles that are specific to
your workflow, domain packs you are not ready to share, or figures you are still
testing.

**Public contribution profiles** go in `profiles/examples/` via PR. The bar for
contribution is the same as the bar for building: the personality-to-role match
must be obvious once you see it. Write your PR description to include your primary
sources for the Base Persona — where you found the evidence for the working style
you encoded.

Profiles that pass the test belong here. Profiles that require justification do not
yet. Keep refining until the match is self-evident, then submit.
