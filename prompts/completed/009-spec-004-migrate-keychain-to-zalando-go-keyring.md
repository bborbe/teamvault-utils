---
status: completed
spec: [004-migrate-keychain-to-zalando-go-keyring]
container: teamvault-utils-exec-009-spec-004-migrate-keychain-to-zalando-go-keyring
dark-factory-version: v0.164.0
created: "2026-05-21T20:00:00Z"
queued: "2026-05-21T20:06:26Z"
started: "2026-05-21T20:06:28Z"
completed: "2026-05-21T20:44:45Z"
branch: dark-factory/migrate-keychain-to-zalando-go-keyring
lastFailReason: 'execute prompt: docker run failed: wait command: exit status 137'
---

<summary>
- `keychain_darwin.go` is rewritten to call `github.com/zalando/go-keyring` instead of constructing `security` REPL scripts. `keychain_other.go` and `keychain_executor.go` are deleted.
- `darwinKeychain` gains a small package-private `keyringClient` interface (Counterfeiter-mockable) wrapping `keyring.Get` and `keyring.Set`. Production code injects the real wrapper; unit tests inject a fake for behavioral assertions.
- A pre-implementation probe runs zalando against newline and NUL-byte passwords once on the host, recorded in the completion report. The probe determines whether the existing `validatePasswordForKeychain` helper stays (zalando passes raw bytes through) or goes (zalando validates internally).
- The probe also resolves how zalando signals "no usable backend" on Linux without Secret Service, so `isNoBackendError` ships with a real implementation rather than a `return false` stub.
- `NewKeychainWithExecutor` is audited for external usage first; deleted outright if no consumers, deprecated-stubbed otherwise.
- Existing public API (`Keychain` interface, `NewKeychain`, `KeychainServiceName`, `ErrKeychainNotSupported`) is unchanged. All 7 CLI binaries build untouched.
- Scenario 002 walks clean end-to-end against the migrated binary. All four other scenarios continue to pass without modification.
</summary>

<objective>
Replace the hand-rolled `security` shell-out in `keychain_darwin.go` with `github.com/zalando/go-keyring`. The library handles all quoting / escaping / platform dispatch internally, eliminating the failure surface that produced two v4.10–v0.9.10 bugs in our own code. Public Go API and existing callers stay byte-identical. The migration ships with a testable abstraction (`keyringClient` interface) so unit tests cover real `darwinKeychain` logic rather than mocking the public `Keychain` it implements.
</objective>

<context>
Project: Go library + CLI tools under module `github.com/bborbe/teamvault-utils/v4`. Working directory is `/workspace` inside the container.

Read first:
- `CLAUDE.md` for project conventions.
- `specs/in-progress/004-migrate-keychain-to-zalando-go-keyring.md` — the full spec including Failure Modes and Acceptance Criteria.
- `keychain.go` — `Keychain` interface, `KeychainServiceName` constant, `ErrKeychainNotSupported` sentinel. Do not change signatures.
- `keychain_darwin.go` — the current implementation to replace. Note `validatePasswordForKeychain` (NUL/newline guard); decide whether to retain based on probe result.
- `keychain_executor.go` — `Executor` interface + `osExecutor`, both deleted in this prompt.
- `keychain_other.go` — non-darwin stub, also deleted in this prompt.
- `keychain_darwin_test.go` — existing Ginkgo suite; the Executor-mocking blocks go away, replaced with `keyringClient`-fake-based tests (see requirement 7).
- `keychain_darwin_integration_test.go` — real-Keychain integration test using unique per-test service name; update to call `teamvault.NewKeychain()` instead of constructing via the deleted Executor seam.
- `mocks/` — Counterfeiter generates files here from `//counterfeiter:generate` directives in source.

Read these coding guides (in-container paths under `/home/node/.claude/plugins/marketplaces/coding/docs/`):
- `go-error-wrapping-guide.md` — wrap with `github.com/bborbe/errors`; never `fmt.Errorf`; never `errors.Wrapf(ctx, nil, ...)`.
- `go-testing-guide.md` — Ginkgo/Gomega + Counterfeiter conventions.
- `go-mocking-guide.md` — `//counterfeiter:generate` directive format.
- `go-mod-dependency-fix-guide.md` — adding a dependency with `go get` + `go mod tidy`.
- `test-pyramid-triggers.md` — when to write unit vs integration tests.

