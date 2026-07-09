---
name: teamvault
description: Fetch secrets (password, username, url, file) from TeamVault via the `teamvault` CLI, and help set it up. Use when the user needs a company-managed credential, references a TeamVault key, wants to replace `op`/1Password for TeamVault secrets, or asks to install/configure/log in to teamvault. Never embeds or echoes secret values into files or the conversation.
---

## What this does

`teamvault` is a single CLI that reads secrets from the company TeamVault by their lookup key. Use it to hand a credential to a command (or to yourself) just-in-time, instead of storing secrets in `.env` files, prompts, or code — the sanctioned alternative to the 1Password `op` CLI for TeamVault-managed secrets.

Full walkthrough for humans: `docs/getting-started.md` in the teamvault-utils repo.

## Prerequisites (check first, set up if missing)

Run `teamvault --help`. If the command is missing:

```bash
go install github.com/bborbe/teamvault-utils/v5@latest   # installs to $(go env GOPATH)/bin
```

Config: `teamvault` needs a URL + username. Check for `~/.teamvault.json`:

```json
{ "url": "https://teamvault.your-company.example", "user": "your-teamvault-username" }
```

Point at it with `--teamvault-config ~/.teamvault.json` or `export TEAMVAULT_CONFIG=~/.teamvault.json`. Every setting also has a flag (`--teamvault-url`, `--teamvault-user`, …) and env var (`TEAMVAULT_URL`, `TEAMVAULT_USER`, …); precedence is flag → env → config file.

Password: prefer the Keychain over the config file. Run `teamvault login` once — it verifies the password and stores it in the macOS Keychain, after which every command reads it automatically.

## Retrieving a secret

A TeamVault secret's **key** is the short alphanumeric ID in its web-UI URL (`…/secret/AbC123/` → `AbC123`). Ask the user for the key if you don't have it — do not guess.

```bash
teamvault password --teamvault-key <KEY>    # the password
teamvault username --teamvault-key <KEY>    # the username
teamvault url      --teamvault-key <KEY>    # the URL
teamvault file     --teamvault-key <KEY>    # a stored file
```

Output is the raw value with **no trailing newline**, so it composes directly:

```bash
curl -u "$(teamvault username --teamvault-key <KEY>):$(teamvault password --teamvault-key <KEY>)" https://api.internal/…
```

## Handling secrets safely (non-negotiable)

- **Never** write a retrieved secret into a file, commit, comment, or the chat transcript. Pipe it directly into the consuming command, or use command substitution as above.
- Prefer `$(teamvault password --teamvault-key <KEY>)` inline over assigning the value to a visible variable.
- Do not log, print, or echo the value to confirm it — confirm success by the consuming command's exit status instead.
- In project setups, resolve secrets at shell-entry via `.envrc` (`export DB_PASSWORD="$(teamvault password --teamvault-key <KEY>)"`) rather than storing them in `.env`.

## Config templating (optional)

Render config files with TeamVault values at build time:

- `teamvault config parse` — reads a template from stdin, writes rendered output to stdout.
- `teamvault config generate --source-dir <DIR> --target-dir <DIR>` — renders a directory tree.

## Success criteria

- `teamvault <sub> --help` runs (CLI installed) and, when needed, config + `teamvault login` are set up.
- The requested secret is retrieved by key and delivered to its consumer without the value ever being written to disk, committed, or printed into the conversation.
