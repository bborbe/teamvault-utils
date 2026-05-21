---
status: draft
spec: [003-bug-keychain-write-empty-password-on-piped-stdin]
created: "2026-05-21T17:35:00Z"
branch: dark-factory/bug-keychain-write-empty-password-on-piped-stdin
---

<summary>
- Fix `teamvault-login` writing an empty password to macOS Keychain when stdin is piped (non-interactive shell)
- Root cause: `security add-generic-password -w` without a positional argument prompts on `/dev/tty`, not stdin; the password passed via `cmd.Stdin` is never read
- Solution: use `security -i` interactive mode with the password passed via stdin to the interactive session, avoiding `/dev/tty` prompt and argv exposure
- If `security -i` stdin approach doesn't work, fall back to direct macOS Keychain Services API via cgo
- Unit tests verify the executor receives the password in a way `security` will actually consume
- Integration test verifies real `security` round-trip with a piped password via temporary service name
- Existing tests continue to pass; existing scenario `scenarios/002-keychain-login-and-retrieve.md` should pass after fix
</summary>

<objective>
Fix `teamvault-login` so it reliably writes the verified password to the macOS Keychain in both interactive and non-interactive (stdin-piped) invocations. The password must arrive at Keychain byte-for-byte identical to what the user provided.
</objective>

<context>
Project: `~/Documents/workspaces/teamvault/teamvault-utils/` — Go library + CLI tools under module `github.com/bborbe/teamvault-utils/v4`.

Read first:
- `CLAUDE.md` for project conventions.
- `keychain_darwin.go` — the bug location. Current `WritePassword` at lines 73-93 passes password via `cmd.Stdin` but `security -w` without a value prompts on `/dev/tty`, not stdin.
- `keychain_executor.go` — the `Executor` interface. `Run(ctx, name, args, stdin) (stdout, stderr, exitCode, err)`.
- `keychain_darwin_test.go` — existing unit tests for `ReadPassword` and `WritePassword`.
- `keychain_darwin_integration_test.go` — existing integration test pattern using unique service names.
- `keychain.go` — `Keychain` interface definition (DO NOT change public API).
- `mocks/executor.go` — Counterfeiter mock for `Executor`.

Read these coding guides:
- `~/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md` — error wrapping with `github.com/bborbe/errors`
- `~/.claude/plugins/marketplaces/coding/docs/go-testing-guide.md` — Ginkgo/Gomega + Counterfeiter conventions
- `~/.claude/plugins/marketplaces/coding/docs/test-pyramid-triggers.md` — which test types to write

**Spike first:** Before committing to an implementation, run a 10-line Go spike to verify `security -i` interactive mode accepts `add-generic-password` via stdin and actually stores the password. See requirement 1.

**Constraint reminder:** Approach #2 (`security -w "$PASS"` positional) is explicitly rejected by the spec — the password would be visible in `ps` output.
</context>

<requirements>

1. **Spike `security -i` approach** (10 lines, temporary file):

   Write a temporary Go file that:
   - Runs `security -i` via `exec.Command`
   - Writes `add-generic-password -U -s teamvault-utils-test-spike -a https://spike.test -w SPIKE_PASSWORD\nquit\n` to stdin
   - Waits for completion and checks exit code
   - Cleans up with `security delete-generic-password -s teamvault-utils-test-spike -a https://spike.test 2>/dev/null`
   - Reads back with `security find-generic-password -s teamvault-utils-test-spike -a https://spike.test -w`
   - Verifies the stored password equals `SPIKE_PASSWORD`

   If the spike works: proceed with the `security -i` approach (requirement 2a).
   If the spike fails: proceed with the cgo approach (requirement 2b).

2. **Implement the fix** — choose the approach that passed the spike:

   **2a. `security -i` interactive mode (if spike succeeded):**

   Modify `keychain_darwin.go` `WritePassword` to use `security -i` interactive mode:

   ```go
   func (d *darwinKeychain) WritePassword(ctx context.Context, url Url, password Password) error {
       // Use security -i interactive mode so the password is sent via stdin,
       // not as a -w positional argument visible in ps, and not prompting on /dev/tty.
       // The stdin pipe works reliably in both interactive and non-interactive shells.
       script := fmt.Sprintf("add-generic-password -U -s %s -a %s -w %s\nquit\n",
           KeychainServiceName, url, password)
       _, stderr, exitCode, err := d.executor.Run(ctx, "security", []string{"-i"}, script)
       if err != nil {
           return errors.Wrapf(ctx, err, "execute security command failed")
       }
       if exitCode != 0 {
           return errors.Errorf(ctx, "security add-generic-password failed with exit code %d: %s", exitCode, stderr)
       }
       glog.V(2).Infof("keychain write succeeded for url %q", url)
       return nil
   }
   ```

   **2b. cgo Keychain Services API (if spike failed):**

   Create `keychain_darwin_cgo.go` with build tag `//go:build darwin && cgo`:

   - Use `github.com/bborbe/errors` for error wrapping
   - Call `SecKeychainAddGenericPassword` directly from the macOS Keychain Services API
   - Use `SecKeychainFindGenericPassword` to check for existing entry and `SecKeychainItemModifyAttributes` to update, or delete and recreate
   - Keep the same `KeychainServiceName` constant and URL as account
   - The cgo file should live alongside `keychain_darwin.go` and be build-constrained to `darwin && cgo`
   - `NewKeychain()` should detect cgo availability and use the appropriate implementation
   - The public `Keychain` interface MUST NOT change — `ReadPassword` and `WritePassword` signatures stay identical

   **Important:** If using cgo, do NOT pass the password as a string to any C function that logs it or stores it in a way visible to `ps`. Use `*C.char` and `C.Func` directly with proper memory management.

