# Security

<div align="center">
<img src="assets/svg/security-classified.svg" alt="CLASSIFIED document style intelligence assessment showing the three identified risk factors (API key exposure, SSH passthrough, profile volume mount) and the Go binary mitigations, with a hidden CLASSIFIED stamp in the corner." width="700">
</div>

<!-- POSTER: Security — Poster 1 — generate from docs/assets/ai-prompts/poster-manifest.md -->
<!-- POSTER: Security — Poster 2 — generate from docs/assets/ai-prompts/poster-manifest.md -->

---

## The Go Binary: No Runtime Attack Surface

Armies v3.0 is a statically compiled Go binary. There is no Python interpreter, no pip packages, no Docker image with CVEs to audit, and no runtime dependency chain that could introduce vulnerabilities.

When you install the binary -- whether via `go install`, a pre-built download, or building from source -- what you get is a single self-contained executable. The binary embeds the bundled example profiles directly; there is no separate installer, no package manager resolving transitive dependencies at runtime, and no virtual environment where a compromised package could execute.

This matters because armies sits at an intersection of risks. It processes Claude API credentials (indirectly, through the environment where it runs), forwards SSH credentials via `armies sync`, and mounts your `~/.armies/` profile store. The attack surface of a single static binary with no external runtime dependencies is dramatically smaller than a Python package that pulls in fifty transitive dependencies.

For supply chain verification, the recommended approach for pre-built binaries is to check the SHA256 checksums published alongside each release:

```bash
# Download the binary and its checksum
curl -L https://github.com/petersimmons1972/armies/releases/latest/download/armies-linux-amd64 -o armies
curl -L https://github.com/petersimmons1972/armies/releases/latest/download/armies-linux-amd64.sha256 -o armies.sha256

# Verify
sha256sum -c armies.sha256
```

A verified checksum means the binary you downloaded matches what was published. Any discrepancy means something changed between the build and your download.

---

## Your Private Profiles

The Armies engine is public. Your profiles are private. This separation is maintained by `~/.armies/` living entirely outside the public repo.

Your profiles never exist in the `~/projects/armies/` directory -- they live in `~/.armies/` on your host machine. The binary contains only the engine and the bundled example profiles (which are public). Your profile data comes in only through the `~/.armies/` directory at runtime.

This separation works as long as you do not accidentally commit `~/.armies/` content to your armies fork. The `.gitignore` in the repo excludes common paths. But if you have added `~/.armies/` to your repo manually (for example, while experimenting with a different layout), check before pushing:

```bash
# Check what is staged
git -C ~/projects/armies status

# Check what is tracked
git -C ~/projects/armies ls-files | grep '.armies'

# If anything from ~/.armies/ appears, remove it from tracking
git -C ~/projects/armies rm --cached path/to/file
echo 'path/to/file' >> ~/projects/armies/.gitignore
git -C ~/projects/armies commit -m "chore: exclude accidentally tracked profile file"
```

If you maintain a private fork of the armies repo that includes your profiles, that is a valid approach -- but be intentional about it. Keep the repo private, protect the branch, and treat any accidental public push as a credential rotation event.

---

## Profile Path Traversal Guards

The binary validates all profile paths before reading them. When you run `armies spawn grace-hopper --role implementer`, the engine resolves the profile path by joining the configured profiles directory with the agent name and `.md`. It does not follow symlinks outside the profiles directory and rejects paths containing `..` components.

This means you cannot accidentally (or intentionally) load a profile from outside `~/.armies/profiles/` by passing a crafted agent name like `../../etc/passwd`. The path is canonicalized and checked before any file read.

---

## YAML Injection in Profiles

Profile frontmatter is parsed as YAML using the Go `gopkg.in/yaml.v3` library. The library is used in safe mode -- it does not execute arbitrary code during parsing. A profile with malformed YAML will fail to parse and be skipped with a warning; it will not cause unexpected behavior or code execution.

That said, the content of a profile becomes part of the spawn prompt that gets pasted into Claude Code. If a profile contains instructions designed to override the agent's behavior in unintended ways -- prompt injection in the Base Persona or Role blocks -- those instructions will be present in the spawn output. You are responsible for reviewing profiles before deploying them, especially profiles sourced from untrusted contributors.

The bundled profiles in `examples/generals/` are reviewed before release. Private profiles you author or receive are your own responsibility.

---

## What Armies Does NOT Do

Be clear about the limits of the security model.

**Armies does not encrypt `~/.armies/` at rest.** Your profiles, malus ledger, service records, and config are plaintext YAML and Markdown on your local filesystem. Anyone with read access to your home directory can read them. If your laptop is stolen or your home directory is accessible to other users, those files are accessible.

**Armies does not authenticate `armies sync` beyond git credentials.** The sync mechanism is a wrapper around `git pull` and `git push`. Your GitHub access is protected by your SSH key (or personal access token). Armies adds no additional authentication layer on top of that. Protect your GitHub token and SSH key as you normally would.

**Armies does not sandbox the spawned Claude agents.** When you paste a spawn prompt into a Claude Code Agent tool call, that agent runs with whatever tools its profile declares (or that Claude Code defaults to). Armies controls which _instructions_ the agent receives, not which _capabilities_ it has at the platform level. An implementer profile that declares `disallowedTools: [Agent]` will have the Agent tool blocked -- but only because Claude Code enforces that, not because Armies enforces it.

**The malus system tracks accountability, not security controls.** Malus is an organizational accountability mechanism. It can block a general from being spawned in certain roles based on past performance. It is not a security gate -- it does not prevent a human from copy-pasting a blocked agent's prompt directly.
