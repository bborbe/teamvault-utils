---
status: completed
spec: [005-consolidate-cli-into-teamvault-command]
summary: Added cobra CLI skeleton at cmd/teamvault/ with Execute/Run pattern, seven shared env-seeded persistent flags, and four secret-reader subcommands (password/username/url/file) that print without trailing newline
execution_id: teamvault-utils-consolidate-cli-exec-011-spec-005-cli-root-and-secret-readers
dark-factory-version: v0.191.0
created: "2026-07-09T15:01:03Z"
queued: "2026-07-09T15:16:04Z"
started: "2026-07-09T15:19:51Z"
completed: "2026-07-09T15:27:08Z"
---

<summary>
- Introduces a single `teamvault` command built from a thin `main.go` that delegates to `pkg/cli.Execute()`.
- Builds the cobra command tree root carrying the seven shared flags (url, user, pass, config, staging, timeout, cache) as persistent flags, each falling back to its existing `TEAMVAULT_*`/`STAGING`/`CACHE` env var when the flag is absent — preserving the current env contract.
- Adds the four secret-reader subcommands `password`, `username`, `url`, `file`, each taking a required `--teamvault-key` and reusing the existing factory + connector unchanged.
- Fixes the documented trailing-newline bug: these four subcommands now print the resolved value with NO trailing newline, so `curl -u` basic-auth stops breaking.
- `teamvault --help` and `teamvault password --help` print only the tool's own flags — no Ginkgo/glog leakage — because cobra uses a private flag set.
- The seven old `cmd/teamvault-*` binaries are left in place and building; `login` and `config` subcommands plus the `cmd/` removal land in the next prompt.
- Adds unit tests: env-var seeding for all seven shared flags, the no-newline assertion via a mocked connector, and a clean-`--help` check.
</summary>

<objective>
Create the cobra CLI skeleton — `main.go` delegating to `pkg/cli.Execute()`, a root command with the seven shared env-seeded persistent flags, and the four secret-reader subcommands (`password`/`username`/`url`/`file`) that reuse `factory.CreateConnectorWithConfigAndTimeout` and print their result with no trailing newline. The old `cmd/` binaries stay in place this prompt; they are removed in prompt 3.
</objective>

<context>
Read `CLAUDE.md` for project conventions. This prompt runs AFTER prompt 1 (module is already `github.com/bborbe/teamvault-utils/v5`).

Read before changing:
- `cmd/teamvault-password/main.go` — canonical secret-reader: builds `httpClient` via `factory.CreateHttpClient(ctx)`, then `factory.CreateConnectorWithConfigAndTimeout(...)`, then `conn.Password(ctx, teamvault.Key(...))`, then `fmt.Printf("%v\n", result)`. The `\n` here is the bug to drop.
- `cmd/teamvault-username/main.go`, `cmd/teamvault-url/main.go`, `cmd/teamvault-file/main.go` — identical shape, calling `conn.User` / `conn.Url` / `conn.File` respectively.
- `factory/factory.go` — the connector factory. DO NOT change it. Signature to call:
  ```go
  factory.CreateConnectorWithConfigAndTimeout(
      ctx context.Context,
      httpClient *http.Client,
      configPath teamvault.TeamvaultConfigPath,
      apiURL teamvault.Url,
      apiUser teamvault.User,
      apiPassword teamvault.Password,
      staging teamvault.Staging,
      cacheEnabled bool,
      currentDateTime libtime.CurrentDateTime,
      keychain teamvault.Keychain,
      cliTimeout libtime.Duration,
  ) (teamvault.Connector, error)
  ```
  and `factory.CreateHttpClient(ctx context.Context) (*http.Client, error)`.
