---
status: active
---

# Scenario 007: hermetic end-to-end read via the fake TeamVault server

Validates `teamvault-cli` end-to-end against `cmd/fakevault` — a fake TeamVault HTTP server with seeded secrets — using a temp config + `TEAMVAULT_PASS` (no live TeamVault, no macOS Keychain, no personal probe key). Exercises the real HTTP connector, Basic-auth header, and JSON parse path that `--staging` (pure fixtures) and unit tests do not. This is the CI-runnable variant of scenario 001; CI runs it via `make e2e`.

The fastest path is `make e2e` (builds both binaries, starts fakevault, asserts, cleans up). The steps below are the manual walk it automates.

## Setup

- [ ] `go build -C ~/Documents/workspaces/sm-teamvault-cli -o /tmp/teamvault-cli .`
- [ ] `go build -C ~/Documents/workspaces/sm-teamvault-cli -o /tmp/fakevault ./cmd/fakevault`
- [ ] `/tmp/fakevault --addr 127.0.0.1:0 >/tmp/fv.log 2>&1 & FV_PID=$!`
- [ ] `URL=$(sed -n 's#^fakevault listening on ##p' /tmp/fv.log)` (repeat until non-empty)
- [ ] `printf '{"url":"%s","user":"test"}\n' "$URL" > /tmp/fv-config.json`
- [ ] `export TEAMVAULT_CONFIG=/tmp/fv-config.json TEAMVAULT_PASS=test`

## Action

- [ ] `PW=$(/tmp/teamvault-cli password --teamvault-key demo)`
- [ ] `USER=$(/tmp/teamvault-cli username --teamvault-key demo)`
- [ ] `URLOUT=$(/tmp/teamvault-cli url --teamvault-key demo)`
- [ ] `FILE=$(/tmp/teamvault-cli file --teamvault-key demo)`
- [ ] `BYTES=$(/tmp/teamvault-cli password --teamvault-key demo | wc -c | tr -d ' ')`

## Expected

- [ ] `[ "$PW" = "demo-pass-123" ]` (password fetched over real HTTP + Basic auth)
- [ ] `[ "$USER" = "demo-user" ]`
- [ ] `[ "$URLOUT" = "https://demo.example/login" ]`
- [ ] `[ "$FILE" = "demo-file-contents" ]` (file goes through the current-revision → `/data` path)
- [ ] `[ "$BYTES" = "13" ]` (no trailing newline — `demo-pass-123` is 13 bytes)

## Cleanup

```bash
kill "$FV_PID" 2>/dev/null; rm -f /tmp/teamvault-cli /tmp/fakevault /tmp/fv.log /tmp/fv-config.json
```
