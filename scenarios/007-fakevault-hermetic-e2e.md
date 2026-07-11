---
status: active
---

# Scenario 007: hermetic end-to-end via the fake TeamVault server

Validates `teamvault-cli` end-to-end against `cmd/fakevault` — a fake TeamVault HTTP server with seeded secrets — using a temp config file that carries url + user + pass (no live TeamVault, no macOS Keychain, no personal probe key). Exercises the real HTTP connector, Basic-auth header, and JSON parse path that `--staging` (pure fixtures) and unit tests do not.

Setup/assert helpers live in `scenarios/helper/lib.sh` (same convention as vault-cli / dark-factory). CI runs the whole thing via `make e2e`; the fastest local path is also `make e2e`.

Covered cases: `password`, `username`, `url`, `file`, `config parse`, `config generate`, `not-found error`, `auth failure (401)`, config-via-env (`TEAMVAULT_CONFIG`), config-via-flag (`--teamvault-config`), no-trailing-newline.

## Setup

```bash
source ~/Documents/workspaces/sm-teamvault-cli/scenarios/helper/lib.sh
build_binaries      # builds teamvault-cli + fakevault to a temp dir, sets $TV
start_fakevault     # starts the server, writes a temp config (url+user+pass), exports TEAMVAULT_CONFIG
```

- [ ] `$TV` and `$WORK_DIR/config.json` exist; `fakevault` is listening (`$FV_URL` non-empty)

## Action + Expected

```bash
assert_eq "password" "demo-pass-123"              "$("$TV" password --teamvault-key demo)"
assert_eq "username" "demo-user"                  "$("$TV" username --teamvault-key demo)"
assert_eq "url"      "https://demo.example/login" "$("$TV" url --teamvault-key demo)"
assert_eq "file"     "demo-file-contents"         "$("$TV" file --teamvault-key demo)"
assert_eq "config parse" "user=demo-user pass=demo-pass-123" \
	"$(printf 'user={{ teamvaultUser "demo" }} pass={{ teamvaultPassword "demo" }}' | "$TV" config parse)"
assert_exit_nonzero "not-found error" "$TV" password --teamvault-key nope-does-not-exist
mkdir -p "$WORK_DIR/gen-src" && printf 'db_pass={{ teamvaultPassword "demo" }}' >"$WORK_DIR/gen-src/app.conf"
"$TV" config generate --source-dir "$WORK_DIR/gen-src" --target-dir "$WORK_DIR/gen-out"
assert_eq "config generate" "db_pass=demo-pass-123" "$(cat "$WORK_DIR/gen-out/app.conf")"
printf '{"url":"%s","user":"test","pass":"wrong"}\n' "$FV_URL" >"$WORK_DIR/badconfig.json"
assert_exit_nonzero "auth failure (401)" "$TV" password --teamvault-config "$WORK_DIR/badconfig.json" --teamvault-key demo
assert_eq "config via env"  "demo-user" "$("$TV" username --teamvault-key demo)"
assert_eq "config via flag" "demo-user" \
	"$(env -u TEAMVAULT_CONFIG "$TV" username --teamvault-config "$WORK_DIR/config.json" --teamvault-key demo)"
assert_eq "no trailing newline" "13" "$("$TV" password --teamvault-key demo | wc -c | tr -d ' ')"
scenario_done   # prints "e2e: PASS" and exits non-zero if any assertion failed
```

- [ ] All assertions print `ok:` and `scenario_done` reports `e2e: PASS`

## Cleanup

`scenarios/helper/lib.sh` installs an EXIT trap that kills `fakevault` and removes `$WORK_DIR` — no manual cleanup needed.
