---
status: completed
spec: [001-macos-keychain-credential-store]
summary: Created teamvault-login binary with credential verification, Keychain write on success, hidden stdin prompt loop (3 attempts), platform-safe non-darwin notice, 18 Ginkgo tests, README Setup section, and CHANGELOG entry
container: teamvault-utils-005-spec-001-teamvault-login-binary
dark-factory-version: v0.156.1-1-g04f3863-dirty
created: "2026-05-12T16:49:55Z"
queued: "2026-05-12T17:16:10Z"
started: "2026-05-12T18:09:45Z"
completed: "2026-05-12T18:22:04Z"
---

<summary>
- Add a new `teamvault-login` binary that verifies credentials against the TeamVault API and then stores the password in the macOS Keychain.
- Resolution chain inputs (`--teamvault-pass`, `TEAMVAULT_PASS`, config file, Keychain via the previous prompt) are reused — `teamvault-login` does not re-implement them.
- On verification failure (wrong password) or missing password, the user is prompted on stdin (hidden), with up to 3 attempts. On success, the password is written to the Keychain.
- On non-macOS platforms the command still verifies but does not store, and prints a notice that Keychain storage is macOS-only in v1.
- README is updated to document `teamvault-login` as the recommended setup path on macOS and to flag putting the password in the config file as insecure.
- CHANGELOG records the new binary.
- One Docker container run produces the new binary, its tests, and the doc updates.
</summary>

<objective>
Deliver the `teamvault-login` binary so that a new macOS user can set up `teamvault-utils` without ever writing a TeamVault password into a plaintext file. The command must verify credentials before persisting them, so a wrong password is never silently saved.
</objective>

<context>
Project: `~/Documents/workspaces/teamvault/teamvault-utils/` — Go library + CLI tools under module `github.com/bborbe/teamvault-utils/v4`.

Read first:
- `CLAUDE.md` for project conventions.
- `docs/dod.md` for Definition of Done.
- `specs/in-progress/001-macos-keychain-credential-store.md` for full spec.
- `keychain.go` (created in the previous prompt) — `Keychain` interface, `NewKeychain()`, `KeychainServiceName`, and the `ErrKeychainNotSupported` sentinel.
- `factory/factory.go` — `CreateConnectorWithConfigAndKeychain` (new constructor from the previous prompt) and the existing `CreateConnectorWithConfig` that delegates to it.
- `cmd/teamvault-password/main.go` — canonical pattern for a teamvault-utils CLI: `libservice.MainCmd`, struct-tag flags/env, `factory.CreateConnectorWithConfig`.
- `cmd/teamvault-url/main.go`, `cmd/teamvault-username/main.go` — additional examples of the same pattern.
- `connector.go` and `remote-connector.go` (or whichever file backs the real `Connector`) — to understand which call to make for credential verification.

Read these guides in `~/.claude/plugins/marketplaces/coding/docs/`:
- `go-cli-guide.md` — CLI structure, flags, env parsing patterns.
- `go-error-wrapping-guide.md` — error wrapping with `github.com/bborbe/errors`.
- `go-testing-guide.md` — Ginkgo/Gomega + Counterfeiter conventions.
- `test-pyramid-triggers.md` — which test types to write.
- `readme-guide.md` — README structure and style for Go libraries.

Verification strategy: the lightest credential-validating call against the TeamVault API. Search the existing `Connector` implementation (`remote-connector.go` or similar) for a method that performs an authenticated GET and surfaces auth failures distinctly from "not found." If no obvious method exists, perform a `Connector.Search(ctx, "")` or `Connector.Search(ctx, "_login_probe_")` — the call hits an authenticated endpoint and returns a clear HTTP 401/403 on bad credentials. Prefer reusing an existing method over adding a new one to the `Connector` interface (the spec freezes that interface).

Hidden stdin input: use `golang.org/x/term` `term.ReadPassword(int(os.Stdin.Fd()))`. If `golang.org/x/term` is already in `go.mod` (it almost certainly is via `github.com/bborbe/service`), reuse it without adding to direct deps. If not present in `go.mod` at all, add it to direct deps — this is a justified new dep for the new binary.
</context>

<requirements>

