# Getting Started with `teamvault-cli`

`teamvault-cli` is a single command-line tool for reading secrets from a [TeamVault](https://github.com/trustedsec/teamvault) instance — passwords, usernames, URLs, and files — by their lookup key. It's built for both humans at a terminal and AI coding agents (e.g. Claude Code), as a company-sanctioned alternative to the 1Password `op` CLI for TeamVault-managed credentials.

One binary, a handful of subcommands, and your secret never has to sit in plaintext in a shell history or a repo.

## 1. Install

Pick the one for your OS. Each drops a single `teamvault-cli` binary on your `PATH`.

**macOS — [Homebrew](https://brew.sh) (recommended):**

```bash
brew install seibert-data/tap/teamvault-cli
```

Update later with `brew upgrade teamvault-cli`. The cask installs an unsigned binary and strips the download quarantine, so it runs without a Gatekeeper prompt.

**Linux — prebuilt release binary (no Go toolchain needed):**

```bash
curl -sSL "https://github.com/Seibert-Data/teamvault-cli/releases/latest/download/teamvault-cli_linux_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz" | tar xz teamvault-cli && sudo install teamvault-cli /usr/local/bin/ && rm teamvault-cli
```

To update, re-run the command. (`x86_64` and `arm64`/`aarch64` are both covered.)

**Any platform — Go toolchain:**

```bash
go install github.com/Seibert-Data/teamvault-cli/v5@latest
```

Puts the binary in `$(go env GOPATH)/bin`.

Verify whichever you picked:

```bash
teamvault-cli --version
```

## 2. Configure

`teamvault-cli` needs to know your TeamVault URL and username. The password is best left out of the config file and stored in your macOS Keychain via `teamvault-cli login` (see step 3).

Create `~/.teamvault.json`:

```json
{
  "url": "https://teamvault.your-company.example",
  "user": "your-teamvault-username"
}
```

Point the tool at it with `--teamvault-config ~/.teamvault.json`, or export `TEAMVAULT_CONFIG=~/.teamvault.json` once (e.g. in your shell profile or a project `.envrc`).

Every setting also has a flag and an environment variable, so you can skip the config file entirely if you prefer:

| Flag | Env var | Meaning |
|------|---------|---------|
| `--teamvault-url` | `TEAMVAULT_URL` | TeamVault base URL |
| `--teamvault-user` | `TEAMVAULT_USER` | your username |
| `--teamvault-pass` | `TEAMVAULT_PASS` | password (prefer Keychain via `login`) |
| `--teamvault-config` | `TEAMVAULT_CONFIG` | path to the JSON config above |
| `--teamvault-timeout` | `TEAMVAULT_TIMEOUT` | HTTP timeout (e.g. `5s`, `30s`) |
| `--cache` | `CACHE` | serve from a local disk cache if TeamVault is unreachable |
| `--staging` | `STAGING` | use fixture values instead of the real API |

Precedence is **flag → environment variable → config file**.

## 3. Log in (store your password in the Keychain)

```bash
teamvault-cli login
```

This prompts for your TeamVault password (input hidden), verifies it against the server, and stores it in your **macOS Keychain**. After that, you never pass `--teamvault-pass` again — every command reads the password from the Keychain automatically. (Keychain storage is macOS-only today; on other platforms, supply the password via `TEAMVAULT_PASS` or the config file.)

## 4. Read a secret

Every secret in TeamVault has a short **lookup key** — the alphanumeric ID in the TeamVault web UI URL when you open a secret (e.g. `https://teamvault.…/secret/AbC123/` → key `AbC123`).

```bash
teamvault-cli password --teamvault-key AbC123
teamvault-cli username --teamvault-key AbC123
teamvault-cli url      --teamvault-key AbC123
teamvault-cli file     --teamvault-key AbC123
```

Output is the raw value with **no trailing newline**, so it drops straight into other commands:

```bash
curl -u "$(teamvault-cli username --teamvault-key AbC123):$(teamvault-cli password --teamvault-key AbC123)" https://api.internal/…
```

## 5. Use it in projects with direnv

Instead of copying secrets into `.env` files, resolve them at shell-entry. In a project's `.envrc`:

```bash
export TEAMVAULT_CONFIG="$HOME/.teamvault.json"
export DB_PASSWORD="$(teamvault-cli password --teamvault-key AbC123)"
```

The secret lives only in memory for the session and never touches disk.

## 6. Use it with an AI agent (Claude Code)

When an agent needs a credential, have it call `teamvault-cli` rather than embedding secrets in prompts or code:

```bash
teamvault-cli password --teamvault-key AbC123
```

The agent gets the value it needs, the secret is resolved just-in-time from TeamVault, and nothing sensitive is written into the conversation or the repository. This is the sanctioned replacement for ad-hoc `op` usage on company-managed secrets.

## 7. Config templating (optional)

Render config files with secrets pulled from TeamVault at build time:

- `teamvault-cli config parse` — reads a template from stdin, writes the rendered result to stdout.
- `teamvault-cli config generate --source-dir templates/ --target-dir out/` — renders every file in a directory tree.

## Command reference

| Command | Purpose |
|---------|---------|
| `teamvault-cli login` | verify credentials and store the password in the macOS Keychain |
| `teamvault-cli password --teamvault-key <KEY>` | print a secret's password |
| `teamvault-cli username --teamvault-key <KEY>` | print a secret's username |
| `teamvault-cli url --teamvault-key <KEY>` | print a secret's URL |
| `teamvault-cli file --teamvault-key <KEY>` | print a secret's file contents |
| `teamvault-cli config parse` | render a template from stdin |
| `teamvault-cli config generate --source-dir <DIR> --target-dir <DIR>` | render a directory of templates |

Run `teamvault-cli <command> --help` for the full flag list on any subcommand.
