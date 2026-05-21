---
status: completed
spec: [002-cache-enable-and-timeout]
summary: Added --teamvault-timeout flag and TEAMVAULT_TIMEOUT env var to all 7 teamvault-* binaries, wired to CreateConnectorWithConfigAndTimeout in 6 binaries and httpClient.Timeout in teamvault-login, added CLI parse contract test, updated README with config/CLI/cache docs, and added CHANGELOG entry.
container: teamvault-utils-exec-007-spec-002-cli-binaries-and-integration-test
dark-factory-version: v0.164.0
created: "2026-05-21T12:10:00Z"
queued: "2026-05-21T12:38:20Z"
started: "2026-05-21T12:49:38Z"
completed: "2026-05-21T13:03:14Z"
branch: dark-factory/cache-enable-and-timeout
---

<summary>
- All 7 `teamvault-*` CLI binaries gain a `--teamvault-timeout` flag and `TEAMVAULT_TIMEOUT` env var of type `libtime.Duration`. Accepts `5s`, `30s`, etc.
- Each binary's `Run` passes the resolved timeout to the new factory function `CreateConnectorWithConfigAndTimeout` (added in prompt 1). `teamvault-login` wires it via the keychain-aware factory path.
- `teamvault-login`'s independent 10s per-probe `context.WithTimeout` inside `loginFlow` stays unchanged (per spec line 104).
- README.md documents the new `timeout` config field, the new CLI flag, and the cache OR-precedence.
- CHANGELOG entry under `## Unreleased` (section created in prompt 1) gets the CLI-side line.
- One small CLI-parse contract test confirms libargument parses `--teamvault-timeout=5s` into a `libtime.Duration` on at least one binary's application struct.
</summary>

<objective>
Add `--teamvault-timeout` to all 7 `teamvault-*` binaries, wire it to the factory function introduced in prompt 1, and document the new behavior in README + CHANGELOG. This is the user-facing layer of the cache-enable-and-timeout spec.
</objective>

<context>
Project: `~/Documents/workspaces/teamvault/teamvault-utils/` — Go library + CLI tools under module `github.com/bborbe/teamvault-utils/v4`.

