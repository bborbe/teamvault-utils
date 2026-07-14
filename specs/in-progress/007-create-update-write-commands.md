---
status: prompted
tags:
    - dark-factory
    - spec
approved: "2026-07-14T13:27:53Z"
generating: "2026-07-14T13:37:33Z"
prompted: "2026-07-14T13:44:23Z"
branch: dark-factory/create-update-write-commands
---

## Summary

- Add two write commands to the currently read-only `teamvault-cli`: `create` (make a new secret) and `update <key>` (change an existing one), against TeamVault's real write API at the configured base URL.
- `create` infers the secret's content type from the value flag (`--password-stdin` / `--generate` / `--password` → password, `--file` → file), requires exactly one value source, and prints the new secret's key.
- `update` takes a positional key, sends only the flags the user passed (metadata-only allowed), and creates a new revision when a value flag is given.
- Secret input is secure by default: `--password-stdin` (read from stdin, never argv) and `--generate` (server-side generated) are the recommended sources; `--password <val>` is convenience with a documented shell-history/`ps` leak risk. Secret values are never printed to stdout or written to logs.
- Because Lockbox (API-compatible with TeamVault, no web UI) is the same wire protocol, these commands manage secrets on both. This makes `teamvault-cli` a complete secret manager — no curl, no UI.

## Problem

`teamvault-cli` reads secrets (`password`/`username`/`url`/`file`/`info`) but cannot create or change them. For TeamVault that means every write still goes through the web UI or raw `curl`. For Lockbox — which is deliberately UI-less and API-compatible — there is no supported way to manage secrets at all except hand-rolled HTTP calls. Read shipped in v5.7.0; write is the missing half. Until the CLI can create and update secrets, it is only half a tool and the Lockbox-personal workflow has no first-class client.

## Goal

After this work, `teamvault-cli create` and `teamvault-cli update <key>` write to the TeamVault write API at the configured base URL, using the same Basic-auth credentials, HTTP client, timeout, and error handling as the read path. `create` produces a new secret of the inferred content type and prints its key (or `{"key","api_url"}` with `--json`); `update` changes only the fields the user passed and creates a new revision when a value flag is given. Password input is secure by default (stdin or server-generated), and no secret value ever reaches stdout or a log line. The existing read commands, flags, env vars, and library API are unchanged, and the change ships as a minor (feature) release.

## Non-goals

- Credit-card (`cc`) content type on create/update — deferred; if added later it is a separate spec. `content_type` stays `password` | `file` for this work.
- `delete` / hard-delete — TeamVault exposes no hard-delete API (soft-delete via status is out of scope).
- An `--access-policy` flag, `otp_key_data`, or any `STANDARD_FIELD` beyond `name`/`username`/`url`/`description` — the server default applies; no named consumer requests overriding them.
- Write-path caching or disk-fallback — writes always hit the live server; the cache/disk-fallback connectors decorate reads only.
- The companion change in the `bborbe/lockbox` repo that makes Lockbox's write API TeamVault-compatible — that is a SEPARATE spec in that repo. It is a dependency for the Lockbox-personal use case, referenced below, not delivered here.
- Do NOT add a `--dry-run`, an interactive confirmation prompt, or a write-cache knob — not requested; if a future consumer demands one, that is a separate spec.

## Desired Behavior

1. `create --name <N> [--username <U>] [--url <URL>] [--description <D>] (--password-stdin | --generate | --password <P> | --file <PATH>)` posts a new secret to the write API. The content type is inferred from the value flag: `--password-stdin`/`--generate`/`--password` → `password`; `--file` → `file`. Exactly one value source is required. On success it prints the new key.
2. `update <key> [--name <N>] [--username <U>] [--url <URL>] [--description <D>] [ (--password-stdin | --generate | --password <P>) | --file <PATH> ]` patches the secret named by the positional key. Only the flags the user passed are sent; a value flag creates a new revision; metadata-only updates (no value flag) are allowed.
3. Secure input: `--password-stdin` reads the password from stdin to EOF (never from argv); `--generate` obtains a server-generated password from the generate-password endpoint; `--password <val>` is convenience whose help text names the shell-history/`ps` leak risk. The value sources are mutually exclusive.
4. `--file <PATH>` reads the file's bytes and base64-encodes them into the file content-type payload.
5. No secret value is ever written to stdout or a log line. `create`/`update` print only the key (and, with `--json`, the key + api_url); a generated or supplied password is retrievable afterward only via `teamvault-cli password <key>`.
6. The library gains a write capability that issues the POST/PATCH calls with a request body, reusing the existing Basic-auth header, configured HTTP client, and timeout, and wrapping non-2xx responses with the same error/authentication messaging as the read path.
7. `content_type` is immutable after create. An `update` that would change the content type (e.g. `--file` on a password secret) surfaces the server's rejection as a non-zero exit with the server error — never a silent no-op.
8. `--json` on `create`/`update` prints a single-line object `{"key":"…","api_url":"…"}`; the default (non-JSON) output prints the key for human/script use.

