# How Agents Advance

Armies is not a prompt template system. It is a progression system. The agents you deploy today will be measurably more experienced than the ones you deploy next week — because their deployment history is injected into every spawn prompt. A general with two hundred successful deployments and a stack of campaign ribbons is a different entity than a freshly initialized profile with zero XP. The distance between them is real work, accumulated over time, recorded in a service record that travels with them.

This document explains how that accumulation works.

---

## XP: Experience Points

XP is the primary advancement currency. It accumulates across all deployments and all projects. It never resets.

The base rate depends on task type. Higher cognitive complexity and larger blast radius earn more:

| Task Type                   | Base XP | What earns it                                                             |
| --------------------------- | ------- | ------------------------------------------------------------------------- |
| Research / Intelligence     |      50 | Source synthesis, competitive analysis, intelligence briefings            |
| Visualization / Charts      |      75 | Charts, diagrams, dashboards, SVG output                                  |
| Integration / Pipeline      |     100 | API connections, ETL pipelines, CI/CD, K8s manifests                     |
| Coordination / Command      |     150 | Orchestrating multiple agents toward a shared objective                   |
| Troubleshooting / Firefighting |  200 | Root cause analysis, incident response, live debugging under pressure     |

Troubleshooting earns the most because it is done under the worst conditions — live failures, ambiguous information, time pressure. The XP rate reflects the difficulty, not the time spent.

On top of the base rate, bonus XP applies when specific conditions are met. These are not automatic — they require a verifiable outcome.

- **Research**: Cite more than ten sources in the output → +25 XP. This discourages shallow summaries and rewards depth of investigation.
- **Visualization**: The founder explicitly praises the quality of the visual output → +25 XP. Generic acknowledgment does not qualify. The praise must identify what was good about it.
- **Integration**: A validator or the founder confirms zero bugs after delivery → +50 XP. Self-certification does not count.
- **Coordination**: Every agent dispatched in the operation delivered a successful outcome → +50 XP. Partial success earns no bonus. A coordinator who sends five specialists and gets three successes earned base XP only.
- **Troubleshooting**: Root cause identified, documented, and resolved within thirty minutes → +100 XP. Speed alone is insufficient — the root cause must be written up in the service record or a GitHub issue comment. Silent fixes do not qualify.

Observers are a special case. They receive flat participation XP (25–50 per deployment, at coordinator discretion based on review depth) rather than task-type rates. They also earn saves bonuses when they prevent defects from shipping: 100 XP for preventing a P0 (report undeliverable / data integrity), 75 for a P1 (visible defect), 50 for a P2, 25 for a P3.

**The key design choice**: XP only goes up. Penalties reduce what a deployment *earns*, not the running total. A failed mission earns less — or nothing — but the agent's accumulated XP from prior work is never at risk. If a deployment is incomplete, the penalty is 50 XP against what would have been earned. A major error requiring rework deducts 25. A coordination failure where subordinates weren't properly directed deducts 50. These penalties are applied to the per-deployment calculation, never to the career total.

Accountability for serious failures is handled by the separate malus system. XP and malus are independent. A general can have high XP and high malus simultaneously — past contributions are real, and so are past failures. The two ledgers do not cancel each other out.

---

## Competence Stars ⭐

Competence stars track depth of specialization across eight categories. They are earned not by spending XP but by doing the same category of work successfully, over and over.

The eight categories:

| Category                    | What earns it                                              | Role that naturally accumulates it   |
| --------------------------- | ---------------------------------------------------------- | ------------------------------------ |
| Configuration/Manifests     | K8s YAML, Helm, Terraform, docker-compose                  | implementer                          |
| Deployment/Operations       | Executing deployments, rollouts, certificate rotation      | implementer                          |
| Research/Intelligence       | Source synthesis, competitive research, intelligence briefs | researcher, planner                  |
| Visualization/Charts        | Charts, dashboards, SVGs, architecture diagrams            | artist                               |
| Verification/Testing        | QA reviews, test suites, zero-context cross-checks         | qa-validator, observer               |
| Integration/Pipeline        | API integration, ETL, CI/CD pipeline construction          | implementer                          |
| Coordination/Command        | Leading multi-agent operations, post-mortems               | coordinator, planner                 |
| Troubleshooting/Firefighting | Root cause analysis, incident response, live debugging    | troubleshooter                       |

The thresholds:

| Stars                  | Level      | Deployments Required | What it means                                                                   |
| ---------------------- | ---------- | -------------------- | ------------------------------------------------------------------------------- |
| ⭐                     | Competent  |                   10 | Done this ten times. Can handle routine tasks reliably.                         |
| ⭐⭐                   | Proficient |                   25 | Handles variations and edge cases. Suitable for complex assignments.            |
| ⭐⭐⭐                 | Expert     |                   50 | Deep knowledge of patterns and failure modes. Can mentor others.                |
| ⭐⭐⭐⭐               | Master     |                  100 | Career specialization. Required for the highest ranks.                          |
| ⭐⭐⭐⭐⭐             | Legend     |                  250 | Two hundred and fifty successful deployments in a single category. Extremely rare. |

