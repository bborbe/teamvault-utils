---
status: active
---

# Scenario 009: htpasswd subcommand via the fake TeamVault server

Validates the `htpasswd <KEY>` subcommand end-to-end against `cmd/fakevault`. Exercises the real HTTP connector reading a secret's username + password and the shared `HtpasswdGenerator` bcrypt path — producing a deploy-ready `user:$2...` htpasswd line — against a live (fake) server, which the unit tests (mocked connector) do not.

Setup/assert helpers live in `scenarios/helper/lib.sh` (same convention as scenarios 007/008). CI runs the whole thing via `make e2e`; the fastest local path is also `make e2e`.

Covered cases: `htpasswd` on a seeded fixture (`demo`) prints `<username>:<bcrypt>` as a single line; the same works on a freshly `create`d secret, confirming the read → hash path over real HTTP (not just the mocked-connector unit tests).

## Setup

```bash
source scenarios/helper/lib.sh
build_binaries      # builds teamvault-cli + fakevault to a temp dir, sets $TV
start_fakevault     # starts the server, writes a temp config, exports TEAMVAULT_CONFIG
```

- [ ] `$TV` exists; `fakevault` is listening (`$FV_URL` non-empty); the `demo` fixture resolves

## Action + Expected

```bash
# Seeded fixture: demo -> username demo-user / password demo-pass-123.
HTP_DEMO="$("$TV" htpasswd --teamvault-key demo)"
assert_contains "htpasswd carries the username prefix" "demo-user:" "$HTP_DEMO"
assert_contains "htpasswd emits a bcrypt hash"         '$2'         "$HTP_DEMO"
assert_eq "htpasswd is a single user:bcrypt line" "1" \
	"$(printf '%s\n' "$HTP_DEMO" | grep -c '^demo-user:\$2')"

# Freshly created secret: htpasswd reads username+password back over real HTTP
# (same write path as scenario 008) and hashes them.
HTP_KEY="$(printf 'reg-pw' | "$TV" create --name htpasswd-e2e-secret --username reg-user --password-stdin)"
HTP_NEW="$("$TV" htpasswd --teamvault-key "$HTP_KEY")"
assert_contains "htpasswd of a created secret carries its username" "reg-user:" "$HTP_NEW"
assert_contains "htpasswd of a created secret is bcrypt"            '$2'        "$HTP_NEW"

scenario_done   # prints "e2e: PASS" and exits non-zero if any assertion failed
```

- [ ] All assertions print `ok:` and `scenario_done` reports `e2e: PASS`

## Cleanup

`scenarios/helper/lib.sh` installs an EXIT trap that kills `fakevault` and removes `$WORK_DIR` — no manual cleanup needed.
