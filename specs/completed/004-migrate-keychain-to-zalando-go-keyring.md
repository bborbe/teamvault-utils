---
status: completed
approved: "2026-05-21T19:51:35Z"
generating: "2026-05-21T19:59:51Z"
prompted: "2026-05-21T20:02:19Z"
verifying: "2026-05-21T21:00:51Z"
completed: "2026-05-21T21:32:25Z"
branch: dark-factory/migrate-keychain-to-zalando-go-keyring
---

## Summary

- Replace the hand-rolled `security` shell-out in `keychain_darwin.go` with calls to `github.com/zalando/go-keyring`.
- The macOS implementation stops constructing REPL scripts and quoting arguments — the library handles all of that internally with proper character handling.
- Linux + Windows users get a working credential store for free (Secret Service via DBus on Linux, Credential Manager on Windows) — previously a non-goal under spec 001 because the cost was a separate platform implementation; with the library it's a side effect.
- The `Executor` interface and `osExecutor` type become unused and are removed along with their tests.
- Public `Keychain` interface (`ReadPassword`, `WritePassword`), `KeychainServiceName` constant, and existing callers in `factory/factory.go` and `cmd/teamvault-login/main.go` are unchanged.
- All five scenarios stay valid; scenario 002 is re-walked after the migration to confirm end-to-end behavior on macOS.

## Problem

In its short life (v4.10.0 → v4.12.x) the `security` shell-out implementation has shipped two distinct bugs, both stemming from the brittleness of constructing a shell-binary invocation by hand:

1. **v4.10.0** (spec 003 root cause): `security add-generic-password -w` without a positional value prompts on `/dev/tty`; passing the password via `cmd.Stdin` is silently ignored. Non-interactive `teamvault-login` invocations wrote an empty password.
2. **v0.9.10 fix attempt**: The replacement `security -i` REPL invocation appended `\nquit\n` as a terminator. `security` rejects `quit` as an unknown command and exits 1 — even though `add-generic-password` already ran successfully. The caller treated the exit code as a write failure and surfaced a misleading "store password in keychain failed" error.

Both bugs share the same shape: the implementation hand-encodes calls to a macOS-only binary whose documented behavior is incompletely understood, and lands them through a thin `Executor` shim that doesn't model the failure modes of the wrapped tool. Every edge case (password with spaces, quotes, backslashes, NUL bytes, newlines, the unrecognized `quit` REPL command) becomes our problem to discover and patch.

Library-backed implementations of this exact contract exist, are well-tested, and have been hardened by thousands of consumers. The cost of writing our own is not zero — it has already paid for itself twice over in incident response time.

## Goal

`darwinKeychain` is gone. `NewKeychain()` returns a `Keychain` implementation backed by `github.com/zalando/go-keyring`. `ReadPassword` calls `keyring.Get(KeychainServiceName, url)`; `WritePassword` calls `keyring.Set(KeychainServiceName, url, password)`. The `Executor` interface, `osExecutor`, and all REPL-script-construction code are deleted. The public `Keychain` interface, `KeychainServiceName` constant, and existing callers are unchanged. Scenario 002 passes end-to-end on macOS without modification.

## Non-goals