## Constraints

- **Read-only design decision is deliberately reversed.** `CLAUDE.md` currently states "Read-only tool — no write-to-TeamVault … Do not add server-mutating commands." This spec overturns that; the implementation MUST update that section of `CLAUDE.md` rather than silently contradict it.
- **Minor (feature) release only.** Adding a method to the existing exported `Connector` interface is a breaking change for external implementers and would force a major bump the releaser refuses. Write MUST be exposed WITHOUT altering the existing `Connector` interface's five read methods — the read API in `docs/library.md` must still compile unchanged.
- Existing read commands, flag names, env-var names (`TEAMVAULT_URL`/`_USER`/`_PASS`/`_CONFIG`/`_TIMEOUT`, `CACHE`, `STAGING`), config resolution precedence, and `--json` semantics are unchanged.
- Credentials resolve through the same factory wiring the read path uses (flag → env → config → Keychain). The CLI always uses the configured URL — no special targeting for Lockbox vs TeamVault (same wire protocol).
- Errors wrapped via `github.com/bborbe/errors` (`Wrapf`/`Errorf`), never `fmt.Errorf`; no `context.Background()` in business logic; `libtime` types over `time.Now()`. All exported items keep GoDoc. Tests use Ginkgo/Gomega with Counterfeiter mocks under `pkg/mocks/`. Follow `docs/dod.md` and the coding-guidelines cobra/go-cli patterns already used by the read commands in `pkg/cli/cli.go`.
- TeamVault write API shapes are confirmed from the Django source at `~/Documents/workspaces/teamvault/teamvault/teamvault/apps/secrets/api/`: create `POST /api/secrets/` with `{content_type, name, username?, url?, description?, secret_data}`; password `secret_data` = `{password}`; file `secret_data` = `{file_content: <base64>}`; update `PATCH /api/secrets/{hashid}/` with any of `name/username/url/description` and/or a new `secret_data`; generate via `POST /api/generate_password/`. Auth: HTTP Basic (same as the read path).

## Security / Abuse Cases

- The user controls the secret value, the `--file` path, and the metadata (`name`/`username`/`url`/`description`). Request bodies are built with `encoding/json`, so values are encoded, not interpolated — no injection into the API path or body.
- `--password <val>` places the secret on argv, where it leaks via shell history and `ps`. Its help text MUST state this and point to `--password-stdin`/`--generate` as the safe defaults.
- Secret material (stdin password, generated password, `--file` bytes) MUST NOT appear in any `glog`/`slog` line or on stdout. `create`/`update` emit only the key. This preserves the Claude Code plugin's "never write a secret into the conversation/file/commit" rule.
- `--file` reads a path with the invoking user's own filesystem permissions (no privilege boundary crossed); the whole file is base64-encoded in memory (see resource-exhaustion failure mode).
- `--password-stdin` reads stdin to EOF. When stdin is an interactive TTY (nothing piped), the command must not hang indefinitely waiting for input — agent decides at impl time whether to detect a TTY and error, or document "pipe input" — but it must not block forever silently.

## Failure Modes

