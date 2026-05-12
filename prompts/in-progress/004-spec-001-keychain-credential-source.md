---
status: committing
spec: [001-macos-keychain-credential-store]
summary: 'Wired macOS Keychain into teamvault-utils as a password fallback: new Keychain interface + darwin implementation via security(1) + non-darwin no-op stub + factory wiring + full Ginkgo/Gomega test suite with Counterfeiter mocks.'
container: teamvault-utils-004-spec-001-keychain-credential-source
dark-factory-version: v0.156.1-1-g04f3863-dirty
created: "2026-05-12T16:49:55Z"
queued: "2026-05-12T17:16:10Z"
started: "2026-05-12T17:57:22Z"
completed: "2026-05-12T17:54:07Z"
lastFailReason: 'execute prompt: docker run failed: wait command: exit status 1'
---

<summary>
- Add the macOS Keychain as a new password source used only when no password reached the factory from any other path.
- The Keychain source is read-only here — writing is the next prompt's concern (`teamvault-login`).
- macOS-only implementation; non-macOS builds get a no-op stub via Go build tags so existing platforms behave byte-identically.
- Implementation uses the system `security(1)` binary (already on every macOS install) via `os/exec` — no new module dependencies on any platform.
- Multi-vault works automatically: the Keychain entry is keyed by the TeamVault URL, so two different config files with two URLs produce two independent entries.
- Public Go API of `Config`, `Connector`, and `TeamvaultConfigPath.Parse()` is unchanged — wiring lives inside `factory.CreateConnectorWithConfig`.
- Unit tests with Ginkgo/Gomega + Counterfeiter cover hit, miss, file-wins precedence, missing URL, and locked Keychain.
</summary>

<objective>
Wire the macOS Keychain into `teamvault-utils` as a fallback password source so that `teamvault-*` binaries continue to work when the config file contains URL + user but no password. Behavior must be additive only — every existing flag/env/config-file path continues unchanged. Implementation must be platform-isolated via build tags so non-macOS builds do not pull in macOS-specific code or dependencies.
</objective>

<context>
Project: `~/Documents/workspaces/teamvault/teamvault-utils/` — Go library + CLI tools under module `github.com/bborbe/teamvault-utils/v4`.

Read first:
- `CLAUDE.md` for project conventions.
- `docs/dod.md` for Definition of Done.
- `specs/in-progress/001-macos-keychain-credential-store.md` for full spec.
- `config.go` — defines `Config` struct (Url, User, Password).
- `config-path.go` — defines `TeamvaultConfigPath` with `Exists()` and `Parse()`.
- `factory/factory.go` — defines `CreateConnectorWithConfig` where wiring lives.
- `cmd/teamvault-password/main.go` — example of how binaries call into the factory.
- `connector.go` — `Connector` interface.

Read these guides in `~/.claude/plugins/marketplaces/coding/docs/`:
- `go-architecture-patterns.md` — package layout and interface design.
- `go-build-args-guide.md` — `//go:build` build-tag conventions for platform-specific code.
- `go-error-wrapping-guide.md` — error wrapping with `github.com/bborbe/errors`.
- `go-testing-guide.md` — Ginkgo/Gomega test suite conventions.
- `go-test-types-guide.md` — which test types (unit/integration) to write for each change.
- `go-composition.md` — dependency injection patterns; avoid package-level mutable state.
- `changelog-guide.md` — `## Unreleased` heading conventions.

Existing patterns in this repo to match:
- Files at the repo root use `package teamvault`.
- Counterfeiter mocks generated via `//counterfeiter:generate -o mocks/<name>.go --fake-name <Fake> . <Interface>`. Hardcoded counterfeiter version `@v6.12.2` per `tools.env`.
- Error wrapping: `errors.Wrapf(ctx, err, "...")` from `github.com/bborbe/errors`.
- Logging: `glog.V(N).Infof(...)` from `github.com/golang/glog`. Never log passwords or anything derived from them.
- Tests: Ginkgo `Describe/Context/It` blocks, files named `<source>_test.go`, suite file `teamvault_suite_test.go` if it exists. Use `gexec`/`gomega` matchers.

