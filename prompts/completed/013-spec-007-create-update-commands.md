---
status: completed
spec: [007-create-update-write-commands]
summary: Added `create` and `update <key>` subcommands to teamvault-cli with secure password input (--password-stdin, --generate), mutual-exclusion validation, metadata-only update support, and key-only/json output
execution_id: sm-teamvault-cli-exec-013-spec-007-create-update-commands
dark-factory-version: dev
created: "2026-07-14T14:00:01Z"
queued: "2026-07-14T13:56:02Z"
started: "2026-07-14T14:04:57Z"
completed: "2026-07-14T14:18:20Z"
branch: dark-factory/create-update-write-commands
---

<summary>
- Adds two new commands to `teamvault-cli`: `create` (make a new secret) and `update <key>` (change an existing one).
- `create` figures out whether the secret is a password or a file from which value flag you pass, requires exactly one value source, and prints the new secret's key.
- `update` takes the key as a positional argument, sends only the flags you actually passed, and allows metadata-only edits (no value flag).
- Password input is secure by default: `--password-stdin` reads from a pipe and `--generate` asks the server for one; `--password <val>` is offered for convenience but its help text warns it leaks via shell history and `ps`.
- The value sources are mutually exclusive, empty piped input is rejected, and an interactive terminal with nothing piped does not hang forever.
- `--file <path>` reads the file and base64-encodes it into the payload; a missing/unreadable file fails before any network call.
- Output is secret-safe: only the key is printed (or `{"key","api_url"}` with `--json`); no password or file value ever reaches stdout or a log line.
</summary>

<objective>
Add `create` and `update <key>` cobra subcommands to `pkg/cli`, wired to the `teamvault.Writer` from prompt 1. `create` infers content type from the value flag, validates exactly-one value source, and prints the new key. `update` takes a positional key, sends only the flags passed, allows metadata-only updates, and creates a new revision when a value flag is given. Password input is secure by default (`--password-stdin` / `--generate`), `--password` warns about its leak risk, value sources are mutually exclusive, empty stdin is rejected, an interactive TTY does not block forever, `--file` base64-encodes bytes, and output is the key only (or `{"key","api_url"}` with `--json`) with no secret value ever on stdout or in a log.
</objective>

<context>
Read `CLAUDE.md` for project conventions. This prompt runs AFTER prompt 1 (the `teamvault.Writer` interface, `NewRemoteWriter`, `factory.CreateRemoteWriter`, `CreateSecret`/`UpdateSecret`/`ContentType` types, and `pkg/mocks/writer.go` all exist). If any of those symbols is missing, STOP and report `status: failed` with `"writer layer from prompt 1 not yet present"` — do NOT re-create them here.

Read before changing:
- `pkg/cli/cli.go` — the cobra command tree. Key things to reuse/follow:
  - `NewRootCommand(ctx)` registers subcommands via `rootCmd.AddCommand(...)`. Add `createCreateCommand(ctx, sf)` and `createUpdateCommand(ctx, sf)` here.
  - `type sharedFlags struct { url, user, pass, configPath string; staging, cache bool; timeout string }` — the persistent shared flags. Your new commands read config/creds through `sf`.
  - `(sf *sharedFlags) buildConnector(ctx) (teamvault.Connector, error)` — the read-path builder. It resolves the timeout, builds the http client via `factory.CreateHttpClient(ctx)`, and calls `factory.CreateConnectorWithConfigAndTimeout(...)`, which internally resolves credentials (flag→env→config→Keychain). You need an ANALOGOUS builder that returns a `teamvault.Writer` with the SAME resolved url/user/pass/timeout. See requirement 1.
  - `createSecretCommand(...)` and `createInfoCommand(...)` — the factory shape to imitate (`&cobra.Command{Use, Short, Args, RunE}`, local flags via `cmd.Flags()`, `cmd.OutOrStdout()`).
  - `writeSecret(...)` — the `--json` output helper for reads (`json.Marshal(map[string]string{field: value})`, single line). Your output helper is analogous but emits `{"key":..,"api_url":..}`.