Key library facts about `github.com/zalando/go-keyring`:
- `keyring.Get(service, user string) (string, error)` — returns password or `keyring.ErrNotFound`.
- `keyring.Set(service, user, password string) error`.
- macOS backend internally invokes `/usr/bin/security` with library-managed argument quoting.
- Linux backend uses Secret Service via DBus when available.
- Linux without Secret Service surfaces an error — exact form must be observed by the probe (requirement 2) before writing the matcher.

`bborbe/errors` exposes `errors.Is`, `errors.As`, `errors.Wrap`, `errors.Wrapf`, `errors.Errorf`, `errors.New`. `errors.Is(err, sentinel)` works on stdlib-style errors including `keyring.ErrNotFound`. If for any reason the wrapper does not expose `Is`, fall back to `stderrors "errors"` import alongside.
</context>

<requirements>

1. **Audit `NewKeychainWithExecutor` external usage** before deleting. Two checks:
   - Open `https://pkg.go.dev/github.com/bborbe/teamvault-utils/v4?tab=importedby` and inspect for external importers. If the page lists any module other than `github.com/bborbe/teamvault-utils/v4` itself, treat as "consumer exists".
   - Run a GitHub code search: `https://github.com/search?q=teamvault.NewKeychainWithExecutor&type=code`. Inspect results for any file outside this repo.

   Record both result counts (verbatim) in the completion report.

   - No external consumers → delete `NewKeychainWithExecutor` outright in requirement 5.
   - At least one external consumer → keep a 4-line deprecation stub that returns `errors.Errorf(ctx, "deprecated: Keychain is no longer Executor-backed; mock the Keychain interface for tests")`. Stub stays in `keychain_darwin.go`.

2. **Pre-implementation probe**: write `/tmp/keyring-probe.go` that exercises zalando against the character classes our existing `validatePasswordForKeychain` rejects, AND captures the exact error type zalando returns when no backend is available. Use a unique service+user pair so it never collides with real credentials.

   ```go
   //go:build ignore
   package main

   import (
       "errors"
       "fmt"
       "os"

       "github.com/zalando/go-keyring"
   )

   func main() {
       svc := "teamvault-utils-probe"
       acc := "https://probe.test"
       defer keyring.Delete(svc, acc)

       cases := []struct{ name, pw string }{
           {"simple", "simple-password-1234"},
           {"comma", "eTdDWGDhgwmhiQibR8pKBqW,fa4tzXQU"},
           {"spaces+quotes", `pass "with" spaces`},
           {"backslash", `pass\\with\\backslash`},
           {"dollar+backtick", "p$$$`cmd`"},
           {"newline", "pass\nword"},
           {"nul", "pass\x00word"},
       }
       for _, c := range cases {
           setErr := keyring.Set(svc, acc, c.pw)
           if setErr != nil {
               fmt.Printf("SET %-16s FAIL  err=%T %v\n", c.name, setErr, setErr)
               continue
           }
           got, getErr := keyring.Get(svc, acc)
           if getErr != nil {
               fmt.Printf("GET %-16s FAIL  err=%T %v\n", c.name, getErr, getErr)
               continue
           }
           if got == c.pw {
               fmt.Printf("%-20s ROUNDTRIP_OK len=%d\n", c.name, len(got))
           } else {
               fmt.Printf("%-20s MUTATED  in_len=%d out_len=%d\n", c.name, len(c.pw), len(got))
           }
       }

       // Probe no-backend behavior (only meaningful on Linux; on darwin this should succeed)
       _, notFoundErr := keyring.Get("teamvault-utils-probe-unused", "https://no-such-entry.test")
       fmt.Printf("NOT_FOUND  err=%T  is_ErrNotFound=%v\n",
           notFoundErr, errors.Is(notFoundErr, keyring.ErrNotFound))

       os.Exit(0)
   }
   ```

   Run: `cd /tmp && go mod init probe 2>/dev/null; go get github.com/zalando/go-keyring 2>&1 | tail; go run keyring-probe.go`.

   **Capture the output verbatim in the completion report.** Decisions to make from the output:
   - If newline / NUL passwords return `MUTATED` or `FAIL`: keep `validatePasswordForKeychain` and call it as the first action of `WritePassword` (early-return on validation error).
   - If both newline and NUL round-trip cleanly: delete `validatePasswordForKeychain` — zalando handles it.
   - From the `NOT_FOUND` line: confirm `errors.Is(err, keyring.ErrNotFound)` returns `true`. If not, use string-match fallback in `ReadPassword`.

   Delete `/tmp/keyring-probe.go` and `/tmp/go.mod`, `/tmp/go.sum` after the probe. The verification grep guards against accidental retention.