Read first:
- `CLAUDE.md` for project conventions.
- `specs/in-progress/002-cache-enable-and-timeout.md` for the full spec.
- `prompts/in-progress/<N>-spec-002-config-and-factory.md` (this prompt's sibling, executed first) — adds `CreateConnectorWithConfigAndTimeout`, `Config.Timeout`, and the OR-logic fix.
- `cmd/teamvault-password/main.go` — canonical CLI pattern: `libservice.MainCmd`, libargument struct tags, calls `factory.CreateConnectorWithConfig`.
- `cmd/teamvault-url/main.go`, `cmd/teamvault-username/main.go`, `cmd/teamvault-file/main.go`, `cmd/teamvault-config-parser/main.go`, `cmd/teamvault-config-dir-generator/main.go` — five binaries follow the same `libservice.MainCmd` pattern.
- `cmd/teamvault-login/main.go` — different shape (own `Run`, builds connector via a closure on `factory.CreateConnector`). Reads `factory.CreateHttpClient(ctx)` at line ~80. The 10s `context.WithTimeout` calls at lines ~128 and ~158 are per-probe verification timeouts and MUST NOT change.
- `factory/factory.go` — after prompt 1, has `CreateConnectorWithConfigAndTimeout(..., cliTimeout libtime.Duration)` and `CreateConnectorWithConfig` delegating to it.
- `README.md` — existing config example block and CLI tool sections.
- `CHANGELOG.md` — `## Unreleased` section created in prompt 1; append CLI-side entry.

Library facts:
- `github.com/bborbe/argument/v2` (libargument): supports any type implementing `encoding.TextUnmarshaler` as a CLI/env field. `libtime.Duration` implements `TextUnmarshaler`, so `arg:"teamvault-timeout"` on a `libtime.Duration` field parses `--teamvault-timeout=5s` correctly.
- Zero `libtime.Duration` is the "use factory default" sentinel — no per-CLI validation needed; the factory rejects negative and substitutes 5s for zero.

Read these coding guides:
- `~/.claude/plugins/marketplaces/coding/docs/go-cli-guide.md`
- `~/.claude/plugins/marketplaces/coding/docs/changelog-guide.md`
- `~/.claude/plugins/marketplaces/coding/docs/readme-guide.md`
</context>

<requirements>

1. **Add `TeamvaultTimeout` field to the six `libservice.MainCmd` binaries** (`teamvault-password`, `teamvault-username`, `teamvault-url`, `teamvault-file`, `teamvault-config-parser`, `teamvault-config-dir-generator`). In each `application` struct add:

   ```go
   TeamvaultTimeout libtime.Duration `required:"false" arg:"teamvault-timeout" env:"TEAMVAULT_TIMEOUT" usage:"HTTP request timeout for TeamVault API calls (e.g. 5s, 30s); 0 = default 5s"`
   ```

   Ensure `libtime "github.com/bborbe/time"` is imported (already present in every cmd file).

   The flag name `--teamvault-timeout` and env var `TEAMVAULT_TIMEOUT` MUST be identical across all binaries.

2. **Wire the timeout through to the factory** in each of the six binaries' `Run` method. Replace the existing call:

   ```go
   teamvaultConnector, err := factory.CreateConnectorWithConfig(
       ctx, httpClient, ..., a.Cache, currentDateTime,
   )
   ```

   with:

   ```go
   teamvaultConnector, err := factory.CreateConnectorWithConfigAndTimeout(
       ctx, httpClient, ..., a.Cache, currentDateTime,
       teamvault.NewKeychain(),
       a.TeamvaultTimeout,
   )
   ```

   Match the parameter order of `CreateConnectorWithConfigAndTimeout` as defined in prompt 1.

3. **`teamvault-login` wiring.** In `cmd/teamvault-login/main.go`:

   a. Add the same `TeamvaultTimeout libtime.Duration` field to the `application` struct with the same struct tag as above.

   b. Replace the call to `factory.CreateConnector` inside the `makeConnector` closure with `factory.CreateConnectorWithConfigAndTimeout` if the existing call signature lets you pass a timeout. If `CreateConnector` itself doesn't take a timeout, instead resolve the timeout once before constructing the closure and apply it to the `*http.Client` directly: `httpClient.Timeout = a.TeamvaultTimeout.Duration()` (followed by `if httpClient.Timeout == 0 { httpClient.Timeout = 5 * time.Second }`). Pick whichever fits the existing flow more cleanly — read `cmd/teamvault-login/main.go` in full first.

   c. DO NOT touch the two `context.WithTimeout(ctx, 10*time.Second)` calls inside `loginFlow` (currently at lines ~128 and ~158). Those are per-probe verification deadlines and are independent of the HTTP request timeout.

4. **CLI parse contract test.** Add a small Ginkgo test for one of the binaries (e.g. `cmd/teamvault-password/main_test.go` if it exists, otherwise create it) that:

   - Constructs a fresh `application{}` struct,
   - Invokes libargument via the standard pattern used in `cmd/teamvault-login/main_test.go` (if a precedent exists; otherwise call `libargument.Parse(ctx, args, "...", "...")` directly per the libargument README),
   - Passes `[]string{"--teamvault-timeout=5s"}` (or sets `TEAMVAULT_TIMEOUT=5s` and parses with no flag),
   - Asserts `app.TeamvaultTimeout.Duration() == 5 * time.Second`.

   This is the boundary-contract test for libargument × libtime.Duration. One binary suffices since the struct tag is identical across all 7.

5. **Update `README.md`**:

   a. In the JSON config example block, add a `timeout` entry. Example shape (adjust to existing block style):

      ```json
      {
          "url": "https://teamvault.example.com",
          "user": "my-user",
          "cacheEnabled": true,
          "timeout": "30s"
      }
      ```

      Add one explanatory line under the example: `"timeout"` sets the HTTP request timeout. Defaults to 5 seconds when absent.

   b. In a relevant CLI Tools section (or a new short subsection), document:

      ```
      --teamvault-timeout=5s   HTTP request timeout for TeamVault API calls (env: TEAMVAULT_TIMEOUT; default: 5s)
      ```

   c. Add a short paragraph near the existing `cacheEnabled` docs (or in a new "Cache behavior" subsection): cache is enabled if EITHER the `--cache` / `CACHE` env var is `true` OR the config file's `cacheEnabled: true` is set. There is no way to force-disable via CLI when config opts in; edit the config to disable.

   d. Evidence: `grep -ni 'timeout' README.md` returns ≥3 lines (config example + CLI flag docs + behavior paragraph); `grep -ni 'cache.*either\|either.*cache\|cacheEnabled.*OR' README.md` returns ≥1 line.

6. **CHANGELOG entry**. Append to the `## Unreleased` section created in prompt 1 (do not duplicate the section header). Add:

   ```
   - feat: `--teamvault-timeout` flag and `TEAMVAULT_TIMEOUT` env var across all `teamvault-*` CLI binaries; threads through to the new factory `CreateConnectorWithConfigAndTimeout`.
   ```

7. **Verify all 7 binaries still build**:

   ```
   go build ./cmd/teamvault-password ./cmd/teamvault-username ./cmd/teamvault-url ./cmd/teamvault-file ./cmd/teamvault-config-parser ./cmd/teamvault-config-dir-generator ./cmd/teamvault-login
   ```

</requirements>

<constraints>
- The CLI flag name `--teamvault-timeout` and env var `TEAMVAULT_TIMEOUT` are identical across all 7 binaries.
- Field type is `libtime.Duration` (NOT `time.Duration`). libargument needs `encoding.TextUnmarshaler` to parse `5s` strings, and `libtime.Duration` provides it; stdlib `time.Duration` does not.
- Pass `app.TeamvaultTimeout` (the `libtime.Duration` value) to `CreateConnectorWithConfigAndTimeout`, NOT `app.TeamvaultTimeout.Duration()`. The factory takes `libtime.Duration`.
- Zero `libtime.Duration` is the "use default" sentinel — no per-CLI validation; the factory handles defaulting and negative rejection.
- `teamvault-login`'s per-probe 10s `context.WithTimeout` calls in `loginFlow` MUST NOT change.
- Use `github.com/bborbe/errors` for any new errors; no `fmt.Errorf`.
- All paths repo-relative.
- Tests use Ginkgo/Gomega.
- Do NOT commit — dark-factory handles git.
- The integration test for timeout × cache fallback lives in prompt 1's deliverables (`factory/factory_integration_test.go`). Do not duplicate it here.
- Existing tests must still pass unchanged.
</constraints>

<verification>
- `make precommit` exits 0.
- All 7 binaries build: `go build ./cmd/teamvault-password ./cmd/teamvault-username ./cmd/teamvault-url ./cmd/teamvault-file ./cmd/teamvault-config-parser ./cmd/teamvault-config-dir-generator ./cmd/teamvault-login`.
- `grep -n 'TeamvaultTimeout' cmd/teamvault-password/main.go cmd/teamvault-username/main.go cmd/teamvault-url/main.go cmd/teamvault-file/main.go cmd/teamvault-config-parser/main.go cmd/teamvault-config-dir-generator/main.go cmd/teamvault-login/main.go` returns ≥7 lines (1 per file minimum).
- `grep -n 'CreateConnectorWithConfigAndTimeout' cmd/` returns ≥6 lines (the six libservice binaries).
- `grep -ni 'timeout' README.md` returns ≥3 lines.
- `grep -ni 'cache.*either\|either.*cache\|cacheEnabled.*OR' README.md` returns ≥1 line.
- `grep -n '## Unreleased' CHANGELOG.md` returns exactly 1 match.
- CLI parse contract test passes (`go test ./cmd/teamvault-password/... -run Timeout` or equivalent).
</verification>
