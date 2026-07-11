---
status: completed
branch: feature/default-config-path
issue: IT-44264
note: "Design record — hand-authored (route B), not executed via dark-factory containers. Supersedes the rejected prompted draft of the same number."
---

## Summary

- Default the config file location when neither `--teamvault-config` nor `TEAMVAULT_CONFIG` is set, checking **two candidates in order**: XDG `${XDG_CONFIG_HOME:-~/.config}/teamvault-cli/config.json`, then legacy `~/.teamvault.json`.
- Adopts the **XDG Base Directory** pattern already shipped across the sibling tools (dark-factory, vault-cli, task-watcher, vault-ui, claude-code-router): **XDG-first, legacy-fallback**, no env var or flag required.
- v5 is a breaking major release — the right moment to make the XDG path primary.
- Backward compatible: an existing `~/.teamvault.json` keeps working (fallback); `--teamvault-config` and `TEAMVAULT_CONFIG` still override; JSON format unchanged.

## Problem

`teamvault-cli` applies **no built-in default config path** (`pkg/cli/cli.go` seeds the `--teamvault-config` flag default from `os.Getenv("TEAMVAULT_CONFIG")`, which is empty when unset → `Exists()` false → no config file read). Users must set `TEAMVAULT_CONFIG` or pass the flag every time. Verified this was never a code default in any version (v4.12.0 struct tag carries no `default:`; `git log --all -S'teamvault.json' -- '*.go'` is empty back to tag 1.1.0) — so a default is a new `feat`, not a regression.

Separately, the tool's config lives at home-root (`~/.teamvault.json`) while the rest of the toolchain has migrated to `~/.config/<tool>/` (XDG). teamvault-cli is the last holdout; aligning it removes home-dir clutter and makes the config discoverable where users now look.

## Goal

When the config path is not explicitly provided (flag absent AND `TEAMVAULT_CONFIG` empty), resolve it to the **first existing** of: the XDG path `${XDG_CONFIG_HOME:-$HOME/.config}/teamvault-cli/config.json`, then the legacy `~/.teamvault.json`. If neither exists, no config file is read (unchanged `Exists()` behaviour). Explicit `--teamvault-config` and `TEAMVAULT_CONFIG` keep precedence; field-level precedence (flag/env beats config-file values) is untouched; format stays JSON.

## Non-goals