3. **Add the dependency** in the project repo:
   ```
   cd /workspace
   go get github.com/zalando/go-keyring@latest
   go mod tidy
   ```
   Record the pinned version (from `go.mod`) in the completion report.

4. **Introduce `keyringClient` interface** at the top of `keychain_darwin.go`. This is the test seam. Production code uses the real wrapper; tests inject a Counterfeiter fake.

   ```go
   //counterfeiter:generate -o mocks/keyring_client.go --fake-name KeyringClient . keyringClient

   // keyringClient is the package-private seam over zalando/go-keyring used by
   // darwinKeychain. It exists so unit tests can drive WritePassword/ReadPassword
   // without touching the real macOS Keychain. NewKeychain wires up the real
   // implementation; tests construct darwinKeychain with a Counterfeiter fake.
   type keyringClient interface {
       Get(service, user string) (string, error)
       Set(service, user, password string) error
   }

   type realKeyringClient struct{}

   func (realKeyringClient) Get(service, user string) (string, error) {
       return keyring.Get(service, user)
   }

   func (realKeyringClient) Set(service, user, password string) error {
       return keyring.Set(service, user, password)
   }
   ```

   Run `make generate` (or whatever the project uses to regenerate Counterfeiter mocks) so `mocks/keyring_client.go` is produced before the unit tests run.

5. **Rewrite `keychain_darwin.go`** end-to-end. The new file has no `//go:build` tag — zalando works on darwin / linux / windows / freebsd / openbsd / dragonfly, and we want all of those to use this implementation. Platforms zalando does not support fall through to its error path which we wrap as `ErrKeychainNotSupported`.

   Final file shape (paste-ready outline; the agent fills in the parts marked PROBE):

   ```go
   // Copyright header (preserved verbatim from the existing file).

   package teamvault

   import (
       "context"

       "github.com/bborbe/errors"
       "github.com/golang/glog"
       "github.com/zalando/go-keyring"
   )

   //counterfeiter:generate -o mocks/keyring_client.go --fake-name KeyringClient . keyringClient

   type keyringClient interface {
       Get(service, user string) (string, error)
       Set(service, user, password string) error
   }

   type realKeyringClient struct{}

   func (realKeyringClient) Get(service, user string) (string, error) { return keyring.Get(service, user) }
   func (realKeyringClient) Set(service, user, password string) error { return keyring.Set(service, user, password) }

   // NewKeychain returns a Keychain backed by the OS credential store.
   // On macOS uses Keychain, on Linux uses Secret Service, on Windows uses Credential Manager.
   // On platforms without a supported backend, ReadPassword returns ("", nil) for missing entries
   // and Read/WritePassword return ErrKeychainNotSupported for no-backend errors.
   func NewKeychain() Keychain {
       return &darwinKeychain{client: realKeyringClient{}}
   }

   type darwinKeychain struct {
       client keyringClient
   }

   func (d *darwinKeychain) ReadPassword(ctx context.Context, url Url) (Password, error) {
       if url == "" {
           glog.V(3).Infof("keychain read skipped: empty URL")
           return "", nil
       }
       pwd, err := d.client.Get(KeychainServiceName, string(url))
       if err != nil {
           if errors.Is(err, keyring.ErrNotFound) {
               glog.V(3).Infof("keychain miss for url %q", url)
               return "", nil
           }
           if isNoBackendError(err) {
               return "", ErrKeychainNotSupported
           }
           glog.V(2).Infof("keychain read error for url %q: %v", url, err)
           return "", errors.Wrapf(ctx, err, "keychain read failed for url %q", url)
       }
       glog.V(3).Infof("keychain hit for url %q", url)
       return Password(pwd), nil
   }

   func (d *darwinKeychain) WritePassword(ctx context.Context, url Url, password Password) error {
       if url == "" {
           glog.V(3).Infof("keychain write skipped: empty URL")
           return nil
       }
       // PROBE-DEPENDENT: if requirement 2's probe shows zalando does NOT internally
       // validate NUL/newline, call validatePasswordForKeychain(ctx, password) here
       // and early-return on error. If zalando does validate them, omit this block.
       if err := d.client.Set(KeychainServiceName, string(url), string(password)); err != nil {
           if isNoBackendError(err) {
               return ErrKeychainNotSupported
           }
           glog.V(2).Infof("keychain write error for url %q: %v", url, err)
           return errors.Wrapf(ctx, err, "keychain write failed for url %q", url)
       }
       glog.V(2).Infof("keychain write succeeded for url %q", url)
       return nil
   }

   // PROBE-DEPENDENT: implementation of isNoBackendError comes from the probe in requirement 2.
   // It returns true when err indicates zalando has no usable credential backend on this platform.
   // Typical matcher (replace with the actual one identified by the probe):
   //   - errors.Is(err, keyring.ErrUnsupportedPlatform)   // if zalando exports a typed sentinel
   //   - errors.Is(err, keyring.ErrNoBackend)             // alternate typed sentinel
   //   - strings.Contains(err.Error(), "no usable") || strings.Contains(err.Error(), "unsupported platform")  // string fallback
   func isNoBackendError(err error) bool { /* implementation from probe */ }

   // PROBE-DEPENDENT: if zalando does NOT validate NUL/newline, retain this helper
   // (copy from the existing keychain_darwin.go) and call it from WritePassword.
   // Otherwise omit it entirely.
   ```

   **Hard requirements for this file:**
   - No `os/exec` import. No reference to `security` binary anywhere.
   - No `Executor` references. The `Executor` type is being deleted in the same prompt.
   - `isNoBackendError` MUST have a real body. Do not ship `return false`. The probe's output dictates the matcher.
   - `validatePasswordForKeychain` is either retained (with its existing body) and called from `WritePassword`, or deleted entirely — depending on the probe.
   - Public symbols `NewKeychain`, `darwinKeychain` (type name kept for git-history continuity, even though it now works on more than darwin — rename is out of scope) keep their signatures.

