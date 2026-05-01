---
status: committing
summary: Migrated from tools.go to tools.env + Makefile @version pattern; go.mod reduced from 461 to 53 lines, replace block removed, all make targets passing.
container: teamvault-utils-003-migrate-tools-go
dark-factory-version: dev
created: "2026-05-01T07:35:35Z"
queued: "2026-05-01T07:35:35Z"
started: "2026-05-01T07:35:36Z"
---

# Migrate from tools.go to tools.env + Makefile @version pattern

<summary>
- This Go library currently pins CLI tool versions via `tools.go` (build tag `tools`), which pollutes `go.mod` with hundreds of unrelated transitive dependencies.
- Migrate to the `tools.env` + Makefile `@version` pattern: each tool invocation becomes `go run pkg@$(VERSION)` driven by a flat `tools.env` file at the repo root.
- After migration, `go.mod` shrinks dramatically (typically 5x to 50x smaller) and contains only real direct/indirect deps of the library.
- The historical `replace` block (typically cellbuf, go-header, go-diskfs, ginkgolinter, runtime-spec — exact entries vary per repo) becomes unnecessary because the conflicting tool deps are no longer in the graph. The `updater` tool (v0.23.2+) will auto-drop these on the next `updater all` run because `tools.go` is gone.
- All `make` targets keep working: `make precommit`, `make test`, `make lint`, `make check`, etc.
- Pilot for this migration: `bborbe/errors v1.5.11` (commit `release v1.5.11` on master). go.mod went from 443 lines to 24 lines.
</summary>

<objective>
Apply the canonical `tools.env` + `@version` pattern to this repo so it stops polluting downstream `go.mod` files via tools.go cascade. Keep all developer-facing make targets working identically. Reduce `go.mod` to its true direct/indirect deps.
</objective>

<context>
The pattern was validated end-to-end on `bborbe/errors`. Key concepts:

- `tools.go` imports CLI tools under build tag `tools`. This pins versions in `go.mod` BUT pulls every transitive dep of every tool into the project. Cascades through library imports.
- `tools.env` declares versions as Make variables. Makefile `include`s it.
- Each Makefile tool invocation uses `go run pkg@$(VERSION)` instead of `go run -mod=mod pkg`. This builds the tool in a temporary module — the host project's `go.mod` is untouched.
- `//go:generate` directives use hardcoded `@version` (counterfeiter is the only common case): `//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6@v6.12.2 -generate`. Hardcoded because `go generate` runs from the package directory and Make variables aren't visible there.
- After deleting `tools.go`, write a minimal known-good `go.mod` (just direct deps + `go 1.x`), then `go mod tidy` repopulates legitimate indirects.
- `osv-scanner` must be pinned to `@v2.3.1` — newer versions are broken upstream (osv-scalibr's `bazelbuild/buildtools/build` package fails to resolve).
- `trivy` is invoked as a SYSTEM binary (not `go run`) — leave its Makefile target unchanged.
- `go vet -mod=mod`, `go test -mod=mod`, `go list -mod=mod`, `go generate -mod=mod` are built-in Go subcommands, NOT third-party tools — leave these unchanged.
</context>

<requirements>
1. **Create `tools.env` at the repo root** with this exact content (this is the canonical version; keep in sync with all other bborbe Go projects):

   ```
   # Canonical tool versions for all bborbe Go projects.
   # Each repo should keep its tools.env in sync with the canonical file.
   # COUNTERFEITER_VERSION must also match all `//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6@<ver>` directives.

   ADDLICENSE_VERSION         ?= v1.2.0
   COUNTERFEITER_VERSION      ?= v6.12.2
   ERRCHECK_VERSION           ?= v1.10.0
   GINKGO_VERSION             ?= v2.28.3
   GOIMPORTS_REVISER_VERSION  ?= v3.12.6
   GOLANGCI_LINT_VERSION      ?= v2.11.4
   GOLINES_VERSION            ?= v0.13.0
   GO_MODTOOL_VERSION         ?= v0.7.1
   GOSEC_VERSION              ?= v2.26.1
   GOVULNCHECK_VERSION        ?= v1.3.0
   OSV_SCANNER_VERSION        ?= v2.3.1
   ```

2. **Update `Makefile`.** Add `include tools.env` near the top (after any `export ROOTDIR ?= ...` style line, before the first target). Replace every `go run -mod=mod pkg` invoking a third-party tool with `go run pkg@$(VERSION_VAR)`:

   - `go run -mod=mod github.com/shoenig/go-modtool` → `go run github.com/shoenig/go-modtool@$(GO_MODTOOL_VERSION)`
   - `go run -mod=mod github.com/incu6us/goimports-reviser/v3` → `go run github.com/incu6us/goimports-reviser/v3@$(GOIMPORTS_REVISER_VERSION)`
   - `go run -mod=mod github.com/segmentio/golines` → `go run github.com/segmentio/golines@$(GOLINES_VERSION)`
   - `go run -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint` (or `/v2/...`) → `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)` (note: explicitly use the `/v2/` import path)
   - `go run -mod=mod github.com/kisielk/errcheck` → `go run github.com/kisielk/errcheck@$(ERRCHECK_VERSION)`
   - `go run -mod=mod golang.org/x/vuln/cmd/govulncheck` → `go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)`
   - `go run -mod=mod github.com/google/osv-scanner/v2/cmd/osv-scanner` → `go run github.com/google/osv-scanner/v2/cmd/osv-scanner@$(OSV_SCANNER_VERSION)`
   - `go run -mod=mod github.com/securego/gosec/v2/cmd/gosec` → `go run github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)`
   - `go run -mod=mod github.com/google/addlicense` → `go run github.com/google/addlicense@$(ADDLICENSE_VERSION)`

   Leave unchanged: `go vet -mod=mod`, `go test -mod=mod`, `go list -mod=mod`, `go generate -mod=mod`, and the `trivy fs ...` invocation (trivy is a system binary).

3. **Update every `//go:generate` counterfeiter directive.** Find every file containing `//go:generate go run -mod=mod github.com/maxbrunsfeld/counterfeiter/v6 -generate` and replace with `//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6@v6.12.2 -generate`. The version is hardcoded because Make variables aren't accessible from `go generate`. The hardcoded value MUST match `COUNTERFEITER_VERSION` in `tools.env`.