- `pkg/cli/login.go` — CREDENTIAL RESOLUTION reference. `createLoginCommand`'s `RunE` shows how to resolve `resolvedURL`/`resolvedUser`/`initialPass` from flags → config file → Keychain, and how to build the http client + resolve the timeout (`libtime.ParseDuration`, `httpClient.Timeout = effective`). Reuse the SAME resolution so `create`/`update` authenticate exactly like the read path. Also note `term.ReadPassword` / `golang.org/x/term` usage (already a dependency) — you'll use `term.IsTerminal(int(os.Stdin.Fd()))` for the TTY guard.
- `pkg/factory/factory.go` — `CreateHttpClient(ctx) (*http.Client, error)`, `CreateConnectorWithConfigAndTimeout(...)`, and (from prompt 1) `CreateRemoteWriter(httpClient, apiURL, apiUser, apiPassword, currentDateTime) teamvault.Writer`.
- `pkg/writer.go` (from prompt 1) — `Writer` interface (`Create(ctx, CreateSecret) (Key, ApiUrl, error)`, `Update(ctx, Key, UpdateSecret) (Key, ApiUrl, error)`, `GeneratePassword(ctx) (Password, error)`), plus `CreateSecret`, `UpdateSecret`, `ContentType`, `ContentTypePassword`, `ContentTypeFile`.
- `pkg/mocks/writer.go` (from prompt 1) — Counterfeiter `mocks.Writer`, used to unit-test the commands without a network.
- `pkg/config.go` / `pkg/config-parser.go` — how config carries url/user/pass so credential resolution matches the read path.

Read these coding guides (in-container paths):
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-cli-guide.md` — cobra command factories, `RunE` returning error, `SilenceUsage`, positional args (`cobra.ExactArgs(1)`), local vs persistent flags.
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-patterns.md` — error wrapping, interfaces.
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-testing-guide.md` — Ginkgo/Gomega + Counterfeiter, coverage ≥80%, error-path tests.
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md` — `bborbe/errors`, no `fmt.Errorf`.

Library facts (verified):
- `golang.org/x/term v0.43.0` IS already a dependency (used in `login.go`). `term.IsTerminal(fd int) bool` and `term.ReadPassword(fd int) ([]byte, error)` are available.
- `libtime.ParseDuration(ctx, value) (*libtime.Duration, error)` parses the `--teamvault-timeout` string; `libtime.NewCurrentDateTime()` builds the `CurrentDateTime`.
- The four secret readers live in `pkg/cli/cli.go`; keep the new commands in the SAME package (`pkg/cli`), a new file per command is fine (`pkg/cli/create.go`, `pkg/cli/update.go`).
</context>

<requirements>

1. **Add a shared writer-builder** on `sharedFlags` (in `pkg/cli`, e.g. a new file `pkg/cli/writer.go` or appended to `cli.go`), analogous to `buildConnector` but returning `teamvault.Writer` with credentials resolved EXACTLY like the read/login path (flag → env → config file → Keychain), the same timeout resolution, and the same http client:
   ```go
   func (sf *sharedFlags) buildWriter(ctx context.Context) (teamvault.Writer, error)
   ```
   Resolve url/user/pass and timeout the same way `login.go`'s `RunE` does (parse config when `configPath.Exists()`, fall back to `teamvault.NewKeychain().ReadPassword(ctx, url)` when pass is empty, resolve the effective timeout, set `httpClient.Timeout`). Then return `factory.CreateRemoteWriter(httpClient, url, user, pass, libtime.NewCurrentDateTime())`. Reject a negative parsed timeout with a wrapped error (mirror `login.go`). To avoid duplicating the ~40-line credential/timeout block, you MAY extract a small shared helper (e.g. `(sf *sharedFlags) resolveCredentialsAndClient(ctx) (url teamvault.Url, user teamvault.User, pass teamvault.Password, httpClient *http.Client, err error)`) used by `buildWriter`; if you extract it, do NOT change the behavior of the existing read path. If extraction risks touching the read path, just inline the resolution in `buildWriter` — do not refactor `buildConnector`/`login.go` in this prompt.