6. **Delete `keychain_executor.go`** — the `Executor` interface and `osExecutor` type are no longer referenced anywhere. Verify with `grep -rn 'Executor' --include='*.go' . | grep -v mocks/` returning 0 matches before deleting `mocks/executor.go`.

7. **Delete `keychain_other.go`** — the new `keychain_darwin.go` has no build tag, so it compiles on every platform zalando supports. Platforms zalando does not support fall through to `isNoBackendError` → `ErrKeychainNotSupported`, preserving the existing sentinel contract.

8. **Delete `mocks/executor.go`** — generated mock for the deleted `Executor` interface. Will be regenerated as `mocks/keyring_client.go` for the new `keyringClient` interface (via `make generate` in requirement 4).

9. **Rewrite unit tests in `keychain_darwin_test.go`**. Drop ALL existing `It` blocks that mocked the `Executor` interface (the seam is gone). The new test suite uses the Counterfeiter `mocks.KeyringClient` fake (generated in requirement 4) injected into a `darwinKeychain{client: fake}` constructed directly in the test. Required `It` blocks:

   a. **Read happy path**: `fake.GetReturns("hello", nil)`. Call `kc.ReadPassword(ctx, "https://example.test")`. Assert returned password is `"hello"`, no error. Assert `fake.GetCallCount() == 1`, and the args passed to `Get` were `("teamvault-utils", "https://example.test")`.

   b. **Read returns empty on `keyring.ErrNotFound`**: `fake.GetReturns("", keyring.ErrNotFound)`. Assert `ReadPassword` returns `("", nil)` (empty Password, nil error — existing semantic).

   c. **Read propagates other errors wrapped**: `fake.GetReturns("", stderrors.New("locked"))`. Assert error returned, wrapped via `bborbe/errors` (contains the original message).

   d. **Read returns `ErrKeychainNotSupported` on no-backend error**: configure `fake.GetReturns("", <the error class isNoBackendError matches>)` (use whatever the probe identified as the sentinel — `keyring.ErrUnsupportedPlatform` or equivalent). Assert `errors.Is(returnedErr, ErrKeychainNotSupported)` is true.

   e. **Empty URL on read**: `kc.ReadPassword(ctx, "")` returns `("", nil)` without calling the fake. Assert `fake.GetCallCount() == 0`.

   f. **Write happy path**: `fake.SetReturns(nil)`. Call `kc.WritePassword(ctx, "https://example.test", "secret")`. Assert no error. Assert `fake.SetCallCount() == 1` and args were `("teamvault-utils", "https://example.test", "secret")`.

   g. **Write propagates errors**: `fake.SetReturns(stderrors.New("locked"))`. Assert wrapped error returned.

   h. **Write returns `ErrKeychainNotSupported` on no-backend error**: parallel to (d).

   i. **Empty URL on write**: `kc.WritePassword(ctx, "", "secret")` returns nil without calling the fake.

   j. **(Conditional) NUL/newline rejection** — include only if requirement 2's probe showed zalando does NOT validate these. Call `WritePassword(ctx, "https://example.test", "foo\x00bar")` and `"foo\nbar"`. Assert non-nil error, `fake.SetCallCount() == 0` (validation must run before the call).

   Test file imports: `stderrors "errors"` for constructing plain test errors, `"github.com/zalando/go-keyring"` for the `ErrNotFound` sentinel, the project's standard Ginkgo/Gomega imports. The `mocks` import path is `github.com/bborbe/teamvault-utils/v4/mocks`.

   **Do not mock `mocks.Keychain` in this file.** `mocks.Keychain` exists for callers of the `Keychain` interface, not for the `darwinKeychain` implementation. Mocking it would test the mock instead of the real code.

