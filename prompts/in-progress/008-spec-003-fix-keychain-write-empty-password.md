---
status: committing
spec: [003-bug-keychain-write-empty-password-on-piped-stdin]
summary: Fixed darwinKeychain.WritePassword to use security -i REPL mode (or cgo Keychain Services API as fallback), eliminating the bug where piping stdin to teamvault-login stored an empty password
container: teamvault-utils-exec-008-spec-003-fix-keychain-write-empty-password
dark-factory-version: v0.164.0
created: "2026-05-21T17:35:00Z"
queued: "2026-05-21T17:39:17Z"
started: "2026-05-21T17:39:18Z"
branch: dark-factory/bug-keychain-write-empty-password-on-piped-stdin
---

<summary>
- Fix `teamvault-login` writing an empty password to the macOS Keychain when invoked with stdin piped (non-interactive shell).
- Root cause: `security add-generic-password -w` with no positional value prompts on `/dev/tty`, not stdin; the password passed via `cmd.Stdin` is silently ignored.
- Fix replaces the `security` subprocess invocation in `darwinKeychain.WritePassword` with `security -i` interactive mode, which reads commands from stdin so the password never appears in the process argv (no `ps` exposure) and the prompt-on-tty path is bypassed.
- A spike script verifies `security -i` actually accepts and stores a password before committing to the approach. If the spike fails on a required character class (notably double-quote or newline inside the password), the prompt falls back to the cgo Keychain Services path.
- Boundary-contract test: an integration test calls the real `security` binary end-to-end with a piped password and reads it back, using a temp service name to avoid clobbering real entries.
- Unit tests cover metacharacter handling, NUL-byte rejection, locked-Keychain error path, and the regression-guard that the executor never receives the old buggy argv shape.
- Existing tests, existing public `Keychain` interface, and existing `KeychainServiceName` constant are unchanged.
</summary>

<objective>
Make `darwinKeychain.WritePassword` store the verified password byte-for-byte in the macOS login Keychain in both interactive AND non-interactive (stdin-piped) `teamvault-login` invocations, without exposing the password via process argv. The fix must keep the public `Keychain` interface unchanged and pass the existing scenario `scenarios/002-keychain-login-and-retrieve.md` end-to-end.
</objective>

<context>
Project: `~/Documents/workspaces/teamvault/teamvault-utils/` — Go library + CLI tools under module `github.com/bborbe/teamvault-utils/v4`.

Read first:
- `CLAUDE.md` for project conventions.
- `specs/in-progress/003-bug-keychain-write-empty-password-on-piped-stdin.md` for the full bug spec including the Alternatives section and Constraint #2 (no `ps`-visible password).
- `keychain.go` — `Keychain` interface and `KeychainServiceName` constant. Public API; do NOT change.
- `keychain_darwin.go` — `darwinKeychain.WritePassword` (the bug) and `darwinKeychain.ReadPassword` (reference). Anchor by function name.
- `keychain_executor.go` — `Executor` interface signature `Run(ctx, name, args, stdin) (stdout, stderr, exitCode, err)`.
- `keychain_darwin_test.go` — existing Ginkgo suite; extend the `Describe("WritePassword", ...)` block, do not modify existing `It` blocks.
- `keychain_darwin_integration_test.go` — existing real-`security` integration test using a unique per-test service name (`teamvault-utils-test-<...>`). Mirror this pattern for the new integration test.
- `mocks/executor.go` — Counterfeiter `Executor` fake.
- `scenarios/002-keychain-login-and-retrieve.md` — manual end-to-end verification scenario; currently fails on the "Keychain entry idempotent" + "password non-empty" assertions because of this bug.

Read these coding guides (in-container paths under `/home/node/.claude/plugins/marketplaces/coding/docs/`):
- `go-error-wrapping-guide.md` — wrap with `github.com/bborbe/errors` (`errors.Wrapf`, `errors.Errorf`); never `fmt.Errorf`.
- `go-testing-guide.md` — Ginkgo/Gomega + Counterfeiter conventions.
- `test-pyramid-triggers.md` — which test types are appropriate.

