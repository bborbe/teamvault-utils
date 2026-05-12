---
status: idea
created: 2026-05-12
---

# Replace github.com/bborbe/argument/v2 with stdlib flag in all teamvault-utils CLI commands

## Summary

- The current `cmd/teamvault-*/main.go` files use `libservice.MainCmd` from `github.com/bborbe/service`, which uses `github.com/bborbe/argument/v2` under the hood to parse struct-tagged fields into CLI args + env vars.
- `argument/v2` calls `flag.Parse()` on the global `flag` set, which means `teamvault-* -h` prints every flag any transitive import has registered — including a wall of Ginkgo test flags (verified on `teamvault-login`, see audit notes below).
- Replace `libservice.MainCmd` + `argument/v2` with a small hand-rolled stdlib `flag` parser per command, using a fresh `flag.NewFlagSet` so global registrations are invisible.
- Net effect: clean `--help` output per binary, no behavioral change to flag/env resolution, one transitive dep dropped.
- Style: follow the minimal pattern used by `~/Documents/workspaces/dark-factory/main.go` (hand-rolled `ParseArgs`, custom `printHelp`) and `~/Documents/workspaces/vault-cli/main.go` (delegates to a `cli.Execute()` factory). Both produce clean help; the dark-factory shape is the closer fit for single-command binaries.

## Problem

`teamvault-login --help` prints **40+ Ginkgo flags** (`--ginkgo.seed`, `--ginkgo.randomize-all`, `--ginkgo.parallel.process`, etc.) before any of the binary's own flags. Reason: `argument/v2` parses via the global `flag` set; Ginkgo's `init()` registers all its CLI flags there as a side effect of being imported transitively (via `golang.org/x/tools` → `golang.org/x/term` → some test package).

The help output is unreadable. New users running `teamvault-login --help` to discover the right syntax see test-runner noise instead of the actual flags they need.

Bonus: `argument/v2` plus `libservice.MainCmd` together are doing very little for this project — the binaries each have ~5 flags. The infrastructure cost (one indirect dep, plus the global-flag-set leakage) outweighs the savings.

## Goal

Every `cmd/teamvault-*/main.go` parses its own CLI args using only `flag` from stdlib, against a fresh `flag.NewFlagSet`. `--help` for each binary prints exactly the flags that binary defines, nothing else. `argument/v2` disappears from `go.mod` (drops out of indirect deps once `libservice.MainCmd` is no longer called).

The 7 binaries in scope:

- `cmd/teamvault-config-dir-generator`
- `cmd/teamvault-config-parser`
- `cmd/teamvault-file`
- `cmd/teamvault-login`
- `cmd/teamvault-password`
- `cmd/teamvault-url`
- `cmd/teamvault-username`

## Non-goals

- Adding subcommands (each binary stays single-purpose).
- Switching to `cobra` or other third-party CLI frameworks — overkill for single-command binaries; the dark-factory hand-rolled style is the model.
- Changing flag names or env-var names — backwards compatible. `--teamvault-config`, `TEAMVAULT_CONFIG`, etc. stay identical.
- Removing `libservice` entirely — the package may still be used elsewhere; only the `MainCmd(ctx, app)` call sites in `cmd/teamvault-*/main.go` change.
- Fixing the Ginkgo flag pollution upstream (e.g., wrapping ginkgo imports in `_test.go` only) — out of scope; the migration fixes it as a side effect.

## Desired Behavior

1. Each `cmd/teamvault-*/main.go` defines a fresh `flag.FlagSet` named after the binary (`flag.NewFlagSet("teamvault-login", flag.ExitOnError)`).
2. Each flag is registered explicitly via `fs.StringVar`, `fs.BoolVar`, etc.
3. Each flag has a corresponding env-var fallback: if the flag is unset on the command line, the env var is consulted before applying the default. Pattern: a small `envDefault(envName, fallback)` helper used in the `StringVar` default argument, OR a post-`fs.Parse` pass that fills empty fields from `os.Getenv`.
4. `--help` (or `-h`) prints only the binary's flags. No Ginkgo / glog / other transitive flags leak through.
5. Existing behavior preserved: `--teamvault-config`, `TEAMVAULT_CONFIG`, `--teamvault-pass`, `TEAMVAULT_PASS`, etc. resolve identically to today (including the precedence order: flag > env > config-file > Keychain).
6. After migration, `grep argument go.mod` returns empty (or only `// indirect` lines unrelated to this dep, which the migration removes via `go mod tidy`).

## Constraints

