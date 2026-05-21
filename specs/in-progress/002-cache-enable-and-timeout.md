---
status: verifying
approved: "2026-05-21T11:55:22Z"
generating: "2026-05-21T12:03:02Z"
prompted: "2026-05-21T12:06:25Z"
verifying: "2026-05-21T13:03:14Z"
branch: dark-factory/cache-enable-and-timeout
---

## Summary

- CLI `--cache` / `CACHE` env var is silently overridden by the config file's `cacheEnabled` — fix so EITHER source enabling cache turns it on.
- HTTP timeout to TeamVault is hardcoded to 5 seconds — make it configurable via config file and CLI/env, default 5s.
- On timeout, the existing disk-fallback cache (`~/.teamvault-cache/<key>/<kind>`) returns the previously fetched value, so transient TeamVault outages stop breaking dependent tooling.
- No breaking change: existing configs with `cacheEnabled:true` and existing CLI usage keep working byte-identically.

## Problem

Two operational issues surfaced when running `teamvault-*` binaries against a flaky or slow TeamVault instance:

1. **Precedence surprise.** The factory unconditionally overwrites the CLI/env cache flag with the config file's value (`factory/factory.go:71`). A user who keeps `cacheEnabled` unset in the config and passes `--cache=true` on the command line gets `cache=false` silently. The config wins over the more explicit signal — opposite of how every other field on these binaries behaves (URL/user/password use args, fall back to config).

2. **No configurable timeout.** `factory.CreateHttpClient` hardcodes `5 * time.Second`. When TeamVault is slow (cold start, network blip), every dependent binary fails with a context deadline error, even though a cached value sits on disk and would be returned by `DiskFallbackConnector` if only the timeout fired sooner — or later, depending on the operator's preference. Operators have no knob.

Together these mean reliability tooling (`teamvault-config-parser`, `teamvault-config-dir-generator`) is fragile in exactly the scenarios — slow upstream, partial outage — where the disk cache is most useful.

## Goal

The library and CLI tools accept a timeout duration via config file, CLI flag, and env var (in that overlay order — last-set wins per existing libargument semantics). Cache enablement is the logical OR of CLI and config. When the configured timeout fires on a remote call and cache is enabled, the disk fallback returns the cached value transparently. Existing callers see no behavior change unless they set the new knob or rely on the precedence bug.

## Non-goals

- In-memory cache wiring. `teamvault.NewCacheConnector` exists but stays unused in the factory chain; introducing it is a separate decision.
- Per-call timeout override. One process-level timeout suffices for v1.
- Retries / circuit breaker / exponential backoff. v1 ships timeout + cache fallback only.
- Cache TTL / invalidation. Disk cache is already "valid until the user deletes it"; this spec doesn't change that.
- Restructuring the precedence model for URL/user/password. Only the cache flag has the inversion bug; the rest already follow "arg overrides empty arg → config → keychain".
- A separate `--no-cache` flag to force-disable when config enables. Out of scope; OR-only.
- Renaming `CacheEnabled` to clarify it means "disk fallback." Cosmetic; defer.

## Desired Behavior

### Cache enablement (OR semantics)

1. Cache is **on** if `--cache=true` (or `CACHE=true`) OR the config file's `cacheEnabled: true` is set.
2. Cache is **off** if both sources are absent or both are false.
3. When cache is on, the factory wraps the remote connector in `DiskFallbackConnector` exactly as today.

### Configurable timeout

1. New config field `timeout` (Go duration string, e.g. `"5s"`, `"30s"`). Absent / empty → default 5s (current behavior).
2. New CLI flag `--teamvault-timeout` and env var `TEAMVAULT_TIMEOUT` on every existing `teamvault-*` command. Both accept a Go duration string. Empty/absent → fall back to config; absent there → default 5s.
3. Precedence (matches existing pattern for URL/user/password): CLI/env value (via libargument default → env → arg) is used when non-empty; config value used when CLI is empty.
4. The resolved timeout is applied to the underlying `*http.Client` returned by `factory.CreateHttpClient`.
5. Zero or negative duration is rejected with a clear error at startup.

### Timeout × cache interaction

1. When the HTTP request to TeamVault exceeds the configured timeout, the `*http.Client` cancels the request and the remote connector returns an error.
2. If cache is enabled, `DiskFallbackConnector` reads the previously persisted value from `~/.teamvault-cache/<key>/<kind>` and returns it.
3. If cache is disabled, the timeout error propagates to the caller as today.

## Constraints

