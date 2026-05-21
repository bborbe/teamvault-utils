---
status: active
---

# Scenario 004: Timeout without cache surfaces a transport error

Validates that when TeamVault is unreachable and cache is disabled, the binary exits non-zero with a timeout-class error rather than hanging or silently returning empty output. Counterpart to scenario 003 — same setup, cache off.

## Setup

- [ ] `go build -C ~/Documents/workspaces/teamvault/teamvault-utils -o /tmp/new-teamvault-password ./cmd/teamvault-password`
- [ ] `WORK_DIR=$(mktemp -d)`
- [ ] `TV_HOME="$WORK_DIR/home"; mkdir -p "$TV_HOME"`
- [ ] `TV_KEY=probe-key-004`
- [ ] Config WITHOUT `cacheEnabled`, timeout 1s, non-routable URL:
      `printf '{"url":"https://10.255.255.1","user":"probe","pass":"probe","timeout":"1s"}' > "$WORK_DIR/no-cache.json"`
- [ ] `! timeout 2 nc -z 10.255.255.1 443 2>/dev/null` (verify 10.255.255.1 is unreachable)
- [ ] `[ ! -d "$TV_HOME/.teamvault-cache" ]` (no cache directory exists — nothing to fall back to even if cache were enabled)

## Action

- [ ] `START=$(date +%s); OUT=$(env HOME="$TV_HOME" /tmp/new-teamvault-password --teamvault-config $WORK_DIR/no-cache.json --teamvault-key $TV_KEY 2>/tmp/scenario-004.err); RC=$?; END=$(date +%s); DURATION=$((END - START))`

## Expected

- [ ] `[ "$RC" != "0" ]` (non-zero exit)
- [ ] `[ -z "$OUT" ]` (stdout empty — no partial output)
- [ ] `grep -qiE 'timeout|deadline|context canceled|Client\.Timeout' /tmp/scenario-004.err` (stderr names the transport failure)
- [ ] `[ "$DURATION" -lt 3 ]` (failure surfaces within timeout window — no unbounded hang)

## Cleanup

```bash
rm -rf "$WORK_DIR" /tmp/new-teamvault-password /tmp/scenario-004.err
```