Key facts about `security -i`:
- `security -i` runs an interactive REPL that reads one subcommand per line from stdin.
- Each line is tokenized by whitespace; double-quoted segments group tokens; backslash escapes `"` and `\`.
- The OS sees only `security -i` in argv — the password on a REPL line lives in the subprocess stdin pipe, NOT in argv, so it is invisible to `ps`.
- Newline inside the password cannot be sent (it terminates the REPL line). Other shell-metacharacters (`$`, backticks, `;`) are NOT shell-interpreted because `security -i` is not a shell — they pass through to `security`'s own tokenizer.

Constraint reminder (spec §Constraints #2): the password MUST NOT appear in any process's argv (`ps` / `/proc`). The `-w "$PASS"` positional form is therefore rejected; `-i` is required.
</context>

<requirements>

1. **Spike `security -i` viability with realistic password shapes.** Write a temporary Go file at `/tmp/security-i-spike.go` (delete after the spike) that:

   - Runs `security -i` via `exec.Command` with its stdin connected to a `strings.Reader` containing the REPL script.
   - For each of the following test passwords, run the REPL with:
     ```
     add-generic-password -U -s teamvault-utils-spike -a https://spike.test -w <PASSWORD>
     quit
     ```
     where `<PASSWORD>` is substituted following these quoting rules: if the password contains a space, double-quote, or backslash, wrap in `"…"` and escape internal `"` and `\` with backslash; otherwise pass raw.
   - Test passwords (run each, verify round-trip via `security find-generic-password -s teamvault-utils-spike -a https://spike.test -w`):
     a. `eTdDWGDhgwmhiQibR8pKBqW,fa4tzXQU` (the real reporter password — contains comma)
     b. `simple-alphanumeric-123`
     c. `pass with spaces`
     d. `pass$with`backticks`and$dollars` (note: backtick-style chars; security is not a shell so these should pass through)
     e. `pass"with"quotes` (must be escape-encoded as `pass\"with\"quotes` inside the `"..."` wrapper)
     f. `pass\with\backslashes` (escape as `pass\\with\\backslashes`)
   - After each test, `security delete-generic-password -s teamvault-utils-spike -a https://spike.test` to clean up.
   - Print PASS/FAIL per test password.

   **Spike acceptance criterion:** all six test passwords above MUST round-trip byte-for-byte. If any fail (especially e or f), report which classes failed and proceed to requirement 2b. If all pass, proceed to requirement 2a.

   Delete `/tmp/security-i-spike.go` after the spike completes. Do NOT commit the spike file.

