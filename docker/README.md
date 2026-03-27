# Armies — Docker Packaging

Run the Armies multi-agent coordination engine in a container without installing
Python or any dependencies on your host machine.

---

## Quick Install (3 commands)

```bash
# 1. Clone the repo
git clone https://github.com/your-org/armies.git ~/projects/armies

# 2. Build the image
cd ~/projects/armies/docker
docker compose build

# 3. Add a shell alias so `armies` works anywhere
echo 'alias armies="docker compose -f ~/projects/armies/docker/docker-compose.yaml run --rm armies"' >> ~/.bashrc
source ~/.bashrc
```

That's it. You now have an `armies` command.

---

## Running Armies

All standard CLI commands work exactly the same way:

```bash
armies roster
armies spawn patton --task "Take the bridge"
armies profile eisenhower
armies sync
```

Or run it interactively (drops you into a shell inside the container):

```bash
docker compose -f ~/projects/armies/docker/docker-compose.yaml run --rm --entrypoint bash armies
```

---

## YOUR DATA IS SAFE

**`~/.armies/` is a volume mount — it lives on your host machine, not inside the container.**

This means:

- Removing the container does **not** delete your profiles or configuration.
- Rebuilding the image does **not** delete your profiles or configuration.
- Running `docker compose down --rmi all` destroys the container and image but your data at `~/.armies/` on your host is completely untouched.

Your agent profiles, XP history, malus ledger, and all configuration exist at:

```
~/.armies/          ← host machine, always safe
```

The container only holds the armies source code and its Python dependencies. Think of it like a CD-ROM: you can throw away the disc and your save files are still on your hard drive.

---

## Setting Up GitHub Sync

Armies can sync your agent profiles to a private GitHub repository. Set this up
from inside the container the first time:

```bash
# Start an interactive session
docker compose -f ~/projects/armies/docker/docker-compose.yaml run --rm --entrypoint bash armies

# Inside the container, initialize sync
armies init --sync

# Follow the prompts to connect your GitHub repo
# Your SSH keys are forwarded from the host (see SSH section below)
```

Once configured, the sync settings are written to `~/.armies/config.yaml` on your
host. Every future `armies sync` picks them up automatically.

---

## Uninstalling

### Remove the container and image only (keep your data)

```bash
cd ~/projects/armies/docker
docker compose down --rmi all
```

Your profiles at `~/.armies/` are untouched.

### Remove everything including your profile data

Only do this if you want a completely clean slate:

```bash
cd ~/projects/armies/docker
docker compose down --rmi all

# WARNING: This deletes all your agent profiles and history
rm -rf ~/.armies/
```

Also remove the alias from your shell config if you added one:

```bash
# Edit ~/.bashrc or ~/.zshrc and delete the armies alias line
```

---

## Troubleshooting

### SSH key passthrough for git sync

Armies uses SSH for GitHub sync. The compose file forwards your host SSH agent
into the container via the `SSH_AUTH_SOCK` socket.

Make sure your SSH agent is running and your key is loaded on the host:

```bash
# Check the agent is running
echo $SSH_AUTH_SOCK

# Load your key if needed
ssh-add ~/.ssh/id_ed25519

# Verify the key is available
ssh-add -l
```

If `SSH_AUTH_SOCK` is not set on your host, the compose file falls back to
`/dev/null` (no SSH forwarding). Git operations that require authentication
will fail until you start your SSH agent.

### Permission errors on ~/.armies/

If armies can't write to `~/.armies/`, the directory may not exist yet:

```bash
mkdir -p ~/.armies
```

The container runs as `nonroot` (UID 65532, non-root). The volume mount maps
`~/.armies` on your host to `/home/nonroot/.armies` inside the container.

### Building fails with "No module named armies"

The Dockerfile installs the package from the repo root. Make sure you're
building from the `docker/` directory, which sets the build context to `..`
(the repo root):

```bash
cd ~/projects/armies/docker
docker compose build
```

Do not run `docker build` directly from the repo root without specifying
`-f docker/Dockerfile` — the build context path assumptions will be wrong.

---

## Architecture Note

The `docker-compose.yaml` sets the build context to the parent directory (`..`)
so the full armies source tree is available to `pip install -e .` during the
build. The `Dockerfile` lives in `docker/` but operates on the entire repo.

---

## Security

The Docker image uses [Chainguard Python](https://edu.chainguard.dev/chainguard/chainguard-images/) as the base image — a minimal, distroless-style image with 0-2 CVEs (vs. 40+ for standard Python images). It runs as non-root user UID 65532 by default.

See [docs/security.md](../docs/security.md) for the full security posture.
