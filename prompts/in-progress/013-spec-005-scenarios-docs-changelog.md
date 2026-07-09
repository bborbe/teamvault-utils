---
status: approved
spec: [005-consolidate-cli-into-teamvault-command]
created: "2026-07-09T15:01:03Z"
queued: "2026-07-09T15:16:04Z"
---

<summary>
- The five end-to-end walkthroughs now build and drive the one consolidated command instead of the seven old separate tools.
- The README teaches a single install and shows every operation as a subcommand of one command; the old "seven binaries" story and the outdated install and import lines are gone.
- The release runbook is updated the same way: one binary to build, one install command, and the new major version everywhere it names how to install or consume the tool.
- The changelog records the consolidation, the framework migration, the trailing-newline fix, the major breaking version bump, and the removal of two environment-variable shortcuts on the config-generation command.
- Documentation and walkthroughs only — no program behavior changes; the command surface is already final from the previous step.
</summary>

<objective>
Bring the documentation and scenario walkthroughs in line with the single `teamvault` binary and the `/v5` module path: rewrite `scenarios/001`–`005`, `README.md`, `docs/releasing-teamvault-utils.md`, and add the `CHANGELOG.md` entry. No Go code changes.
</objective>

<context>
Read `CLAUDE.md` for project conventions. Runs LAST, AFTER prompt 3 (the single `teamvault` binary is final; `cmd/` is gone; `main.go` builds from the repo root).

Read before changing:
- `scenarios/001-teamvault-password-happy-path.md` through `scenarios/005-negative-timeout-rejected.md` — each has `## Setup` steps that `go build -C ~/Documents/workspaces/teamvault/teamvault-utils -o /tmp/new-teamvault-<verb> ./cmd/teamvault-<verb>` and `## Action`/`## Expected`/`## Cleanup` steps that invoke `/tmp/new-teamvault-<verb> …`. The frontmatter is `status: active` — keep it.
- Current per-scenario binary usage:
  - 001: builds `teamvault-password` + `teamvault-username`; invokes both with `--teamvault-config`/`--teamvault-key`.
  - 002: builds `teamvault-login` + `teamvault-password` + `teamvault-username`; runs `login` (stdin-piped) then `password`/`username`.
  - 003: builds `teamvault-password`; two runs with `--cache`.
  - 004: builds `teamvault-password`; one unreachable-no-cache run.
  - 005: builds `teamvault-password`; negative-timeout runs (flag + config).
- `README.md` — has `teamvault-login`/`teamvault-password`/… usage examples and `go get github.com/bborbe/teamvault-utils/v4/cmd/teamvault-*` install lines (around lines 45–68 and 300–400).
- `docs/releasing-teamvault-utils.md` — describes "a Go library + seven CLI binaries (…)", `go install …/v4/cmd/teamvault-*@…` lines, and per-binary `go build -C … ./cmd/teamvault-…` build lines.
- `CHANGELOG.md` — `## Unreleased` section (created in prompt 1, appended in prompts 1–3).

Read this coding guide (in-container path):
- `/home/node/.claude/plugins/marketplaces/coding/docs/changelog-guide.md`

Facts about the new surface:
- Build once from the repo root: `go build -C ~/Documents/workspaces/teamvault/teamvault-utils -o /tmp/new-teamvault .`
- Invoke as: `teamvault password …`, `teamvault username …`, `teamvault url …`, `teamvault file …`, `teamvault login …`, `teamvault config parse`, `teamvault config generate`.
- Flags and env vars are unchanged: `--teamvault-config`, `--teamvault-key`, `--cache`, `--teamvault-timeout`, etc.
- Install: `go install github.com/bborbe/teamvault-utils/v5@latest` produces the `teamvault` binary.
</context>

<requirements>

