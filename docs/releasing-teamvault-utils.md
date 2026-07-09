# Releasing teamvault-utils

How to ship a new version of teamvault-utils. Mandatory reading before tagging.

## One surface, one version stream

Unlike `vault-cli` / `dark-factory` / `coding` which ship both a binary and a Claude Code plugin, `teamvault-utils` is **binary-only**: a Go library + a single `teamvault` CLI binary (spf13/cobra) with subcommands (`teamvault login`, `teamvault password`, `teamvault username`, `teamvault url`, `teamvault file`, `teamvault config parse`, `teamvault config generate`) distributed via Go modules.

| Surface | Versioned by | Consumed by | Bumped how |
|---------|--------------|-------------|------------|
| **Binary + library** | git tag `vX.Y.Z` + matching `## vX.Y.Z` section in `CHANGELOG.md` | `go install github.com/bborbe/teamvault-utils/v5/cmd/teamvault@latest`; downstream Go projects importing the library | Auto-tagged by the project's own dark-factory daemon (`autoRelease: true`) when a prompt completes and updates `## Unreleased` |

There is no plugin to maintain. No marketplace JSONs to align. The CHANGELOG top entry and the git tag are the only two version sources, and `autoRelease` keeps them in lockstep.

## The release gate (run BEFORE approving binary-surface prompts)

`make precommit` does NOT cover real macOS Keychain behavior, real TeamVault API behavior, or libargument-CLI-flag wiring. Unit tests pass while runtime behavior is broken — that has bitten this project twice in a month:

- **v4.10.1** shipped a `security add-generic-password -w` invocation that prompted on `/dev/tty` instead of reading the password from stdin. Result: piped-stdin `teamvault login` silently stored an empty password. All unit tests passed because the agent mocked the `Executor` interface; the real `security` binary contract was never exercised.
- **v4.12.1** replaced the broken shell-out with a `security -i` REPL invocation that appended `\nquit\n` as a terminator. `security` rejects `quit` as an unknown command and exits 1, even though `add-generic-password` already ran successfully. The Go caller treated the exit code as a write failure and surfaced "store password in keychain failed" despite the keychain having been written correctly.

Both bugs were caught only when a scenario was walked against a built binary against the real macOS Keychain. The rule: **before approving any dark-factory prompt that touches `keychain*.go`, `pkg/factory/`, or `cmd/`, walk the relevant active scenarios against a freshly built binary.** Surface-scoped skipping is acceptable when the diff is genuinely empty (see below).

### Active scenarios at time of writing

```bash
awk '/^status:/{print FILENAME": "$2; nextfile}' scenarios/*.md
```

Expected output as of v4.13.0 baseline:

```
scenarios/001-teamvault-password-happy-path.md: active
scenarios/002-keychain-login-and-retrieve.md: active
scenarios/003-timeout-and-cache-fallback.md: active
scenarios/004-no-cache-timeout-propagates-error.md: active
scenarios/005-negative-timeout-rejected.md: active
```

All five are macOS-only and require the user's real `~/.teamvault.json` config + a working keychain entry. There is no `scenarios/helper/run-all.sh` yet — walk each markdown file by hand.

### Preflight

```bash
# Ensure jq + security binary available (both ship with macOS)
command -v jq && command -v security || echo "missing macOS tools"

# Confirm the operator's real keychain entry is intact BEFORE running any scenario
security find-generic-password -s teamvault-utils -a "$(jq -r .url ~/.teamvault.json)" -w >/dev/null && echo "keychain OK"
```

If the keychain entry is missing or empty, restore it before walking scenarios:

```bash
REAL_PASS=$(jq -r .xpass ~/.teamvault.json)
security add-generic-password -U -s teamvault-utils -a "$(jq -r .url ~/.teamvault.json)" -w "$REAL_PASS"
```

