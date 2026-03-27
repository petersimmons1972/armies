# Troubleshooting

Common problems, their causes, and how to fix them. Problems are organized by category.

---

## Setup Problems

### "armies command not found"

**Symptom:** Running `armies` in the terminal returns `command not found` or `armies: No such file or directory`.

**Cause:** The `armies` package is not installed, or the Python environment's `bin/` directory is not in your `PATH`.

**Fix:**
```bash
# Install via pip
pip install armies

# Or install in development mode from the repo
cd ~/projects/armies
pip install -e .

# If installed but not found, check where pip puts executables
python -m armies --help   # bypass PATH issue entirely

# Find where the armies binary landed
python -c "import sysconfig; print(sysconfig.get_path('scripts'))"
# Add that directory to your PATH in ~/.bashrc or ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"
```

---

### "armies roster shows nothing"

**Symptom:** Running `armies roster` prints:
```
No profiles found in /home/user/.armies/profiles
Run armies init to create the directory structure.
```

**Cause:** `~/.armies/profiles/` is empty or doesn't exist yet.

**Fix:**
```bash
# Create the directory structure
armies init

# Then either copy an existing profile or run armies research to create one
cp my-profile.md ~/.armies/profiles/
# or
armies research implementer
# Feed the generated draft to a Claude Code agent
```

---

### "armies init: directory already exists" (or similar)

**Symptom:** Running `armies init` on an existing installation shows `config.yaml already exists — skipping prompt`.

**Cause:** `~/.armies/` was already initialized in a previous session.

**Fix:** This is not an error. `armies init` is idempotent — it creates missing directories and skips creation for anything that already exists. If you want to change your remote URL, edit `~/.armies/config.yaml` directly:

```bash
# Edit config manually
nano ~/.armies/config.yaml

# Config fields:
# remote_url: git@github.com:you/armies-profiles.git
# default_model: sonnet
# profiles_dir: /home/user/.armies/profiles
```

---

### "armies sync: no remote_url configured"

**Symptom:** `armies sync` prints:
```
Error: remote_url not configured
```

**Cause:** `~/.armies/config.yaml` has `remote_url: ""` or the key is missing.

**Fix:**
```bash
# Edit config.yaml and set your GitHub repo URL
nano ~/.armies/config.yaml

# Set:
# remote_url: git@github.com:yourname/your-private-profiles.git

# Then make sure git is initialized and the remote is set
git -C ~/.armies remote add origin git@github.com:yourname/your-private-profiles.git
# or if origin already exists:
git -C ~/.armies remote set-url origin git@github.com:yourname/your-private-profiles.git
```

---

## Profile Problems

### "Role block 'X' not found in profile Y. Available roles: ..."

**Symptom:** `armies spawn grace-hopper --role researcher` prints:
```
Role block '## Role: researcher' not found in grace-hopper.md
Available role blocks:
  Role: implementer
```

**Cause:** The profile's Markdown body does not contain a `## Role: researcher` section. The frontmatter may declare `roles.primary: implementer` only, with no secondary role.

**Fix:** Either use a role that the profile declares, or edit the profile to add the role block. Each role listed in frontmatter must have a corresponding body section:

```markdown
---
roles:
  primary: implementer
  secondary: researcher   # ← declare the secondary role here
---

## Role: implementer
...

## Role: researcher       # ← and add the matching body section
...
```

Run `armies roster` to see which profiles exist, then `armies spawn <agent> --role <role>` with a role that matches the profile.

---

### "Profile not appearing in roster"

**Symptom:** A `.md` file is in `~/.armies/profiles/` but `armies roster` doesn't list it (or the roster shows fewer agents than expected).

**Cause:** The most common cause is a YAML parse error in the profile's frontmatter. YAML is whitespace-sensitive and silently fails on certain formatting errors.

**Fix:** Validate the frontmatter manually:

```bash
# Test whether the file's frontmatter parses cleanly
python3 -c "
import yaml
content = open('~/.armies/profiles/my-profile.md').read()
fm = content.split('---')[1]
print(yaml.safe_load(fm))
"
```

Common frontmatter errors:
- **Tabs instead of spaces** — YAML requires spaces for indentation, not tabs
- **Unquoted special characters** — colons, brackets, or `#` in values need quoting:
  ```yaml
  # Wrong:
  description: Planner: strategic thinking
  # Right:
  description: "Planner: strategic thinking"
  ```
- **Inconsistent indentation** — all fields at the same level must use the same indentation depth
- **Missing closing `---`** — frontmatter must be delimited by `---` on both sides

---

### "armies spawn produces empty output"

**Symptom:** `armies spawn` exits without printing anything, or prints only the frontmatter with no body sections.

**Cause:** The profile file exists but is empty, or the body contains no `## Base Persona` or matching `## Role:` sections.

**Fix:** Open the profile and verify it has:

```markdown
---
name: my-agent
...
---

## Base Persona

[Content here — required]

## Role: implementer

[Content here — required if implementer is declared in roles]
```

If the file is truly empty, the profile was likely created as a placeholder. Either populate it or remove it from the profiles directory.

---

### "XP field rejected" or XP not updating

**Symptom:** XP doesn't change, or you see an error when trying to set XP manually.

**Cause:** XP is a read-only field managed by the service record system. It is never set directly in the profile frontmatter. Starting XP for a new profile is always `0`.

**Fix:** Do not set XP manually. The correct workflow is:
1. Complete a deployment
2. Append a service record entry to the agent's service record YAML (in `~/.armies/service-records/<name>.yaml` or inline in the profile)
3. The service record entry specifies `xp_earned`, which is then summed to compute total XP
4. Commit the service record update

