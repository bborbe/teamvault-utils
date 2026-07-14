---
status: active
---

# Scenario 008: create / search / update / read-back via the fake TeamVault server

Validates the `create`, `update <key>`, and `search <query>` subcommands end-to-end against `cmd/fakevault` — the same fake TeamVault HTTP server used by scenario 007, extended with an in-memory write + search surface. Exercises the real HTTP writer (`POST`/`PATCH /api/secrets/...`), the real HTTP connector reading the created secret back, and the search endpoint's `results[].api_url` → `Key` parsing that unit tests (which mock the writer/connector) do not.

Setup/assert helpers live in `scenarios/helper/lib.sh` (same convention as scenario 007). CI runs the whole thing via `make e2e`; the fastest local path is also `make e2e`.

Covered cases: `create --password-stdin` (primary create path), read-back of the created secret, `search` finds the new secret by name substring, `update` changes a field (password) and the read-back reflects it, `update` metadata-only change (username) does not disturb the password.

## Setup

```bash
source scenarios/helper/lib.sh
build_binaries      # builds teamvault-cli + fakevault to a temp dir, sets $TV
start_fakevault     # starts the server, writes a temp config (url+user+pass), exports TEAMVAULT_CONFIG
```

- [ ] `$TV` and `$WORK_DIR/config.json` exist; `fakevault` is listening (`$FV_URL` non-empty)

## Action + Expected

```bash
NEW_KEY="$(printf 'first-secret-pw' | "$TV" create --name write-search-e2e-secret --username wse-user --password-stdin)"
assert_eq "create returned a non-empty key" "nonempty" "$([ -n "$NEW_KEY" ] && echo nonempty || echo empty)"

assert_eq "read back created password" "first-secret-pw" "$("$TV" password --teamvault-key "$NEW_KEY")"
assert_eq "read back created username" "wse-user"         "$("$TV" username --teamvault-key "$NEW_KEY")"

assert_contains "search finds the new secret" "$NEW_KEY" "$("$TV" search write-search-e2e-secret)"

printf 'updated-secret-pw' | "$TV" update "$NEW_KEY" --password-stdin >/dev/null
assert_eq "read back updated password" "updated-secret-pw" "$("$TV" password --teamvault-key "$NEW_KEY")"

"$TV" update "$NEW_KEY" --username wse-user-renamed >/dev/null
assert_eq "read back updated username"        "wse-user-renamed"  "$("$TV" username --teamvault-key "$NEW_KEY")"
assert_eq "password unchanged by metadata-only update" "updated-secret-pw" "$("$TV" password --teamvault-key "$NEW_KEY")"

scenario_done   # prints "e2e: PASS" and exits non-zero if any assertion failed
```

- [ ] All assertions print `ok:` and `scenario_done` reports `e2e: PASS`

## Cleanup

`scenarios/helper/lib.sh` installs an EXIT trap that kills `fakevault` and removes `$WORK_DIR` — no manual cleanup needed.
