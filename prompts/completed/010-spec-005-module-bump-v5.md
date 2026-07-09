---
status: completed
spec: [005-consolidate-cli-into-teamvault-command]
summary: 'Bumped Go module path from github.com/bborbe/teamvault-utils/v4 to v5 across go.mod, all .go imports, and regenerated mocks; added ## Unreleased changelog entry'
execution_id: teamvault-utils-consolidate-cli-exec-010-spec-005-module-bump-v5
dark-factory-version: v0.191.0
created: "2026-07-09T15:01:03Z"
queued: "2026-07-09T15:16:04Z"
started: "2026-07-09T15:17:19Z"
completed: "2026-07-09T15:19:50Z"
---

<summary>
- Bumps the Go module major version from `/v4` to `/v5` — a breaking change because later prompts remove the seven binaries and rename the module path.
- Rewrites every internal Go import from `github.com/bborbe/teamvault-utils/v4` to `.../v5` across the whole repo (root package, `factory/`, `cmd/`, tests, generated mocks).
- No behavior changes: the seven `cmd/teamvault-*` binaries still exist and still build after this prompt.
- Regenerates Counterfeiter mocks so their internal import path matches the new module path.
- Leaves an `## Unreleased` CHANGELOG section that later prompts append to.
- `make precommit` stays green — this is a mechanical rename that keeps the tree building.
</summary>

<objective>
Rename the Go module from `github.com/bborbe/teamvault-utils/v4` to `github.com/bborbe/teamvault-utils/v5` and rewrite all internal `.go` import paths accordingly, with zero behavior change. The seven `cmd/teamvault-*` binaries remain present and building; this prompt only prepares the module path for the CLI consolidation that follows.
</objective>

<context>
Read `CLAUDE.md` for project conventions.

Project: `github.com/bborbe/teamvault-utils` — a Go library plus CLI tools. Currently module `.../v4`.

Read before changing:
- `go.mod` — line 1 is `module github.com/bborbe/teamvault-utils/v4`.
- Any one `cmd/teamvault-*/main.go` (e.g. `cmd/teamvault-password/main.go`) — shows the `github.com/bborbe/teamvault-utils/v4` and `.../v4/factory` import forms to be rewritten.
- `factory/factory.go` — imports `github.com/bborbe/teamvault-utils/v4`.
- `Makefile` — `generate` target runs `go generate` to (re)build mocks under `mocks/`; `format`/`ensure` targets. Note the `goimports-reviser` invocation uses `-project-name github.com/bborbe/teamvault-utils` (no version suffix) — leave that flag as-is.
- `CHANGELOG.md` — check whether an `## Unreleased` section already exists at the top.

Read this coding guide:
- `/home/node/.claude/plugins/marketplaces/coding/docs/changelog-guide.md`

Library facts:
- The module path version suffix (`/v4`, `/v5`) is part of every import path for major versions ≥ 2, per Go module semantics. Changing `module …/v5` in `go.mod` requires rewriting every self-import in the repo, including in generated `mocks/`.
</context>

<requirements>

1. **Bump the module path in `go.mod`.** Change line 1 from:
   ```
   module github.com/bborbe/teamvault-utils/v4
   ```
   to:
   ```
   module github.com/bborbe/teamvault-utils/v5
   ```
   Do NOT change the `go 1.26.5` line or any `require` block.

2. **Rewrite every internal import across all `.go` files.** Replace every occurrence of the literal string `github.com/bborbe/teamvault-utils/v4` with `github.com/bborbe/teamvault-utils/v5` in all `*.go` files in the repository (root package files, `factory/`, all seven `cmd/teamvault-*/`, all `*_test.go`, and generated files under `mocks/`). This covers both the base import (`github.com/bborbe/teamvault-utils/v4`) and the subpackage import (`github.com/bborbe/teamvault-utils/v4/factory`).

   A safe mechanical approach:
   ```
   grep -rl 'teamvault-utils/v4' --include='*.go' . | xargs sed -i '' 's#teamvault-utils/v4#teamvault-utils/v5#g'
   ```
   (On the container this may be GNU sed — use `sed -i 's#…#…#g'` without the empty `''` argument if the BSD form errors. Verify with the grep in step 5 regardless of which sed form is used.)

3. **Regenerate mocks** so the Counterfeiter-generated files under `mocks/` carry the new import path. Run `make generate`. If `make generate` is unavailable in the environment, the sed rewrite in step 2 already covers `mocks/*.go`; either path is acceptable as long as the verification greps pass.

4. **Tidy the module graph.** Run `go mod tidy` (or `make ensure`) so `go.sum` and the require graph stay consistent. Do NOT add or remove any third-party dependency in this prompt — `libservice`, `argument/v2`, `glog`, cobra, etc. are all out of scope here (cobra is added in the next prompt).

5. **Verify no `/v4` import remains.** `grep -rn 'teamvault-utils/v4' --include='*.go' .` MUST return 0 matches. `head -1 go.mod` MUST equal `module github.com/bborbe/teamvault-utils/v5`.

6. **Ensure the seven binaries still build unchanged.** Do NOT delete, rename, or restructure anything under `cmd/`. After the rename, all seven binaries must compile:
   ```
   go build ./cmd/teamvault-password ./cmd/teamvault-username ./cmd/teamvault-url ./cmd/teamvault-file ./cmd/teamvault-config-parser ./cmd/teamvault-config-dir-generator ./cmd/teamvault-login
   ```

7. **CHANGELOG.** Ensure `CHANGELOG.md` has an `## Unreleased` section at the top (create it directly under the title if absent). Add one bullet:
   ```
   - build: bump module path to `github.com/bborbe/teamvault-utils/v5` (breaking — major version bump preparing the single-`teamvault`-binary consolidation).
   ```
   Do NOT add a dated release section — the operator cuts `v5.0.0` after merge.

</requirements>

<constraints>
- Mechanical rename only — NO behavior change, NO file moves, NO deletions under `cmd/`.
- Errors, factory signatures, connector APIs, typed wrappers — all untouched.
- Do NOT add cobra or remove `libservice` in this prompt (next prompts handle the CLI migration).
- Keep the `goimports-reviser -project-name github.com/bborbe/teamvault-utils` flag in the Makefile unchanged (it has no version suffix).
- Use `github.com/bborbe/errors` for any errors (none should be added here).
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass unchanged.
</constraints>

<verification>
- `make precommit` exits 0.
- `head -1 go.mod` equals `module github.com/bborbe/teamvault-utils/v5`.
- `grep -rn 'teamvault-utils/v4' --include='*.go' .` returns 0 matches.
- All seven binaries build: `go build ./cmd/teamvault-password ./cmd/teamvault-username ./cmd/teamvault-url ./cmd/teamvault-file ./cmd/teamvault-config-parser ./cmd/teamvault-config-dir-generator ./cmd/teamvault-login`.
- `grep -n '## Unreleased' CHANGELOG.md` returns ≥1.
</verification>