3. **Unit tests** in `keychain_darwin_test.go`:

   Extend the `Describe("WritePassword", ...)` block with new `Context` cases:

   a. **Password with shell metacharacters** (`$`, backticks, spaces):
      - Call `WritePassword` with a password containing `$FOO`, `` `cmd` ``, and spaces
      - Assert executor was called with the password appearing verbatim in the command script (not truncated at whitespace)
      - Assert no error returned

   b. **Password containing a newline** (`foo\nbar`):
      - Call `WritePassword` with `Password("foo\nbar")`
      - Assert executor was called with the newline appearing verbatim in the script
      - Assert no error returned OR assert a clear error is returned before invoking `security` (either is acceptable per spec)
      - Document which behavior was chosen

   c. **Password containing NUL byte**:
      - Call `WritePassword` with `Password("foo\x00bar")`
      - Assert a non-nil error is returned BEFORE calling the executor (NUL bytes are rejected at the Go layer per spec Failure Modes table)
      - Assert error message mentions NUL or is otherwise actionable

   d. **Executor receives script with password via stdin** (regression guard for the bug fix):
      - Call `WritePassword` with a known password
      - Assert the 4th argument (stdin) to the executor's `Run` call contains the password in a form `security -i` will consume
      - Assert the args do NOT contain `-w` without a positional password argument (the old buggy pattern)

   e. **Keychain locked error path** (extend existing):
      - Simulate executor returning exit code 36 with "could not be unlocked" in stderr
      - Assert wrapped error contains "Keychain locked" or "unlock" hint

4. **Integration test** in `keychain_darwin_integration_test.go`:

   Add a new `It` test:

   ```go
   It("round-trips a password written via stdin-piped invocation", func() {
       // This test verifies the fix for the non-interactive stdin-piped bug.
       // Uses a unique service name to avoid clobbering real entries.
       const testPwd = teamvault.Password("stdin-test-password-12345")
       const testService = fmt.Sprintf("teamvault-utils-integration-test-stdin-%d", time.Now().UnixNano())
       const testURL = teamvault.Url(testService)

       By("writing the password via the normal (non-spike) code path")
       // The integration test uses the real NewKeychain(), not the spike
       realKeychain := teamvault.NewKeychain()
       // But we use a temp service name to avoid polluting real entries
       // We need to test via the actual code path, so we write to a test service
       // then read back. Since NewKeychain() uses the real service name,
       // we test via a temporary workaround: write to test service, verify via executor observation.
       // Actually: use the real keychain but a test URL that won't conflict.
       // The real fix uses security -i; the integration test verifies round-trip.
       Expect(realKeychain.WritePassword(ctx, testURL, testPwd)).To(Succeed())

       By("reading it back")
       got, err := realKeychain.ReadPassword(ctx, testURL)
       Expect(err).NotTo(HaveOccurred())
       Expect(got).To(Equal(testPwd))
   })
   ```

   Note: The integration test above uses the real `NewKeychain()` which writes to `teamvault-utils` service. If this conflicts with the existing integration test, create a second `BeforeEach`/`AfterEach` pair that uses a distinct service name and clean up after both tests.

   **Alternative simpler integration test:** If the above is too complex, just add `stdin-test-password-12345` as an additional hardcoded test password in the existing integration test block and verify round-trip.

5. **Verify existing tests still pass:**

   After implementing, run `go test ./... -run KeychainDarwin -v` to confirm existing test cases still pass.

6. **Re-walk scenario (manual verification):**

   After `make precommit` passes, manually run the scenario at `scenarios/002-keychain-login-and-retrieve.md` against the fixed binary and confirm all checkboxes turn green. This is a manual verification step — do not automate it in the prompt.

</requirements>

<constraints>
- Password MUST NOT appear as a literal command-line argument visible in `ps` output (Constraint #2 from spec). The `-w` flag without a positional value prompts on `/dev/tty`; that's the bug. The fix must send the password via stdin to the `security` process in a way that works in both interactive and non-interactive shells.
- Public Go API of `Keychain` interface (`ReadPassword`, `WritePassword`) MUST NOT change.
- `KeychainServiceName` constant stays `"teamvault-utils"`.
- All changes confined to `keychain_darwin.go`, `keychain_executor.go` (if needed for the spike), and their tests. New darwin-specific files are allowed if build-tagged appropriately.
- Existing unit tests in `keychain_darwin_test.go` must still pass unchanged.
- Existing integration test `keychain_darwin_integration_test.go` must still pass unchanged.
- Use `github.com/bborbe/errors` for all error wrapping; never `fmt.Errorf`.
- Tests use Ginkgo/Gomega; mocks via Counterfeiter under `mocks/`.
- Do NOT commit — dark-factory handles git.
</constraints>

<verification>
- `make precommit` exits 0.
- `go test ./... -run KeychainDarwin` passes all existing and new tests.
- `go test ./... -tags integration` passes including the new integration test.
- `grep -n 'security.*-i' keychain_darwin.go` returns at least 1 match (if using -i approach) OR `grep -n 'SecKeychainAddGenericPassword' keychain_darwin_cgo.go` returns at least 1 match (if using cgo approach).
- `grep -n 'NUL\|Nul\|\\x00' keychain_darwin.go` returns at least 1 match (NUL rejection at Go layer).
- Scenario `scenarios/002-keychain-login-and-retrieve.md` passes when manually re-walked (checkbox evidence required in completion report).
- CHANGELOG `## Unreleased` entry added per spec requirement: "fix: teamvault-login now reliably stores the password in macOS Keychain when invoked with stdin piped (non-interactive shell), fixing a bug where the password was silently stored as empty."
</verification>