Why Path B (shell out to `security(1)`):
- macOS ships with `/usr/bin/security` on every install — no missing-binary failure mode in practice.
- Zero new module dependencies — preserves the spec constraint "no new mandatory deps for Linux users."
- Trivial test seam: wrap `exec.Command` behind an interface, fake it in unit tests.
- Password never crosses argv: writes use `-w "-"` reading from stdin; reads use `-w` flag on `find-generic-password`.
</context>

<requirements>

1. **Create the Keychain credential source interface and sentinel.**

   Create `keychain.go` in the repo root (package `teamvault`):

   ```go
   // Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
   // Use of this source code is governed by a BSD-style
   // license that can be found in the LICENSE file.

   package teamvault

   import (
       "context"

       "github.com/bborbe/errors"
   )

   //counterfeiter:generate -o mocks/keychain.go --fake-name Keychain . Keychain

   // KeychainServiceName is the constant service name used for all teamvault-utils
   // Keychain entries. The account key is the TeamVault URL, which keeps multi-vault
   // setups isolated automatically.
   const KeychainServiceName = "teamvault-utils"

   // ErrKeychainNotSupported indicates the current platform has no supported
   // credential store backend. Callers may match this with errors.Is to
   // differentiate "no Keychain on this platform" from real Keychain failures.
   var ErrKeychainNotSupported = errors.New(context.Background(), "keychain storage is supported on macOS only in v1")

   // Keychain reads and writes TeamVault passwords from the OS credential store.
   // On macOS it backs onto the login Keychain via the `security(1)` binary.
   // On other platforms it is a no-op: ReadPassword returns ("", nil); WritePassword
   // returns ErrKeychainNotSupported.
   type Keychain interface {
       // ReadPassword returns the password stored for the given TeamVault URL,
       // or ("", nil) if no entry exists. A non-nil error indicates a real
       // failure (Keychain locked, security binary error, etc.) — callers
       // should surface this to the user, not fall through silently.
       ReadPassword(ctx context.Context, url Url) (Password, error)

       // WritePassword stores or overwrites the password for the given URL.
       // On non-darwin platforms it returns ErrKeychainNotSupported.
       WritePassword(ctx context.Context, url Url, password Password) error
   }
   ```

2. **Implement the darwin backend** at `keychain_darwin.go`:

   ```go
   //go:build darwin
   ```

   Implementation contract:

   - `NewKeychain() Keychain` factory returns a darwin-backed instance.
   - `ReadPassword`: invokes `security find-generic-password -s teamvault-utils -a <url> -w` via `os/exec`. Captures stdout. On exit code 44 (kSecItemNotFound) or stderr matching "could not be found" → return `("", nil)`. On exit code 36 / stderr mentioning "user interaction is not allowed" or "could not be unlocked" → return wrapped error: "TeamVault password requires Keychain unlock; unlock your Keychain and retry". On any other non-zero exit → wrap the error. Trim a trailing newline from stdout.
   - `WritePassword`: invokes `security add-generic-password -U -s teamvault-utils -a <url> -w` and writes the password to the child process's stdin (so the password never appears in argv). `-U` makes the call idempotent (overwrites existing entry). Wrap any non-zero exit error.
   - Both methods must respect `ctx` via `exec.CommandContext`.
   - If `url` is the empty string, `ReadPassword` returns `("", nil)` without invoking `security` — the chain will proceed to the error step in the factory.
   - Password value must never appear in `glog` output. Log only the URL (the account key) and outcome ("hit", "miss", "locked", "error").

3. **Implement the non-darwin stub** at `keychain_other.go`:

   ```go
   //go:build !darwin
   ```

   - `NewKeychain() Keychain` returns an instance whose `ReadPassword` always returns `("", nil)` and whose `WritePassword` returns `ErrKeychainNotSupported` (the sentinel defined in `keychain.go`). This preserves byte-identical resolution behavior on Linux/Windows and gives callers a stable `errors.Is` target to detect platform limitations.