4. **Delete `tools.go`** from the repo root. If `tools.go` imports a CLI that has no Makefile counterpart (historical example: `golang.org/x/lint/golint` in some libs — Makefile doesn't invoke `golint` anywhere), simply drop that import along with `tools.go` — do NOT add a `tools.env` entry for it. Only add `tools.env` entries for tools the Makefile actually invokes.

5. **Bump every `bborbe/*` direct dep to `@latest`.** Critical step — even ONE pre-migration bborbe dep (e.g. `metrics v0.5.0`) cascades back the full tools.go pollution (~360 indirect requires). Use this script-driven recipe (no manual enumeration):

   ```bash
   # Every bborbe direct dep → @latest, in one shot
   grep '^	github.com/bborbe/' go.mod | grep -v '// indirect\|=>' | awk '{print $1}' | xargs -I {} go get {}@latest
   ```

   Notes:
   - This is `@latest` not `@<specific-version>` — the migrated versions are already on the proxy, `@latest` picks them up automatically. No need for the prompt to enumerate every version.
   - For multi-module repos (e.g. `agent/lib`'s sub-modules) where the lib is referenced via local `replace`, the local replace stays; `@latest` only affects the published require directive.
   - The constraint "no `go get -u`" still holds for *non-bborbe* deps; the bborbe bump is intentional and required for the migration.

6. **Reset `go.mod` to a minimal known-good state, then tidy.** Running `go mod tidy -e` on the polluted go.mod can truncate it. Instead:

   a. Identify the real direct deps. The reliable recipe: list every non-`tools.go`, non-`vendor/` `.go` file, extract their `import` blocks, filter out stdlib (anything not containing a `.`), and the remaining external imports are the direct deps. The library's primary `require (...)` block in the new go.mod should contain ONLY these (with their existing versions taken from the current go.mod, or the bumped `@latest` versions from step 5).
   b. Manually rewrite `go.mod` as: `module ...`, `go 1.x`, then a single `require (...)` block listing only direct deps. Drop the entire `replace (...)` block (EXCEPT in multi-module repos: keep the `replace github.com/bborbe/<repo>/<sub> => ../<sub>` directives that resolve sister sub-modules locally). Drop the `// indirect` requires block — `go mod tidy` will repopulate it.
   c. Run `go mod tidy`. Verify the new `go.mod` is dramatically smaller (target: under 30 lines for a library; services may have more depending on real deps).
   d. **Verify zero tools.go-era pollution remains:**

      ```bash
      grep -E '(cellbuf|go-header|go-diskfs|golangci-lint|osv-scanner|ginkgolinter|charmbracelet/x|denis-tingaikin)' go.mod
      ```

      Must return zero matches. If matches appear, run `go mod why <package>` to find the cascade source — it's almost always an unbumped `bborbe/*` dep still pulling in the old tools.go. Re-run step 5 to bump it.
   e. Verify `go.sum` was regenerated.

7. **Clean up stale CVE suppressions** (the post-migration dep graph is dramatically smaller, so existing suppressions for tool-only transitives become dead):

   a. Open `.osv-scanner.toml` if it exists. Each entry pins a CVE in a specific dep version. Remove entries that are pinned to deps no longer present in the slimmed `go.mod` (typically docker, etcd, bbolt, aws-sdk, kubernetes, charmbracelet — anything that was a tools.go transitive). Re-run `make osv-scanner` after each removal; if the scanner still passes, the entry was dead and stays removed. If new CVEs surface in real production deps, leave them visible — do NOT add fresh suppressions in this migration.
   b. Inspect the `make vulncheck` target in `Makefile`. If it has a `jq -e 'select(... .finding.osv != "GO-..." ...)'` filter listing specific GO-IDs, attempt to drop each ID one at a time and re-run `make vulncheck`. Drop IDs that are no longer triggered (the corresponding vulnerable dep is gone). Preserve the `jq` filter structure verbatim, only adjusting the list of OSV IDs.

8. **Run `make precommit`.** Must pass end-to-end. If `make osv-scanner` reports actual vulnerabilities in REAL production deps (post-cleanup), do NOT invent suppressions — leave the failure visible and surface it in the commit message / changelog as a follow-up.

9. **Verify `mocks/` regeneration works.** `make generate` should run successfully. The diff between old and new mocks is acceptable if it's only header-level (license, generation timestamp) — counterfeiter via `@version` produces semantically identical mocks to the previous tools.go-bound version. If mock content drifts, investigate before committing.

10. **Do NOT touch the existing `replace (...)` block manually beyond removing it as part of step 6b.** The `updater` tool (v0.23.2+) will also auto-drop any tools.go-era replaces from migrated projects on subsequent runs because `tools.go` is gone — this is a safety net, not the primary mechanism.

11. **Commit + tag.** Use the existing release workflow on master. The CHANGELOG entry should describe: "Migrate to tools.env + Makefile @version pattern; remove tools.go and obsolete replace block. go.mod reduced from <N> to <M> lines."
</requirements>

<verification>
After migration, all of these must hold:

- `tools.env` exists at the repo root with the 11 version variables
- `tools.go` does NOT exist
- `Makefile` includes `tools.env` near the top
- `grep -c 'go run -mod=mod ' Makefile` returns `0` (every third-party tool invocation has been migrated; `go vet -mod=mod`, `go test -mod=mod`, etc. are unaffected because they don't match the `go run -mod=mod ` pattern)
- All `//go:generate` directives invoking counterfeiter use `@v6.12.2` syntax (no `-mod=mod`)
- `go.mod` does not contain a `replace (` block (or contains at most one truly-needed replace if a non-tools.go reason exists — extremely rare)
- `go.mod` line count is dramatically reduced compared to before (typically 5x to 50x smaller)
- `make precommit` passes end-to-end
- `make test` passes with the same coverage as before the migration
- `git diff --stat go.mod` shows the file shrinking by hundreds of lines
- `git status` shows `tools.env` added, `tools.go` deleted, `Makefile` and `go.mod` modified
</verification>

<constraints>
- Don't bump non-bborbe dependencies (no `go get -u` of upstream deps). The `bborbe/*` deps ARE bumped to `@latest` per step 5 — that's intentional and required to drop tools.go-era cascade pollution.
- Don't refactor production code or test code logic
- Don't touch `vendor/` (gitignored in the canonical layout; `make ensure` deletes it anyway)
- Don't add new linters or remove existing ones from `make check`
- Don't change Go language version (`go 1.x` directive in `go.mod`)
- Don't replace `trivy fs ...` with `go run` — trivy is a system binary, not a Go tool
- Don't invent `.osv-scanner.toml` suppressions to make CVE findings disappear — surface real vulns instead
</constraints>