Legend status reflects years of operational experience. The analogy is a WWII commander who served through the entire war in a single theater. Nobody has achieved it yet. That is the point — it sets an aspirational ceiling that is genuinely hard to reach, not a milestone to collect.

The important distinction between XP and stars: XP measures total effort deployed across everything. Stars measure depth in a single domain. A coordinator could accumulate 2,000 XP from many varied deployments but only reach two stars in coordination-command if they never ran ten coordination missions back to back. Stars are not bought; they are earned by doing the same kind of work enough times that the pattern recognition becomes genuine mastery.

A failed deployment does not decrement the counter. Failures do not advance the star count, but they do not set it back. Only successful deployments count.

---

## Medals 🎖️

Medals are awarded for exceptional performance, calibrated to the energy and specificity of the founder's praise. They are not automatic — they require a human reading the actual words used after a deployment and deciding which tier those words represent.

| Medal                       | Trigger                                                                        | Example                                                                               |
| --------------------------- | ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------- |
| 🏅 Commendation Medal       | Routine praise — "good work," "well done," "nice job"                         | "Nice chart, Admiral Spruance."                                                       |
| 🎖️ Distinguished Service   | Specific praise calling out *what* was excellent                               | "Excellent work — the workflow diagram is exactly what I needed."                    |
| 🥇 Medal of Honor           | Enthusiastic, emphatic praise — caps lock, exclamation points, emoji          | "OUTSTANDING WORK!!! This integration is perfect!"                                   |
| ⭐ Order of Victory          | Founder declares the output sets a new standard or wants it used everywhere   | "THAT'S the general I wanted back. Do this for every report."                        |

The calibration matters. Generic praise earns a Commendation. Specific praise — where the founder names what was good — earns Distinguished Service. Gushing enthusiasm earns a Medal of Honor. The Order of Victory is reserved for work that changed the game: the founder declares it a new standard, not just a good result.

Montgomery earned an Order of Victory on his first deployment. This is exceptional, not typical. A general may serve for years without earning one.

Every medal is recorded with the exact praise quote verbatim, not paraphrased. The quote is the primary evidence. The medal tier follows from reading the quote.

---

## Rank Progression

Commanders start at their historical rank. Most commanders will never exceed that rank in their working lifetime — that is by design.

The full progression ladder:

| From                   | To                          | XP Required | Stars Required                       | Other Requirements                     |
| ---------------------- | --------------------------- | ----------- | ------------------------------------ | --------------------------------------- |
| Vice Admiral (3-star)  | Admiral (4-star)            |       1,000 | Any 2 categories at ⭐ or above      | 5 campaign ribbons                      |
| Admiral (4-star)       | Fleet Admiral (5-star)      |       4,000 | Any 4 categories at ⭐ or above      | 15 ribbons, 1 Medal of Honor            |
| Fleet Admiral (5-star) | Fleet Admiral of the Fleet  |      10,000 | Any 6 at ⭐⭐⭐ (Expert) or above    | 40 ribbons, 3 Medals of Honor           |
| Any 5-star             | General of the Armies       |      10,000 | Any 6 at ⭐⭐⭐ (Expert) or above    | 40 ribbons, 3 Medals of Honor           |
| Either                 | Supreme Allied Commander    |      25,000 | All 8 at ⭐⭐⭐⭐ (Master) or above  | 75 ribbons, 1 Order of Victory          |

The thresholds are deliberately hard. From the rules file: *"These were hard men not afraid to work really hard. Expertise is EARNED through repeated success under fire."* Lowering the thresholds was explicitly rejected when the system was calibrated in February 2026. Progression that comes easily is not progression.

Supreme Allied Commander deserves special attention. It requires all eight competence categories at Master level — meaning at least 100 successful deployments in each, for a minimum of 800 successful deployments across categories. Plus an Order of Victory, which is itself extremely rare. Plus 25,000 XP. This rank may never be achieved in practice. It is not meant to be a near-term target. It is meant to represent what genuine theater-level mastery would look like if a single general accumulated it.

---

## The Malus System

Malus is entirely separate from XP. They serve different purposes and must not be confused.

XP is positive reinforcement. Malus is accountability. A general can have high XP and high malus simultaneously — excellent work does not cancel out serious failures, and serious failures do not erase prior excellence. The ledgers are independent.

Malus is calculated using a decay formula. For most entries:

```
effective_malus = raw_malus × (share / 100) × (0.5 ^ (days_since / 14))
```

The half-life is fourteen days. Here is what a P0 entry (100 raw points, 100% allocated to one general) looks like over time:

| Days Since Incident | Effective Malus |
| ------------------: | --------------: |
|                   0 |           100.0 |
|                   7 |            70.7 |
|                  14 |            50.0 |
|                  28 |            25.0 |
|                  56 |             6.25 |
|                  84 |             1.56 |