10. **Rewrite `keychain_darwin_integration_test.go`** to call `teamvault.NewKeychain()` (the real zalando-backed implementation) instead of constructing through the deleted Executor seam. Pattern (one new `It` is sufficient — replaces the existing integration tests; keep the unique-service-name and `DeferCleanup` patterns):

    ```go
    It("round-trips a password through zalando go-keyring against the real OS keychain", func() {
        svc := fmt.Sprintf("teamvault-utils-it-%d", time.Now().UnixNano())
        url := teamvault.Url(svc)
        pwd := teamvault.Password(`integration "test" password with , and \\ chars`)
        kc := teamvault.NewKeychain()
        DeferCleanup(func() {
            _ = exec.Command("security", "delete-generic-password", "-s", svc, "-a", svc).Run()
        })

        // Skip if Keychain is locked
        if err := kc.WritePassword(ctx, "", pwd); err != nil {
            Skip(fmt.Sprintf("keychain probe failed; skipping: %v", err))
        }

        Expect(kc.WritePassword(ctx, url, pwd)).To(Succeed())
        got, err := kc.ReadPassword(ctx, url)
        Expect(err).NotTo(HaveOccurred())
        Expect(got).To(Equal(pwd))
    })
    ```

    The `Skip` probe writes to empty URL (which returns nil without touching the keychain) — replace with a real probe that detects lock: try a `keyring.Get` on a nonexistent unique service+account; if the returned error contains "locked" or "unlock", `Skip`.

11. **Update `CHANGELOG.md` `## Unreleased` section**. Append (do not duplicate the heading):

    ```
    - refactor: Migrate `keychain_darwin.go` from hand-rolled `security` shell-out to `github.com/zalando/go-keyring`. Eliminates the REPL-script construction and quoting logic that produced two bugs in v4.10–v0.9.10. As a side effect, Linux and Windows users now have a working credential store via Secret Service and Credential Manager. The internal `Executor` interface and `osExecutor` type are removed; `NewKeychainWithExecutor` is removed (or deprecated-stubbed if external consumers exist — see prompt completion report). See spec 004.
    ```

12. **Verify all 7 CLI binaries still build untouched**:
    ```
    cd /workspace
    go build ./cmd/teamvault-login ./cmd/teamvault-password ./cmd/teamvault-username ./cmd/teamvault-url ./cmd/teamvault-file ./cmd/teamvault-config-parser ./cmd/teamvault-config-dir-generator
    git diff --name-only cmd/   # must return empty
    ```

</requirements>

<constraints>
- Public Go API of `Keychain` interface, `NewKeychain`, `KeychainServiceName`, `ErrKeychainNotSupported` MUST NOT change.
- Backward-compatible read: Keychain entries written by `teamvault-login` from v4.10–v4.12 (service `teamvault-utils`, account = URL) must remain readable after migration.
- `NewKeychainWithExecutor`: audit external usage before deleting; if external consumers exist, leave a 4-line deprecation stub. Decision is recorded in the completion report.
- Errors wrapped via `github.com/bborbe/errors` (`errors.Wrapf`, `errors.Errorf`); never `fmt.Errorf` and never `errors.Wrapf(ctx, nil, ...)`.
- All exported items keep their GoDoc comments per `docs/dod.md`.
- Tests use Ginkgo/Gomega; the new `keyringClient` interface generates a Counterfeiter mock at `mocks/keyring_client.go` via `make generate`. The existing `mocks/keychain.go` (Counterfeiter mock of the `Keychain` interface) is the test seam for CONSUMERS of Keychain — never imported by `keychain_darwin_test.go`.
- All paths in `<context>` and code are repo-relative or container-absolute (`/workspace`, `/home/node/.claude/...`). No host paths (`~/`, `/Users/`).
- All five scenarios (`scenarios/001`–`005`) continue to pass without modification.
- Do NOT commit — dark-factory handles git.
- Do NOT leave the probe file or its `go.mod`/`go.sum` on disk after the probe completes.
- `go.mod` must declare `github.com/zalando/go-keyring` at a pinned version.
- The Go build must succeed under default (CGO enabled) and with `CGO_ENABLED=0`. No new cgo dependencies.
</constraints>

