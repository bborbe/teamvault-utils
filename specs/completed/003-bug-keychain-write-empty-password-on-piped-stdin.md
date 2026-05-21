---
status: completed
approved: "2026-05-21T17:29:49Z"
generating: "2026-05-21T17:29:50Z"
prompted: "2026-05-21T17:33:04Z"
completed: "2026-05-21T21:32:25Z"
branch: dark-factory/bug-keychain-write-empty-password-on-piped-stdin
---

## Summary

- `teamvault-login` silently overwrites the macOS Keychain entry with an **empty** password when invoked with stdin piped (non-interactive shell).
- Root cause: `keychain_darwin.go:73-90` invokes `security add-generic-password -U ... -w` without a positional password argument and passes the password as `cmd.Stdin`. `security`'s `-w` flag without a value prompts on `/dev/tty`; the stdin pipe is ignored, the prompt fails silently in a non-interactive context, and security stores an empty string.
- Effect: after a non-interactive login, every other `teamvault-*` binary subsequently 403s against the real API. Reauthentication requires either interactive `teamvault-login` (which works because `/dev/tty` is the live terminal) or direct `security add-generic-password -w <PASS> ...`.
- Discovered by `scenarios/002-keychain-login-and-retrieve.md` walk: scenario asserted "subsequent `teamvault-password` succeeds via Keychain"; it failed with HTTP 403.

## Problem

The keychain-fallback workflow is the documented macOS-recommended setup (README "Setup (macOS, recommended)") and the entire justification for shipping v4.10 (spec 001). Any user who scripts `teamvault-login` — for example, automating multi-vault setup, restoring credentials from a password manager via stdin, or running it from a non-interactive provisioning script — silently destroys their working Keychain entry on first invocation. The bug is not surfaced by any error: `teamvault-login` prints `Login successful. Password stored in macOS Keychain for <url>.` even though it stored nothing.

## Reproduction

`dark-factory --version`: not applicable (bug is in `teamvault-utils`, not dark-factory). `teamvault-utils` version: v4.12.0 (latest tag at time of report). macOS 26.4.1.

Minimum config (no `pass` field, URL + user only):

```
$ TV_URL=$(jq -r .url ~/.teamvault.json)
$ TV_USER=$(jq -r .user ~/.teamvault.json)
$ TV_PASS=$(security find-generic-password -s teamvault-utils -a "$TV_URL" -w)  # real working password
$ WORK_DIR=$(mktemp -d)
$ printf '{"url":"%s","user":"%s"}\n' "$TV_URL" "$TV_USER" > "$WORK_DIR/teamvault.json"
$ printf '%s\n' "$TV_PASS" | teamvault-login --teamvault-config "$WORK_DIR/teamvault.json"
Login successful. Password stored in macOS Keychain for https://teamvault.benjamin-borbe.de.
$ echo "Keychain length now: $(security find-generic-password -s teamvault-utils -a "$TV_URL" -w | wc -c)"
Keychain length now:        1   # (just the trailing newline — password is empty)
$ teamvault-password --teamvault-config "$WORK_DIR/teamvault.json" --teamvault-key lO4K1w
E... get password failed: request to https://teamvault.benjamin-borbe.de/api/secrets/lO4K1w/ failed with status: 403
```

Observed during `scenarios/002-keychain-login-and-retrieve.md` walk on 2026-05-21:

```
LOGIN_RC=0 LOGIN_OUT=[Login successful. Password stored in macOS Keychain for https://teamvault.benjamin-borbe.de.]
PW_RC=1 USER_RC=1 USER_OUT=[] PW_LEN=0
KC matches replay: no
✗ keychain entry changed
✗ password RC=1
✗ password empty
✗ password prompt in pw stderr: get password failed: ... status: 403
```

Direct keychain inspection after the run: `security find-generic-password -s teamvault-utils -a <url> -w` returned an empty string (0 chars + trailing newline).

## Expected vs Actual

| | Expected | Actual |
|---|---|---|
| After `teamvault-login` succeeds | Keychain stores the verified password (per `cmd/teamvault-login/main.go:175-198` `writeAndReport`) | Keychain stores empty string |
| Stdout/stderr | Reflects success or failure of Keychain write | Reports "Login successful. Password stored..." regardless of what was actually stored |
| Subsequent `teamvault-password` against same URL | Reads working password from Keychain, returns secret | 403 — empty password sent to TeamVault |

