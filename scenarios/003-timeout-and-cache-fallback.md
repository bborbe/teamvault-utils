---
status: active
---

# Scenario 003: Configured timeout fires; disk cache returns cached value

Validates that when TeamVault is unreachable and `cacheEnabled` is on (via config OR `--cache=true` flag), the binary returns the cached value within the configured timeout. Locks the v4.11+ timeout × disk-fallback path and the OR-logic precedence fix.

## Setup

- [ ] `go build -C ~/Documents/workspaces/teamvault/teamvault-utils -o /tmp/teamvault .`
- [ ] `WORK_DIR=$(mktemp -d)`
- [ ] `TV_HOME="$WORK_DIR/home"; mkdir -p "$TV_HOME"`
- [ ] `TV_KEY=probe-key-003`
- [ ] `CACHED_VALUE="cached-secret-from-scenario-003"`
- [ ] Pre-populate disk cache (mirrors `diskfallback-connector.go` layout):
      `mkdir -p "$TV_HOME/.teamvault-cache/$TV_KEY" && printf '%s' "$CACHED_VALUE" > "$TV_HOME/.teamvault-cache/$TV_KEY/password"`
- [ ] Config with `cacheEnabled` + `timeout` in JSON:
      `printf '{"url":"https://10.255.255.1","user":"probe","pass":"probe","cacheEnabled":true,"timeout":"1s"}' > "$WORK_DIR/with-cache-enabled.json"`
- [ ] Config WITHOUT `cacheEnabled` (CLI flag will provide it):
      `printf '{"url":"https://10.255.255.1","user":"probe","pass":"probe","timeout":"1s"}' > "$WORK_DIR/no-cache-field.json"`
- [ ] `! timeout 2 nc -z 10.255.255.1 443 2>/dev/null` (verify 10.255.255.1 is unreachable)
- [ ] `[ -s "$TV_HOME/.teamvault-cache/$TV_KEY/password" ]` (cache file populated)

## Action

### Path A — config sets cacheEnabled

- [ ] `START_A=$(date +%s); OUT_A=$(env HOME="$TV_HOME" /tmp/teamvault password --teamvault-config $WORK_DIR/with-cache-enabled.json --teamvault-key $TV_KEY 2>/tmp/scenario-003-A.err); RC_A=$?; END_A=$(date +%s); DURATION_A=$((END_A - START_A))`

### Path B — CLI `--cache=true` overrides absent config flag (OR-logic regression check)

- [ ] `START_B=$(date +%s); OUT_B=$(env HOME="$TV_HOME" /tmp/teamvault password --teamvault-config $WORK_DIR/no-cache-field.json --teamvault-key $TV_KEY --cache=true 2>/tmp/scenario-003-B.err); RC_B=$?; END_B=$(date +%s); DURATION_B=$((END_B - START_B))`

## Expected

### Path A

- [ ] `[ "$RC_A" = "0" ]`
- [ ] `[ "$OUT_A" = "$CACHED_VALUE" ]`
- [ ] `[ "$DURATION_A" -lt 3 ]` (timeout fired and fallback ran — proves no 5s hardcode regression, no unbounded hang)

### Path B

- [ ] `[ "$RC_B" = "0" ]`
- [ ] `[ "$OUT_B" = "$CACHED_VALUE" ]` (CLI `--cache=true` enabled cache despite config omitting `cacheEnabled`)
- [ ] `[ "$DURATION_B" -lt 3 ]`

## Cleanup

```bash
rm -rf "$WORK_DIR" /tmp/teamvault /tmp/scenario-003-*.err
```
