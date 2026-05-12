---
status: prompted
approved: "2026-05-12T15:11:44Z"
generating: "2026-05-12T17:52:51Z"
branch: dark-factory/macos-keychain-credential-store
---

# macOS Keychain password fallback for teamvault-utils

## Summary

- Add the macOS Keychain as a new password source in `teamvault-utils`, used only when `--teamvault-pass`, `TEAMVAULT_PASS`, and the config-file `password` field are all empty.
- Ship a new `teamvault-login` binary that verifies credentials against the TeamVault API and stores the password in the Keychain on success.
- macOS only in v1 — Linux/Windows behavior byte-identical to today. Username and URL stay in flag/env/config.
- Unblocks public sharing of teamvault-utils: new macOS users never need to write a TeamVault password into a dotfile.

## Problem

Today, `teamvault-utils` resolves the TeamVault password from one of: a `--teamvault-pass` flag, `TEAMVAULT_PASS` env var, or the `password` field in the config file at `--teamvault-config` / `TEAMVAULT_CONFIG`. For interactive desktop users, the only ergonomic option is the config file — which means the password sits in plain text on disk (and in dotfile-sync, backups, screen-shares, accidental `cat`s).

This is the single hardest adoption barrier. The sharing initiative ([[AI Knowledge Sharing at Seibert]] in the Brogrammers vault) cannot promote teamvault-utils to coworkers until the password lives somewhere safer than `~/.teamvault.json`.

## Goal

On macOS, `teamvault-utils` reads the TeamVault password from the user's login Keychain when no other source provides it. A new `teamvault-login` command verifies the password works against the TeamVault API and stores it in the Keychain on success. Existing flag / env / config-file flows are unchanged. Multi-vault setups (e.g. personal + work) are disambiguated by URL as the Keychain account key.

## Non-goals

- Linux Secret Service integration — file a fresh spec when an actual Linux desktop user surfaces.
- Windows Credential Manager — file a fresh spec when a Windows user surfaces.
- Storing username or URL in the Keychain — password-only in v1; revisit if writing the config file becomes friction.
- `teamvault-logout` / delete-credential command — workaround is `security delete-generic-password -s teamvault-utils -a <url>`.
- Debug command reporting which source supplied the password — add only if multi-vault resolution surprises someone in practice.
- Auto-migration of existing config-file users — opt-in only. Users keep the file until they run `teamvault-login`.
- Encrypting the config file as an alternative — rejected; Keychain is the chosen path.
- iCloud-synced Keychain items, ACLs, certificates — outside MVP.
- Rotation, team-shared Keychain entries — v2.

## Desired Behavior

### Password resolution (all `teamvault-*` binaries)

1. `--teamvault-pass` flag — *exists*
2. `TEAMVAULT_PASS` env var — *exists*
3. `password` field in config file at `--teamvault-config` / `TEAMVAULT_CONFIG` — *exists*
4. **macOS Keychain** — service `teamvault-utils`, account = TeamVault URL resolved from flag / env / config — *new*
5. Error: `TeamVault password not found; run teamvault-login to set it`

Each step runs only if the previous one returned nothing. File still wins over Keychain.

### `teamvault-login` binary

New binary at `cmd/teamvault-login/main.go`, mirroring the existing `cmd/teamvault-password/` layout. Flow:

1. Resolve URL, user, password via the same chain above (flag/env/config/Keychain).
2. Call the TeamVault API to verify the resolved credentials.
3. If verification succeeds and the password did **not** come from the Keychain → store it in the Keychain (idempotent: overwrites any existing entry for the same URL).
4. If verification fails (wrong password) or password is missing → prompt the user on stdin (hidden input) for a password, call the API again, repeat up to 3 attempts. On success, store in Keychain. On exhaustion or interrupt, exit non-zero with a clear error.
5. On non-macOS platforms: perform steps 1–2 only; skip Keychain storage and print a notice that Keychain storage is macOS-only in v1.

### Multi-vault

- `--teamvault-config ~/.teamvault.json` and `--teamvault-config ~/.teamvault-sm.json` resolve to different URLs and therefore different Keychain entries.
- No new flag introduced. The existing `--teamvault-config` selector covers multi-vault.

