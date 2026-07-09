# teamvault-utils

[![Go Reference](https://pkg.go.dev/badge/github.com/bborbe/teamvault-utils/v5.svg)](https://pkg.go.dev/github.com/bborbe/teamvault-utils/v5)
[![CI](https://github.com/bborbe/teamvault-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/bborbe/teamvault-utils/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bborbe/teamvault-utils/v5)](https://goreportcard.com/report/github.com/bborbe/teamvault-utils/v5)

A single command-line tool for reading secrets from [TeamVault](https://github.com/trustedsec/teamvault) — passwords, usernames, URLs, and files — by their lookup key. Built for humans at a terminal **and** AI coding agents (e.g. Claude Code), as a sanctioned alternative to the 1Password `op` CLI for TeamVault-managed credentials.

## Install

```bash
go install github.com/bborbe/teamvault-utils/v5/cmd/teamvault@latest
```

## Quick start

1. Create `~/.teamvault.json` with your vault URL + user (leave the password out — store it in the Keychain):

   ```json
   { "url": "https://teamvault.your-company.example", "user": "your-username" }
   ```

2. Log in once — verifies your password and stores it in the macOS Keychain:

   ```bash
   teamvault login --teamvault-config ~/.teamvault.json
   ```

3. Read a secret by its key (the alphanumeric ID in the TeamVault web-UI URL, e.g. `…/secret/AbC123/` → `AbC123`):

   ```bash
   teamvault password --teamvault-config ~/.teamvault.json --teamvault-key AbC123
   ```

   Output has **no trailing newline**, so it composes directly:

   ```bash
   curl -u "$(teamvault username … --teamvault-key AbC123):$(teamvault password … --teamvault-key AbC123)" https://api.internal/…
   ```

**→ Full walkthrough: [docs/getting-started.md](docs/getting-started.md)** (config, env vars, direnv, and using it with AI agents).

## Commands

| Command | Purpose |
|---------|---------|
| `teamvault login` | verify credentials and store the password in the macOS Keychain |
| `teamvault password --teamvault-key <KEY>` | print a secret's password |
| `teamvault username --teamvault-key <KEY>` | print a secret's username |
| `teamvault url --teamvault-key <KEY>` | print a secret's URL |
| `teamvault file --teamvault-key <KEY>` | print a secret's file contents |
| `teamvault config parse` | render a template from stdin (`{{ "<KEY>" \| teamvaultPassword }}` placeholders) |
| `teamvault config generate --source-dir <DIR> --target-dir <DIR>` | render a directory of templates |

Shared flags (persistent on every subcommand) — each also reads an env var: `--teamvault-url` (`TEAMVAULT_URL`), `--teamvault-user` (`TEAMVAULT_USER`), `--teamvault-pass` (`TEAMVAULT_PASS`), `--teamvault-config` (`TEAMVAULT_CONFIG`), `--teamvault-timeout` (`TEAMVAULT_TIMEOUT`), `--cache` (`CACHE`), `--staging` (`STAGING`). Run `teamvault <command> --help` for details.

## Using it with AI agents

Have the agent call `teamvault` for credentials instead of embedding secrets in prompts or code — the value is resolved just-in-time and never written to the conversation or the repo. See the [getting-started guide](docs/getting-started.md#6-use-it-with-an-ai-agent-claude-code).

## Claude Code Plugin

teamvault-utils ships a Claude Code plugin — a `teamvault` skill that helps you (or an agent) set up the CLI and fetch secrets from a Claude Code session, with a hard rule to never write a secret into the conversation, a file, or a commit.

```bash
# Install
claude plugin marketplace add bborbe/teamvault-utils
claude plugin install teamvault-utils

# Update
claude plugin marketplace update teamvault-utils
claude plugin update teamvault-utils@teamvault-utils
```

| Command | Description |
|---------|-------------|
| `/teamvault` | Fetch a secret (password/username/url/file) or set up the `teamvault` CLI |

## Go library

`teamvault-utils` is also a Go library (`github.com/bborbe/teamvault-utils/v5`). See **[docs/library.md](docs/library.md)** and the [API reference](https://pkg.go.dev/github.com/bborbe/teamvault-utils/v5).

## Development

```bash
make precommit   # format, generate, test, lint, security checks
```

## License

BSD-style — see [LICENSE](LICENSE).
