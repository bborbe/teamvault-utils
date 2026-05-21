---
status: active
---

# Scenario 001: teamvault-password and teamvault-username read a secret end-to-end

Validates that `teamvault-password` and `teamvault-username` resolve config + keychain + remote and print the resolved value to stdout. Smoke test proving libargument parsing, factory wiring, and remote call work in the shipped binaries.

Assumes a working `~/.teamvault.json` (url + user, no `pass`) with the password already in the macOS Keychain via `teamvault-login`. Probe key `lO4K1w` (personal vault, username "longhorn"). Override via `TV_PROBE_KEY` for other setups.

## Setup

- [ ] `go build -C ~/Documents/workspaces/teamvault/teamvault-utils -o /tmp/new-teamvault-password ./cmd/teamvault-password`
- [ ] `go build -C ~/Documents/workspaces/teamvault/teamvault-utils -o /tmp/new-teamvault-username ./cmd/teamvault-username`
- [ ] `TV_CONFIG=~/.teamvault.json`
- [ ] `TV_KEY=${TV_PROBE_KEY:-lO4K1w}`
- [ ] `[ -f "$TV_CONFIG" ]` (config file exists)
- [ ] `security find-generic-password -s teamvault-utils -a "$(jq -r .url $TV_CONFIG)"` returns an entry (Keychain has the password)

## Action

- [ ] `PASS_OUT=$(/tmp/new-teamvault-password --teamvault-config $TV_CONFIG --teamvault-key $TV_KEY 2>/tmp/scenario-001-pw.err); PASS_RC=$?`
- [ ] `USER_OUT=$(/tmp/new-teamvault-username --teamvault-config $TV_CONFIG --teamvault-key $TV_KEY 2>/tmp/scenario-001-user.err); USER_RC=$?`

## Expected

- [ ] `[ "$PASS_RC" = "0" ]` (password command exit 0)
- [ ] `[ -n "$PASS_OUT" ]` (password stdout non-empty)
- [ ] `[ ! -s /tmp/scenario-001-pw.err ]` (password stderr empty)
- [ ] `[ "$USER_RC" = "0" ]` (username command exit 0)
- [ ] `[ "$USER_OUT" = "longhorn" ]` (username stdout exactly `longhorn`)
- [ ] `[ ! -s /tmp/scenario-001-user.err ]` (username stderr empty)

## Cleanup

```bash
rm -f /tmp/new-teamvault-password /tmp/new-teamvault-username /tmp/scenario-001-pw.err /tmp/scenario-001-user.err
```
