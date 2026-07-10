---
status: active
---

# Scenario 005: Negative timeout is rejected at startup

Validates that `--teamvault-timeout=-1s` (or a negative `timeout` field in the config) causes the binary to exit non-zero with a clear validation error before any network call is attempted. Locks the v4.11+ negative-duration rejection at the factory boundary.

## Setup

- [ ] `go build -C ~/Documents/workspaces/sm-teamvault-cli -o /tmp/teamvault-cli .`
- [ ] `WORK_DIR=$(mktemp -d)`
- [ ] Throwaway config (real URL/user not needed — validation fires before any call):
      `printf '{"url":"https://example.invalid","user":"probe","pass":"probe"}' > "$WORK_DIR/config.json"`

## Action

### Path A — negative timeout via CLI flag

- [ ] `START_A=$(date +%s); OUT_A=$(/tmp/teamvault-cli password --teamvault-config $WORK_DIR/config.json --teamvault-key any-key --teamvault-timeout=-1s 2>/tmp/scenario-005-A.err); RC_A=$?; END_A=$(date +%s); DURATION_A=$((END_A - START_A))`

### Path B — negative timeout via config file

- [ ] `printf '{"url":"https://example.invalid","user":"probe","pass":"probe","timeout":"-5s"}' > "$WORK_DIR/negative-config.json"`
- [ ] `START_B=$(date +%s); OUT_B=$(/tmp/teamvault-cli password --teamvault-config $WORK_DIR/negative-config.json --teamvault-key any-key 2>/tmp/scenario-005-B.err); RC_B=$?; END_B=$(date +%s); DURATION_B=$((END_B - START_B))`

## Expected

### Path A

- [ ] `[ "$RC_A" != "0" ]`
- [ ] `[ -z "$OUT_A" ]`
- [ ] `grep -q 'invalid timeout' /tmp/scenario-005-A.err`
- [ ] `grep -q -- '-1s' /tmp/scenario-005-A.err` (error mentions the offending value)
- [ ] `[ "$DURATION_A" -lt 2 ]` (validation fires fast — no network attempt)

### Path B

- [ ] `[ "$RC_B" != "0" ]`
- [ ] `[ -z "$OUT_B" ]`
- [ ] `grep -q 'invalid timeout' /tmp/scenario-005-B.err`
- [ ] `grep -q -- '-5s' /tmp/scenario-005-B.err`
- [ ] `[ "$DURATION_B" -lt 2 ]`

## Cleanup

```bash
rm -rf "$WORK_DIR" /tmp/teamvault-cli /tmp/scenario-005-*.err
```
