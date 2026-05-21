---
status: committing
spec: [002-cache-enable-and-timeout]
summary: Implemented Config.Timeout field, CreateConnectorWithConfigAndTimeout factory function with cache OR-logic fix, and corresponding tests
container: teamvault-utils-exec-006-spec-002-config-and-factory
dark-factory-version: v0.164.0
created: "2026-05-21T12:10:00Z"
queued: "2026-05-21T12:38:20Z"
started: "2026-05-21T12:38:22Z"
branch: dark-factory/cache-enable-and-timeout
---

<summary>
- `Config` struct gains a `Timeout` field of type `libtime.Duration`. JSON accepts `"5s"`, `"30s"`, etc.; absent/empty deserializes to zero.
- Factory gains a new function `CreateConnectorWithConfigAndTimeout` that accepts a `libtime.Duration` timeout. The existing `CreateConnectorWithConfig` and `CreateConnectorWithConfigAndKeychain` keep their current signatures and delegate to the new function with zero timeout (preserves source compatibility for external callers).
- Cache OR semantics fixed: `cacheEnabled = cacheEnabled || config.CacheEnabled` replaces the unconditional overwrite at `factory/factory.go:71`.
- Timeout precedence inside the factory: CLI/programmatic timeout > config file `timeout` > 5s default. The resolved timeout is applied to the passed-in `*http.Client` via `httpClient.Timeout = resolved` (request-level deadline, not just dial timeout).
- Zero is the "use default" sentinel. Negative timeouts return a wrapped error from the factory.
- Unit tests cover: cache OR-logic (4 combinations), timeout precedence (cli > config > default), negative-value error, Config JSON round-trip for `Timeout` field, factory-built client honors the resolved timeout.
- Integration test (Ginkgo, no build tag): stub HTTP server sleeps longer than configured timeout + pre-populated disk cache → connector returns cached value. Without cache → connector returns timeout error.
</summary>

<objective>
Implement the library-level (Config + factory + tests) part of the cache-enable-and-timeout spec. This is the layer the CLI binaries consume in prompt 2. Public API must stay source-compatible — existing call sites of `CreateConnectorWithConfig` keep working without modification.
</objective>

<context>
Project: `~/Documents/workspaces/teamvault/teamvault-utils/` — Go library + CLI tools under module `github.com/bborbe/teamvault-utils/v4`.

Read first:
- `CLAUDE.md` for project conventions.
- `docs/dod.md` for Definition of Done.
- `specs/in-progress/002-cache-enable-and-timeout.md` for the full spec.
- `config.go` — `Config` struct. Add the new `Timeout` field here.
- `factory/factory.go` — current factory functions. Read in full to confirm the actual `CreateHttpClient` signature uses `libhttp.NewClientBuilder().WithTimeout(...).Build(ctx)`.
- `connector.go` and `diskfallback-connector.go` — unchanged by this prompt; consumed for tests.
- `remote-connector.go` — uses `r.httpClient.Do(req)` with `ctx`; respecting `httpClient.Timeout` works as expected.

Key library facts:
- `github.com/bborbe/time` `libtime.Duration` is `type Duration stdtime.Duration` with `MarshalJSON`/`UnmarshalJSON` (accepts both `"5s"` strings and numeric ns). Method `Duration()` returns the underlying `stdtime.Duration`. Already imported throughout the project.
- `github.com/bborbe/http` `libhttp.NewClientBuilder().WithTimeout(t).Build(ctx)` — **NOTE**: this `WithTimeout` sets the dial timeout only (per `http_client-builder_test.go` comment: `// Timeout affects the DialContext function, not client.Timeout`). For a request-level deadline, the factory must set `client.Timeout` directly after `Build`.
- `github.com/bborbe/errors` — wrap with `errors.Wrapf(ctx, err, "...")` or surface a fresh error with `errors.Errorf(ctx, "...")`. Never `fmt.Errorf`.

Read these coding guides:
- `~/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md`
- `~/.claude/plugins/marketplaces/coding/docs/go-factory-pattern.md`
- `~/.claude/plugins/marketplaces/coding/docs/go-testing-guide.md`
- `~/.claude/plugins/marketplaces/coding/docs/go-time-injection.md`
- `~/.claude/plugins/marketplaces/coding/docs/test-pyramid-triggers.md`

Existing test patterns to match:
- Ginkgo `Describe`/`Context`/`It` blocks (see `factory/factory_test.go`).
- Counterfeiter mocks in `mocks/` (do not need new mocks for this prompt).
- `httptest.NewServer` for stub HTTP servers.
</context>

<requirements>

1. **Add `Timeout` field to `Config`** in `config.go`:

   ```go
   import libtime "github.com/bborbe/time"

   // Config holds the configuration for connecting to a TeamVault instance.
   type Config struct {
       Url          Url               `json:"url"`
       User         User              `json:"user"`
       Password     Password          `json:"pass"`
       CacheEnabled bool              `json:"cacheEnabled,omitempty"`
       Timeout      libtime.Duration  `json:"timeout,omitempty"`
   }
   ```

   GoDoc on the new field: `// Timeout sets the HTTP request timeout for TeamVault API calls. Zero or absent means the factory default (5s).`