## Constraints

- Public Go API of `Config`, `Connector`, `TeamvaultConfigPath.Parse()` is unchanged. The Keychain source is wired in below the existing resolution code.
- No new mandatory dependencies for Linux users. Keychain code is darwin-only (build tag); Linux is a no-op stub returning "not found" so the chain proceeds to the error.
- macOS support works on Apple Silicon and Intel from a single binary.
- No daemons, no helper agents. Single binary per command.
- No telemetry, no network calls beyond TeamVault itself.
- Password never logged or printed (audit `glog` / log call sites).
- `teamvault-login` reads the password from stdin, never argv, when prompting.

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---|---|---|
| Keychain locked (user logged out, requires auth) | Return clear error: "TeamVault password requires Keychain unlock; unlock your Keychain and retry" | User unlocks login Keychain |
| Keychain entry missing for the resolved URL | Step 4 returns "not found"; chain proceeds to step 5 error pointing at `teamvault-login` | Run `teamvault-login` |
| TeamVault URL unset when reaching step 4 | Step 4 returns "not found" (account key would be empty); chain proceeds to error | User supplies URL via flag / env / config |
| `teamvault-login` API verification fails after 3 prompts | Exit non-zero with clear error; Keychain unchanged | User checks credentials, retries |
| TeamVault API unreachable from `teamvault-login` | Exit non-zero with network error; Keychain unchanged | User checks network, retries |

## Do-Nothing Option

Keep the current flag/env/config-file resolution. Result: the AI Knowledge Sharing initiative cannot promote teamvault-utils to coworkers, because every install attempt stalls at "write your TeamVault password into a JSON dotfile." Every user who installs anyway carries an ongoing low-grade leak risk (backups, dotfile sync, screen-shares, accidental `cat`). Cost is rising one frustrated install at a time, not catastrophic — so this is "do soon," not "do this week."

## Acceptance Criteria

- [ ] Password resolution adds a macOS Keychain step between config-file and error, keyed by service `teamvault-utils` + account `<TeamVault URL>`.
- [ ] On non-macOS platforms, the Keychain step is a no-op stub (build tag); existing binaries behave byte-identically to today.
- [ ] New binary `cmd/teamvault-login/main.go` verifies credentials against the TeamVault API and stores the password in the macOS Keychain on success.
- [ ] `teamvault-login` prompts for the password on stdin (hidden) when current credentials don't verify; stores after successful re-verification.
- [ ] Two distinct `--teamvault-config` paths with different URLs produce two independent Keychain entries.
- [ ] Public Go API of `Config`, `Connector`, `TeamvaultConfigPath.Parse()` is unchanged.
- [ ] Unit tests (Ginkgo/Gomega + Counterfeiter mocks, per project DoD) cover: Keychain hit, Keychain miss, file wins over Keychain, missing URL, locked Keychain (error path), API verify success, API verify failure, prompt-then-store flow.
- [ ] README documents the new flow as the recommended setup on macOS and flags the config-file password as insecure.
- [ ] CHANGELOG entry added.
- [ ] `make precommit` passes.

## Verification

- `make precommit` in `~/Documents/workspaces/teamvault/teamvault-utils/`.
- Manual on macOS: with no password in flag / env / config, run any `teamvault-*` binary → expect error pointing at `teamvault-login`. Run `teamvault-login`, type password, command succeeds → re-run the original binary → it works.
- Manual on macOS: edit `~/.teamvault.json` to remove the `password` field → existing binaries still work via Keychain.
- Manual on macOS: with two config files for two vaults, run `teamvault-login` for each → `security find-generic-password -s teamvault-utils -a <url-1>` and `<url-2>` both return entries.
- Manual on macOS: with `password` in the config file, file wins (Keychain not consulted).
- Manual on Linux: behavior byte-identical to today's release.

## Notes

- Linked from [[AI Knowledge Sharing at Seibert]] (Brogrammers vault Supply Matrix entry for teamvault-utils — currently flagged as blocked on this work).
- Library choice for the Keychain layer (cgo + Security.framework vs shell-out to `security(1)` vs `zalando/go-keyring`) is an implementation detail to decide in the prompt phase, not in this spec.