- `connector.go` — `teamvault.Connector` interface: `Password(ctx, Key) (Password, error)`, `User(ctx, Key) (User, error)`, `Url(ctx, Key) (Url, error)`, `File(ctx, Key) (File, error)`, `Search(ctx, string) ([]Key, error)`.
- Typed wrappers (all string-based): `teamvault.Url`, `teamvault.User`, `teamvault.Password`, `teamvault.Key`, `teamvault.File`, `teamvault.Staging` (bool-based), `teamvault.TeamvaultConfigPath`, `teamvault.NewKeychain()`.
- `mocks/connector.go` — Counterfeiter `mocks.Connector` (fake for `teamvault.Connector`), used for the no-newline test.
- `github.com/bborbe/vault-cli` at `pkg/cli/cli.go` is the reference cobra implementation in this ecosystem: note `NewRootCommand(ctx)` building `&cobra.Command{Use, Short, SilenceUsage: true}`, `PersistentFlags().StringVar(...)`, `create<Xxx>Command(ctx, …) *cobra.Command` factories, `RunE` returning `error`, and `Execute()` owning `context.Background()`. (Reference only — vault-cli is a different repo; do not import from it.)

Read these coding guides (in-container paths):
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-cli-guide.md` — Multi-Command Binary Pattern, the `Execute()`/`Run()` split, signal handling, `SilenceUsage: true`, `MarkFlagRequired`, `slog` not glog.
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-package-layout-guide.md`
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-testing-guide.md`

Library facts (verified against the vendored source):
- cobra/pflag is NOT yet a dependency. Add it: `go get github.com/spf13/cobra@latest` then `go mod tidy`.
- `libtime.Duration` (`github.com/bborbe/time`) is `type Duration stdtime.Duration`; `.Duration()` returns the stdlib duration. To parse a string into it use `libtime.ParseDuration(ctx context.Context, value interface{}) (*libtime.Duration, error)`. `libtime.CurrentDateTime` is created with `libtime.NewCurrentDateTime()`.
- The factory validates and defaults the timeout: a zero `libtime.Duration` means "use default"; a negative value is rejected by the factory (do NOT re-validate in the CLI).
</context>

<requirements>

1. **Add cobra.** `go get github.com/spf13/cobra@latest` and `go mod tidy`. cobra pulls in `github.com/spf13/pflag` transitively.

2. **Create `main.go` at the repo root** — a thin delegator, GoDoc-commented:
   ```go
   package main

   import "github.com/bborbe/teamvault-utils/v5/pkg/cli"

   func main() {
       cli.Execute()
   }
   ```
   Include the standard BSD license header (copy the header block from any existing `.go` file in the repo).

3. **Create `pkg/cli/cli.go`** with `Execute()`, `NewRootCommand(ctx) *cobra.Command`, and `Run(ctx, args) error`, following the go-cli-guide Multi-Command pattern:
   - `Execute()` owns the sole `context.Background()`, sets up signal handling (`SIGINT`/`SIGTERM` → cancel), configures `slog` to stderr, calls `Run(ctx, os.Args[1:])`, and on error prints `Error: %v` to stderr and `os.Exit(1)`. Use the exact `Execute()` shape shown in the go-cli-guide "Multi-Command Binary Pattern" section (context + signal goroutine + `Run` + `os.Exit(1)`; the `Execute()`/`Run()` split is identical for single- and multi-command binaries).
   - `Run(ctx, args) error` builds `NewRootCommand(ctx)`, calls `rootCmd.SetArgs(args)`, and returns `rootCmd.ExecuteContext(ctx)`.
   - `NewRootCommand(ctx)` returns `&cobra.Command{Use: "teamvault", Short: "…", SilenceUsage: true}` with the persistent flags from step 4 registered and the subcommands from step 5 added.
   - Use `log/slog` to stderr for any logging. Do NOT import `github.com/golang/glog`, `github.com/bborbe/service`, or stdlib `flag` in `pkg/cli`.

4. **Register the seven shared flags as `PersistentFlags` on the root command, each defaulting from its env var.** Hold the bound values in a struct so subcommand factories can read them — define in `pkg/cli`:
   ```go
   type sharedFlags struct {
       url        string
       user       string
       pass       string
       configPath string
       staging    bool
       cache      bool
       timeout    string // raw string; parsed to libtime.Duration when building the connector
   }
   ```
   Register with env-derived defaults so an unset flag falls back to the env var and an explicit flag overrides it (flag > env):
   ```go
   sf := &sharedFlags{}
   pf := rootCmd.PersistentFlags()
   pf.StringVar(&sf.url,        "teamvault-url",     os.Getenv("TEAMVAULT_URL"),     "teamvault url")
   pf.StringVar(&sf.user,       "teamvault-user",    os.Getenv("TEAMVAULT_USER"),    "teamvault user")
   pf.StringVar(&sf.pass,       "teamvault-pass",    os.Getenv("TEAMVAULT_PASS"),    "teamvault password")
   pf.StringVar(&sf.configPath, "teamvault-config",  os.Getenv("TEAMVAULT_CONFIG"),  "teamvault config file path")
   pf.BoolVar(&sf.staging,      "staging",           envBool("STAGING"),             "staging status")
   pf.BoolVar(&sf.cache,        "cache",             envBool("CACHE"),               "enable teamvault secret cache")
   pf.StringVar(&sf.timeout,    "teamvault-timeout", os.Getenv("TEAMVAULT_TIMEOUT"), "HTTP request timeout for TeamVault API calls (e.g. 5s, 30s); 0 = default 5s")
   ```
   Add a helper `func envBool(name string) bool { return os.Getenv(name) == "true" }` (matches the old libargument `default:"false"` + `env` behavior: env `"true"` → true, anything else → false). Never log `sf.pass`.

   NOTE: because defaults are captured from the environment at `NewRootCommand` call time, tests that seed env vars must build a fresh root command per case (see tests below).

5. **Add the four secret-reader subcommands** via `create<Xxx>Command(ctx, sf *sharedFlags) *cobra.Command` factories, all added to the root command in `NewRootCommand`. Each:
   - Has `Use: "password"` / `"username"` / `"url"` / `"file"`, a one-line `Short`, `Args: cobra.NoArgs`.
   - Declares a LOCAL required flag `--teamvault-key` (NOT persistent — it must not appear on other subcommands), default empty, enforced by cobra's `MarkFlagRequired`:
     ```go
     var key string
     cmd.Flags().StringVar(&key, "teamvault-key", "", "teamvault key")
     _ = cmd.MarkFlagRequired("teamvault-key")
     ```
     This matches the spec's Failure Mode exactly: a missing `--teamvault-key` yields a non-zero exit with `required flag(s) "teamvault-key" not set` and no network call. `TEAMVAULT_KEY` is NOT one of the seven env-preserved shared flags (the env contract covers only url/user/pass/config/staging/timeout/cache), so it is intentionally NOT seeded from the environment here — `--teamvault-key` must be passed explicitly.
   - In `RunE`: build the connector via the shared helper from step 6, call the matching connector method, and write the result to `cmd.OutOrStdout()` through the shared `writeSecret` helper from req 7 (NO trailing newline):
     ```go
     result, err := conn.Password(ctx, teamvault.Key(key))
     if err != nil {
         return errors.Wrapf(ctx, err, "get password failed")
     }
     return writeSecret(cmd.OutOrStdout(), result)
     ```
     The inline `fmt.Fprintf(...)` shown in req 7 is illustrative of what `writeSecret` does — do NOT inline a second print path here; all four secret readers funnel their output through the single `writeSecret` helper so there is exactly one output code path and no trailing newline. `username` → `conn.User` / "get user failed"; `url` → `conn.Url` / "get url failed"; `file` → `conn.File` / "get file failed".

6. **Add a shared connector-builder helper** in `pkg/cli` that the four subcommands (and, in prompt 3, `config`) reuse:
   ```go
   func (sf *sharedFlags) buildConnector(ctx context.Context) (teamvault.Connector, error) {
       httpClient, err := factory.CreateHttpClient(ctx)
       if err != nil {
           return nil, errors.Wrapf(ctx, err, "create httpClient failed")
       }
       var timeout libtime.Duration
       if sf.timeout != "" {
           d, err := libtime.ParseDuration(ctx, sf.timeout)
           if err != nil {
               return nil, errors.Wrapf(ctx, err, "parse teamvault-timeout %q failed", sf.timeout)
           }
           timeout = *d
       }
       conn, err := factory.CreateConnectorWithConfigAndTimeout(
           ctx,
           httpClient,
           teamvault.TeamvaultConfigPath(sf.configPath),
           teamvault.Url(sf.url),
           teamvault.User(sf.user),
           teamvault.Password(sf.pass),
           teamvault.Staging(sf.staging),
           sf.cache,
           libtime.NewCurrentDateTime(),
           teamvault.NewKeychain(),
           timeout,
       )
       if err != nil {
           return nil, errors.Wrapf(ctx, err, "create connector failed")
       }
       return conn, nil
   }
   ```
   A negative `--teamvault-timeout` (e.g. `-1s`) parses successfully here and is rejected by the factory before any network call — preserving scenario 005 behavior. Do NOT add CLI-side timeout validation.

7. **Extract the value-print step into a testable helper** so the no-newline behavior can be unit-tested with a mocked connector without hitting the real factory. For example:
   ```go
   func writeSecret(out io.Writer, value fmt.Stringer) error {
       _, err := fmt.Fprintf(out, "%v", value)
       return err
   }
   ```
   (Or a small function per field.) The four `RunE`s call this helper. The unit test in step 8b targets it directly with a `mocks.Connector`-sourced value.

8. **Tests** (`pkg/cli/cli_test.go`, Ginkgo/Gomega). Add a suite bootstrap file if none exists (`pkg/cli/cli_suite_test.go` with `RegisterFailHandler(Fail)` + `RunSpecs`). Cover:

   a. **Env-var seeding — table-driven over all seven shared flags.** For each `(envName, flagName, envValue)` pair — `TEAMVAULT_URL`→`teamvault-url`, `TEAMVAULT_USER`→`teamvault-user`, `TEAMVAULT_PASS`→`teamvault-pass`, `TEAMVAULT_CONFIG`→`teamvault-config`, `STAGING`→`staging` (value `"true"`), `TEAMVAULT_TIMEOUT`→`teamvault-timeout`, `CACHE`→`cache` (value `"true"`) — set the env var, build a fresh `NewRootCommand(ctx)`, and assert `rootCmd.PersistentFlags().Lookup(flagName).Value.String()` equals the expected resolved value (the env value; for bool flags `"true"`). Use `os.Setenv`/`os.Unsetenv` (or `GinkgoT().Setenv`) and clean up between cases so pairs don't leak.

   b. **Flag beats env (precedence).** Set `TEAMVAULT_URL=from-env`, build a fresh `NewRootCommand(ctx)`, then call `rootCmd.PersistentFlags().Parse([]string{"--teamvault-url=from-flag"})` directly (this binds the persistent flags WITHOUT invoking any `RunE`, so no connector is built and no network call happens). Assert `rootCmd.PersistentFlags().Lookup("teamvault-url").Value.String() == "from-flag"` (the explicit flag overrides the env default). Unset `TEAMVAULT_URL` afterward.

   c. **No trailing newline.** Using a `mocks.Connector` returning `teamvault.Password("secret")`, invoke the password read path (via the `writeSecret` helper or by wiring the mock connector into a `RunE` you can exercise) capturing into a `bytes.Buffer`. Assert the buffer equals exactly `"secret"` — `buf.String() == "secret"` with no `\n`. Do the same shape for `username`/`url`/`file` (at least `password` is mandatory; cover all four if cheap).

   d. **Clean `--help` — no test-runner leakage.** Build `NewRootCommand(ctx)`, set args `["--help"]` and also `["password", "--help"]`, execute with output captured via `cmd.SetOut`/`SetErr` into buffers, and assert the captured help text contains no case-insensitive `ginkgo` substring. (This proves the private pflag set does not leak global flags.)

   e. **Missing `--teamvault-key` is rejected before any network call (headline failure mode).** Build `NewRootCommand(ctx)`, `rootCmd.SetArgs([]string{"password"})` (no `--teamvault-key`), capture output via `cmd.SetOut`/`SetErr`, and call `rootCmd.Execute()`. Assert the returned error is non-nil and its message contains `required flag(s) "teamvault-key" not set`. Also assert NO connector/network call occurred — either inject a `mocks.Connector`-backed connector builder and assert its call count is 0, or structure the test so `RunE`/`buildConnector` is never reached (cobra enforces required flags before invoking `RunE`, so a plain assertion that the error is the required-flag error — not a connector/factory error — is sufficient). This is the spec's headline Failure Mode and MUST be covered by a test, not only implemented.

9. **Leave `cmd/` untouched.** The seven old binaries must still build after this prompt. Do NOT delete or edit `cmd/` here.

</requirements>

<constraints>
- The root-package public API (`Connector`, `factory.Create*`, typed wrappers, `teamvault.NewKeychain`) MUST NOT change. Only `main.go` + `pkg/cli` are added.
- Every flag name and env-var name is preserved exactly: `--teamvault-url`/`TEAMVAULT_URL`, `--teamvault-user`/`TEAMVAULT_USER`, `--teamvault-pass`/`TEAMVAULT_PASS`, `--teamvault-config`/`TEAMVAULT_CONFIG`, `--staging`/`STAGING`, `--teamvault-timeout`/`TEAMVAULT_TIMEOUT`, `--cache`/`CACHE`, `--teamvault-key`/`TEAMVAULT_KEY`.
- Precedence preserved: explicit flag > env > (config-file/keychain, handled inside the factory).
- `password`/`username`/`url`/`file` print the value with NO trailing newline. No `fmt.Printf("%v\n", …)` anywhere in `pkg/cli`.
- cobra + pflag only; NO stdlib `flag`, NO glog in `pkg/cli`. Use `slog` to stderr.
- `--teamvault-pass` is never logged.
- Errors wrapped via `github.com/bborbe/errors` (`errors.Wrapf`/`errors.Errorf`); never `fmt.Errorf`, never `errors.Wrapf(ctx, nil, …)`.
- Layout: flat `pkg/cli/` — one `create<Xxx>Command` factory per subcommand, no per-subcommand sub-packages. `Execute()` owns the sole `context.Background()` + signal handling; `RunE` carries business logic and returns `error` (testable without `os.Exit`).
- All exported items keep GoDoc comments per `docs/dod.md`.
- `pkg/cli` must not import test-only packages at package scope (no re-introducing global-flag pollution).
- Tests use Ginkgo/Gomega; mocks via Counterfeiter under `mocks/`.
- Do NOT commit — dark-factory handles git.
- Existing `cmd/` tests must still pass.
</constraints>

<verification>
- `make precommit` exits 0.
- `test -f main.go` succeeds and `grep -c 'cli.Execute()' main.go` returns ≥1.
- `grep -rn 'cobra.Command' pkg/cli/ | wc -l` returns ≥5 (root + 4 secret readers) and `grep -rn 'SilenceUsage: *true' pkg/cli/` returns ≥1.
- `go run . --help 2>&1` lists `password`, `username`, `url`, `file`.
- `go run . password --help 2>&1` shows `--teamvault-url`, `--teamvault-user`, `--teamvault-pass`, `--teamvault-config`, `--staging`, `--teamvault-timeout`, `--cache`, and `--teamvault-key`.
- `go run . --help 2>&1 | grep -ci ginkgo` returns 0 and `go run . password --help 2>&1 | grep -ci ginkgo` returns 0.
- `grep -rn 'Printf("%v\\n"' pkg/cli/` returns 0 matches.
- `go test ./pkg/cli/...` passes (env seeding, precedence, no-newline, clean-help).
- The seven binaries still build: `go build ./cmd/teamvault-password ./cmd/teamvault-username ./cmd/teamvault-url ./cmd/teamvault-file ./cmd/teamvault-config-parser ./cmd/teamvault-config-dir-generator ./cmd/teamvault-login`.
</verification>