4. **Wire the Keychain into the factory via a new constructor.**

   The existing `factory.CreateConnectorWithConfig` signature must NOT change (spec constraint). Instead, introduce a new exported constructor that accepts an explicit `Keychain` and keep the old one as a thin delegate. This avoids the package-level-mutable-state test seam that `go-composition.md` flags as an anti-pattern.

   In `factory/factory.go`:

   ```go
   // CreateConnectorWithConfigAndKeychain is the dependency-injected variant of
   // CreateConnectorWithConfig. Production callers use CreateConnectorWithConfig,
   // which delegates to this with teamvault.NewKeychain(). Tests inject a fake
   // Keychain to drive resolution-chain scenarios.
   func CreateConnectorWithConfigAndKeychain(
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
   ) (teamvault.Connector, error) {
       // existing file-or-args resolution (lifted from CreateConnectorWithConfig)
       // ...
       // After resolution: if apiPassword is empty AND apiURL is non-empty,
       // consult the Keychain.
       if apiPassword == "" && apiURL != "" {
           pwd, err := keychain.ReadPassword(ctx, apiURL)
           if err != nil {
               return nil, errors.Wrapf(ctx, err, "read password from keychain for url %q failed", apiURL)
           }
           if pwd != "" {
               apiPassword = pwd
           }
       }
       return CreateConnector(httpClient, apiURL, apiUser, apiPassword, staging, cacheEnabled, currentDateTime), nil
   }

   // CreateConnectorWithConfig keeps its existing signature; delegates to the
   // keychain-aware variant with the real macOS-backed Keychain.
   func CreateConnectorWithConfig(
       ctx context.Context,
       httpClient *http.Client,
       configPath teamvault.TeamvaultConfigPath,
       apiURL teamvault.Url,
       apiUser teamvault.User,
       apiPassword teamvault.Password,
       staging teamvault.Staging,
       cacheEnabled bool,
       currentDateTime libtime.CurrentDateTime,
   ) (teamvault.Connector, error) {
       return CreateConnectorWithConfigAndKeychain(
           ctx, httpClient, configPath, apiURL, apiUser, apiPassword,
           staging, cacheEnabled, currentDateTime, teamvault.NewKeychain(),
       )
   }
   ```

   - The Keychain step only runs when `apiPassword` is empty AND `apiURL` is non-empty. A locked Keychain or other real error must be wrapped and returned (not silently swallowed). A "not found" miss returns `("", nil)` and resolution continues with empty password — existing downstream behavior preserved.
   - The factory itself does NOT surface a "password not found" error. That is `teamvault-login`'s job in the next prompt.