1. **Create `cmd/teamvault-login/main.go`** mirroring the layout of `cmd/teamvault-password/main.go`. Application struct fields:

   ```go
   type application struct {
       TeamvaultUrl        string `required:"false" arg:"teamvault-url"    env:"TEAMVAULT_URL"    usage:"teamvault url"`
       TeamvaultUser       string `required:"false" arg:"teamvault-user"   env:"TEAMVAULT_USER"   usage:"teamvault user"`
       TeamvaultPass       string `required:"false" arg:"teamvault-pass"   env:"TEAMVAULT_PASS"   usage:"teamvault password" display:"length"`
       TeamvaultConfigPath string `required:"false" arg:"teamvault-config" env:"TEAMVAULT_CONFIG" usage:"teamvault config"`
       Staging             bool   `required:"false" arg:"staging"          env:"STAGING"          usage:"staging status" default:"false"`
   }
   ```

   No `TeamvaultKey` field — login does not look up a secret, it verifies and stores the master credential.

2. **`Run(ctx context.Context)` flow** in `cmd/teamvault-login/main.go`:

   a. Resolve URL + User + Password via the existing chain:
      - Build the connector with `factory.CreateConnectorWithConfig(...)` exactly as `teamvault-password` does. This handles flag/env/config-file/Keychain transparently.
      - Separately re-resolve the URL (config-file URL, else flag, else env) for the post-success `WritePassword` step. (The factory consumes URL internally; the login command needs it again to address the Keychain entry.)

   b. Attempt verification using the resolved connector. Wrap the call in `context.WithTimeout(ctx, 10*time.Second)` so an unreachable TeamVault doesn't hang the login flow indefinitely. See "Verification strategy" in `<context>` for the call to use.

   c. If verification succeeds, call `keychain.WritePassword(ctx, url, pass)` unconditionally. The write is idempotent (`-U` overwrites), so the spec's "do not re-write if password came from Keychain" optimization is not worth its complexity — always-write keeps the code straight and behavior predictable. On the non-darwin platforms the call returns the `ErrKeychainNotSupported` sentinel; see step 2f. On macOS write success, print `Login successful. Password stored in macOS Keychain for <url>.` to stderr and exit 0.

   d. If verification fails with an authentication error (HTTP 401/403) or no password was resolved at all, enter the prompt loop:

      ```
      for attempt := 1; attempt <= 3; attempt++ {
          // print "TeamVault password for <user>@<url>: " to stderr (NOT stdout — keeps stdout scriptable)
          // read password from stdin via term.ReadPassword
          // rebuild a one-shot connector with this password (do NOT touch the Keychain yet)
          // wrap verify call in context.WithTimeout(ctx, 10*time.Second)
          // verify
          // on success: break
          // on auth failure: print "Invalid password, try again." to stderr (suppress on last attempt)
      }
      ```

      On exhaustion of 3 attempts: return a wrapped error `login failed: 3 invalid password attempts`. `libservice.MainCmd` will exit non-zero.

      On `io.EOF` from stdin (Ctrl-D) or `context.Canceled` (Ctrl-C): return a wrapped error `login aborted`. Exit non-zero.

   e. On a non-auth verification error (network unreachable, TLS error, malformed URL, context-deadline-exceeded, etc.): return the wrapped error immediately — do NOT prompt for a password, the credential isn't the problem. Make the error message actionable: include the URL and the underlying error class.

   f. After a successful prompt loop, call `keychain.WritePassword(ctx, url, pass)`:
      - On macOS success → print `Login successful. Password stored in macOS Keychain for <url>.` to stderr and exit 0.
      - On `errors.Is(err, teamvault.ErrKeychainNotSupported)` (non-darwin platforms) → print `Login successful. (Keychain storage is macOS-only in v1; password not persisted.)` to stderr and exit 0. This branch is treated as a non-fatal info notice.
      - On any other write error (real Keychain failure: locked Keychain, security-binary error, etc.) → return the wrapped error. The verification already succeeded; the user knows the password is correct, just couldn't be saved. The wrapped error must mention the URL and suggest unlocking the Keychain if applicable.

   g. Never print the password to stdout, stderr, or `glog`.