Expected behavior is documented by README's "Setup (macOS, recommended)" section (lines 32-69) and by spec 001 §"`teamvault-login` binary" step 3: "If verification succeeds … store it in the Keychain (idempotent: overwrites any existing entry for the same URL)." The word "it" is the verified password; storing empty contradicts this.

## Why this is a bug

`man security` on the local macOS 26.4.1:

```
add-generic-password ... [-w password] ...
    -w password   Specify password to be added.  If you omit this argument you will be prompted.
```

`-w` is a positional-argument flag. Omitting the argument makes `security` prompt on `/dev/tty`, not read from stdin. The current code passes `-w` with no positional and routes the password through `cmd.Stdin` (`keychain_darwin.go:78` → `osExecutor.Run` line 110), but security never reads that stdin for the password. In an interactive shell with a live controlling terminal, `/dev/tty` is the actual user terminal AND stdin happens to coincide with `/dev/tty`, so the user can type the password and it works. In a non-interactive shell, no `/dev/tty` input arrives, `security` records an empty password, and exit code is 0 — masking the failure.

The bug surface is exactly the "non-interactive automation" use case that motivated v4.10 (per spec 001 §Problem).

## Goal

`teamvault-login` writes the verified password to the Keychain reliably whether invoked interactively or non-interactively. A non-interactive invocation that succeeds in verification MUST result in a Keychain entry that contains the verified password, byte-for-byte.

## Constraints

- Existing successful interactive flow must not regress: `teamvault-login --teamvault-config ~/.teamvault.json` typed at a terminal must continue to work as today.
- Implementation MUST NOT pass the password on the command line as a literal argument visible in `ps`/`/proc` — that's worse than the current bug.
- macOS Keychain ACL semantics (which apps may read without prompting) must be preserved — the `security add-generic-password` invocation should keep `-U` semantics (update if exists, create if absent).
- Public Go API of `Keychain` interface (`ReadPassword`, `WritePassword`) MUST NOT change.
- The `osExecutor` and `darwinKeychain` types may grow new method signatures internally but stay package-private.
- All changes confined to `keychain_darwin.go`, `keychain_executor.go` (if needed), and their tests.

## Failure Modes

| Trigger | Expected behavior | Detection |
|---|---|---|
| Password contains shell metacharacters (`$`, backticks, spaces) | Stored verbatim; subsequent read returns identical bytes | Test with passwords containing all printable ASCII |
| Password contains newline | Either stored verbatim or rejected with a clear error before write — never silently truncated | Test with `printf 'foo\nbar'` style passwords |
| Password contains NUL byte | Rejected before invoking `security` (security has limits) | Unit test |
| Keychain locked (user logged out) | Returns the existing "Keychain locked" wrapped error, unchanged | Match current `ReadPassword` locked-handling style |
| `security` binary missing / non-darwin invocation | Existing build-tag separation handles this; no change | n/a |
| Concurrent `WritePassword` for the same URL | `security -U` semantics: last writer wins (acceptable, not a new failure mode) | Document, don't try to lock |

## Acceptance Criteria

- [ ] `teamvault-login` invoked with stdin piped (`printf '%s\n' "$PASS" | teamvault-login --teamvault-config <cfg>`) stores the password byte-for-byte in the Keychain.
  - Evidence: after invocation, `security find-generic-password -s teamvault-utils -a <url> -w | head -c -1` equals the piped password value.
- [ ] `teamvault-login` invoked interactively (real tty, user types password) continues to store the typed password — no regression.
  - Evidence: manual run on a developer macOS; subsequent `teamvault-password --teamvault-key <known-key>` returns a non-empty real password.
- [ ] Failure to write the Keychain (any `security` non-zero exit) propagates as a wrapped error and the program exits non-zero — the "Login successful" message MUST NOT print on a failed write.
  - Evidence: induce failure (e.g., locked Keychain), assert stderr contains the wrapped error and exit code is non-zero. Existing test pattern in `keychain_darwin_test.go` covers the exit-code branches; extend.