| Trigger | Expected behavior | Detection | Reversibility | Concurrency |
|---|---|---|---|---|
| TeamVault/Lockbox unreachable during POST/PATCH | non-zero exit, wrapped transport error; no secret created/changed | stderr + exit ≠ 0 | reversible — nothing written | single request; no partial state |
| 401/403 (bad credentials) | reuse existing auth error + `run teamvault-cli login` hint; no write | stderr + exit ≠ 0 | reversible | n/a |
| Server rejects payload (invalid/immutable `content_type`, missing required field) | non-zero exit surfacing the server's status/body; no silent no-op | stderr + exit ≠ 0 | reversible — server rejected | n/a |
| `--file` path missing/unreadable | non-zero exit before any network call | stderr + exit ≠ 0 | reversible — nothing sent | n/a |
| No value source on `create` | validation error before any network call | stderr + exit ≠ 0 | reversible | n/a |
| Two value sources given (e.g. `--generate --password`) | mutual-exclusion validation error before any network call | stderr + exit ≠ 0 | reversible | n/a |
| `--password-stdin` with empty stdin | reject empty password with non-zero exit (mirrors bug-003 empty-password guard) | stderr + exit ≠ 0 | reversible | n/a |
| `--file` far larger than server max | base64 built in memory then server 413/400; large files risk client memory | stderr + exit ≠ 0 on server reject | reversible | n/a |
| Rate limited (429) | surface the 429 status; no infinite retry loop | stderr + exit ≠ 0 | reversible | n/a |
| `--generate` endpoint call fails before the create POST | non-zero exit; no create attempted. Generate is server-stateless (returns a value, stores nothing) → no orphaned secret | stderr + exit ≠ 0 | reversible — nothing created | generate then create are two calls; a crash between them leaves no server state |
| Two concurrent `update <key>` | last write wins; each creates a new revision (TeamVault revisioning) | prior revision retained server-side | partial — both revisions kept | revisions are append-only; no lost-update corruption |

Clock skew is not applicable: no client-supplied timestamp is sent (cc expiration is out of scope).

## Acceptance Criteria

- [ ] `create` and `update` are registered subcommands — evidence: `go run . --help 2>&1` lists both `create` and `update`.
- [ ] `create --help` shows metadata flags (`--name`/`--username`/`--url`/`--description`) and the four value flags, with `--password` help naming the history/`ps` leak risk — evidence: `go run . create --help 2>&1 | grep -c -- '--password'` ≥1 and the leak wording present (`grep -ni 'history\|ps ' `).
- [ ] Content-type inference is correct — evidence: `pkg/cli` unit test asserts `--password`/`--password-stdin`/`--generate` build a `password` payload and `--file` builds a `file` payload.
- [ ] `create` with no value source exits non-zero before any network call — evidence: exit code ≠ 0 from `go run . create --name x` (mocked/no server), stderr names the missing value source.
- [ ] Value sources are mutually exclusive — evidence: exit code ≠ 0 from `go run . create --name x --generate --password p`.
- [ ] `update` allows metadata-only and sends no secret_data when no value flag is passed — evidence: unit test asserts the PATCH body for `update K --description d` contains `description` and NO `secret_data` key.
- [ ] `update` uses a positional key — evidence: `go run . update --help 2>&1` shows `update <key>` usage; unit test drives `update K …` and asserts the request targets the `K` secret path.
- [ ] `--password-stdin` reads the password from stdin, not argv — evidence: `pkg/cli` integration test feeds a password on stdin and asserts it lands in the request body's `secret_data.password`; the value never appears in `os.Args`.
- [ ] `--password-stdin` with empty stdin is rejected before any network call (mirrors bug-003 empty-password guard) — evidence: `printf '' | go run . create --name x --password-stdin` exits ≠ 0 and stderr names the empty password.
- [ ] `--password-stdin` does not hang on an interactive TTY (no piped input) — evidence: unit test with a non-TTY/closed stdin returns promptly (does not block); on a detected interactive TTY the command errors or prompts, never blocks silently forever.
- [ ] `--file` base64-encodes file contents into the file payload — evidence: unit test writes a temp file and asserts the request body `secret_data.file_content` equals the base64 of the bytes.
- [ ] `create`/`update` print only the key (default) or `{"key","api_url"}` (`--json`); no secret value in stdout — evidence: unit test captures stdout and asserts it equals the key (default) / parses to an object with exactly `key`+`api_url` (`--json`), and does NOT contain the input password.
- [ ] No secret value is logged — evidence: `grep -rn 'glog\|slog' pkg/` in the new write code shows no log statement taking a password/file value as an argument (reviewer-checkable) AND the stdout-equality test above.
- [ ] The write library path issues POST (create) / PATCH (update) with a JSON body and the Basic-auth header — evidence: a `pkg` unit/integration test against an `httptest` server asserts method, `Authorization: Basic …` header, and body shape.
- [ ] The existing exported `Connector` interface is unchanged (minor bump preserved) — evidence: `grep -A8 'type Connector interface' pkg/connector.go` still lists exactly `Password`, `User`, `Url`, `File`, `Search` and nothing else.
- [ ] `CLAUDE.md` read-only section is updated to reflect write support — evidence: `grep -c 'no write-to-TeamVault' CLAUDE.md` returns 0 and `grep -ni 'create\|update\|write' CLAUDE.md` returns ≥1 in the design-decisions area.
- [ ] `README.md` command reference and `docs/library.md` document `create`/`update` — evidence: `grep -c 'teamvault-cli create' README.md` ≥1 and `grep -c 'teamvault-cli update' README.md` ≥1.
- [ ] `CHANGELOG.md` `## Unreleased` documents the two new commands as a feature — evidence: `grep -n '## Unreleased' CHANGELOG.md` ≥1 with `create`/`update` beneath it.
- [ ] `make precommit` exits 0.