2. **`create` command** — add `createCreateCommand(ctx context.Context, sf *sharedFlags) *cobra.Command` and register it in `NewRootCommand`. Shape:
   - `Use: "create"`, `Short: "Create a new secret in TeamVault"`, `Args: cobra.NoArgs`.
   - Local flags:
     - `--name` (string, required — a create needs a name): `cmd.Flags().StringVar(&name, "name", "", "secret name")`.
     - `--username`, `--url`, `--description` (string metadata, optional).
     - Value flags (mutually exclusive, exactly one required):
       - `--password-stdin` (bool): read the password from stdin to EOF.
       - `--generate` (bool): obtain a server-generated password via `writer.GeneratePassword(ctx)`.
       - `--password` (string): convenience. Its help text MUST name the leak risk, e.g. `"secret password (WARNING: visible in shell history and process list (ps); prefer --password-stdin or --generate)"`.
       - `--file` (string): path to a file whose bytes are base64-encoded into a file secret.
     - `--json` (bool): print `{"key","api_url"}` instead of the bare key.
   - `RunE` logic (in this order, all validation BEFORE any network call):
     1. Determine the value source. Exactly one of `--password-stdin` / `--generate` / `--password` / `--file` must be set. Zero → error naming the missing value source. Two+ → mutual-exclusion error naming the conflicting flags. Both are non-zero exits BEFORE any network call.
     2. Infer content type: `--file` → `ContentTypeFile`; any of `--password-stdin`/`--generate`/`--password` → `ContentTypePassword`.
     3. Resolve the value:
        - `--file`: read the file with `os.ReadFile(path)`. On error (missing/unreadable) return a wrapped non-zero error BEFORE any network call. Put the raw bytes in `CreateSecret.FileContent` (the writer base64-encodes).
        - `--password-stdin`: read stdin to EOF via a stdin reader (see req 4 for the TTY guard + empty-input guard). Reject empty input with a non-zero error BEFORE any network call (mirrors bug-003 empty-password guard). Put the value in `CreateSecret.Password`.
        - `--generate`: build the writer, call `writer.GeneratePassword(ctx)`. If that fails, return a wrapped non-zero error and do NOT attempt the create (generate is server-stateless, so nothing is orphaned). Put the returned password in `CreateSecret.Password`.
        - `--password`: use the flag value directly as `CreateSecret.Password`.
     4. Build the `CreateSecret{ContentType, Name, Username, Url, Description, Password/FileContent}` and call `writer.Create(ctx, secret)`. Wrap errors with `errors.Wrap(ctx, err, "create secret failed")`.
     5. Print output via the shared key-output helper (req 5). Default: the bare key, NO trailing newline (consistent with the read commands' no-newline contract). `--json`: `{"key":"…","api_url":"…"}` single line with trailing newline.

3. **`update <key>` command** — add `createUpdateCommand(ctx, sf)` and register it. Shape:
   - `Use: "update <key>"`, `Short: "Update an existing TeamVault secret"`, `Args: cobra.ExactArgs(1)` (the positional key is required; take it from `args[0]`).
   - Local flags: same metadata flags (`--name`, `--username`, `--url`, `--description`), same value flags (`--password-stdin`, `--generate`, `--password`, `--file`), and `--json`. NONE of the value/metadata flags are required for update (metadata-only update, or even value-only, is allowed).
   - `RunE` logic:
     1. Read the key from `args[0]`; validate non-empty (`teamvault.Key(args[0]).Validate(ctx)`).
     2. Enforce value-source mutual exclusion: at most ONE of the four value flags may be set (zero is allowed for update — metadata-only). Two+ → mutual-exclusion error before any network call.
     3. Build an `UpdateSecret` setting ONLY the fields the user actually passed. Use `cmd.Flags().Changed("name")` etc. to distinguish "flag passed" from "default zero value", and set the corresponding pointer field only when `Changed(...)` is true. This is what makes a metadata-only PATCH omit `secret_data` and a `--description d`-only update send just `description`.
     4. Resolve the value flag the same way as `create` (file read + base64, stdin read + empty guard + TTY guard, generate, or literal `--password`), setting `UpdateSecret.Password` (pointer) or `UpdateSecret.FileContent`.
     5. Call `writer.Update(ctx, key, secret)`. Wrap errors with `errors.Wrap(ctx, err, "update secret failed")`. A server rejection of an immutable content-type change (e.g. `--file` on a password secret) surfaces as the writer's non-2xx error — do NOT swallow it; return it so the exit is non-zero.
     6. Print output via the shared key-output helper (same as create).

4. **Secure stdin reader with TTY + empty guards** — factor a helper used by both commands:
   ```go
   // readPasswordFromStdin reads a password from stdin to EOF. It rejects empty
   // input and does not block forever on an interactive terminal.
   func readPasswordFromStdin(ctx context.Context, in *os.File) (teamvault.Password, error)
   ```
   - If `term.IsTerminal(int(in.Fd()))` is true (interactive TTY, nothing piped), return a non-zero error like `"--password-stdin requires piped input (e.g. echo -n pw | teamvault-cli create --password-stdin); refusing to block on an interactive terminal"` — this satisfies the spec's "must not block forever silently" requirement by erroring instead of hanging.
   - Otherwise read ALL of stdin (`io.ReadAll(in)`), trim a single trailing newline/CR (mirror `login.go`'s `strings.TrimRight(line, "\n\r")` intent — but for stdin, trim only trailing `\n`/`\r` from the read bytes; do NOT trim interior characters).
   - If the resulting password is empty, return a non-zero error naming the empty password (mirrors bug-003). NEVER log the value.
   - For testability, accept the stdin source as a parameter (`*os.File` for prod = `os.Stdin`); the RunE passes `os.Stdin`. Tests can call the underlying command with a pipe (see tests below) — OR structure so the read function is a package var seam you can override. Prefer passing `cmd.InOrStdin()` where possible; if `cmd.InOrStdin()` returns an `io.Reader` you cannot `IsTerminal`, gate the TTY check on `os.Stdin` specifically and read from `cmd.InOrStdin()`. Choose ONE approach and implement it consistently; document the choice in a comment.

5. **Shared key-output helper** — one testable function both commands funnel through, so there is exactly one output code path and no secret ever leaks:
   ```go
   // writeKey prints the created/updated secret's key. Default: the bare key
   // with NO trailing newline. --json: {"key":"…","api_url":"…"} single line.
   func writeKey(ctx context.Context, out io.Writer, key teamvault.Key, apiURL teamvault.ApiUrl, asJSON bool) error
   ```
   Default mode: `fmt.Fprintf(out, "%s", key.String())` (no `\n`). JSON mode: `json.Marshal(map[string]string{"key": key.String(), "api_url": apiURL.String()})` then `fmt.Fprintf(out, "%s\n", encoded)`. NEVER include the password/file value in either branch.

6. **No secret value in any log or on stdout.** Do not pass a password/file value to any `glog`/`slog` call. The only stdout output is via `writeKey` (key + optional api_url).

7. **Tests** — `pkg/cli/create_test.go` and `pkg/cli/update_test.go` (Ginkgo/Gomega), driving the commands with a `mocks.Writer` injected so no network is hit. To inject the mock, add a seam: e.g. make the command factory accept an optional writer-builder override, OR add an unexported package var `newWriter func(...) (teamvault.Writer, error)` defaulting to `sf.buildWriter` that tests can stub. Choose ONE seam and document it. Cover ALL of:

   a. **Command registration** — `NewRootCommand(ctx)` has subcommands `create` and `update` (assert via `rootCmd.Commands()` names, or drive `["--help"]` and grep the output).
   b. **`create --help` shows the leak warning** — capture `["create", "--help"]` output; assert it contains `--password` AND a case-insensitive `history` or `ps` mention.
   c. **Content-type inference** — table over `--password`, `--password-stdin`, `--generate` → assert the `mocks.Writer.CreateArgsForCall(0)` `CreateSecret.ContentType == ContentTypePassword`; `--file` → `ContentTypeFile`.
   d. **No value source** — `["create", "--name", "x"]` returns a non-nil error whose message names the missing value source, and `mocks.Writer.CreateCallCount() == 0` (no network attempted).
   e. **Two value sources** — `["create", "--name", "x", "--generate", "--password", "p"]` returns a mutual-exclusion error and `CreateCallCount() == 0`.
   f. **Update metadata-only omits secret_data** — drive `["update", "K", "--description", "d"]`; assert `mocks.Writer.UpdateArgsForCall(0)` has `Description != nil` (`*Description == "d"`) and `Password == nil` and `FileContent == nil` (no value → no secret_data). Also assert the key argument is `"K"`.
   g. **Update positional key** — `["update", "--help"]` output shows `update <key>`; and `["update", "K", "--description", "d"]` calls `UpdateArgsForCall(0)` with `key == teamvault.Key("K")`.
   h. **`--password-stdin` reads from stdin, not argv** — feed a password via `cmd.SetIn(strings.NewReader("stdin-pw\n"))` (or the chosen stdin seam) and assert `CreateArgsForCall(0).CreateSecret.Password == "stdin-pw"`; assert the value is NOT present in `os.Args`. Note: for the TTY guard, ensure the test's stdin source reports non-TTY (a `strings.Reader`/pipe is non-TTY) so it does not error.
   i. **Empty stdin rejected** — feed `""` on stdin; assert a non-nil error naming the empty password and `CreateCallCount() == 0`.
   j. **TTY guard does not block** — call `readPasswordFromStdin` (or the command) with a non-TTY closed/empty reader and assert it returns PROMPTLY (either the empty-input error or a value) and never blocks. (A unit test on `readPasswordFromStdin` with a `strings.Reader` suffices; do not attempt to allocate a real PTY.)
   k. **`--file` base64** — write a temp file with `os.WriteFile`, drive `["create", "--name", "x", "--file", path]`, assert `CreateArgsForCall(0).CreateSecret.FileContent == <raw bytes>` (the CLI passes raw bytes; base64 happens in the writer — verified in prompt 1). If your design base64-encodes in the CLI instead, assert the encoded form; keep it consistent with prompt 1's writer (prompt 1 base64-encodes inside the writer, so the CLI passes RAW bytes).
   l. **Output is key-only / json** — with the mock returning `(teamvault.Key("NEWKEY"), teamvault.ApiUrl("http://h/api/secrets/NEWKEY/"), nil)`, capture stdout: default mode asserts `buf.String() == "NEWKEY"` (no newline, no password); `--json` mode asserts the output parses to an object with exactly `key` and `api_url` and does NOT contain the input password string.
   m. **No secret logged** (reviewer-checkable + test) — the stdout-equality test in (l) already proves the password is absent from stdout.

   Aim for ≥80% statement coverage of the new command files. Verify with:
   `go test -coverprofile=/tmp/cover.out ./pkg/cli/... && go tool cover -func=/tmp/cover.out | grep -E 'create|update'`.

</requirements>

<constraints>
- Do NOT alter the existing exported `Connector` interface or the read commands, flags, env-var names, config resolution precedence, or `--json` semantics.
- Credentials resolve through the same factory wiring the read path uses (flag → env → config → Keychain). The CLI always uses the configured URL — no special targeting for Lockbox vs TeamVault.
- The four value sources (`--password-stdin`, `--generate`, `--password`, `--file`) are mutually exclusive. On `create` exactly one is required; on `update` zero-or-one (metadata-only allowed).
- `--password <val>` help text MUST name the shell-history/`ps` leak risk and point to `--password-stdin`/`--generate` as safe defaults.
- `--password-stdin` reads from stdin (never argv); empty stdin is rejected with a non-zero exit; an interactive TTY with no piped input must NOT block forever silently — error instead.
- `--file` reads the path and passes raw bytes (base64 happens in the writer from prompt 1); a missing/unreadable file fails with a non-zero exit BEFORE any network call.
- No secret value (stdin password, generated password, `--password`, `--file` bytes) ever appears on stdout or in any `glog`/`slog` line. `create`/`update` print only the key (default) or `{"key","api_url"}` (`--json`).
- `content_type` is immutable after create — an update that would change it surfaces the server's rejection as a non-zero exit; never a silent no-op (do not swallow the writer error).
- Errors wrapped via `github.com/bborbe/errors` (`errors.Wrap`/`errors.Wrapf`/`errors.Errorf`); never `fmt.Errorf`; never `context.Background()` in `pkg/cli` business logic.
- Keep everything in package `pkg/cli` (new files OK: `create.go`, `update.go`, `writer.go`); no per-subcommand sub-packages. `RunE` carries logic and returns `error` (testable without `os.Exit`).
- All exported items keep GoDoc comments per `docs/dod.md`.
- Tests use Ginkgo/Gomega; mocks via Counterfeiter under `pkg/mocks/` (`mocks.Writer`).
- Include the BSD license header block at the top of every new `.go` file.
- Do NOT touch fakevault, scenarios, or docs — that is prompt 3.
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass.
</constraints>

<verification>
- `make test` passes (fast loop).
- `make precommit` exits 0 (final validation).
- `go run . --help 2>&1` lists both `create` and `update`.
- `go run . create --help 2>&1 | grep -c -- '--password'` returns ≥1, and `go run . create --help 2>&1 | grep -Eni 'history|ps '` matches (leak warning present).
- `go run . update --help 2>&1 | grep -c 'update <key>'` returns ≥1.
- `go run . create --name x 2>&1; echo "exit=$?"` prints `exit=1` (no value source; non-zero before any network call) and stderr names the missing value source.
- `go run . create --name x --generate --password p 2>&1; echo "exit=$?"` prints `exit=1` (mutual exclusion).
- `printf '' | go run . create --name x --password-stdin 2>&1; echo "exit=$?"` prints `exit=1` and stderr names the empty password.
- `grep -rn 'fmt.Errorf' pkg/cli/create.go pkg/cli/update.go 2>/dev/null` returns 0 matches.
- `go test ./pkg/cli/...` passes (all create/update cases above).
- `go test -coverprofile=/tmp/cover.out ./pkg/cli/... && go tool cover -func=/tmp/cover.out | grep -E 'create|update'` shows ≥80% for the new files.
</verification>