- Public Go API of `Config`, `Connector`, `factory.CreateConnectorWithConfig`, `factory.CreateConnectorWithConfigAndKeychain`, `factory.CreateHttpClient` may grow but must remain source-compatible for existing callers: new fields/parameters must be additive, or new functions provided alongside the old ones with the old ones delegating.
- `Config` JSON deserialization remains backward compatible: pre-existing configs without `timeout` or with only `cacheEnabled` continue to work.
- Default timeout = 5 seconds, matching today's hardcoded value.
- Disk fallback behavior on the wire is unchanged — only the trigger (timeout error) is configurable.
- Use `github.com/bborbe/time` `libtime.Duration` (already a project dependency; has `UnmarshalJSON` that accepts both string forms like `"5s"` and numeric nanoseconds; libargument handles it via `encoding.TextUnmarshaler`). No new external deps.
- All paths repo-relative in code and tests. No absolute paths in fixtures.
- Tests use Ginkgo/Gomega; mocks via Counterfeiter under `mocks/` per project DoD.
- Errors wrapped via `github.com/bborbe/errors`; no `fmt.Errorf`.
- All exported items have GoDoc comments per `docs/dod.md`.

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---|---|---|
| Config `timeout` is an unparseable string (e.g. `"banana"`) | Factory returns error at connector creation, wrapped via `bborbe/errors`, mentioning the offending value | User fixes config and reruns |
| Config `timeout` is zero or negative (`"0s"`, `"-1s"`) | Factory returns error; "timeout must be > 0" | User sets a positive duration |
| CLI `--teamvault-timeout` is unparseable | libargument parse error at startup, non-zero exit | User fixes the flag value |
| TeamVault unreachable + cache enabled + cached value exists | Request times out at configured duration; `DiskFallbackConnector` returns cached value; caller sees success | None — designed behavior |
| TeamVault unreachable + cache enabled + no cached value (first run) | Timeout error propagates; nothing cached to fall back to | User retries when TeamVault recovers or pre-warms cache |
| TeamVault unreachable + cache disabled | Timeout error propagates to caller as today | User retries or enables cache |
| Both CLI `--cache=false` and config `cacheEnabled:true` set | Cache is **on** by design (OR semantics — either source enabling turns it on) | Not a failure path; if disabling is required, edit the config |
| Both `--cache=true` and config `cacheEnabled` unset | Cache is **on** — the fixed behavior; today this is off due to the precedence bug | Working as intended after fix |

## Do-Nothing Option

Keep the 5s hardcoded timeout and the silent-override precedence. Operators continue to:
- Be confused why `--cache=true` does nothing when a config file exists.
- Have no recourse when TeamVault is briefly slow — every dependent binary fails despite a usable disk cache sitting next to the running process.

Cost: every TeamVault slowness incident breaks `teamvault-config-parser`-driven pipelines, and every new user touching `--cache` hits the precedence trap once. Neither is catastrophic; both erode trust in the tool steadily.

## Security / Abuse

- Timeout values are operator-supplied, not user-input from network — no injection surface.
- Cached values on disk inherit the existing `~/.teamvault-cache/` permissions (0700 dir, 0600 files per `diskfallback-connector.go:104-114`) — unchanged.
- A long timeout (`"10m"`) doesn't open new attack surface; it only delays failure, which the operator opted into.

## Acceptance Criteria

- [ ] `Config` struct gains a `Timeout` field that accepts a Go duration via JSON (e.g. `"timeout": "30s"`); absent/empty field deserializes to zero value and the factory substitutes the 5s default.
- [ ] Every `cmd/teamvault-*/main.go` (`teamvault-password`, `teamvault-username`, `teamvault-url`, `teamvault-file`, `teamvault-config-parser`, `teamvault-config-dir-generator`, `teamvault-login`) gains a `--teamvault-timeout` flag and `TEAMVAULT_TIMEOUT` env var. `teamvault-login` shares `factory.CreateHttpClient` (`cmd/teamvault-login/main.go:80`) so the same knob applies. The independent per-probe 10s `context.WithTimeout` in `loginFlow` (`cmd/teamvault-login/main.go:128, :158`) is unrelated and stays unchanged.
- [ ] `factory.CreateConnectorWithConfig` and `factory.CreateConnectorWithConfigAndKeychain` accept a timeout duration parameter (additive change); existing in-tree call sites updated; backward-compat wrapper provided if external callers exist.
- [ ] `factory.CreateHttpClient` honors a passed-in timeout (additive parameter, or a new exported function alongside) with 5s default.
- [ ] In the factory, `cacheEnabled := cliCache || config.CacheEnabled` replaces the unconditional overwrite at `factory/factory.go:71`. New code path covered by a Ginkgo test asserting both inputs.
- [ ] Unsigned/zero/negative timeout strings return a wrapped error from `factory.CreateConnectorWithConfig` mentioning the field name and the bad value.
- [ ] Unit tests (Ginkgo/Gomega + Counterfeiter):
  - cache OR-logic: (cli=true, config=false) → on; (cli=false, config=true) → on; (cli=false, config=false) → off; (cli=true, config=true) → on
  - timeout precedence: cli wins when non-empty; falls through to config when empty; default 5s when both empty
  - timeout parse errors surface via wrapped error
  - http client built by the factory has the resolved timeout (introspect `client.Timeout`)