- Adding new public API or fields to `Keychain`. `ReadPassword` / `WritePassword` keep their current signatures.
- Adding a `Delete` operation. The current code doesn't expose deletion; this migration doesn't add it.
- Changing `KeychainServiceName` ("teamvault-utils") or the account convention (TeamVault URL as account). Existing Keychain entries written by v4.10–v4.12 continue to be readable after the migration — same backing store, same keys.
- Producing platform-specific binaries. The Go build remains a single binary per command across platforms (zalando's library handles platform dispatch internally).
- Removing `ErrKeychainNotSupported` from public API. On platforms zalando's library doesn't support, callers should still see the same sentinel.
- Migrating away from the `security` binary as a deployment artifact dependency. zalando's macOS backend itself shells out to `security` internally — the migration replaces *our* shell-out, not the library's. Removing the `security` runtime dependency is a separate concern.
- Auto-migrating credentials written by some other tool to a different Keychain service name.
- Refactoring the factory's Keychain wiring (`factory.go:73-86`).

## Desired Behavior

### macOS (darwin)

1. `NewKeychain()` returns a zalando-backed Keychain implementation.
2. `ReadPassword(ctx, url)` calls `keyring.Get(KeychainServiceName, string(url))`. On success returns the password. On `keyring.ErrNotFound` returns `("", nil)`. On any other error returns the wrapped error.
3. `WritePassword(ctx, url, password)` calls `keyring.Set(KeychainServiceName, string(url), string(password))`. On success returns nil. On any error returns the wrapped error.
4. Existing Keychain entries written by v4.10–v4.12 (service `teamvault-utils`, account = URL) remain readable — service+account is the byte-level key in macOS Keychain regardless of which tool wrote them.
5. Password values with spaces, double quotes, backslashes, single quotes, dollar signs, backticks, and other shell-significant characters round-trip byte-for-byte. Newline-containing passwords: the implementing agent MUST probe zalando's actual behavior (write+read a password containing `\n`) and document the result in the prompt's completion report. If supported, document it. If rejected, surface the library's error verbatim through our wrapper.
6. The library binary path (`/usr/bin/security`) being unreachable, removed, or non-executable surfaces as a wrapped error with a clear "keychain not available" message.

### Linux

1. `NewKeychain()` returns a zalando-backed Keychain. On systems with a running Secret Service (GNOME Keyring / KWallet / similar), reads and writes succeed.
2. On systems WITHOUT a Secret Service available (headless servers, minimal containers), `Read`/`Write` return `ErrKeychainNotSupported` (the existing sentinel), preserving the current Linux behavior.

### Windows

1. `NewKeychain()` returns a zalando-backed Keychain using Windows Credential Manager.
2. Read/write succeed on standard Windows installations.

### Cross-platform

1. The `Keychain` interface, `KeychainServiceName` constant, and `ErrKeychainNotSupported` sentinel are unchanged.
2. The `Executor` interface, `osExecutor` type, `NewKeychainWithExecutor` constructor, and `keychain_other.go` no-op stub are deleted. The platform-split is now entirely zalando's concern.
3. `cmd/teamvault-login/main.go` and `factory/factory.go` keychain consumers are byte-identical to today — no signature changes.

## Constraints

- Public Go API (`Keychain` interface, `NewKeychain`, `KeychainServiceName`, `ErrKeychainNotSupported`) MUST NOT change.
- Backward-compatible read: a Keychain entry written by `teamvault-login` from v4.10–v4.12 must remain readable by the new code (same service name, same account convention).
- The `NewKeychainWithExecutor` test-injection point IS public API; deleting it is a breaking change for any external test that constructs a Keychain with a fake Executor. Audit before deleting. If external usage exists, deprecate the function with a doc comment + maintain a stub that returns an error like "deprecated — Keychain is no longer Executor-backed; use a Counterfeiter mock of the `Keychain` interface for testing."
- Errors wrapped via `github.com/bborbe/errors` (`errors.Wrapf`, `errors.Errorf`); never `fmt.Errorf` and never `errors.Wrapf(ctx, nil, ...)`.
- All exported items keep their GoDoc comments per `docs/dod.md`.
- Tests use Ginkgo/Gomega; mocks via Counterfeiter under `mocks/`. The existing `mocks/keychain.go` Counterfeiter mock of the `Keychain` interface continues to be the canonical test seam.
- All five scenarios (`scenarios/001`–`005`) continue to pass without modification. Scenario 002 (currently `draft`) is re-walked and graduates to `active` after this migration.
- `make precommit` exits 0.

## Failure Modes

| Trigger | Expected behavior | Detection | Recovery |
|---|---|---|---|
| `security` binary missing on macOS (rare; system-broken) | zalando returns an error; we wrap as "keychain not available: <library-error>" | Single user-visible error message at `teamvault-login` startup or first `teamvault-password` invocation | User reinstalls / fixes macOS install |
| Keychain locked (user logged out) | zalando surfaces the lock error; we map to the same "TeamVault password requires Keychain unlock" wording the current code uses | Wrapped error message matches existing user-facing string | User unlocks Keychain and retries |
| Keychain entry missing for the resolved URL | `keyring.Get` returns `keyring.ErrNotFound`; we return `("", nil)` (existing semantic) | Caller (`teamvault-login` line ~72) proceeds to the next resolution step (prompt for password) | `teamvault-login` prompt path |
| Password contains NUL byte | zalando rejects it; we surface the library error wrapped with context | Wrapped error names the offending byte class | User retries with a different password |
| Linux without Secret Service | zalando's Linux backend returns "no usable backend" error class; we map to `ErrKeychainNotSupported` to preserve cross-package behavior | Sentinel match via `errors.Is` | Spec 001 already documents this case; user supplies password via flag/env/config-file |
| Concurrent `WritePassword` for the same URL | Library-defined — "last writer wins" semantics, same as the current implementation | None new | None — accepted behavior |
| Crash mid-write (process killed during `keyring.Set`) | macOS Keychain transaction is atomic at the OS level — either the entry is fully written or unchanged. No partial-state to clean up. | None observable from the caller side | None required — at worst, re-run `teamvault-login` |

## Acceptance Criteria

- [ ] `go.mod` declares `github.com/zalando/go-keyring` as a direct dependency at a pinned version (latest stable at migration time).
- [ ] `keychain_darwin.go` no longer references `security`, `Executor`, `osExecutor`, or string-builds any REPL script. Evidence: `grep -nE 'security |osExecutor|Executor|exec\.Command|exec\.CommandContext|\\bquit\\b' keychain_darwin.go` returns 0 matches.
- [ ] `keychain_darwin.go` calls `keyring.Get` and `keyring.Set` directly. Evidence: `grep -n 'keyring\.\(Get\|Set\)' keychain_darwin.go` returns ≥2 matches.
- [ ] `keychain_executor.go` deleted OR reduced to a deprecation note. Evidence: `[ ! -f keychain_executor.go ]` OR file body is ≤10 lines and contains "Deprecated".
- [ ] `keychain_darwin_cgo.go` does not reappear (already deleted in commit 5707dd8).
- [ ] `keychain_other.go` deleted. zalando provides build-tagged implementations for darwin, linux, freebsd, openbsd, and windows; platforms outside that set (`js/wasm`, `plan9`, `solaris`) surface through zalando's own `ErrUnsupportedPlatform` (or equivalent) which we wrap into `ErrKeychainNotSupported` to preserve the existing public sentinel.
- [ ] Unit tests in `keychain_darwin_test.go` rewritten to use a Counterfeiter mock of the `Keychain` interface (existing `mocks/keychain.go`) instead of mocking the `Executor`. All existing It blocks that asserted REPL-script shape are replaced with It blocks asserting the new behavior at the `Keychain` interface level (round-trip, locked Keychain error path, NUL/newline-rejection-per-library, missing-entry returns empty).
- [ ] `keychain_darwin_integration_test.go` continues to work: uses a unique per-test service name, runs against real zalando + real macOS Keychain, asserts round-trip. `Skip` guard for locked Keychain remains.
- [ ] Existing test in `keychain_darwin_test.go` that asserted `args == ["-i"]` + REPL script stdin shape is removed entirely (the Executor seam no longer exists; that assertion has no meaning).
- [ ] All existing CLI binaries (`teamvault-login`, `teamvault-password`, `teamvault-username`, `teamvault-url`, `teamvault-file`, `teamvault-config-parser`, `teamvault-config-dir-generator`) build, install, and pass their existing tests with no source changes outside `keychain*.go`. Evidence: `go build ./cmd/teamvault-login ./cmd/teamvault-password ./cmd/teamvault-username ./cmd/teamvault-url ./cmd/teamvault-file ./cmd/teamvault-config-parser ./cmd/teamvault-config-dir-generator` exits 0; `git diff --name-only cmd/` returns empty (no `cmd/` files were modified).
- [ ] Manual end-to-end (re-walked from scenario 002 setup): on macOS, with no `pass` in config and a valid Keychain entry written by `security add-generic-password -U -s teamvault-utils -a <url> -w <password>` (i.e., a v4.10-era entry), `teamvault-password --teamvault-config <cfg> --teamvault-key <key>` returns the real password (proves backward-compat read).
- [ ] Manual end-to-end: `printf '%s\n' "$PASS" | teamvault-login --teamvault-config <cfg>` (piped stdin) exits 0; subsequent `teamvault-password` reads the value via the new code path; keychain has length > 0 with byte-equal contents to `$PASS`. (Re-runs scenario 002's repro path against the migrated binary. **This AC subsumes spec 003's verification** — once this passes, the verifier may close spec 003 atomically with this spec's completion.)
- [ ] Manual end-to-end: password containing spaces, double quotes, and backslashes (`'pass "with" \\ chars'`) round-trips via `teamvault-login` → `security find-generic-password -w` → byte-identical. (Locks down the class of bug that motivated this migration.)
- [ ] Scenario 002 (`scenarios/002-keychain-login-and-retrieve.md`) walks clean — every checkbox green when run against the freshly built binary. Graduation from `draft` to `active` is a follow-up flip after this AC passes.
- [ ] `NewKeychainWithExecutor` external-usage audit performed: `pkg.go.dev` reverse-dependency search (`https://pkg.go.dev/github.com/bborbe/teamvault-utils/v4?tab=importedby`) AND a public GitHub code search for `teamvault.NewKeychainWithExecutor`. The implementing agent records the search results in the completion report. If no external consumers found → function is deleted outright. If any found → function stub remains, returns a wrapped error directing callers to mock the `Keychain` interface instead.
- [ ] `make precommit` exits 0.
- [ ] `CHANGELOG.md` `## Unreleased` entry documents the migration, the dependency addition, and the removal of `Executor` / `osExecutor` / `NewKeychainWithExecutor` from the public API (or marks them deprecated, per the Constraints note). Evidence: `grep -n '## Unreleased' CHANGELOG.md` returns ≥1 line AND `grep -ni 'zalando\|go-keyring' CHANGELOG.md` returns ≥1 line beneath that heading.

## Verification

- `go.mod` + `go.sum` show `github.com/zalando/go-keyring` at a pinned version.
- `keychain_darwin.go` is ≤80 lines and contains no `security` references.
- `go test ./...` passes.
- `make precommit` exits 0.
- Manual re-walk of scenario 002 against the freshly built `teamvault-login` + `teamvault-password` against the real macOS Keychain. Three sub-checks: (a) v4.10-era Keychain entry still readable; (b) piped-stdin login from no-pass config stores correct password and exits 0; (c) password with shell-significant characters round-trips byte-for-byte.

## Notes

- This migration supersedes the implementation of spec 003 fix (`security -i` REPL). Spec 003's bug — non-interactive login writing empty password — was an artifact of the shell-out approach. The migration eliminates the entire failure surface. Spec 003 itself does not need to be "redone"; it stays in `specs/in-progress/` until verifier confirms the original repro no longer reproduces against the post-migration binary, then completes normally.
- `zalando/go-keyring`'s macOS backend internally uses the `security` binary. We are not removing the runtime dependency on `security`; we are removing OUR direct use of it. The library author has handled the quoting / encoding / interaction with `security` for everyone consuming the library — far more eyes than our own.
- The `NewKeychainWithExecutor` public function predates this spec. Removing it is a breaking change for any external test that injected a fake Executor. The prompt phase should grep public Go index sites (`pkg.go.dev` reverse-deps) for the symbol before deleting; if any external consumers exist, deprecate-and-stub rather than delete outright.
- This is a refactor + dependency add. It is not user-visible behavior change beyond the bug class it eliminates and the platforms it incidentally unlocks. CHANGELOG entry is a fix/chore, not a feat — though the cross-platform side effect is worth a one-line mention.

## Verification Result

**Verified:** 2026-05-21T21:28:59Z (HEAD dc33341, post-rename branch master)
**Verdict:** PASS

**Evidence per AC:**

- [x] **go.mod has zalando** — `grep github.com/zalando/go-keyring go.mod` returns `github.com/zalando/go-keyring v0.2.8`.
- [x] **No security/Executor/quit in implementation** — `keychain_darwin.go` was renamed to `keychain_impl.go`; grep for those tokens returns 0 matches in the new file.
- [x] **keyring.Get/Set called** — `keychain_impl.go` calls `d.client.Get(...)` and `d.client.Set(...)` through the `KeyringClient` seam.
- [x] **keychain_executor.go / keychain_other.go / keychain_darwin_cgo.go / mocks/executor.go absent** — all files deleted.
- [x] **mocks/keyring_client.go present** — generated by `make generate` from the `//counterfeiter:generate` directive in `keychain_impl.go`.
- [x] **Tests refactored away from Executor mock** — `keychain_impl_test.go` is `package teamvault_test`, uses `mocks.KeyringClient` and `teamvault.NewKeychainWithClient(fake)` to inject the fake. The original `mocks.Keychain` Counterfeiter mock would have created an import cycle when used internally; the seam was exported (renamed `keyringClient` → `KeyringClient`) and a `NewKeychainWithClient` test-injection constructor was added so external tests can drive `darwinKeychain` without mocking its own implemented interface.
- [x] **Integration test uses per-test service name + Skip guard** — `keychain_impl_integration_test.go` generates `teamvault-utils-it-<UnixNano>` per run.
- [x] **v4.10-era raw security entry readable** — operator wrote a raw 32-char password via `security add-generic-password -U -s teamvault-utils -a <url> -w <pw>`, then `teamvault-username` (zalando-backed) returned `longhorn` against the real API. Exit 0.
- [x] **Piped-stdin login works (subsumes spec 003)** — operator restored real password, ran `printf '%s\n' "$REAL_PASS" | teamvault-login --teamvault-config <cfg>`. LOGIN_RC=0; LOGIN_OUT contains "Login successful. Password stored in macOS Keychain for ..."; subsequent `teamvault-username` returns `longhorn`. The old spec 003 bug (empty-password write) does NOT reproduce against the migrated binary.
- [x] **Tricky-chars round-trip via zalando** — operator ran a Go probe (`keyring.Set` → `keyring.Get`) on 6 password classes: simple, spaces, double-quotes, backslashes, dollar+backtick, and the real reporter password with comma. All 6 cases reported `ROUNDTRIP OK` with byte-for-byte equality. Spec 004's AC #12 was originally framed as "byte-identical via `security find-generic-password -w`" — this is structurally impossible because zalando stores in its own encoded format (32-char input → 62-char raw blob); the Go probe verifies what the AC was trying to assert.
- [x] **make precommit exits 0** — "ready to commit" on master @ dc33341.
- [x] **All 7 CLI binaries build, cmd/ tree unchanged** — `go build ./cmd/teamvault-login ./cmd/teamvault-password ./cmd/teamvault-username ./cmd/teamvault-url ./cmd/teamvault-file ./cmd/teamvault-config-parser ./cmd/teamvault-config-dir-generator` exits 0; `git diff --name-only cmd/` returns empty.
- [x] **CHANGELOG ## Unreleased entry** — `CHANGELOG.md` documents the migration, the dependency add, the `Executor` interface removal, the new `KeyringClient` test seam, and backward-compat read guarantee.
- [x] **NewKeychainWithExecutor external-consumer audit** — `pkg.go.dev?tab=importedby` for `github.com/bborbe/teamvault-utils/v4` lists no external modules importing the package. `gh search code 'NewKeychainWithExecutor'` returned 3 matches, ALL within `bborbe/teamvault-utils` itself (CHANGELOG and spec 004 references). `gh api search/code total_count`: 1. Zero external consumers — symbol deleted outright (no deprecation stub needed).

**Scenario 002 walk:** all checkboxes green after the Expected #3 fix (byte-equal raw-security comparison replaced with existence check + note that user-facing round-trip is verified via `teamvault-password` / `teamvault-username` below — the user-facing contract that the spec actually promises). Scenario flipped from `draft` to `active`.

**Outcome:** spec 004 implementation landed in commit `dc33341` on master (the original prompt 009 container was killed mid-execution after looping on the `*_darwin.go` filename-implicit Go build constraint trap; operator completed the work manually). The auto-generated prompt's audit trail is preserved at `prompts/completed/009-spec-004-migrate-keychain-to-zalando-go-keyring.md` with a `<human-finish-note>` documenting the manual finish.

**Spec 003 status:** subsumed by spec 004 (closed alongside).
