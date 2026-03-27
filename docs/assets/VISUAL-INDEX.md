# Visual Index — Armies Documentation Art Program

Complete catalog of all artwork created for the Armies documentation site.
This index is the authoritative reference for art directors and contributors.

**Art Direction**: WWII propaganda poster style mixed with magazine illustration.
Period: 1940s US War Department, Soviet propaganda, British "Keep Calm" series, Life magazine.
**No historical faces depicted** — metaphor, silhouette, and symbol only.

---

## SVG Illustrations (Created — Ready to Use)

All SVGs are in `docs/assets/svg/`. They are committed to the repo and render natively.

| Filename | Page | Dimensions | Description |
|----------|------|------------|-------------|
| `readme-command-table.svg` | README.md (hero) | 900×420 | Commander silhouette at a war room table covered in agent profile cards (GH, VB, JG, Eisenhower BLOCKED). Mission arrows connect the tokens. Lamp overhead casts dramatic light. **Easter egg**: A pocket watch hangs from the commander's coat pocket (Eisenhower's malus entry carries no decay — time has not erased it). |
| `readme-role-insignia.svg` | README.md (Role Classes section) | 900×160 | Eight octagonal role class insignia badges in WWII military style, each with a role symbol: compass (coordinator), wrench (implementer), magnifying glass (qa-validator), triangle (planner), open book (researcher), lightning bolt (troubleshooter), paintbrush (artist), eye (observer). |
| `getting-started-field-manual.svg` | getting-started.md (hero) | 700×500 | WWII field manual cover design with eagle emblem, five setup commands listed in field-manual monospace format, wire binding on left edge. "FM 1-00" document number. UNCLASSIFIED stamp in corner. |
| `how-it-works-spawn-flow.svg` | how-it-works.md (hero) | 900×380 | War room infographic showing the progressive loading spawn flow: profile.md on disk → merged spawn prompt → Claude executes → armies record closes the feedback loop. Color-coded blocks (navy=frontmatter, green=base persona, blue=role). The troubleshooter role shown as dimmed/staying on disk. |
| `progression-medal-rack.svg` | progression.md (hero) | 800×360 | Left: Four medal displays (Commendation, Distinguished Service, Medal of Honor, Order of Victory) with period-accurate ribbon bars. Right: rank ladder from Vice Admiral to Supreme Allied Commander with XP thresholds. "Expertise is earned through repeated success under fire" at the bottom. |
| `progression-malus-decay.svg` | progression.md (malus section) | 700×320 | Wartime infographic chart showing P0 malus decay curve (100→70.7→50→25→6.3→1.6 over 84 days) with data point annotations. A flat dashed red line shows permanent entries that never decay. Eisenhower's 160.0 permanent malus annotated. |
| `creating-profiles-persona-craft.svg` | creating-profiles.md (hero) | 600×500 | Aged parchment field guide showing the five elements of a Base Persona as numbered entries with oak leaf border decoration. **Easter egg**: A moth (Grace Hopper's debugging moth from 1947) is hidden in the top-right border corner with a tiny logbook tape labeled "BUG". |
| `team-templates-formation.svg` | team-templates.md (hero) | 900×440 | Top-down tactical formation map showing all five team templates as military unit compositions with color-coded agent role tokens (C=coordinator navy, I=implementer orange-red, Q=validator green, T=troubleshooter fire-red, P=planner purple, R=researcher slate, O=observer dark blue). Observer rule callout at bottom. |
| `coordinator-command-room.svg` | coordinator-guide.md (hero) | 800×450 | WWII command room with coordinator silhouette at a dispatch desk. Three operational windows on the wall: BRIEF (scroll/seal icon), VERIFY (magnifying glass/checkmark), RECORD (ledger/quill). **Easter egg**: Vannevar Bush's "As We May Think" (The Atlantic Monthly, 1945) visible as a small document on the desk corner. |
| `accountability-ledger.svg` | accountability.md (hero) | 800×500 | Two heavy ledger books: XP Ledger (gold-edged, open, showing XP entries) and Malus Ledger (red-edged, showing MAL-001/002 with "NEVER" decay status). Two founding precedents inscribed as tablets at the bottom. **Easter egg**: A stopped pocket watch near Eisenhower's permanent entries — "no decay" — time has not passed for these marks. |
| `cli-reference-field-card.svg` | cli-reference.md (hero) | 750×440 | WWII soldier's field pocket reference card on khaki/olive cover. Left column: all primary commands (roster, spawn, record, eligible, init). Right column: supporting commands + eligibility tier quick-reference table + key paths. "If it's not committed — it didn't happen" at the bottom. |
| `troubleshooting-fox-tracks.svg` | troubleshooting.md (hero) | 700×360 | Desert fox silhouette trots across sand dunes that gradually resolve into terminal output text at the bottom. Fox paw prints appear in the command lines. Amber/orange/red alert palette. **Easter egg**: A tiny desert fox paw print hidden in the bottom-right corner at very low opacity. |
| `security-classified.svg` | security.md (hero) | 700×400 | CLASSIFIED intelligence document format showing three risk factors (API key exposure, SSH passthrough, profile access) and Chainguard mitigations. Red classification bars at top and bottom. **Easter egg**: A tiny CLASSIFIED stamp buried in the document corner at very low opacity. |
| `divider-gold-stars.svg` | All pages (section dividers) | 800×40 | Decorative section divider: gold rule lines with center star ornament, flanking stars, and diamond markers. For README and progression pages (warm palette). |
| `divider-army-green.svg` | All pages (section dividers) | 800×40 | Decorative section divider: olive green chevron center with flanking bullet points. For getting-started and creating-profiles pages (field palette). |

**Total SVG files**: 15

---

## AI Poster Prompts (Pending Generation)

All prompts are in `docs/assets/ai-prompts/poster-manifest.md`. Images go in `docs/assets/posters/`.

| Filename | Page | Size | Subject Brief |
|----------|------|------|---------------|
| `poster-readme-identity.png` | README.md (hero) | 900px wide | Silhouetted coordinator at glowing command table, role cards arrayed like a battle map, J. Howard Miller Rosie style |
| `poster-readme-profiles.png` | README.md (Role Classes) | 350px portrait | Three soldiers in formation carrying role tools — map, wrench, binoculars — marching toward viewer |
| `poster-getting-started-init.png` | getting-started.md | 700px wide | Soldier reaching into glowing field crate labeled ~/.armies/, pulling out a profile card |
| `poster-getting-started-spawn.png` | getting-started.md (spawn section) | 400px portrait | Military teleprinter producing spawn prompt scroll with profile data fragments |
| `poster-how-it-works-loop.png` | how-it-works.md (hero) | 900px wide | WWII unit insignia patch design showing the four-phase feedback loop |
| `poster-how-it-works-eisenhower.png` | how-it-works.md (Eisenhower Precedent) | 350px portrait | Coordinator silhouette with ghostly prohibited tools (WRITE/EDIT/BASH) and red prohibition symbols |
| `poster-progression-xp-ladder.png` | progression.md (hero) | 900px wide | Career ladder of service ribbons from recruit to Supreme Allied Commander, summit in clouds |
| `poster-progression-malus.png` | progression.md (What Never Decays) | 700px wide | Two stone tablets side by side: CISO Precedent and Eisenhower Precedent, "PERMANENT" |
| `poster-creating-profiles-match.png` | creating-profiles.md (hero) | 900px wide | Victorian natural history specimen board with three profile cards pinned (Hopper, Goodall, Bush) and unlabeled moth |
| `poster-creating-profiles-failuremode.png` | creating-profiles.md (failure mode section) | 350px portrait | Field notebook with bad entry crossed out in red ink, good failure mode entry circled in green |
| `poster-team-templates-formation.png` | team-templates.md (hero) | 900px wide | Aerial top-down military formation with role-tool silhouettes in V-formation |
| `poster-team-observer-rule.png` | team-templates.md (Observer section) | 700px wide | Observer in isolation circle while other agents exchange whisper-lines around them |
| `poster-coordinator-vannevar.png` | coordinator-guide.md (hero) | 900px wide | Coordinator at hub of radial constructivist dispatch diagram, specialists at edges |
| `poster-coordinator-brief.png` | coordinator-guide.md (Brief section) | 350px portrait | Gloved hand holding complete mission brief with five labeled sections and wax seal |
| `poster-accountability-two-ledgers.png` | accountability.md (hero) | 900px wide | Two ledger books under hanging interrogation lamp, balance scale showing independence |
| `poster-accountability-precedents.png` | accountability.md (What Never Decays) | 700px wide | Two stone monoliths with precedents engraved, clock with no hands between them |
| `poster-cli-field-kit.png` | cli-reference.md (hero) | 900px wide | Field kit inspection from above: terminal, scroll, notebook, ledger, manual on khaki cloth |
| `poster-troubleshooting-firefighter.png` | troubleshooting.md (hero) | 900px wide | Firefighter silhouette running toward burning terminal screen with error output flames |
| `poster-security-submarine.png` | security.md (hero) | 900px wide | Nuclear submarine beneath calm ocean surface — minimum attack surface, maximum capability |
| `poster-security-no-shell.png` | security.md (No shell section) | 350px portrait | Welded-shut gate labeled /bin/bash with no entry point, shadowed figure finds nothing to hold |

**Total poster prompts**: 20

---

## Generation Instructions

1. Open `docs/assets/ai-prompts/poster-manifest.md`
2. Each entry has a **Prompt** section — copy the full prompt text
3. For Midjourney: add `--ar 16:9` (wide), `--ar 1:2` (portrait), or `--ar 7:4` (medium)
4. Add `--stylize 750 --style raw` for period-accurate graphic design output
5. Save generated images as PNG to `docs/assets/posters/[filename]`
6. Uncomment the `<!-- POSTER: ... -->` comment lines in each doc and replace with actual `<img>` tags

---

## Easter Eggs Summary

For readers who look closely:

| Location | What to Find | Reference |
|----------|-------------|-----------|
| `creating-profiles-persona-craft.svg` top-right border | A moth with antennae, tiny logbook tape reading "BUG" | Grace Hopper's 1947 debugging moth in the Mark II relay logbook |
| `accountability-ledger.svg` near Eisenhower entries | A stopped pocket watch labeled "no decay" | Eisenhower's MAL-001/002 are permanent — time has not erased them |
| `readme-command-table.svg` commander's coat pocket | A tiny pocket watch on a chain | Same permanence motif — a coordinator's accountability is measured by the work, not the clock |
| `coordinator-command-room.svg` desk corner | "As We May Think — Vannevar Bush, 1945 — The Atlantic Monthly" | Bush wrote this essay between acts of administration — "information flow" was his operational model |
| `security-classified.svg` corner | A tiny CLASSIFIED stamp at near-invisible opacity | Only visible in source code or on very close inspection — buried in SVG metadata area |
| `troubleshooting-fox-tracks.svg` bottom-right | A fox paw print at 18% opacity | The desert fox finds what's actually there |

---

## Color Palette Reference

| Page | Primary | Secondary | Accent |
|------|---------|-----------|--------|
| README | `#1a0a02` (deep night) | `#c8a050` (gold) | `#d4b483` (cream) |
| getting-started | `#3d4a1a` (olive drab) | `#d8c880` (khaki) | `#8b6010` (brass) |
| how-it-works | `#0a1428` (deep navy) | `#4070c0` (steel blue) | `#e8f0ff` (white-blue) |
| progression | `#1a0f00` (dark brown) | `#c8a050` (gold) | `#c84040` (alert red for malus) |
| creating-profiles | `#1a3a10` (forest green) | `#e8d8a0` (aged parchment) | `#2a5a18` (deep green) |
| team-templates | `#1e2820` (operational grey-green) | `#a0c080` (field green) | `#3a5030` (dark tactical) |
| coordinator-guide | `#050c1e` (deep royal blue) | `#c8a050` (gold) | `#1e3a6e` (command blue) |
| accountability | `#0a0404` (near-black) | `#c84040` (blood red) | `#8b6010` (faded gold) |
| cli-reference | `#5a5020` (khaki cover) | `#d8c880` (field tan) | `#5a4800` (dark khaki) |
| troubleshooting | `#3a1800` (dark amber) | `#c06010` (orange alert) | `#8b0000` (alarm red) |
| security | `#1a1a1a` (charcoal) | `#f0ece0` (document paper) | `#8b0000` (classified red) |