**Scenario coverage — extend existing scenario 007, no new scenario file.** The create→read round-trip needs a real HTTP POST/PATCH with Basic-auth and a JSON body plus revision creation — unit tests cannot exercise that end to end, and it is the load-bearing user journey. `cmd/fakevault` (the existing in-process hermetic server) gains POST `/api/secrets/` + PATCH `/api/secrets/{key}/` + POST `/api/generate_password/` handlers, and `scenarios/007` is extended with a round-trip: create a secret, then read its password back and assert equality.

- [ ] `scenarios/007` create→read round-trip passes — evidence: `make e2e` prints `e2e: PASS`; the extended scenario creates a secret via `create`, reads it back with `password <key>`, and asserts the value equals the input.

## Verification

```
make precommit
make e2e
go run . --help
go run . create --help
go run . update --help
```

Expected: `make precommit` and `make e2e` exit 0 (`make e2e` prints `e2e: PASS`); `--help` lists `create` and `update`; the extended scenario 007 round-trip passes.

## Suggested Decomposition

| # | Prompt focus | Covers DBs | Covers ACs | Depends on |
|---|---|---|---|---|
| 1 | Library write capability: POST/PATCH-with-body + generate-password call, typed create/update payloads, reused Basic-auth + timeout + non-2xx error mapping; expose WITHOUT touching the `Connector` interface; `pkg` unit/`httptest` tests | 6, 7 | write-library, Connector-unchanged, POST/PATCH-body ACs | — |
| 2 | `create` + `update` cobra commands: content-type inference, mutual-exclusion + required-value validation, `--password-stdin`/`--generate`/`--password`/`--file` (base64), secret-safe key-only + `--json` output; `pkg/cli` unit/integration tests | 1, 2, 3, 4, 5, 8 | command-registered, help/leak-warning, inference, no-value, mutual-exclusion, metadata-only, positional-key, stdin, file-base64, output, no-log ACs | 1 |
| 3 | fakevault write handlers + extend scenario 007 round-trip; update `README.md`, `docs/library.md`, `CLAUDE.md` (reverse read-only), `CHANGELOG.md` | — | fakevault/scenario, docs, CLAUDE.md, README, CHANGELOG ACs | 2 |

Rationale: prompt 1 builds the wire layer with hermetic `httptest` coverage before any CLI exists; prompt 2 layers the two commands and all input/output validation on top; prompt 3 is server-fixture + docs, once the behavior is final. No cycles — each prompt depends only on the previous.

## Do-Nothing Option

Leave `teamvault-cli` read-only. Cost: TeamVault writes keep going through the web UI or raw `curl`; Lockbox — which has no UI by design — has no first-class client at all, so the Lockbox-personal workflow stays blocked on hand-rolled HTTP. The CLI remains half a tool: it can read every secret but change none. Not acceptable given Lockbox is UI-less and this CLI is meant to be its primary management surface.

## Notes

- Dependency (not delivered here): the Lockbox-personal use case also needs the `bborbe/lockbox` write API to be TeamVault-compatible — a SEPARATE spec in that repo. These commands work against TeamVault and any already-compatible Lockbox today; full Lockbox-personal parity waits on that sibling change.
- Candidate new project doc: the confirmed TeamVault write-API field shapes (content types, `secret_data` per type, generate endpoint) could become `docs/teamvault-write-api.md` so future write work references it instead of re-reading the Django source. Flagging, not creating.