1. **Rewrite all five scenarios to the single binary.** In each of `scenarios/001`–`005`:
   - Replace the multiple `go build -C … -o /tmp/new-teamvault-<verb> ./cmd/teamvault-<verb>` setup lines with a SINGLE build line:
     ```
     - [ ] `go build -C ~/Documents/workspaces/teamvault/teamvault-utils -o /tmp/new-teamvault .`
     ```
     (Keep the `-C` repo path as-is — scenarios run on the host against the real repo. Only the target changes to one binary built from `.`.)
   - Replace every `/tmp/new-teamvault-<verb> …` invocation with `/tmp/new-teamvault <verb> …` — e.g. `/tmp/new-teamvault-password --teamvault-config …` → `/tmp/new-teamvault password --teamvault-config …`; `/tmp/new-teamvault-login …` → `/tmp/new-teamvault login …`; `/tmp/new-teamvault-username …` → `/tmp/new-teamvault username …`.
   - Update `## Cleanup` blocks to `rm -f /tmp/new-teamvault …` (single binary) instead of removing the seven per-verb paths.
   - Preserve each scenario's intent, keys, env vars, timeouts, and Expected assertions — only the build+invocation surface changes, not what is validated.
   - Scenario 001 note: it asserts the password stdout is non-empty and (per the spec) the no-trailing-newline fix; if 001 does not already assert the last byte is not `0a`, add an Expected step: `/tmp/new-teamvault password <key> … | xxd | tail -1` shows the last byte is NOT `0a`. (Keep existing assertions; add this one so the newline fix is covered end-to-end.)
   - Update any prose/headers that name the old binaries (e.g. scenario 002's title "teamvault-login persists password") to the `teamvault login` form.

2. **Verify no old binary name remains in scenarios.** After editing, `grep -rln 'teamvault-password\|teamvault-username\|teamvault-login\|teamvault-url\|teamvault-file\|teamvault-config-parser\|teamvault-config-dir-generator' scenarios/` MUST return 0 files. (Note: the substring `teamvault password` with a SPACE is correct and allowed; the hyphenated forms are the old binary names and must be gone.)

3. **Update `README.md`:**
   - Replace every `go get`/`go install github.com/bborbe/teamvault-utils/v4/cmd/teamvault-*` line with the single install:
     ```
     go install github.com/bborbe/teamvault-utils/v5@latest
     ```
   - Rewrite every usage example from `teamvault-<verb> …` to `teamvault <verb> …` (login, password, username, url, file, config parse, config generate).
   - Remove any prose describing "seven CLI binaries" / seven install lines; describe the single `teamvault` command with subcommands instead.
   - CRITICAL — sweep ALL `/v4` module references to `/v5`. Replace every occurrence of `github.com/bborbe/teamvault-utils/v4` in `README.md` with `github.com/bborbe/teamvault-utils/v5`, including: shields.io/pkg.go.dev badge URLs, the `go get github.com/bborbe/teamvault-utils/v4` library-import line, every code-block `import` line (e.g. `github.com/bborbe/teamvault-utils/v4`, `.../v4/factory`, `.../v4/mocks`), and any pkg.go.dev / godoc link. After this, `grep -c 'teamvault-utils/v4' README.md` MUST return 0 and `grep -rc '/v4/cmd/teamvault-' README.md` MUST return 0.

4. **Update `docs/releasing-teamvault-utils.md`:**
   - Replace the "a Go library + seven CLI binaries (…)" prose with a single-binary description (a Go library + one `teamvault` CLI binary). `grep -c 'seven CLI binaries' docs/releasing-teamvault-utils.md` MUST return 0.
   - Replace the per-binary `go build -C … ./cmd/teamvault-…` verification build lines with the single `go build -C … -o /tmp/new-teamvault .`.
   - Replace `go install …/v4/cmd/teamvault-*@…` lines with `go install github.com/bborbe/teamvault-utils/v5@…`. `grep -rc '/v4/cmd/teamvault-' docs/ README.md` (combined) MUST return 0.
   - Sweep every `github.com/bborbe/teamvault-utils/v4` reference in `docs/releasing-teamvault-utils.md` to `/v5` — install commands, build lines, and consumption/import examples all move to `/v5`. Past release-version STRINGS that are pure history (e.g. a table row literally naming a shipped `v4.11.0` tag as a past event) may remain as history, but any `teamvault-utils/v4` module-path reference that tells the reader how to install, build, or import the tool MUST become `/v5`. Target: `grep -c 'teamvault-utils/v4' docs/releasing-teamvault-utils.md` returns 0 (rewrite any remaining historical `/v4` module paths to `/v5` too if that is the only way to reach 0 — prefer 0 over leaving stale module paths).

5. **Add the `CHANGELOG.md` `## Unreleased` entry.** Append bullets under the existing `## Unreleased` section (created in prompt 1; do NOT add a second header):
   ```
   - feat: consolidate the seven `teamvault-*` binaries into a single `teamvault` command with subcommands (`login`, `password`, `username`, `url`, `file`, `config parse`, `config generate`).
   - refactor: migrate the CLI from `libservice.MainCmd` + `argument/v2` to `spf13/cobra` (`main.go` → `pkg/cli`), giving a clean `--help` with no Ginkgo/glog flag leakage.
   - fix: `password`/`username`/`url`/`file` no longer print a trailing newline (fixes `curl -u` basic-auth).
   - build: BREAKING — module path is now `github.com/bborbe/teamvault-utils/v5`; the seven `teamvault-*` binaries are removed. Migrate invocations from `teamvault-<verb>` to `teamvault <verb>`.
   - build: BREAKING (minor) — `config generate` no longer reads the `SOURCE_DIR`/`TARGET_DIR` environment variables; `--source-dir` and `--target-dir` are now required flags. direnv/`.envrc` consumers that relied on those env vars must pass the flags explicitly.
   ```
   Adjust wording to match the changelog-guide style if the existing entries use a different convention, but cover all five: consolidation, cobra migration, newline fix, v5 breaking change, and the `SOURCE_DIR`/`TARGET_DIR` env-var removal.

</requirements>

<constraints>
- Docs and scenarios ONLY — no Go code, no Makefile, no go.mod changes in this prompt.
- Preserve each scenario's `status: active` frontmatter and its validation intent; change only the build+invocation surface.
- `teamvault <verb>` (space) is the new form; the hyphenated `teamvault-<verb>` binary names must be gone from scenarios, README, and docs install/usage.
- The `-C ~/Documents/workspaces/teamvault/teamvault-utils` build path in scenarios stays (scenarios run on the host against the real repo, not the worktree).
- Install command is `go install github.com/bborbe/teamvault-utils/v5@latest`.
- Do NOT commit — dark-factory handles git.
</constraints>

<verification>
- `make precommit` exits 0 (no code changed, but keep the tree green).
- `grep -rln 'teamvault-password\|teamvault-username\|teamvault-login\|teamvault-url\|teamvault-file\|teamvault-config-parser\|teamvault-config-dir-generator' scenarios/` returns 0 files.
- Each scenario builds one binary from `.`: `grep -rn 'go build.*-o /tmp/new-teamvault .' scenarios/` returns ≥5 lines (one per scenario).
- `grep -c 'seven CLI binaries' docs/releasing-teamvault-utils.md` returns 0.
- `grep -rc '/v4/cmd/teamvault-' docs/ README.md` returns 0 (combined).
- `grep -c 'teamvault-utils/v4' README.md` returns 0.
- `grep -c 'teamvault-utils/v4' docs/releasing-teamvault-utils.md` returns 0.
- `grep -c 'go install github.com/bborbe/teamvault-utils/v5' README.md` returns ≥1.
- `grep -n '## Unreleased' CHANGELOG.md` returns ≥1 and `grep -niE 'consolidat|cobra|v5' CHANGELOG.md` returns ≥1 line beneath it.
- `grep -niE 'SOURCE_DIR|TARGET_DIR' CHANGELOG.md` returns ≥1 (the env-var removal is recorded).
</verification>