2. **Implement the fix in `darwinKeychain.WritePassword`** based on the spike outcome:

   **2a. `security -i` approach (if spike fully passed):**

   Replace the body of `darwinKeychain.WritePassword` (locate the function by name, not line number) with a `security -i` invocation. Outline (final code should follow project style):

   - Build a single REPL script: `add-generic-password -U -s <SERVICE> -a <URL> -w <QUOTED_PASSWORD>\nquit\n`.
   - Quote `<QUOTED_PASSWORD>` per the rule established in requirement 1 (wrap in `"…"` and escape internal `"` and `\` whenever the raw password contains space / `"` / `\`; pass raw otherwise).
   - Reject passwords containing a NUL byte (`\x00`) OR a newline (`\n`) BEFORE invoking `security` — return a wrapped error from `github.com/bborbe/errors` mentioning the unsupported character class. Use a small helper (`func validatePasswordForKeychain(ctx, p Password) error`) and call it as the first action of `WritePassword`. Place the helper in `keychain_darwin.go`.
   - Invoke `d.executor.Run(ctx, "security", []string{"-i"}, script)`.
   - On non-zero exit code, wrap the existing keychain-locked detection and the generic failure return — match the style of the current `ReadPassword` error branches (lines around `exitCode == 36` / "could not be unlocked").
   - On success, `glog.V(2).Infof("keychain write succeeded for url %q", url)`.

   The `URL` is already type-checked at compile time and contains only printable ASCII per existing usage; do not add new validation for the service or account values beyond the empty-string guard the existing code uses.

   **2b. cgo Keychain Services API (only if the spike failed):**

   Create a new file `keychain_darwin_cgo.go` with `//go:build darwin && cgo`. Reimplement `WritePassword` using `SecKeychainAddGenericPassword` / `SecKeychainItemModifyAttributesAndData` from `<Security/Security.h>`. Constraints:

   - The public `Keychain` interface stays as-is — `NewKeychain()` returns the cgo-backed implementation on darwin, the no-op stub on other platforms.
   - Same NUL/newline pre-validation as 2a.
   - Use `C.CString` with `C.free` deferred for every C-allocated string.
   - On error, wrap via `github.com/bborbe/errors`; do NOT log the password.
   - Move the existing `keychain_darwin.go` `WritePassword` body to a new file `keychain_darwin_security.go` (with the same `//go:build darwin` tag) so cgo and non-cgo darwin implementations can co-exist if needed; gate selection inside `NewKeychain()`.

   In either branch, **`ReadPassword` is unchanged.**

3. **Unit tests** in `keychain_darwin_test.go`. Add new `Context` blocks under the existing `Describe("WritePassword", ...)`. Do not modify existing `It` blocks.

   a. **Regression guard against the old buggy argv shape.** Call `WritePassword` via the fake `Executor` with any non-trivial password. Assert:
      - The recorded `args` slice equals `[]string{"-i"}` (no `-w` token in args).
      - The recorded `stdin` argument contains the literal `add-generic-password -U -s teamvault-utils -a` prefix.
      - The recorded `stdin` argument contains the password content (after quoting rules) on the same REPL line as `-w`.

   b. **Metacharacter passthrough.** Call `WritePassword` with `Password("pass$with`backticks")`. Assert the stdin contains the literal `$` and backtick characters byte-for-byte (no shell interpretation), and no error is returned.

   c. **Space and quote-bearing password is correctly quoted.** Call `WritePassword` with `Password(`hello "world" foo`)`. Assert the stdin contains the REPL line with the wrapper `"hello \"world\" foo"` (or whatever exact escape form the implementation chose — assert against the implementation's documented quoting rule, not a guess).

   d. **NUL byte rejected before invoking security.** Call `WritePassword` with `Password("foo\x00bar")`. Assert the executor was NOT called (`Run` call count is 0). Assert the returned error wraps something mentioning "NUL" or "null byte" or similar — exact wording at implementation's discretion.

   e. **Newline rejected before invoking security.** Same shape as (d), with `Password("foo\nbar")`. Assert executor call count is 0 and an error is returned naming "newline" or similar.

   f. **Locked Keychain error path.** Configure the executor fake to return exit code 36 with stderr containing `"could not be unlocked"`. Assert the returned error message contains "Keychain" + "unlock" (matching the existing `ReadPassword` style at line referenced by `exitCode == 36`).

   g. **Successful write.** Executor returns exit code 0; assert `WritePassword` returns nil and `glog` is called at verbosity 2 (skip if glog inspection isn't already a pattern in existing tests).

4. **Integration test** in `keychain_darwin_integration_test.go`. Add ONE new `It` block under the existing `Describe`. Use the existing per-test unique-service-name pattern (`var` not `const`; `fmt.Sprintf` cannot initialize a Go `const`):

   ```go
   It("round-trips a password written via the new code path", func() {
       svc := fmt.Sprintf("teamvault-utils-it-%d", time.Now().UnixNano())
       url := teamvault.Url(svc)
       pwd := teamvault.Password("integration,test,password,with,commas,and \"quotes\"")
       kc := teamvault.NewKeychain()
       DeferCleanup(func() {
           _ = exec.Command("security", "delete-generic-password", "-s", svc, "-a", svc).Run()
       })
       Expect(kc.WritePassword(ctx, url, pwd)).To(Succeed())
       got, err := kc.ReadPassword(ctx, url)
       Expect(err).NotTo(HaveOccurred())
       Expect(got).To(Equal(pwd))
   })
   ```

   - Use a unique service-name per test run (timestamp + nano) — never the production `teamvault-utils` service.
   - `DeferCleanup` deletes the temp entry regardless of test outcome.
   - The Keychain must be unlocked for the test to run. Add a `BeforeEach` (or extend the existing one) that probes whether `security` can write to a no-op entry; on lock detection (`exit code 36`), call `Skip("Keychain locked; skipping integration test")` so the test is unattended-safe.

5. **CHANGELOG entry.** Append under the existing `## Unreleased` heading in `CHANGELOG.md`:

   ```
   - fix: `teamvault-login` now reliably stores the password in the macOS Keychain when invoked with stdin piped (non-interactive shell). The previous implementation silently stored an empty password because `security add-generic-password -w` without a positional value prompts on `/dev/tty`. Fix uses `security -i` REPL mode (or the Keychain Services API via cgo as fallback) so the password is sent via stdin and never appears in `ps` output. See spec 003.
   ```

6. **Verify existing tests still pass:** `go test ./... -run KeychainDarwin -v` exits 0 and all existing `It` blocks remain green.

7. **Re-walk scenario manually** (operator action, NOT automatable in the prompt): after the daemon's commit, walk `scenarios/002-keychain-login-and-retrieve.md` against the freshly-built binary and confirm every checkbox turns green. The spec-verifier will demand this evidence at verification time.

</requirements>

<constraints>
- Password MUST NOT appear in any process's argv (visible to `ps` / `/proc`). The `-w "$PASS"` positional form is rejected; `security -i` (REPL stdin) or cgo Keychain Services are the only acceptable channels.
- Public Go API of the `Keychain` interface (`ReadPassword`, `WritePassword` signatures) MUST NOT change.
- `KeychainServiceName` constant stays `"teamvault-utils"`. Tests use temp service names; production code does not.
- All changes confined to `keychain.go`, `keychain_darwin.go`, optional new `keychain_darwin_security.go` and/or `keychain_darwin_cgo.go`, `keychain_executor.go` (if absolutely necessary; prefer leaving alone), and their test files. The CHANGELOG is the only file outside the keychain layer that must change.
- Existing unit tests in `keychain_darwin_test.go` and existing integration test cases in `keychain_darwin_integration_test.go` must still pass unchanged.
- Use `github.com/bborbe/errors` exclusively for error construction and wrapping (`errors.Errorf`, `errors.Wrapf`). Never `fmt.Errorf`.
- Tests use Ginkgo/Gomega; mocks via Counterfeiter under `mocks/`.
- All paths in code and tests must be repo-relative (no host-absolute paths).
- Integration test must be unattended-safe: `Skip` when Keychain is locked or when not running on darwin.
- Do NOT commit — dark-factory handles git.
- Do NOT leave the spike file (`/tmp/security-i-spike.go`) on disk after the spike completes.
</constraints>

<verification>
- `make precommit` exits 0.
- `go test ./... -run KeychainDarwin -v` passes (existing + new unit tests).
- `go test ./... -tags integration -run KeychainDarwin` passes (existing + new integration test) on darwin with an unlocked Keychain; `Skip`s gracefully otherwise.
- `grep -n 'WritePassword' keychain_darwin.go` returns the method definition; the body uses `security -i` (per requirement 2a) — verify with `grep -n '"-i"' keychain_darwin.go` returning ≥1 match, OR cgo branch chosen — verify with `grep -n 'SecKeychainAddGenericPassword' keychain_darwin_cgo.go` returning ≥1 match.
- `grep -n -- '-w" *,' keychain_darwin.go` returns 0 matches (no leftover buggy `-w` invocation in `WritePassword`).
- `grep -n 'NUL\|null byte\|0x00\|\\x00' keychain_darwin.go` returns ≥1 match in the new validation helper (NUL rejection at the Go layer).
- CHANGELOG: `grep -n '## Unreleased' CHANGELOG.md` returns ≥1 line; `grep -n -i 'keychain.*stdin\|stdin.*keychain\|security -i' CHANGELOG.md` returns ≥1 line under that heading.
- The completion report (DARK-FACTORY-REPORT) MUST explicitly state which branch (2a `security -i` or 2b cgo) the spike steered the implementation toward, and the per-password spike results for the six test cases in requirement 1.
</verification>