<verification>
- `cd /workspace && go mod tidy` exits 0; `grep -n 'zalando/go-keyring' go.mod` returns ≥1 match with a pinned version.
- `grep -nE 'security |osExecutor|Executor|exec\.Command|exec\.CommandContext|REPL|quit' keychain_darwin.go` returns 0 matches.
- `grep -n 'keyring\.Get\|keyring\.Set\|d\.client\.Get\|d\.client\.Set' keychain_darwin.go` returns ≥4 matches (two real wrapper calls + two darwinKeychain method calls into the interface).
- `grep -n 'isNoBackendError' keychain_darwin.go` returns ≥3 matches AND the function body is NOT `return false` — verify via `grep -A1 'func isNoBackendError' keychain_darwin.go` showing real logic.
- `[ ! -f keychain_executor.go ]` true.
- `[ ! -f keychain_other.go ]` true.
- `[ ! -f mocks/executor.go ]` true.
- `[ -f mocks/keyring_client.go ]` true.
- `[ ! -f /tmp/keyring-probe.go ]` true (cleanup verified); `[ ! -f /tmp/go.mod ]` true.
- `go test ./... -run KeychainDarwin -v` passes; the new `mocks.KeyringClient`-based unit tests are present.
- `go test ./... -tags integration` passes on darwin (or `Skip`s gracefully when Keychain locked).
- `cd /workspace && go build ./cmd/teamvault-login ./cmd/teamvault-password ./cmd/teamvault-username ./cmd/teamvault-url ./cmd/teamvault-file ./cmd/teamvault-config-parser ./cmd/teamvault-config-dir-generator` exits 0.
- `git diff --name-only cmd/` returns empty.
- `grep -n '## Unreleased' CHANGELOG.md` returns ≥1; `grep -ni 'zalando\|go-keyring' CHANGELOG.md` returns ≥1 line under that heading.
- `make precommit` exits 0.
- **Completion report MUST record**:
  - The probe output for all 7 password cases (PASS/FAIL/MUTATED) AND the `NOT_FOUND` error type observed.
  - The decision: was `validatePasswordForKeychain` retained or dropped, and why (cite probe row).
  - The `isNoBackendError` matcher chosen and the rationale (cite probe row or library doc).
  - The `NewKeychainWithExecutor` audit result: external consumers count from pkg.go.dev and from GitHub code search. Decision: deleted or deprecation-stubbed.
  - The pinned `github.com/zalando/go-keyring` version.
- Operator action (not automatable in this prompt): walk `scenarios/002-keychain-login-and-retrieve.md` against the freshly built binary and confirm every checkbox green; flag failures back to the spec for verification phase.
</verification>

<human-finish-note>
Container hit the Go *_darwin.go filename-implicit-build-constraint trap and looped on tag widening for 37min before being killed. Human (Benjamin) completed the work manually in commit dc33341: renamed keychain_darwin.go → keychain_impl.go (the only way to widen the platform set past the filename-implicit darwin constraint), exported KeyringClient interface to avoid the mocks-package import cycle, added NewKeychainWithClient test-injection constructor, and moved tests to package teamvault_test. All ACs from spec 004 are covered by the shipped code: zalando/go-keyring v0.2.8 in go.mod, security/REPL references gone from keychain_impl.go, validatePasswordForKeychain retained (zalando does NOT validate NUL/newline), isNoBackendError maps ErrUnsupportedPlatform + dbus-launch missing to ErrKeychainNotSupported, NewKeychainWithExecutor deleted (no external audit performed — risk accepted, lib is small), keychain_other.go deleted, mocks/executor.go deleted, mocks/keyring_client.go regenerated, CHANGELOG ## Unreleased entry added. Backward-compat read verified manually against a v4.10-era raw security entry. make precommit exits 0. Scenario 002 re-walk pending.
</human-finish-note>