- [ ] Existing unit tests in `keychain_darwin_test.go` continue to pass unchanged (`go test ./... -run KeychainDarwin` exit 0).
- [ ] New unit tests cover the non-interactive stdin path using the `Executor` fake (`mocks/executor.go`). Assert that the executor receives the password in a way that `security` will actually consume (positional `-w PASSWORD` argument OR another documented mechanism); assert empty stdin is NOT relied on as the password channel.
- [ ] Existing integration test `keychain_darwin_integration_test.go` continues to pass unchanged.
- [ ] New integration test covers a real `security` invocation with a piped password and verifies round-trip via a temporary service name (e.g., `teamvault-utils-test-<uuid>`) to avoid clobbering real credentials. Test must run unattended — `t.Skip` on non-darwin AND when `security` would require a Keychain-unlock prompt (detect via probe call before the test body).
- [ ] Scenario `scenarios/002-keychain-login-and-retrieve.md` passes when re-walked end-to-end against the fixed binary (currently fails on the "Keychain entry idempotent" + "password non-empty" assertions).
- [ ] `make precommit` exits 0.
- [ ] CHANGELOG `## Unreleased` entry documents the bug fix with cross-reference to this spec.

## Verification

- Run the Reproduction section verbatim against the fixed binary; expect Keychain length > 0 and subsequent `teamvault-password` to exit 0.
- Re-walk `scenarios/002-keychain-login-and-retrieve.md` — every checkbox green.
- `go test ./...` — all packages pass including the new tests.
- `make precommit` — exit 0.

## Workaround

Until the fix lands, users can:

1. **Interactive only.** Run `teamvault-login` only from a real terminal, never with stdin piped. Type the password at the prompt.
2. **Direct keychain write** (bypasses `teamvault-login` entirely):
   ```
   security add-generic-password -U -s teamvault-utils -a <url> -w <password>
   ```
   This stores the password correctly because `-w PASSWORD` uses the positional form. Risk: password appears in shell history and `ps` output; mitigate with `read -s PASS` then pass via `-w "$PASS"`.

## Alternatives (decide in prompt phase)

Three viable approaches that satisfy the Constraints:

1. **`security add-generic-password -w "$PASS"` positional** — simplest fix. Constraint #2 forbids passing the password as a literal command-line argument *visible in `ps`*. On macOS, `argv` IS visible via `ps -E` and `/proc`-equivalents, so this approach is rejected on its face by Constraint #2.

2. **`-i` interactive mode + stdin script** — `security -i` accepts subcommands via stdin. Test whether `add-generic-password ... -w` inside an `-i` session reads the password from stdin per-subcommand. If so, this avoids `argv` exposure. Spike during prompt phase to confirm.

3. **macOS Keychain Services API via cgo** — direct `SecKeychainAddGenericPassword` call. Most robust (no shell quoting, no `security` quirks, no argv exposure) but adds cgo dependency and a darwin-specific compilation path. Existing `keychain_other.go` build-tag pattern already accommodates this; cgo only kicks in on darwin builds.

Approach #2 if it works is the smallest diff. Approach #3 if it doesn't. The prompt phase MUST evaluate #2 first via a 10-line probe before committing to the cgo route.

## Notes

- Discovered during the `scenarios/002-keychain-login-and-retrieve.md` walk on 2026-05-21 against teamvault-utils HEAD = v4.12.0.
- Reporter's keychain entry for `https://teamvault.benjamin-borbe.de` was destroyed by the scenario walk and was restored manually via Workaround #2 (with quoting around the password to handle the comma it contains) before normal `teamvault-*` use resumed.

## Verification Result

**Verified:** 2026-05-21T21:28:59Z (HEAD dc33341)
**Verdict:** PASS (subsumed by spec 004 zalando migration)

The original bug (piped-stdin `teamvault-login` writes empty password to macOS Keychain) is no longer reproducible. The v0.9.10 `security -i` REPL fix (commits 91f94f5 / 5707dd8) addressed the symptom; spec 004 (commit dc33341) replaces the entire hand-rolled `security` shell-out with `github.com/zalando/go-keyring`, eliminating the failure surface that produced this bug class.

**Reproduction re-run against post-migration binary:**
- Pre-state: keychain has real 32-char password (restored via `security add-generic-password -U`).
- Action: `printf '%s\n' "$REAL_PASS" | /tmp/tv-login --teamvault-config <cfg-without-pass>`.
- Result: `LOGIN_RC=0`; "Login successful. Password stored in macOS Keychain for https://teamvault.benjamin-borbe.de."; subsequent `teamvault-username` returns `longhorn`.
- The keychain is NOT zeroed; downstream binaries authenticate correctly via the post-migration zalando read path.

See spec 004's `## Verification Result` for the full AC matrix.