5. **Unit tests (Ginkgo/Gomega + Counterfeiter).**

   a. **Boundary contract** — create `keychain_test.go` (no build tag, package `teamvault_test` or `teamvault`):
      - Assert `KeychainServiceName == "teamvault-utils"` (locks the constant — the value is part of the wire contract with `security(1)`).
      - Assert `errors.Is(ErrKeychainNotSupported, ErrKeychainNotSupported)` returns true (sentinel-shape regression guard).
      - Drive the Counterfeiter `Keychain` fake through Read/Write happy paths to lock the interface shape.

   b. **Darwin backend** — create `keychain_darwin_test.go` with `//go:build darwin`:
      - Factor the `security(1)` invocation behind a small `Executor` interface (`Run(ctx, name string, args []string, stdin string) (stdout string, stderr string, exitCode int, err error)`). The real implementation wraps `os/exec`. Tests inject a Counterfeiter fake of `Executor` and verify the argv shape passed to it. Do NOT use the PATH/`t.TempDir()` stub-script approach — it is fragile and platform-dependent.
      - Test scenarios: hit (exit 0 → password returned with trailing newline trimmed), miss (exit 44 → `("", nil)`), locked (exit 36 / "could not be unlocked" → wrapped error mentioning "unlock"), other non-zero (generic wrapped error), empty URL (no Executor call, `("", nil)`), context cancellation propagated to the Executor call.

   c. **Real-binary integration test** — create `keychain_darwin_integration_test.go` with `//go:build darwin && integration`:
      - Skip with `t.Skip` if `/usr/bin/security` is not present.
      - Use a one-off service name (e.g. `teamvault-utils-integration-test-<random>`) to avoid clobbering real entries. Write, read back, delete. Verify the round-trip end-to-end against the real binary. This catches argv-shape typos that any number of fake-based tests would miss.
      - Tag the test with `Pending` or guard with an env var if the project's `make precommit` does not run integration tests by default — match whatever convention `Makefile` already uses for integration tests (check `make test-integration` or similar target). If none exists, the build tag alone is sufficient — `make precommit` won't run it.

   d. **Factory wiring** — create `factory/factory_test.go` (new file; the existing `factory/factory_suite_test.go` is the Ginkgo bootstrap and stays untouched). Required scenarios as `It` blocks:

      - Password resolved from args (no config file, no Keychain consulted) — fake's `ReadPasswordCallCount()` is 0.
      - Password resolved from config file (file present, Keychain not consulted) — `ReadPasswordCallCount()` is 0.
      - Config file has URL + user, no password → Keychain returns hit → connector built with Keychain password.
      - Config file has URL + user, no password → Keychain returns miss → connector built with empty password (existing behavior preserved).
      - Config file has URL + user, no password → Keychain returns locked error → factory returns wrapped error mentioning the URL.
      - No URL anywhere (empty args, no config file) → Keychain not consulted (`ReadPasswordCallCount()` is 0).

      Tests call `CreateConnectorWithConfigAndKeychain` directly with the Counterfeiter `Keychain` fake. The thin `CreateConnectorWithConfig` delegate is exercised by at least one smoke test that confirms it builds a connector against the real `NewKeychain()` (no assertions on Keychain content — just that the call returns a non-nil connector for a fully-specified args case).

6. **Regenerate mocks.** Run `make generate` so `mocks/keychain.go` is produced by counterfeiter.

7. **CHANGELOG.** Add an entry under `## Unreleased` in `CHANGELOG.md` (create the section if missing):

   ```
   - Added macOS Keychain as a password fallback source. When the resolved config provides URL + user but no password, the library now looks up the password from the login Keychain (service `teamvault-utils`, account = URL). On non-macOS platforms this step is a no-op.
   ```

8. **Do NOT** create the `teamvault-login` binary in this prompt. Do NOT update README in this prompt. Both are the next prompt's responsibility. This prompt is the library layer only.

</requirements>

<constraints>
- Public Go API of `Config`, `Connector`, `TeamvaultConfigPath.Parse()`, and `factory.CreateConnectorWithConfig` signatures must NOT change.
- No new mandatory module dependencies on any platform (use stdlib `os/exec` only).
- macOS code must compile and link on Apple Silicon and Intel from a single binary.
- No daemons, no helper agents, no telemetry, no network calls beyond what already exists.
- The password value must never be passed via argv to `security` and must never appear in any log line.
- The Keychain read step must respect `ctx` cancellation.
- Use `github.com/bborbe/errors` for all wrapping; `glog` for all logging.
- Counterfeiter version is `@v6.12.2` per `tools.env` — match existing `//go:generate` syntax in this repo.
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass.
</constraints>

<verification>
- `make precommit` exits 0.
- `make generate` produces `mocks/keychain.go` without diff drift on a re-run.
- `grep -rn "Password" --include='*.go' . | grep -E 'glog|fmt\.Print|os\.Stdout'` shows no new lines emitting the resolved password value.
- On a non-darwin GOOS (`GOOS=linux go build ./...`), the build succeeds — verifies the build tags compile cleanly.
- The new `Keychain` interface, the darwin implementation, the non-darwin stub, and the factory wiring all have unit tests as listed in requirement 5.
</verification>