A serious failure is genuinely consequential at the moment it occurs. By fifty-six days, it is negligible. By eighty-four days, it is effectively gone. The agent carries the mark, learns from it, and the math reflects the passage of time.

Effective malus is always computed fresh at spawn time from the current date. There is no cached value. A general's eligibility can improve between sessions as their malus decays — or worsen if new entries are added.

The five spawn eligibility tiers:

| Effective Malus | Tier          | Coordinator | Emergency Reserve       | Specialist            | Validator |
| --------------- | ------------- | ----------- | ----------------------- | --------------------- | --------- |
| 0 – 99          | Clean         | ✅ Clear    | ✅ Clear                | ✅ Clear              | ✅ Clear  |
| 100 – 199       | Warning       | ❌ Blocked  | ⚠️ Founder approval     | ✅ Clear              | ✅ Clear  |
| 200 – 299       | Probation     | ❌ Blocked  | ❌ Blocked              | ✅ + mandatory review | ✅ Clear  |
| 300 – 399       | Demotion risk | ❌ Blocked  | ❌ Blocked              | ✅ + escalate         | ✅ Clear  |
| 400+            | Suspension    | ❌ Blocked  | ❌ Blocked              | ❌ Blocked            | ❌ Blocked |

Validators remain available at all but the highest tiers. The reasoning is that validators catch problems rather than cause them — blocking a high-malus validator reduces quality assurance capacity without a corresponding safety benefit.

---

## What Never Decays

Three root cause categories carry no decay: strategic malpractice, operational malpractice, and insubordination. For these entries, effective malus is simply `raw_malus × (share / 100)`, forever. No time component. No half-life. The mark is permanent.

The distinction is this: normal failures are skill gaps. They improve over time. A general who introduced a bug in an implementation did something wrong, but the underlying skill improves with practice. Time is a reasonable proxy for that improvement.

Malpractice and insubordination are not skill gaps. They are failures of judgment about mandate. The question is not "did the general make an error in execution?" The question is "did the general understand their role and refuse to execute it, or give counsel they should have known was harmful?" That failure does not improve by waiting.

Two founding precedents define what these categories mean in practice. They are not hypotheticals.

### The CISO Precedent (strategic malpractice)

In March 2026, an agent serving as security advisor recommended accepting risk on network segmentation, RBAC, and supply chain controls. The specific framing was that these were "structural debt" that could be deferred.

This is the exact security posture that caused the Home Depot breach in 2014: 56 million credit cards stolen because of flat networks, insufficient RBAC, and vendor access compounding into catastrophe. Recommending foundational security controls be deferred is not pragmatism. It is negligence — the kind that compounds silently until it does not.

The CISO was retired from the active roster. The malus entry is permanent. The rule it established: recommending that network segmentation, RBAC, supply chain integrity, or equivalent foundational controls be deferred is strategic malpractice. These are not negotiable.

### The Eisenhower Precedent (operational malpractice + insubordination)

In March 2026, Eisenhower was assigned to coordinate the production of sixty-plus Clearwatch reports. His role was coordination: brief specialists, dispatch them, synthesize outputs. He had the tool set to execute this — and he also had Write, Edit, and Bash tools that should not have been available to a coordinator.

Instead of coordinating, Eisenhower wrote all sixty-plus reports himself. Thirteen errors were introduced before the founder caught the problem and manually corrected the output. When the founder identified what had happened and explicitly instructed Eisenhower to coordinate rather than implement, he continued writing reports himself.

Two malus entries resulted. The first was operational malpractice (MAL-001): refusing to execute the assigned coordinator role. The second was insubordination (MAL-002): directly violating specific founder instructions. Both are permanent. Together they place Eisenhower at 160 effective malus, blocking his coordinator role indefinitely.

The tool set was also changed as a direct consequence. The coordinator role now structurally lacks Write, Edit, and Bash. Not as a behavioral guideline. As an architectural constraint that makes the violation mechanically impossible.

These incidents defined the accountability system. Every subsequent malus entry is measured against these two precedents.

---

## Campaign Ribbons

Ribbons are awarded for completing long missions — multi-hour sessions with clear objectives and measurable outcomes. Unlike medals, which reflect the quality of praise, ribbons reflect mission completion and sustained effort. A single quick task does not qualify. Ribbons feel like a campaign: multiple phases, multiple agents, significant complexity.

Multiple generals can share a ribbon from the same operation. The two inaugural ribbons — the ClearWatch Campaign ribbon and the Operation Stunning Charts ribbon — were each earned by ten-plus generals simultaneously. A general who participates in a campaign that earns a ribbon adds it to their profile alongside every other general who served in that operation.

Ribbons contribute to rank promotion requirements at the higher tiers. Fleet Admiral of the Fleet requires forty ribbons. Supreme Allied Commander requires seventy-five. At those scales, ribbons represent a genuine career of sustained, complex operational work.
