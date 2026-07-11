---
status: completed
approved: "2026-07-09T14:31:51Z"
generating: "2026-07-09T15:15:36Z"
prompted: "2026-07-09T15:15:36Z"
verifying: "2026-07-11T12:45:37Z"
completed: "2026-07-11T12:45:51Z"
branch: dark-factory/consolidate-cli-into-teamvault-command
issue: IT-44264
---

## Summary

- Replace the seven separate `cmd/teamvault-*` binaries with a single `teamvault` command that exposes the same operations as subcommands.
- The subcommand tree is: `teamvault login`, `teamvault password`, `teamvault username`, `teamvault url`, `teamvault file`, `teamvault config parse`, `teamvault config generate`.
- Migrate the CLI framework from `libservice.MainCmd` + `argument/v2` (which parses via the process-global `flag` set and pollutes `--help` with 40+ Ginkgo flags) to `spf13/cobra` following the coding-plugin go-cli-guide (Multi-Command Binary Pattern) and go-package-layout-guide — `main.go` → `pkg/cli`.
- The `TEAMVAULT_*` env-var + flag contract is preserved byte-for-byte so existing `.envrc`/direnv consumers keep working after they switch the invocation from `teamvault-password …` to `teamvault password …`.
- Fix the documented trailing-newline bug: `password`/`username`/`url`/`file` print the resolved value with **no** trailing newline (currently `fmt.Printf("%v\n", …)` breaks `curl -u` basic-auth).
- Ship as a major version bump to `/v5` (removing the seven binaries and changing the module path are both breaking).

## Problem

`teamvault-utils` ships seven single-purpose binaries. Installing and teaching the tool means seven `go install …/cmd/teamvault-*` lines and seven binary names to remember — friction for humans and, per IT-44264, for AI agents that should use it as the sanctioned TeamVault secret CLI instead of the 1Password `op` CLI. A single `teamvault <verb>` command is one install, one entry point, and one discoverable `--help`. Two concrete pain points compound the ergonomic case: (1) `teamvault-login --help` currently prints 40+ Ginkgo test-runner flags because `argument/v2` parses against the global `flag` set (documented in `specs/ideas/replace-argument-v2-with-stdlib-flag.md`); (2) `teamvault-password`/`teamvault-username` emit a trailing newline that silently breaks `curl -u` basic-auth (documented gotcha). Consolidating onto cobra fixes both as a side effect of the migration.

## Goal

The repo builds a single `teamvault` binary from a root `main.go` that delegates to `pkg/cli.Execute()`. `teamvault --help` lists exactly seven subcommands and only the flags the tool defines (no Ginkgo/glog leakage). Each subcommand reproduces the exact behavior of its former binary — same flags, same env vars, same precedence, same connector call — with `password`/`username`/`url`/`file` emitting the value with no trailing newline. The `cmd/` directory is gone. `go install github.com/bborbe/teamvault-utils/v5@latest` produces the `teamvault` binary. `make precommit` passes and all five scenarios walk clean against the new binary.

## Non-goals