- A company-specific env var (`TEAMVAULT_SEIBERT`, `TEAMVAULT_SM`) — `TEAMVAULT_CONFIG` already overrides the location.
- Auto-migrating or auto-creating either config file. Absent stays absent (no write).
- Candidate paths beyond these two (no `./.teamvault.json`, no `/etc/...`, no other XDG dirs).
- XDG-compliant **cache/data** dirs — only `~/.config` scope (matches the XDG goal's scope).
- Changing the config file format — JSON stays JSON.
- Field-level precedence changes.

## Desired Behavior

Config-path resolution (location only — field precedence unchanged). XDG path = `${XDG_CONFIG_HOME:-$HOME/.config}/teamvault-cli/config.json`.

| `--teamvault-config` | `TEAMVAULT_CONFIG` | XDG file exists | Resolved config path |
|---|---|---|---|
| set | (any) | (any) | the flag value |
| unset | non-empty | (any) | the env value |
| unset | empty | yes | XDG path |
| unset | empty | no | legacy `~/.teamvault.json` |

1. The resolved path passes through the existing `TeamvaultConfigPath.NormalizePath()` (`~` expansion).
2. The file is read only when `Exists()` is true. If neither XDG nor legacy exists, nothing is read — identical to today.
3. Applies to every subcommand via the shared persistent flag (`password`, `username`, `url`, `file`, `config parse`, `config generate`, `login`).
4. Explicit `--teamvault-url`/`TEAMVAULT_URL` (etc.) still override config-file values — the default only decides *which file* is read.

## Constraints

- Resolver is a small, unit-testable helper: given the `TEAMVAULT_CONFIG` env value, return it when non-empty; else return the XDG path if that file exists; else the legacy path. It performs a filesystem existence check on the XDG candidate — testable by injecting temp `HOME` + `XDG_CONFIG_HOME`.
- Respect `XDG_CONFIG_HOME` (default `$HOME/.config` when unset) — proper XDG behaviour.
- XDG directory name is `teamvault-cli` (matches the sibling `~/.config/vault-cli/` convention), file `config.json`.
- Default strings stay in tilde/`$HOME`-relative form until `NormalizePath` expands them at use time — no pre-expanded absolute paths baked into the flag default.
- No new external dependencies. Public Go API stays source-compatible (CLI-layer change in `pkg/cli`).
- Errors wrapped via `github.com/bborbe/errors`; no `fmt.Errorf`. Exported items have GoDoc per `docs/dod.md`. Tests use Ginkgo/Gomega.

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---|---|---|
| Neither XDG nor legacy present, no flag/env | No config read; url/user/pass from flags/env/Keychain as today | None — unchanged |
| XDG present but malformed JSON | Same wrapped parse error as a malformed explicit config | Fix the file, or point elsewhere via `TEAMVAULT_CONFIG` |
| XDG absent, legacy present | Legacy `~/.teamvault.json` is read (fallback) | None — intended |
| Both present | XDG wins (first candidate); legacy ignored | Remove/relocate XDG, or set `TEAMVAULT_CONFIG` |
| `XDG_CONFIG_HOME` set to a custom dir | XDG candidate resolves under it | — |
| `HOME` and `XDG_CONFIG_HOME` both unset (rare CI) | Both candidates fail to resolve → `Exists()` false → no config read | Provide flags/env, or set `TEAMVAULT_CONFIG` |

## Acceptance Criteria

- [ ] Resolver (`pkg/cli`): env non-empty → env value; env empty + XDG file exists → XDG path; env empty + XDG absent → legacy `~/.teamvault.json`. Respects `XDG_CONFIG_HOME` (default `$HOME/.config`), dir `teamvault-cli`, file `config.json`.
- [ ] Explicit `--teamvault-config <path>` overrides env, XDG, and legacy (cobra flag precedence).
- [ ] Default applies across subcommands via the shared persistent flag (verified for `password` + `login`).
- [ ] Backward compatible: legacy `~/.teamvault.json` is read when XDG is absent; when neither exists, no config read (the `Exists()`-false branch).
- [ ] Unit tests (Ginkgo/Gomega) inject temp `HOME`/`XDG_CONFIG_HOME` and cover all four resolution-table rows + the `Exists()`-false fallthrough.
- [ ] Scenario `scenarios/006-config-path-resolution.md` (`status: active`) validates the shipped binary: XDG-present, XDG-absent-legacy-present, both-absent (no silent default read), env-override, flag-beats-env — keyed on probe `lO4K1w` → `longhorn`.
- [ ] `make precommit` exits 0.
- [ ] Docs updated (README, `docs/getting-started.md`, `skills/teamvault/SKILL.md`): XDG `~/.config/teamvault-cli/config.json` is the primary default, `~/.teamvault.json` the fallback, `TEAMVAULT_CONFIG` overrides both.
- [ ] `CHANGELOG.md` `## Unreleased` `feat` entry describing the XDG-first + legacy-fallback default.

## Verification

- `make precommit` → exit 0.
- XDG-present: with `~/.config/teamvault-cli/config.json` valid, no env/flag → `teamvault-cli --staging username --teamvault-key lO4K1w` reads it.
- Fallback: remove the XDG file, keep `~/.teamvault.json` → same command still resolves (legacy read).
- Override: `TEAMVAULT_CONFIG=/tmp/other.json` and `--teamvault-config /tmp/other.json` each win over the defaults.
- Custom XDG: `XDG_CONFIG_HOME=/tmp/xdg` with `/tmp/xdg/teamvault-cli/config.json` present → read.

## Notes

- Hand-authored (route B) after the dark-factory container path hit repeated friction on this small change (stale spec-005 auto-run, reviewer-bot malfunction, lifecycle tangles). This spec is the design/why record; the code + tests + scenario + docs ship in the same PR.
- Implementation is CLI-layer only: replace the `os.Getenv("TEAMVAULT_CONFIG")` flag default in `pkg/cli/cli.go` with the resolver. `factory.go`/`login.go`/`config-path.go` are untouched — the `Exists()` gate gives "absent file → no-op" for free.
- Aligns with `[[Migrate Configs to XDG Base Directory]]` (XDG-first, legacy-fallback across the toolchain).
