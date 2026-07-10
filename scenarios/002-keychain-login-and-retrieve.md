---
status: active
---

# Scenario 002: teamvault-cli login persists password; subsequent reads use Keychain

Validates that `teamvault login` stores the password in the macOS Keychain and that subsequent `teamvault-cli` subcommand calls read it from there without a stdin prompt or a plaintext `pass` in the config. macOS-only.

## Setup

- [ ] `[ "$(uname -s)" = "Darwin" ]` (else skip)
- [ ] `go build -C ~/Documents/workspaces/sm-teamvault-cli -o /tmp/teamvault-cli .`
- [ ] `TV_URL=$(jq -r .url ~/.teamvault.json)` (vault URL from existing config)
- [ ] `TV_USER=$(jq -r .user ~/.teamvault.json)` (vault user from existing config)
- [ ] `TV_PASS=$(security find-generic-password -s teamvault-cli -a "$TV_URL" -w)` (replay existing Keychain entry)
- [ ] `WORK_DIR=$(mktemp -d)`
- [ ] Write throwaway config without `pass`:
      `printf '{"url":"%s","user":"%s"}\n' "$TV_URL" "$TV_USER" > "$WORK_DIR/teamvault.json"`
- [ ] `TV_CONFIG="$WORK_DIR/teamvault.json"`
- [ ] `TV_KEY=${TV_PROBE_KEY:-lO4K1w}`
- [ ] `jq 'has("pass")' "$TV_CONFIG"` prints `false` (no plaintext password in the test config)

## Action

- [ ] `LOGIN_OUT=$(printf '%s\n' "$TV_PASS" | /tmp/teamvault-cli login --teamvault-config $TV_CONFIG 2>&1); LOGIN_RC=$?`
- [ ] `PW_OUT=$(/tmp/teamvault-cli password --teamvault-config $TV_CONFIG --teamvault-key $TV_KEY </dev/null 2>/tmp/scenario-002-pw.err); PW_RC=$?`
- [ ] `USER_OUT=$(/tmp/teamvault-cli username --teamvault-config $TV_CONFIG --teamvault-key $TV_KEY </dev/null 2>/tmp/scenario-002-user.err); USER_RC=$?`

## Expected

- [ ] `[ "$LOGIN_RC" = "0" ]` (login succeeded)
- [ ] `echo "$LOGIN_OUT" | grep -qi 'login successful'` (login confirmation message)
- [ ] `[ -n "$(security find-generic-password -s teamvault-cli -a "$TV_URL" -w 2>/dev/null)" ]` (Keychain entry exists after login; raw value is zalando-encoded post-v0.9.10+ migration, so a byte-equal comparison to the input password is no longer meaningful — user-facing round-trip is verified via `teamvault password` / `teamvault username` below)
- [ ] `[ "$PW_RC" = "0" ]` (password retrieval via Keychain succeeded)
- [ ] `[ -n "$PW_OUT" ]` (password stdout non-empty — proves Keychain read from a no-`pass` config)
- [ ] `! grep -qi 'password' /tmp/scenario-002-pw.err` (no password prompt appeared)
- [ ] `[ "$USER_RC" = "0" ]`
- [ ] `[ "$USER_OUT" = "longhorn" ]` (resolved password authenticated against the real API)
- [ ] `! grep -qi 'password' /tmp/scenario-002-user.err`

## Cleanup

```bash
rm -rf "$WORK_DIR" /tmp/teamvault-cli /tmp/scenario-002-*.err
```

Keychain entry is intentionally left in place — it was reused, not created by this scenario.