- Public Go API (`Config`, `Connector`, `factory.CreateConnectorWithConfig`, `factory.CreateConnectorWithConfigAndKeychain`, `TeamvaultConfigPath.Parse()`) is unchanged. Only `cmd/teamvault-*/main.go` files change.
- Backwards compatibility: every flag name, every env var, every default value, every `usage:` string stays identical (modulo stylistic rewording where it improves readability — keep the wire-level CLI contract identical).
- No new direct dependencies. Stdlib `flag` only.
- Unit tests: each `main.go`'s flag parsing must be testable. Factor the flag-parsing function out of `main()` so it can be called with a custom `[]string` and `io.Writer` in tests.
- `make precommit` must pass after migration.
- `glog` flags currently registered globally by the `glog` package must still work (`--logtostderr`, `-v=2`, etc. are used in some workflows per the README). Decision: keep `glog` flags registered on the global `flag` set, but route the binary's own flags through a separate `FlagSet`. The binary's `--help` only shows its own flags; `glog` flags work because the binary calls `flag.Parse()` once at startup on the global set (after the per-binary `fs.Parse(...)` succeeded) — OR more cleanly, register glog flags into the per-binary FlagSet via `glog.InitFlags(fs)`.

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---|---|---|
| User passes unknown flag | `flag.ExitOnError` prints usage + exits 2 | User reads usage |
| User passes `-h` / `--help` | Print custom usage (binary name + per-flag table); exit 0 | None |
| Required env var missing AND flag absent (e.g. `--teamvault-key` for `teamvault-password`) | Same as today: connector call fails downstream with the same error | User supplies the value |
| `--teamvault-config` points at non-existent file | Same as today: `TeamvaultConfigPath.Exists()` returns false; factory falls through to flag/env/Keychain | User fixes path |
| Existing scripts in CI / Makefiles passing `--teamvault-pass=...` | Continue to work — flag names unchanged | None |

## Do-Nothing Option

Keep `argument/v2`. Cost: `--help` stays cluttered with Ginkgo flags for every binary. New users hit it on first run. The recently-added `teamvault-login` is the worst affected because its UX explicitly invites the user to discover the flow via `--help`. The unreadable help erodes the "Setup (macOS, recommended)" experience the keychain spec just shipped.

Cost is low per-incident but high in aggregate because `--help` is the single most-used discovery surface for any CLI.

## Acceptance Criteria

- [ ] All 7 `cmd/teamvault-*/main.go` files use stdlib `flag` with `flag.NewFlagSet(<binary-name>, flag.ExitOnError)`.
- [ ] Each binary's `--help` lists only that binary's flags — no Ginkgo, no unrelated test flags.
- [ ] Each binary's flag names + env-var names + default values match today's behavior exactly (verified by `dark-factory:audit-prompt`-style diff of usage strings or by a checked-in `--help` golden file per binary).
- [ ] `glog` flags (`--logtostderr`, `-v`) continue to work — verify with `teamvault-config-parser --teamvault-config ... --logtostderr -v=2 < template.txt` against a known template.
- [ ] `libservice.MainCmd` is no longer called from any `cmd/teamvault-*/main.go`. `grep -rn 'libservice.MainCmd' cmd/` returns empty.
- [ ] `github.com/bborbe/argument/v2` is no longer in `go.mod` after `go mod tidy` (unless pulled in by some other dep — if so, it's at least no longer a direct consequence of teamvault-utils' own code).
- [ ] Unit tests for each binary's flag-parsing function covering: all flags present, flags from env, flags from command line override env, missing required flag (where applicable), `-h` prints to stderr and returns a sentinel.
- [ ] `make precommit` passes.

## Verification

- `make precommit`
- Manual: `teamvault-login --help` prints only the 5 documented flags, no Ginkgo entries.
- Manual: repeat for each of the other 6 binaries — each `--help` is clean.
- Manual: `TEAMVAULT_CONFIG=~/.teamvault.json teamvault-username --teamvault-key=lO4K1w` still works (env-var path).
- Manual: `teamvault-username --teamvault-config ~/.teamvault.json --teamvault-key=lO4K1w` still works (flag path).
- Manual: `teamvault-config-parser --teamvault-config ~/.teamvault.json --logtostderr -v=2 < template.txt` still works (glog flags still registered).
- `git diff --stat go.mod go.sum` shows `argument/v2` removed.

## Notes

- Reference style A: `~/Documents/workspaces/dark-factory/main.go` — hand-rolled `ParseArgs` + custom `printHelp()`. Closest fit for single-command binaries like ours.
- Reference style B: `~/Documents/workspaces/vault-cli/main.go` + `pkg/cli/cli.go` — uses `github.com/spf13/cobra`. Overkill for single-command binaries; included here only as a fallback if cobra-style subcommands ever come up.
- The Ginkgo-leak was discovered during manual smoke-testing of `teamvault-login` after the macOS Keychain feature shipped. See `specs/in-progress/001-macos-keychain-credential-store.md` for context.
- Promotion path: when this idea is picked up, audit it via `/dark-factory:audit-spec`, decompose into one prompt per binary (or one prompt for all 7 if the per-file change is mechanical enough), and execute via dark-factory in the usual flow.
