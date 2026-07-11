---
status: active
---

# Scenario 006: config-path resolution (XDG-first, legacy fallback, overrides)

Validates that the shipped `teamvault-cli` binary resolves its config file by the documented precedence — explicit flag → `TEAMVAULT_CONFIG` env → XDG (`$XDG_CONFIG_HOME/teamvault-cli/config.json`) → legacy (`~/.teamvault.json`) — and that an absent config is not silently replaced by a default read.

Assumes a working legacy `~/.teamvault.json` (url + user, no `pass`) with the password already in the macOS Keychain via `teamvault-cli login`. Probe key `lO4K1w` (personal vault, username "longhorn"). Override via `TV_PROBE_KEY`. Uses a **temp** `XDG_CONFIG_HOME` so the real `~/.config` is never touched.

## Setup

- [ ] `go build -C ~/Documents/workspaces/sm-teamvault-cli-defaultconfig -o /tmp/teamvault-cli .`
- [ ] `TV_KEY=${TV_PROBE_KEY:-lO4K1w}`
- [ ] `[ -f ~/.teamvault.json ]` (legacy config exists)
- [ ] `TV_XDG=$(mktemp -d)` (temp XDG root)
- [ ] `mkdir -p "$TV_XDG/teamvault-cli" && cp ~/.teamvault.json "$TV_XDG/teamvault-cli/config.json"` (XDG config = copy of legacy)

## Action

- [ ] `LEGACY_OUT=$(env -u TEAMVAULT_CONFIG XDG_CONFIG_HOME=/tmp/tv-empty-xdg /tmp/teamvault-cli username --teamvault-key $TV_KEY 2>/tmp/s006-legacy.err); LEGACY_RC=$?` (no XDG file present → legacy fallback)
- [ ] `XDG_OUT=$(env -u TEAMVAULT_CONFIG XDG_CONFIG_HOME="$TV_XDG" /tmp/teamvault-cli username --teamvault-key $TV_KEY 2>/tmp/s006-xdg.err); XDG_RC=$?` (XDG file present → XDG wins)
- [ ] `FLAG_OUT=$(XDG_CONFIG_HOME=/tmp/tv-empty-xdg TEAMVAULT_CONFIG=/nonexistent-env.json /tmp/teamvault-cli username --teamvault-config ~/.teamvault.json --teamvault-key $TV_KEY 2>/tmp/s006-flag.err); FLAG_RC=$?` (flag beats env + XDG + legacy)
- [ ] `BADENV_OUT=$(env -u XDG_CONFIG_HOME TEAMVAULT_CONFIG=/definitely/nonexistent.json /tmp/teamvault-cli username --teamvault-key $TV_KEY 2>/tmp/s006-badenv.err); BADENV_RC=$?` (env points at a missing file → NOT silently replaced by a default read)

## Expected

- [ ] `[ "$LEGACY_RC" = "0" ] && [ "$LEGACY_OUT" = "longhorn" ]` (legacy `~/.teamvault.json` read when no XDG file)
- [ ] `[ "$XDG_RC" = "0" ] && [ "$XDG_OUT" = "longhorn" ]` (XDG `config.json` read when present)
- [ ] `[ "$FLAG_RC" = "0" ] && [ "$FLAG_OUT" = "longhorn" ]` (explicit `--teamvault-config` wins over env + XDG)
- [ ] `[ "$BADENV_RC" != "0" ]` (a `TEAMVAULT_CONFIG` pointing at a missing file does NOT fall back to the default — the env value is authoritative; command fails for lack of url/user)

## Cleanup

```bash
rm -rf "$TV_XDG" /tmp/teamvault-cli /tmp/s006-*.err
```