3. **Unit tests** at `cmd/teamvault-login/main_test.go`:

   The `Run` method is heavy — factor the prompt loop, verification call, and write step into a separate function (e.g. `func loginFlow(ctx context.Context, in io.Reader, errOut io.Writer, connector Connector, keychain Keychain, url Url, initialPass Password) error`) and test that function with Ginkgo/Gomega. Use Counterfeiter fakes for both `Connector` and `Keychain`.

   Required scenarios (each as its own `It`):

   - Resolved credentials already verify → no prompt → Keychain write called once with the URL + password (always-write semantics).
   - No password resolved → prompts once → user types correct password → connector verifies → Keychain write called once with correct URL + password.
   - No password resolved → prompts three times wrong → returns "3 invalid password attempts" error → Keychain write NOT called.
   - User hits Ctrl-D after one wrong attempt → returns "login aborted" error → Keychain write NOT called.
   - Verification returns network error (non-auth) → returns wrapped error immediately → no prompts → no Keychain write.
   - Verification times out (simulate by returning `context.DeadlineExceeded`) → returns wrapped error immediately → no prompts.
   - Keychain write returns `ErrKeychainNotSupported` (non-darwin path simulated by fake) → `loginFlow` returns nil → stderr contains the "macOS-only" notice → stdout is empty.
   - Keychain write returns a real "locked Keychain" error → returns wrapped error → stderr error mentions the URL.
   - Successful flow: stdout is empty (machine-scriptable), confirmation message goes to stderr.
   - Sentinel-shape regression guard: assert `errors.Is(stubErr, teamvault.ErrKeychainNotSupported)` returns true when the `Keychain` fake returns `teamvault.ErrKeychainNotSupported` directly.

4. **Update README.md.** Add a top-level section titled exactly `## Setup (macOS, recommended)` that documents:

   - The recommended flow: write a config file with `url` and `user` only (no `pass` field), then run `teamvault-login` once to store the password in the Keychain.
   - A short note that putting `pass` in the config file is the legacy path — still works, but the password is stored in plaintext on disk; prefer the Keychain.
   - A short note that multi-vault is supported: run `teamvault-login --teamvault-config ~/.teamvault.json` and `teamvault-login --teamvault-config ~/.teamvault-sm.json` for each.
   - On non-macOS: `teamvault-login` verifies credentials but does not persist; users still rely on flag/env/config for now.
   - Mention `security delete-generic-password -s teamvault-utils -a <url>` as the way to remove a stored password until a dedicated subcommand exists.

5. **CHANGELOG.** Append under `## Unreleased`:

   ```
   - Added `teamvault-login` command: verifies TeamVault credentials against the API and stores the password in the macOS Keychain on success. Replaces the need to write the TeamVault password into the config file as plaintext.
   ```

6. **Wire the new binary into the build.** If the `Makefile` builds each `cmd/<name>/` automatically, no change is needed. If it lists binaries explicitly (e.g. a `BINARIES :=` variable), add `teamvault-login` to the list. Match the existing pattern.

7. **Do NOT modify `keychain.go` or `keychain_other.go`.** The `ErrKeychainNotSupported` sentinel and `KeychainServiceName` constant were defined in the previous prompt's output and are part of its public surface. This prompt only consumes them via `errors.Is` and direct reference.

</requirements>

<constraints>
- Public Go API of `Config`, `Connector`, `TeamvaultConfigPath.Parse()`, and `factory.CreateConnectorWithConfig` signatures must NOT change.
- No new mandatory module dependencies on Linux beyond what `github.com/bborbe/service` already pulls in. `golang.org/x/term` is acceptable if not already present (it is required for the hidden-stdin prompt).
- The password value must never be printed to stdout, stderr, or `glog`. The user input is consumed only by the connector and (on macOS) by the Keychain backend.
- All user-facing messages from `teamvault-login` go to stderr — stdout is reserved for scriptable output (currently nothing). This matches the existing `teamvault-password` pattern where the secret is the only stdout content.
- Maximum 3 password prompts per invocation. Ctrl-C / Ctrl-D abort cleanly.
- Each verification call is wrapped in a 10-second `context.WithTimeout`. An unreachable TeamVault server must not hang `teamvault-login`.
- The verification call must use the same `Connector` returned by `factory.CreateConnectorWithConfig` — do not bypass the existing connector layer.
- Use `github.com/bborbe/errors` for all wrapping; `glog` for any debug-level logging (never include the password in log values).
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass.
</constraints>

<verification>
- `make precommit` exits 0.
- `go build ./cmd/teamvault-login` produces a binary.
- Running `./teamvault-login --help` (or whatever `libservice.MainCmd` produces for `-h`) lists the documented flags without `--teamvault-key`.
- README has a `## Setup (macOS, recommended)` section (exact title from requirement 4) describing the new flow.
- CHANGELOG `## Unreleased` lists both this prompt's entry and the previous prompt's entry.
- Unit tests cover the scenarios listed in requirement 3.
- `grep -rn "Password\|TeamvaultPass" cmd/teamvault-login/` shows no `fmt.Print*` / `os.Stdout.Write` / `glog.*` lines that emit the password value.
- `GOOS=linux go build ./cmd/teamvault-login` builds cleanly — the sentinel-based non-darwin path compiles.
</verification>