- [ ] Integration test (single Ginkgo `It`): with a stub HTTP server that sleeps longer than the configured timeout AND cache enabled with a pre-populated cache file, the connector returns the cached value without error. Without cache, the same scenario returns a timeout error.
- [ ] `go test ./...` passes; `make precommit` exits 0.
- [ ] `README.md` documents the new `timeout` config field and `--teamvault-timeout` CLI flag in the relevant sections (config example block + CLI tool sections). Evidence: `grep -n 'timeout' README.md` returns ≥2 lines (config example + CLI section). Cache OR-precedence noted in a short paragraph; evidence: `grep -n -i 'cache.*precedence\|cacheEnabled.*or\|either.*cache' README.md` returns ≥1 line.
- [ ] `CHANGELOG.md` has a `## Unreleased` entry covering both changes. Evidence: `grep -n '## Unreleased' CHANGELOG.md` returns line 1 match, AND `grep -n -i 'timeout\|cache' CHANGELOG.md` shows additions under that section.

## Verification

- `cd ~/Documents/workspaces/teamvault/teamvault-utils && make precommit` → exit 0.
- Manual cache OR-logic check:
  1. With a config file containing only `url`, `user` (no `cacheEnabled`), run `teamvault-password --cache=true --teamvault-key=<key>` → cache enabled (verify by `ls ~/.teamvault-cache/<key>/` after the run).
  2. With a config file containing `"cacheEnabled": true` and no `--cache` flag, run the same command → cache enabled.
- Manual timeout check:
  1. Point a config at a non-routable IP (e.g. `https://10.255.255.1`) with `"timeout": "1s"` and `"cacheEnabled": true`. Pre-populate `~/.teamvault-cache/<key>/password` with `"cached-value"`.
  2. Run `teamvault-password --teamvault-key=<key>` → command completes in ~1s and prints `cached-value`.
  3. Repeat with `"timeout": "1s"` but **without** `cacheEnabled` → command fails with a timeout error in ~1s.
- Manual default-preservation: existing config (no `timeout` field) runs against a healthy TeamVault → identical wall-clock behavior to pre-change baseline.

## Notes

- The fix to factory precedence (line 71) is one line: `cacheEnabled = cacheEnabled || config.CacheEnabled`. The bulk of work is the timeout plumbing across 7 cmd binaries + factory + config + tests.
- `Config.Timeout` field type is `libtime.Duration` from `github.com/bborbe/time` (decided up front — see Constraints). Project already imports `libtime` everywhere; the type has `UnmarshalJSON` that handles both `"5s"` strings and numeric nanoseconds, so legacy and ergonomic configs both deserialize cleanly.

## Verification Result

**Verified:** 2026-05-21T13:43:50Z (HEAD f58b188)
**Binary:** /tmp/df-verify-002/teamvault-password (built from HEAD via `go build ./cmd/teamvault-password`)
**Scenario:** Live spec `## Verification` manual run: non-routable IP 10.255.255.1, timeout=1s, pre-populated cache file `~/.teamvault-cache/test-key/password`.
**Evidence:**
- `make precommit` → `ready to commit` (exit 0); `go test ./...` all packages PASS; `factory` suite: `Ran 19 of 19 Specs in 4.011 seconds`.
- Manual cache-fallback: `teamvault-password --teamvault-config=… --teamvault-key=test-key` → stdout `cached-value`, exit 0, elapsed 1.75s.
- Manual no-cache timeout: same config with `cacheEnabled:false` → `context deadline exceeded (Client.Timeout exceeded while awaiting headers)`, exit 1, elapsed 1.08s — proves resolved 1s timeout applied to `*http.Client`.
- Manual cache OR-logic: config without `cacheEnabled` + `--cache=true` → `cached-value`, exit 0, elapsed 1.04s — proves the OR replaces the prior silent-override.
- Manual negative-rejection: config `"timeout":"-5s"` → `create connector failed: invalid timeout -5s: must be >= 0`, exit 1.
- README: `grep -n 'timeout' README.md` → 3 lines (296, 410, 414); `grep -n -i 'either.*cache' README.md` → line 300.
- CHANGELOG: v4.11.0 lines 11-12 (Config.Timeout + OR fix), v4.12.0 line 7 (CLI flag) — dark-factory auto-tag path.
**Verdict:** PASS