2. **Fix cache OR semantics** in `factory/factory.go`. At the line that currently reads:

   ```go
   cacheEnabled = config.CacheEnabled
   ```

   (current `factory.go:71`, inside the `if configPath.Exists()` block of `CreateConnectorWithConfigAndKeychain`), change to:

   ```go
   cacheEnabled = cacheEnabled || config.CacheEnabled
   ```

   This is the single-line bug fix that makes CLI `--cache=true` win over a config that omits `cacheEnabled`.

3. **Add new factory function `CreateConnectorWithConfigAndTimeout`** in `factory/factory.go`. It takes the same parameters as `CreateConnectorWithConfigAndKeychain` plus a `cliTimeout libtime.Duration` at the end:

   ```go
   // CreateConnectorWithConfigAndTimeout is like CreateConnectorWithConfigAndKeychain
   // but also accepts a CLI-supplied timeout. Resolution order: cliTimeout > config.Timeout > 5s default.
   // Negative cliTimeout returns a wrapped error.
   func CreateConnectorWithConfigAndTimeout(
       ctx context.Context,
       httpClient *http.Client,
       configPath teamvault.TeamvaultConfigPath,
       apiURL teamvault.Url,
       apiUser teamvault.User,
       apiPassword teamvault.Password,
       staging teamvault.Staging,
       cacheEnabled bool,
       currentDateTime libtime.CurrentDateTime,
       keychain teamvault.Keychain,
       cliTimeout libtime.Duration,
   ) (teamvault.Connector, error)
   ```

   Behavior:
   - Parse config if `configPath.Exists()` (same as today).
   - `cacheEnabled = cacheEnabled || config.CacheEnabled` (per requirement 2).
   - Reject negative cliTimeout: `if cliTimeout.Duration() < 0 { return nil, errors.Errorf(ctx, "invalid timeout %v: must be >= 0", cliTimeout.Duration()) }`.
   - Reject negative `config.Timeout`: same error shape.
   - Resolve timeout: `effective := cliTimeout.Duration(); if effective == 0 { effective = config.Timeout.Duration() }; if effective == 0 { effective = 5 * time.Second }`.
   - Apply: `httpClient.Timeout = effective`. (libhttp's `WithTimeout` is dial-only; setting `client.Timeout` directly gives a full-request deadline.)
   - Read the keychain password (existing logic) and call `CreateConnector` (existing logic).

4. **Keep existing factory functions source-compatible.** `CreateConnectorWithConfig` and `CreateConnectorWithConfigAndKeychain` retain their current signatures and delegate to `CreateConnectorWithConfigAndTimeout` with `cliTimeout = libtime.Duration(0)`:

   ```go
   func CreateConnectorWithConfig(/* existing params */) (teamvault.Connector, error) {
       return CreateConnectorWithConfigAndTimeout(/* same */, teamvault.NewKeychain(), libtime.Duration(0))
   }

   func CreateConnectorWithConfigAndKeychain(/* existing params */) (teamvault.Connector, error) {
       return CreateConnectorWithConfigAndTimeout(/* same */, libtime.Duration(0))
   }
   ```

   No external caller needs to change. In-tree callers in `cmd/` stay untouched in this prompt — prompt 2 wires the real CLI timeout.

5. **Do not change `CreateHttpClient`'s signature.** It still returns `(*http.Client, error)` and uses `libhttp.NewClientBuilder().WithTimeout(5 * time.Second).Build(ctx)`. The factory mutates `httpClient.Timeout` after receiving the client (requirement 3). Rationale: callers in `cmd/` still build a single `*http.Client` once and hand it to the factory; the factory owns timeout resolution because only the factory knows the config-file value.

6. **Unit tests** in `factory/factory_test.go`. Add new `Describe("CreateConnectorWithConfigAndTimeout", ...)` block; do not modify existing `It` blocks. Cover:

   a. **Cache OR-logic** (4 sub-cases):
      - cli `cacheEnabled=true`, config `cacheEnabled=false` → returned connector is the disk-fallback variant.
      - cli `cacheEnabled=false`, config `cacheEnabled=true` → disk-fallback variant.
      - cli `cacheEnabled=false`, config `cacheEnabled=false` → raw remote variant (no disk fallback).
      - cli `cacheEnabled=true`, config `cacheEnabled=true` → disk-fallback variant.

      Identify variant by type-asserting the returned `teamvault.Connector` against the unexported type or, if unexported, by behavior (e.g. call `Search` against a stub and observe disk read on failure). If `DiskFallbackConnector` has an exported constructor, check by constructing a sentinel and comparing via interface method behavior. Prefer behavior over type-assertions when the underlying type is unexported.

   b. **Timeout precedence**:
      - cliTimeout=`3s`, config.Timeout=`7s` → `httpClient.Timeout == 3s` after the call.
      - cliTimeout=`0`, config.Timeout=`7s` → `httpClient.Timeout == 7s`.
      - cliTimeout=`0`, config absent / config.Timeout=`0` → `httpClient.Timeout == 5s`.

   c. **Negative-value rejection**:
      - cliTimeout=`-1s` → wrapped error containing `"invalid timeout"` and `"-1s"`.
      - cliTimeout=`0`, config.Timeout=`-1s` → same error shape.

7. **Config JSON round-trip tests** in `config-parser_test.go` (or a new `config_test.go` next to `config.go`):

   - JSON `{"url":"...","user":"u","pass":"p","timeout":"5s"}` → `ParseTeamvaultConfig` returns `Config` with `cfg.Timeout.Duration() == 5*time.Second`.
   - JSON without `timeout` field → `cfg.Timeout.Duration() == 0`.
   - JSON with unparseable timeout `{"timeout":"banana"}` → `ParseTeamvaultConfig` returns a non-nil error (the error comes from `libtime.Duration.UnmarshalJSON`; do not re-wrap).
   - JSON with negative timeout `{"timeout":"-5s"}` → parsed successfully (negative validation is the factory's job, not config parse).

8. **Integration test** (in `factory/factory_integration_test.go`, plain Ginkgo, no build tag — runs under `make test`):

   - Spin up a stub via `httptest.NewServer` whose handler `time.Sleep(2 * time.Second)` then writes a valid TeamVault JSON response (`{"current_revision":"...", "password":"live-value"}`).
   - Create a temp `HOME` (`t.TempDir()` mapped via `os.Setenv("HOME", ...)`) and pre-populate `<HOME>/.teamvault-cache/<key>/password` with `cached-value` (mirror `diskfallback-connector.go`'s path layout).
   - Build connector via `CreateConnectorWithConfigAndTimeout` pointing `apiURL` at the stub, `cacheEnabled=true`, `cliTimeout=200ms`. Call `conn.Password(ctx, key)`.
   - Assert: no error; returned password == `cached-value`; total wall-clock < 1s (proves the timeout fired and the fallback returned without waiting the full 2s).
   - Second scenario: same setup but `cacheEnabled=false`. Assert: returns a non-nil error within ~500ms (timeout error, not the cached value).
   - Restore `HOME` and clean up via `t.Cleanup`.

9. **CHANGELOG**. Create `## Unreleased` section in `CHANGELOG.md` (insert immediately after the top-level `# Changelog` heading and intro lines, before the first existing version section). Add two bullets:

   ```
   ## Unreleased

   - feat: Add configurable HTTP timeout via `Config.Timeout` (`libtime.Duration`); resolution order is CLI > config > 5s default. The factory applies the resolved timeout to `httpClient.Timeout` for full-request deadlines.
   - fix: Cache enable is now the logical OR of CLI `--cache` / `CACHE` and config `cacheEnabled` — previously the config silently overrode the CLI value at `factory/factory.go:71`.
   ```

10. **Regenerate mocks if any interface changed**. `Connector`, `Keychain` are unchanged in this prompt, so `make generate` should produce no diff. Verify after implementing.

</requirements>

<constraints>
- Public Go API of `Config`, `Connector`, `TeamvaultConfigPath.Parse()`, and existing `CreateConnectorWithConfig` / `CreateConnectorWithConfigAndKeychain` signatures MUST stay source-compatible. New behavior is exposed via the new `CreateConnectorWithConfigAndTimeout` function.
- `Config` JSON deserialization stays backward compatible: pre-existing configs without `timeout` continue to load and produce zero `Timeout`.
- Default timeout = `5 * time.Second`, matching today's hardcoded value.
- Use `github.com/bborbe/time` `libtime.Duration` for the Config field — no new external deps.
- Use `github.com/bborbe/errors` for all wrapping; never `fmt.Errorf`.
- Set `httpClient.Timeout` directly (post-build) — do NOT rely on `libhttp`'s `WithTimeout` for the request deadline (it's dial-only).
- All exported items have GoDoc comments per `docs/dod.md`.
- Tests use Ginkgo/Gomega; mocks (if needed) via Counterfeiter under `mocks/`.
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass unchanged.
</constraints>

<verification>
- `make precommit` exits 0.
- `make generate` produces no diff (no interface changes).
- `go build ./...` compiles cleanly.
- `grep -n 'cacheEnabled = cacheEnabled || config.CacheEnabled' factory/factory.go` returns exactly 1 match.
- `grep -n 'Timeout' config.go` returns at least 1 match (the new field).
- `grep -n 'CreateConnectorWithConfigAndTimeout' factory/factory.go` returns at least 1 match (definition + delegation calls).
- `grep -n '## Unreleased' CHANGELOG.md` returns 1 match.
- `go test ./factory/... ./...` passes (factory + integration tests).
</verification>