- Changing the resolution logic in the root package or `factory/` (connector construction, config parsing, keychain, cache/timeout precedence). This is a CLI-surface refactor only; `factory.*` and `teamvault.*` APIs are untouched.
- Renaming or changing flag names / env-var names (`--teamvault-key`, `TEAMVAULT_CONFIG`, `--staging`, etc. stay identical at the wire level).
- Adding any new capability, flag, or output format beyond the newline fix.
- Keeping backward-compatible `teamvault-*` shim binaries — they are removed; consumers migrate to `teamvault <verb>` (that migration lands in the consuming repos, outside this repo's scope).
- Transferring the repo to the `seibert-media` org — that is a separate sibling task; v5 is cut under `github.com/bborbe` now.
- Adding a new end-to-end scenario — the existing `scenarios/001`–`005` already cover the behaviors; they are updated to invoke the single binary, not added to.

## Acceptance Criteria

- [ ] Root `main.go` exists and is a thin delegator. Evidence: `test -f main.go` succeeds AND `grep -c 'cli.Execute()' main.go` returns ≥1.
- [ ] `pkg/cli` builds the cobra tree. Evidence: `grep -rn 'cobra.Command' pkg/cli/ | wc -l` returns ≥8 (root + 7 subcommands) AND `grep -rn 'SilenceUsage: *true' pkg/cli/` returns ≥1.
- [ ] The subcommand tree is exactly the seven verbs. Evidence: `go run . --help 2>&1` lists `login`, `password`, `username`, `url`, `file`, and `config`; `go run . config --help 2>&1` lists `parse` and `generate`.
- [ ] Shared flags are persistent on root. Evidence: `go run . password --help 2>&1` shows `--teamvault-url`, `--teamvault-user`, `--teamvault-pass`, `--teamvault-config`, `--staging`, `--teamvault-timeout`, `--cache`, and `--teamvault-key`.
- [ ] Env-var contract preserved: each of the seven shared flags falls back to its env var when the flag is unset. Evidence: a table-driven `pkg/cli` unit test iterates all seven pairs — `TEAMVAULT_URL`→`--teamvault-url`, `TEAMVAULT_USER`→`--teamvault-user`, `TEAMVAULT_PASS`→`--teamvault-pass`, `TEAMVAULT_CONFIG`→`--teamvault-config`, `STAGING`→`--staging`, `TEAMVAULT_TIMEOUT`→`--teamvault-timeout`, `CACHE`→`--cache` — sets each env var, parses with the flag absent, and asserts the resolved value equals the env value.
- [ ] `password`/`username`/`url`/`file` print the value with **no trailing newline**. Evidence: `pkg/cli` unit test captures stdout for the `password` subcommand (Connector mocked to return `secret`) and asserts the buffer equals `secret` exactly (no `\n`); `grep -rn 'Printf("%v\\\\n"' pkg/cli/` returns 0 matches.
- [ ] `login` behavior preserved: no `--teamvault-key`, resolves via flag→config→keychain, probes via `Connector.Search`, writes to keychain, all user-facing output on **stderr**. Evidence: `go run . login --help 2>&1` does NOT list `--teamvault-key`; `pkg/cli` test asserts the login command writes its success message to stderr (stdout empty).
- [ ] `config parse` reads a template from stdin and writes rendered output to stdout unchanged; `config generate` requires `--source-dir` and `--target-dir`. Evidence: `go run . config generate --help 2>&1` shows both flags marked required; `pkg/cli` integration test pipes a template into `config parse` and asserts the rendered stdout.
- [ ] Clean `--help` — no test-runner leakage. Evidence: `go run . password --help 2>&1 | grep -ci ginkgo` returns 0 AND `go run . --help 2>&1 | grep -ci ginkgo` returns 0.
- [ ] Module path bumped to `/v5`. Evidence: `head -1 go.mod` equals `module github.com/bborbe/teamvault-utils/v5` AND `grep -rn 'teamvault-utils/v4' --include='*.go' .` returns 0 matches.
- [ ] `cmd/` directory removed. Evidence: `test ! -d cmd`.
- [ ] Makefile installs the single binary. Evidence: `grep -c 'cmd/teamvault-' Makefile` returns 0 AND `grep -c 'go build -o $(GOPATH)/bin/teamvault ' Makefile` returns ≥1; `make install` produces a `teamvault` binary (`test -x $(go env GOPATH)/bin/teamvault`).
- [ ] `libservice.MainCmd` no longer referenced. Evidence: `grep -rn 'libservice.MainCmd' --include='*.go' .` returns 0.
- [ ] Scenarios updated to the single-binary surface. Evidence: `grep -rln 'teamvault-password\|teamvault-username\|teamvault-login\|teamvault-url\|teamvault-file\|teamvault-config-parser\|teamvault-config-dir-generator' scenarios/` returns 0 files; each scenario builds one binary (`go build -o /tmp/… .`) and invokes it as `teamvault <verb>`.
- [ ] `docs/releasing-teamvault-utils.md` and `README.md` updated: no "seven CLI binaries" prose, no `/v4/cmd/teamvault-*` install lines. Evidence: `grep -c 'seven CLI binaries' docs/releasing-teamvault-utils.md` returns 0 AND `grep -rc '/v4/cmd/teamvault-' docs/ README.md` returns 0.
- [ ] `make precommit` exits 0.
- [ ] `CHANGELOG.md` `## Unreleased` documents the consolidation, the cobra migration, the newline fix, and the v5 breaking change. Evidence: `grep -n '## Unreleased' CHANGELOG.md` returns ≥1 AND `grep -niE 'consolidat|cobra|v5' CHANGELOG.md` returns ≥1 line beneath it.

## Verification

### Container-executable (runs inside the YOLO container at prompt time)

- `make precommit` — format + lint + vet + test + security, exits 0.
- `go run . --help 2>&1` and `go run . config --help 2>&1` — subcommand tree assertions above.
- `go run . password --help 2>&1 | grep -ci ginkgo` returns 0 — clean help.
- `grep -rn 'teamvault-utils/v4' --include='*.go' .` returns 0 — import rewrite complete.
- `test ! -d cmd` — old binaries removed.
- `go test ./pkg/cli/...` — subcommand unit/integration tests pass (env seeding, no-newline, login-to-stderr, config parse round-trip).

### Operator-executable (runs on the host — mandatory release gate per `docs/releasing-teamvault-utils.md`)

Touching `cmd/`, `factory/` scope triggers the FULL scenario walk; `make precommit` is explicitly insufficient (two prior releases shipped green-but-broken).

- `go build -o /tmp/new-teamvault .` — single binary builds.
- Walk `scenarios/001`→`005` by hand against `/tmp/new-teamvault` using the real `~/.teamvault.json` + macOS Keychain:
  - 001: `/tmp/new-teamvault password <key>` and `… username <key>` return the real values; `/tmp/new-teamvault password <key> | xxd | tail -1` shows the last byte is NOT `0a` (no trailing newline).
  - 002: `/tmp/new-teamvault login` stores the password in Keychain; subsequent `… password <key>` reads it with no stdin prompt.
  - 003: cache fallback returns the disk-cached value within the timeout when TeamVault is unreachable and `--cache` is set.
  - 004: no-cache + unreachable → non-zero exit with a timeout-class error.
  - 005: `--teamvault-timeout=-1s` → non-zero validation error before any network call.
- `go install github.com/bborbe/teamvault-utils/v5@latest` (post-merge, post-tag) produces a working `teamvault` binary.

## Desired Behavior

1. A single `teamvault` binary is built from `main.go` (3-line delegator to `pkg/cli.Execute()`); `Execute()` owns the one `context.Background()`, signal handling, and `rootCmd.ExecuteContext(ctx)`.
2. The root command carries the seven shared flags as `PersistentFlags`, each defaulting from its env var (`--teamvault-url`←`TEAMVAULT_URL`, `--teamvault-user`←`TEAMVAULT_USER`, `--teamvault-pass`←`TEAMVAULT_PASS`, `--teamvault-config`←`TEAMVAULT_CONFIG`, `--staging`←`STAGING`, `--teamvault-timeout`←`TEAMVAULT_TIMEOUT`, `--cache`←`CACHE`). `--teamvault-pass` is never logged.
3. `password`/`username`/`url`/`file` each take a required `--teamvault-key`, build the connector via `factory.CreateConnectorWithConfigAndTimeout(...)`, call the matching `Connector` method (`Password`/`User`/`Url`/`File`), and write the result to stdout with **no trailing newline**.
4. `login` takes no key and no cache, resolves url/user/pass manually (flag → config file → keychain), probes credentials via `Connector.Search`, writes the validated password to the macOS Keychain, and emits all status/prompt output on stderr (never stdout) — identical to today's `teamvault-login`.
5. `config parse` reads a template from stdin and writes the rendered output to stdout via `teamvault.NewConfigParser(conn).Parse`; `config generate` takes required `--source-dir`/`--target-dir` and runs `teamvault.NewConfigGenerator(NewConfigParser(conn)).Generate`.
6. The module path is `github.com/bborbe/teamvault-utils/v5`; all internal imports are rewritten; `cmd/` is deleted; the Makefile `install` target builds the one `teamvault` binary.
7. `teamvault --help` and every `teamvault <verb> --help` print only the tool's own flags — no Ginkgo/glog/global-flag pollution — because cobra/pflag use a private flag set.

## Constraints

- The root-package public API (`Connector` interface, `factory.Create*` functions, `teamvault.NewConfigParser`/`NewConfigGenerator`/`NewKeychain`, typed wrappers `Key`/`Url`/`User`/`Password`/`Staging`/`TeamvaultConfigPath`/`SourceDirectory`/`TargetDirectory`) MUST NOT change. Only the `cmd/` → `main.go`+`pkg/cli` surface changes.
- Every flag name, env-var name, default, and resolution precedence (flag > env > config-file > keychain) is preserved exactly.
- Errors wrapped via `github.com/bborbe/errors` (`errors.Wrapf`/`errors.Errorf`); never `fmt.Errorf`, never `errors.Wrapf(ctx, nil, …)`.
- cobra + pflag only (a private flag set — this is what keeps `--help` clean); no stdlib `flag`; no glog in the new binary — use `slog` to stderr. Layout: a 3-line `main.go` delegating to `pkg/cli.Execute()`; `pkg/cli` holds `Execute()` + `NewRootCommand(ctx)` + one `createXxxCommand(ctx) *cobra.Command` factory per subcommand, flat in `pkg/cli/` (no per-subcommand sub-packages). `Execute()` owns the sole `context.Background()` + signal handling; `RunE` carries business logic. Reference guides live in-container at `/home/node/.claude/plugins/marketplaces/coding/docs/go-cli-guide.md` and `go-package-layout-guide.md`. All exported items keep GoDoc comments per `docs/dod.md`.
- Tests use Ginkgo/Gomega; mocks via Counterfeiter under `mocks/`. Subcommand handlers must be testable without `os.Exit` (factory functions return `*cobra.Command`; run logic returns `error`).
- The five scenarios continue to pass; they are edited to invoke the single binary, not rewritten in intent.
- `pkg/cli` must not import test-only packages at package scope (avoid re-introducing global-flag pollution).

## Failure Modes

| Trigger | Expected behavior | Detection | Recovery |
|---|---|---|---|
| Required `--teamvault-key` missing on `password`/`username`/`url`/`file` | cobra `MarkFlagRequired` → non-zero exit with `required flag(s) "teamvault-key" not set`, no network call | stderr message + exit code ≠ 0 | User supplies `--teamvault-key` |
| Required `--source-dir`/`--target-dir` missing on `config generate` | non-zero exit naming the missing flag | stderr + exit ≠ 0 | User supplies both flags |
| Negative `--teamvault-timeout` | factory-boundary validation rejects before any network call (scenario 005) | non-zero exit, validation error | User passes a non-negative duration |
| Env var set but flag also passed | flag wins (precedence preserved) | unit test asserts flag overrides env | n/a — intended |
| TeamVault unreachable, `--cache` on | disk-cached value returned within timeout (scenario 003) | value on stdout, exit 0 | n/a — intended fallback |
| TeamVault unreachable, no cache | non-zero exit, timeout-class error (scenario 004) | stderr + exit ≠ 0 | User retries when reachable, or enables cache |
| Keychain locked during `login` | wrapped "keychain unlock" error on stderr, exit ≠ 0 | stderr message | User unlocks Keychain, re-runs `login` |

## Suggested Decomposition

Prompts should be generated in this order — each row is a single prompt with a clear scope.

| # | Prompt focus | Covers DBs | Covers ACs | Depends on |
|---|---|---|---|---|
| 1 | Module bump to `/v5` (go.mod + rewrite all internal imports); no behavior change yet, cmd/ still present and building | 6 (partial) | v5-path AC, import-rewrite AC | — |
| 2 | `pkg/cli` root + persistent flags (env-seeded) + `password`/`username`/`url`/`file` subcommands with no-newline fix; `main.go` delegator; unit tests | 1, 2, 3, 7 | main.go, pkg/cli tree, persistent flags, env seeding, no-newline, clean-help ACs | 1 |
| 3 | `login` + `config parse` + `config generate` subcommands; delete `cmd/`; Makefile install target; tests | 4, 5, 6 | login, config, cmd-removed, Makefile, libservice-gone ACs | 2 |
| 4 | Update `scenarios/001`–`005`, `docs/releasing-teamvault-utils.md`, `README.md`, `CHANGELOG.md` to the single-binary/v5 surface | — | scenarios, docs, changelog ACs | 3 |

Rationale: prompt 1 is a mechanical rename that keeps the tree green; prompt 2 establishes the cobra skeleton + the highest-value subcommands (the four secret-readers) alongside the newline fix; prompt 3 finishes the remaining subcommands and removes the old surface; prompt 4 is docs/scenarios only, once the CLI is final.

## Do-Nothing Option

Leave the seven binaries as-is. Cost: IT-44264's request (a single sanctioned TeamVault CLI for AI agents) is unmet; agents keep using `op` or juggling seven binaries; the Ginkgo-polluted `--help` and the `curl -u` newline bug persist; the sibling agent-skill task (which wraps a single `teamvault` command) is blocked. The idea spec `specs/ideas/replace-argument-v2-with-stdlib-flag.md` would still fix the `--help` pollution but not deliver the one-command goal — this spec supersedes it.

## Notes

- This spec **supersedes** `specs/ideas/replace-argument-v2-with-stdlib-flag.md`: that idea kept seven single-purpose binaries and swapped `argument/v2` for stdlib `flag`; the cobra consolidation delivers the same clean-`--help` fix AND the single-command goal. Remove or mark that idea superseded when this spec approves.
- cobra is mandated over stdlib `flag` by `docs/go-cli-guide.md` (Multi-Command Binary Pattern) precisely because it uses a private flag set, which is what eliminates the global-flag / Ginkgo leakage.
- v5 is a major bump (breaking): the seven binaries are removed and the module path changes. Release is delegated to the github-releaser (`.dark-factory.yaml autoRelease: false`, `.maintainer.yaml autoApprove: true`); the operator cuts `v5.0.0` per the manual major-bump procedure in `docs/releasing-teamvault-utils.md` after merge.
- Consumer `.envrc`/direnv migration (`teamvault-password …` → `teamvault password …`) lands in the consuming repos (Brogrammers, sm-octopus) as separate direct edits — out of this repo's scope.