(Workaround: `~/.teamvault.json` keeps the password under the non-standard key `xpass` so the Config parser doesn't read it — it's only there as a recovery backup. Don't rename to `pass`; that re-introduces the plaintext-config attack surface that motivated spec 001.)

### Steps

```bash
# 1. Build a fresh binary (NOT the installed one) — one binary, all subcommands
go build -C ~/Documents/workspaces/teamvault/teamvault-utils -o /tmp/new-teamvault .

# 2. Walk each active scenario by hand
#    Each file's Setup → Action → Expected must pass.
ls scenarios/00[1-5]-*.md
```

If any active scenario fails: do NOT proceed to install or tag. Fix the regression first, then rerun the gate.

### When a scenario fails — where to look first

| Symptom | Most likely surface |
|---------|---------------------|
| `teamvault login` returns "inappropriate ioctl for device" when piped | the `login` subcommand's `termReader.Read` — `term.ReadPassword(stdin.Fd())` requires a real tty; the bufio prompt path is not reachable from piped stdin. The pre-existing behavior is "keychain has a password OR user types at tty" |
| Keychain entry stored but raw `security -w` shows different bytes than input | Expected post-v4.13.0. `zalando/go-keyring` stores in its own encoded format (a 32-char input shows as ~62 chars when read back via `security -w`). The user-facing round-trip via `teamvault password` / `teamvault username` is what matters — the raw security CLI value is library implementation detail |
| `teamvault username` returns 403 | Keychain was zeroed by an earlier scenario or test. Restore via the preflight command above. |
| Timeout test fails on a non-routable IP | macOS network stack quirks — 10.255.255.1 should hang on connect, but some VPN configurations route or RST it. Switch to another non-routable address (RFC 5737 test-net: 192.0.2.1, 198.51.100.1, 203.0.113.1) |

### Scenario-walk side effects

Scenarios 002 and any test that calls `teamvault login` will OVERWRITE the operator's existing keychain entry (zalando-encoded format post-v4.13.0). The downstream binaries read it back correctly, but a subsequent `security find-generic-password -w` shows zalando's storage shape, not the raw password. This is benign — the keychain works — but if you script automation around `security find-generic-password -w` extraction, that breaks. Restore the raw format if needed via the preflight command.

Scenario 003 pre-populates `~/.teamvault-cache/<key>/password` with a fake value and uses a temp `$HOME`. The temp HOME is cleaned up by the scenario; the real `~/.teamvault-cache` is not touched.

### When the diff is empty

The one valid skip: nothing on the binary surface changed since the latest tag.

```bash
LAST_TAG=$(git describe --tags --abbrev=0)
git diff "$LAST_TAG"..HEAD --name-only | grep -E '\.(go|mod|sum)$|^Makefile$'
# empty output → installed binary is byte-equivalent to /tmp/new-teamvault → skip
```

This is the ONLY documented skip. Do not invent others ("doc-only changes shouldn't break anything") — surface mappings are fragile.

## Binary release (automatic via dark-factory)

`.dark-factory.yaml` has `autoRelease: true`. Every successful prompt that touches `## Unreleased` triggers, in order:

1. Stage all changes (including the agent's `## Unreleased` entry)
2. Determine bump (patch/minor) from changelog content
3. Rename `## Unreleased` → `## vX.Y.Z`
4. Commit `release vX.Y.Z`
5. Tag `vX.Y.Z`, push tag and commit
6. Move the prompt file to `prompts/completed/` and push that commit too

The operator's responsibility is **run the release gate before approving any prompt that may produce a binary change.** Once the prompt is approved, the daemon ships whatever the agent produced — there is no second checkpoint.

| Prompt touches | Gate cadence |
|----------------|--------------|
| `keychain*.go`, `pkg/factory/`, `cmd/`, `*.go` outside test files | Full scenario walk before approving |
| `mocks/`, `*_test.go` only | Skip the scenarios; `make precommit` is sufficient |
| `docs/`, `README.md`, `CHANGELOG.md`-only edits | Skip |
| `prompts/`, `specs/`, frontmatter-only | Skip (pipeline metadata, not shipped) |

To verify a release shipped:

```bash
git fetch --tags
git describe --tags --abbrev=0           # latest tag, e.g. v4.13.0
git log "$(git describe --tags --abbrev=0)"..HEAD --oneline   # any unpushed commits beyond it
```

After a successful auto-release, both `git status` (clean) and `git rev-list @{u}..HEAD --count` (zero) should hold.

## Manual release (when autoRelease did NOT fire)

This is the case when:
- The work happened on a stuck-prompt-killed-then-finished-manually path (`<human-finish-note>` in the completed prompt)
- The daemon was not running when the work shipped
- The release-gate scenarios were run after the commit, not before
- A doc-only commit needs to be associated with a release (rare)

Today (v4.13.0 prep) is the canonical example: the migration prompt 009 was killed at 37 min after hitting the Go `*_darwin.go` filename-implicit build-constraint trap; the operator finished the work manually in commit `dc33341` and pushed to master without going through the daemon's commit pipeline → no auto-tag was created.

Procedure:

```bash
# 1. Pick the next version. Read CHANGELOG.md `## Unreleased` and decide:
#    - removed-public-API or behavior change → minor (X.Y+1.0)
#    - additive features / fixes only        → patch (X.Y.Z+1)
LAST_TAG=$(git describe --tags --abbrev=0)              # e.g. v4.12.1
NEXT_TAG=v4.13.0                                        # picked manually

# 2. Convert `## Unreleased` → `## vX.Y.Z` in CHANGELOG.md
sed -i '' "s|^## Unreleased\$|## ${NEXT_TAG#v}|" CHANGELOG.md

# 3. Commit + tag
git add CHANGELOG.md
git commit -m "release $NEXT_TAG"
git tag -a "$NEXT_TAG" -m "$NEXT_TAG"

# 4. Push
git push origin master "$NEXT_TAG"

# 5. Verify on github.com → Tags
gh release view "$NEXT_TAG" 2>/dev/null || echo "no GitHub Release yet (optional — see below)"
```

The git tag is sufficient for `go install github.com/bborbe/teamvault-utils/v5@vX.Y.Z`. A GitHub Release object is separate (see next section).

## GitHub Release (manual — when to surface a milestone)

`autoRelease` (and the manual procedure above) creates a `vX.Y.Z` git tag. Tags are sufficient for `go install`, `git describe`, and any tag-aware consumer.

A **GitHub Release** is a separate, deliberate act — distinct from the tag. It adds release notes, an entry on the repo's Releases tab, an RSS/atom feed for subscribers, and optional binary assets. Create one only after:

1. All active `scenarios/` pass against the current source tree.
2. The `CHANGELOG.md` entry summarises what users should care about — not the internal commit log.
3. The release is worth announcing (new feature, breaking change, security fix).

Skip the GitHub Release for internal refactors, dependency bumps, pre-release / experimental work, or chains of small tags. It is fine to skip several auto-tags and cumulate them into a single milestone Release later.

How:

```bash
TAG=$(git describe --tags --abbrev=0)
gh release create "$TAG" \
  --target master \
  --title "$TAG" \
  --notes "$(awk "/^## ${TAG#v}/,/^## v/" CHANGELOG.md | head -n -1)"
```

Verify on github.com → Releases tab. The Release object can be edited (notes, draft state) without retagging.

## Install (the moment a new version reaches consumers)

```bash
# Operator local install (single binary, all subcommands)
go install github.com/bborbe/teamvault-utils/v5/cmd/teamvault@latest
teamvault --help 2>&1 | head -1  # sanity check

# Downstream / fresh machine:
go install github.com/bborbe/teamvault-utils/v5@vX.Y.Z

# Library import (downstream Go projects):
#   go.mod: require github.com/bborbe/teamvault-utils/v5 vX.Y.Z
#   import: github.com/bborbe/teamvault-utils/v5
```

This is the step that bites downstream consumers if the gate was skipped. Other Go projects that import this library and any user's `teamvault` CLI on the next `go install` pick up the new binary. A regression in the new binary surfaces in their workflow, not yours.

## Release session checklist (operator template)

Copy this into a scratch note at the start of a release session. Flip `⬜` → `✅` as items tick off.

```text
⬜ Release gate: walk active scenarios 001–005 against /tmp/new-teamvault
⬜ Real keychain is intact (preflight `security find-generic-password ... | wc -c` > 0)
⬜ CHANGELOG.md `## Unreleased` describes user-facing changes (not commit log)
⬜ Decide bump: patch / minor / major based on diff
⬜ Rename `## Unreleased` → `## vX.Y.Z`
⬜ `git commit -m "release vX.Y.Z"`
⬜ `git tag -a vX.Y.Z -m "vX.Y.Z"`
⬜ `git push origin master vX.Y.Z`
⬜ `git status` clean; `git describe --tags --abbrev=0` returns the new tag
⬜ (optional) `gh release create vX.Y.Z ...` for milestone announcements
⬜ Restore keychain if scenarios re-stored in zalando-encoded format (raw security write)
```

When the daemon's `autoRelease` runs, only the bottom half of this list is needed — `make` and `git commit/tag/push` already happened. The gate (top item) still belongs to the operator.

## See also

- `CHANGELOG.md` — the source of truth for what shipped in each version
- `scenarios/` — the regression suite this gate runs
- `.dark-factory.yaml` — `autoRelease: true` field controls the auto-tag behavior
- `specs/completed/` — the behavioral contracts for everything in the latest version
- `~/Documents/workspaces/dark-factory/docs/release-process.md` — how `autoRelease` works in projects that use dark-factory (this is one of them)
