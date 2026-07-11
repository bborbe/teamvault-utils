# teamvault-cli

[![Go Reference](https://pkg.go.dev/badge/github.com/Seibert-Data/teamvault-cli/v5.svg)](https://pkg.go.dev/github.com/Seibert-Data/teamvault-cli/v5)
[![CI](https://github.com/Seibert-Data/teamvault-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/Seibert-Data/teamvault-cli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Seibert-Data/teamvault-cli/v5)](https://goreportcard.com/report/github.com/Seibert-Data/teamvault-cli/v5)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Seibert-Data/teamvault-cli)

Read secrets from [TeamVault](https://github.com/trustedsec/teamvault) — passwords, usernames, URLs, and files — by their lookup key. A single `teamvault-cli` binary for humans at a terminal, shell scripts, deployment tooling, **and** AI coding agents (e.g. Claude Code), as a sanctioned alternative to the 1Password `op` CLI for TeamVault-managed credentials.

## Install the CLI

**macOS (recommended)** — via [Homebrew](https://brew.sh):

```bash
brew install seibert-data/tap/teamvault-cli
```

Update later with `brew upgrade teamvault-cli`. The cask installs an unsigned binary and strips the download quarantine, so it runs without a Gatekeeper prompt.

**Any platform** — via Go:

```bash
go install github.com/Seibert-Data/teamvault-cli/v5@latest
```

Installs a `teamvault-cli` binary into `$(go env GOPATH)/bin`.

Check either install: `teamvault-cli --version`.

## Install the Claude Code plugin

Lets Claude Code (or an agent) set up the CLI and fetch secrets from a session, with a hard rule to never write a secret into the conversation, a file, or a commit.

```bash
# Install
claude plugin marketplace add Seibert-Data/teamvault-cli
claude plugin install teamvault-cli

# Update
claude plugin marketplace update teamvault-cli
claude plugin update teamvault-cli@teamvault-cli
```

Then use `/teamvault` in Claude Code to fetch a secret or set up the CLI.

## Configure

`teamvault-cli` reads its server URL and username from `~/.teamvault.json`. Leave the password **out** of the file — store it in the macOS Keychain instead.

```json
{ "url": "https://teamvault.your-company.example", "user": "your-username" }
```

Log in once to verify the password and store it in the Keychain:

```bash
teamvault-cli login
```

Every flag also reads an env var, so config-less use works too: `--teamvault-url`/`TEAMVAULT_URL`, `--teamvault-user`/`TEAMVAULT_USER`, `--teamvault-pass`/`TEAMVAULT_PASS`, `--teamvault-config`/`TEAMVAULT_CONFIG`, `--teamvault-timeout`/`TEAMVAULT_TIMEOUT`, `--cache`/`CACHE`, `--staging`/`STAGING`.

The secret **key** is the alphanumeric ID from the TeamVault web-UI URL (e.g. `…/secret/AbC123/` → `AbC123`).

## Use in shell scripts

Reads print the raw value with **no trailing newline**, so they compose directly in command substitution:

```bash
# Inject a secret into a process's environment
export DB_PASSWORD="$(teamvault-cli password --teamvault-key AbC123)"

# Basic-auth for an API call
curl -u "$(teamvault-cli username --teamvault-key AbC123):$(teamvault-cli password --teamvault-key AbC123)" \
  https://api.internal/…
```

With [direnv](https://direnv.net), put the lookups in `.envrc` so a repo's secrets load on `cd`:

```bash
# .envrc
export DB_PASSWORD="$(teamvault-cli password --teamvault-key AbC123)"
```

## Use in deployments (config templating)

For k8s manifests, config files, or any templated config that needs secrets, keep templates with placeholders in source control and render them at deploy time — the secret values never touch the repo.

A template uses the `teamvaultPassword` / `teamvaultUser` / `teamvaultUrl` functions with a key:

```yaml
# templates/db-secret.yaml
apiVersion: v1
kind: Secret
metadata: { name: db }
stringData:
  password: {{ "AbC123" | teamvaultPassword }}
  username: {{ "AbC123" | teamvaultUser }}
```

Render one template via stdin/stdout, or a whole directory tree:

```bash
# single file
teamvault-cli config parse < templates/db-secret.yaml > out/db-secret.yaml

# whole tree (templates/ → out/, structure preserved)
teamvault-cli config generate --source-dir templates/ --target-dir out/
```

Pipe rendered output straight to `kubectl` if you'd rather not write secrets to disk:

```bash
teamvault-cli config parse < templates/db-secret.yaml | kubectl apply -f -
```

## Use with an AI agent

Have the agent call `teamvault-cli` for credentials instead of embedding secrets in prompts or code — the value is resolved just-in-time and never written to the conversation or the repo. The Claude Code plugin's `/teamvault` skill enforces this. See the [getting-started guide](docs/getting-started.md#6-use-it-with-an-ai-agent-claude-code).

## Command reference

| Command | Purpose |
|---------|---------|
| `teamvault-cli login` | verify credentials and store the password in the macOS Keychain |
| `teamvault-cli password --teamvault-key <KEY>` | print a secret's password |
| `teamvault-cli username --teamvault-key <KEY>` | print a secret's username |
| `teamvault-cli url --teamvault-key <KEY>` | print a secret's URL |
| `teamvault-cli file --teamvault-key <KEY>` | print a secret's file contents |
| `teamvault-cli config parse` | render a template from stdin to stdout |
| `teamvault-cli config generate --source-dir <DIR> --target-dir <DIR>` | render a directory of templates |

Run `teamvault-cli <command> --help` for all flags. Full walkthrough (config, env vars, direnv, agents): **[docs/getting-started.md](docs/getting-started.md)**.

## Go library

`teamvault-cli` is also a Go library — import `github.com/Seibert-Data/teamvault-cli/v5/pkg` (package `teamvault`). See **[docs/library.md](docs/library.md)** and the [API reference](https://pkg.go.dev/github.com/Seibert-Data/teamvault-cli/v5/pkg).

## Development

```bash
make precommit   # format, generate, test, lint, security checks
```

See [CLAUDE.md](CLAUDE.md) for architecture and contributor notes.

## License

BSD-style — see [LICENSE](LICENSE).