If you believe XP is wrong, file a GitHub Issue documenting the discrepancy. Do not edit the XP field directly.

---

## Docker Problems

### "docker compose run: volume mount permission denied"

**Symptom:** Running any `armies` command via Docker produces:
```
PermissionError: [Errno 13] Permission denied: '/home/nonroot/.armies/'
```

**Cause:** `~/.armies/` on the host is owned by root or by a different user than the one Docker is mapping the volume to. The Chainguard-based container runs as UID 65532 (nonroot). If the host directory is owned by root, the container user cannot write to it.

**Fix:**
```bash
# Check current ownership
ls -la ~/.armies/

# Fix ownership — give your host user ownership
sudo chown -R $(whoami):$(whoami) ~/.armies/

# Verify
ls -la ~/.armies/
```

If the directory doesn't exist yet, create it first:
```bash
mkdir -p ~/.armies
armies init   # or run via Docker: docker compose run --rm armies init
```

---

### "SSH key not forwarded" / "git@github.com: Permission denied (publickey)"

**Symptom:** `armies sync` fails inside Docker with a permission denied error from GitHub, even though your SSH key works on the host.

**Cause:** The Docker Compose file forwards your host SSH agent via `SSH_AUTH_SOCK`, but if the environment variable is not set on the host, no key is forwarded.

**Diagnosis:**
```bash
# Check whether SSH agent is running on the host
echo $SSH_AUTH_SOCK
# Should print a path like: /run/user/1000/keyring/ssh
# If it's empty, the agent is not running

# Check which keys are loaded
ssh-add -l
```

**Fix:**

On **macOS**:
```bash
# Load key into the macOS keychain-backed agent
ssh-add --apple-use-keychain ~/.ssh/id_ed25519
```

On **Linux**:
```bash
# Start the agent if not running
eval $(ssh-agent -s)

# Add your key
ssh-add ~/.ssh/id_ed25519

# Confirm SSH_AUTH_SOCK is now set
echo $SSH_AUTH_SOCK
```

Once the host agent has your key loaded, restart the Docker session:
```bash
docker compose run --rm armies sync
```

---

### "armies sync fails inside Docker"

**Symptom:** `armies sync` works on the host but fails inside the container.

**Cause:** Almost always the SSH key passthrough issue above. Also possible: the `SSH_AUTH_SOCK` path from the host doesn't exist at the same path inside the container.

**Fix:** See SSH key section above. If SSH is working but sync still fails, check whether `~/.armies/` is initialized as a git repo inside the container's view of the volume:

```bash
docker compose run --rm armies bash
# (drops into container — only works if using -dev image or base has shell)
# Check git status
git -C /home/nonroot/.armies status
```

If git is not initialized, run `armies init` with your remote URL from inside the container or on the host.

---

## Eligibility Problems

### "BLOCKED: effective malus too high"

**Symptom:** `armies eligible <agent>` shows `BLOCKED` for all roles, or `armies roster` shows an agent as `BLOCKED`.

**Cause:** The agent's computed effective malus is 400 or higher, placing them in the Suspension tier. This means their ledger entries — weighted by share and decay — sum to at least 400 points.

**Fix:**
```bash
# See the full breakdown
armies eligible <agent>
# Shows: effective malus, tier, and per-role gate status

# Inspect the actual ledger entries
cat ~/.armies/accountability/malus-ledger.yaml
```

Non-decaying entries (operational malpractice, insubordination) do not reduce over time. Decaying entries halve every 14 days. The effective malus will naturally decrease as decaying entries age. For non-decaying entries, resolution requires explicit founder action — mark the entry as resolved in the ledger.

---

### "armies eligible: malus-ledger.yaml not found"

**Symptom:** `armies eligible <agent>` prints:
```
Note: Ledger not found at /home/user/.armies/accountability/malus-ledger.yaml.
Showing gates for zero malus.
```

**Cause:** No malus has been recorded yet — the ledger file doesn't exist.

**Fix:** This is normal for new installs. The note is informational, not an error. Effective malus is 0 and all roles show CLEAR.

If you expected the file to exist:
```bash
# Verify the accountability directory was created
ls ~/.armies/accountability/

# Run armies init if the directory is missing
armies init
```

The ledger file is created automatically when a malus entry is first written. Its schema is:
```yaml
- id: MAL-001
  date: 2026-03-01
  raw_malus: 100
  decays: false
  allocation:
    - agent: eisenhower
      share: 100
  type: operational_malpractice
  description: "Brief description of the incident"
```

---

## Git Sync Problems

### "git push rejected: remote has changes"

**Symptom:** `armies sync` shows:
```
✗ push: ! [rejected] main -> main (non-fast-forward)
```

**Cause:** The remote repository has commits that your local `~/.armies/` does not have. This happens when profiles are updated on another machine or directly on GitHub.

**Fix:**
```bash
cd ~/.armies
git pull --rebase
# Resolve any conflicts if present
# Then push manually or re-run armies sync
git push
# or:
armies sync
```

---

### "git remote not set"

**Symptom:** `armies sync` returns:
```
Error: remote_url not configured
```
or git operations fail with `fatal: 'origin' does not appear to be a git repository`.

**Cause:** `armies init` was not completed with a remote URL, or was run without one.

**Fix:**
```bash
# Initialize git in ~/.armies/ if not already done
git -C ~/.armies init

# Add the remote
git -C ~/.armies remote add origin git@github.com:yourname/armies-profiles.git

# Also update config.yaml so armies sync knows the URL
nano ~/.armies/config.yaml
# Set: remote_url: git@github.com:yourname/armies-profiles.git
```
