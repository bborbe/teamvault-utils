---
status: cancelled
spec: [005-consolidate-cli-into-teamvault-command]
created: "2026-07-09T15:01:03Z"
queued: "2026-07-09T15:16:04Z"
cancelled: "2026-07-11T12:44:19Z"
---

<summary>
- Adds the remaining subcommands to the `teamvault` command: `login`, and a `config` group with `parse` and `generate`.
- `login` reproduces the old `teamvault-login` exactly: no key, resolves url/user/pass via flag → config file → keychain, probes credentials, writes the validated password to the macOS Keychain, and emits ALL status/prompt output on stderr.
- `config parse` reads a template from stdin and writes rendered output to stdout unchanged; `config generate` requires `--source-dir` and `--target-dir`.
- Deletes the entire `cmd/` directory — the seven old binaries are gone; `teamvault <verb>` is the only entry point.
- Rewrites the Makefile `install` target to build the single `teamvault` binary from the repo root.
- Removes the `github.com/bborbe/service` (`libservice.MainCmd`) dependency, which no code references anymore.
- Adds tests: login writes its success message to stderr (stdout empty), and a config-parse round-trip.
</summary>

<objective>
Finish the cobra subcommand tree by adding `login`, `config parse`, and `config generate`, then remove the old surface: delete `cmd/`, rewrite the Makefile `install` target to build the single `teamvault` binary, and drop the now-unused `libservice` dependency.
</objective>

<context>
Read `CLAUDE.md` for project conventions. Runs AFTER prompt 2 (module `/v5`; `main.go` + `pkg/cli` with root, persistent flags, and the four secret-reader subcommands already exist; `sharedFlags` and its `buildConnector(ctx)` helper exist in `pkg/cli`).

Read before changing:
- `cmd/teamvault-login/main.go` — the FULL login implementation to port. Its `Run` resolves url/user/pass manually: reads flag values, then if the config file exists overrides url/user (and pass if empty) from it, then falls back to the keychain for pass; builds an `*http.Client` with a resolved timeout (5s default); then calls `loginFlow(ctx, &termReader{}, os.Stderr, makeConnector, kc, url, user, initialPass)`. The file also defines `loginFlow`, `writeAndReport`, `isAuthError`, the `termReader` type, and the `connectorFactory` type. ALL of this business logic must move into `pkg/cli` unchanged in behavior. Note the two `context.WithTimeout(ctx, 10*time.Second)` per-probe deadlines inside `loginFlow` — keep them.
- `cmd/teamvault-config-parser/main.go` — `config parse`: builds the connector, then `configParser := teamvault.NewConfigParser(conn)`, reads all of stdin (`io.ReadAll(os.Stdin)`), `output, err := configParser.Parse(ctx, content)`, writes `output` to stdout unchanged.
- `cmd/teamvault-config-dir-generator/main.go` — `config generate`: builds the connector, then `teamvault.NewConfigGenerator(teamvault.NewConfigParser(conn)).Generate(ctx, teamvault.SourceDirectory(src), teamvault.TargetDirectory(dst))`. Old struct used required `--source-dir`/`--target-dir` with env `SOURCE_DIR`/`TARGET_DIR`.
- `cmd/teamvault-login/main_test.go` — existing `loginFlow` Ginkgo tests using `mocks.Connector` + `mocks.Keychain` and a `connectorFactory` closure. Port these tests into `pkg/cli` (adjust package + import paths); they are the login regression net.
- `factory/factory.go` — `factory.CreateConnector(httpClient, url, user, pass, staging, cacheEnabled, currentDateTime) teamvault.Connector` (used by login's `makeConnector` closure) and `factory.CreateHttpClient(ctx)`. DO NOT change factory.
- `pkg/cli/cli.go` (from prompt 2) — `sharedFlags`, `buildConnector`, `NewRootCommand`, the four `create<Xxx>Command` factories.
- `Makefile` — the `install` target currently has seven `go build -o $(GOPATH)/bin/teamvault-… cmd/teamvault-…/*` lines to replace.
- `go.mod` — currently requires `github.com/bborbe/service`.

Read these coding guides (in-container paths):
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-cli-guide.md`
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md`
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-testing-guide.md`

Library facts:
- `teamvault.NewKeychain()`, `teamvault.Keychain` interface, `teamvault.ErrKeychainNotSupported`, and `mocks.Keychain` all exist and are unchanged.
- `teamvault.SourceDirectory` / `teamvault.TargetDirectory` are string-based typed wrappers.
- `libtime.ParseDuration(ctx, value) (*libtime.Duration, error)` and `libtime.NewCurrentDateTime()` are available (used by prompt 2's `buildConnector`).
</context>

<requirements>

1. **Port login into `pkg/cli`** (new file `pkg/cli/login.go`, plus `createLoginCommand(ctx, sf *sharedFlags) *cobra.Command`, added to the root in `NewRootCommand`):
   - `Use: "login"`, one-line `Short`, `Args: cobra.NoArgs`. NO `--teamvault-key` flag (login takes no key). It reuses the persistent shared flags for url/user/pass/config/staging/timeout; it ignores `--cache`.
   - Move `loginFlow`, `writeAndReport`, `isAuthError`, the `termReader` type, and the `connectorFactory` type from `cmd/teamvault-login/main.go` into `pkg/cli` UNCHANGED in behavior. Keep the two `context.WithTimeout(ctx, 10*time.Second)` per-probe deadlines.
   - Port the resolution logic from the old `Run` into `createLoginCommand`'s `RunE`, reading url/user/pass/config/staging/timeout from `sf`:
     - Start with `teamvault.Url(sf.url)`, `teamvault.User(sf.user)`, `teamvault.Password(sf.pass)`.
     - If `teamvault.TeamvaultConfigPath(sf.configPath).Exists()`, parse it and override url/user (and pass if the flag pass is empty) — identical to the old logic.
     - Validate url and user are non-empty (same error messages as the old `Run`).
     - If pass is still empty, read it from the keychain via `teamvault.NewKeychain().ReadPassword(ctx, resolvedURL)`.
     - Build the `*http.Client` via `factory.CreateHttpClient(ctx)`, resolve the timeout (parse `sf.timeout` via `libtime.ParseDuration` when non-empty, else default; `if timeout == 0 { timeout = 5*time.Second }`), set `httpClient.Timeout = timeout`.
     - Build the `makeConnector` closure over `factory.CreateConnector(...)` exactly as today (staging from `sf.staging`, cache `false`).
   - **Write ALL user-facing output to `cmd.ErrOrStderr()`, never stdout.** The old code passed `os.Stderr` into `loginFlow`; pass `cmd.ErrOrStderr()` instead so tests can capture it. `loginFlow`/`writeAndReport` already take an `errOut io.Writer` — thread `cmd.ErrOrStderr()` through. The `termReader` reads the typed password from the real terminal (`os.Stdin`) as today.
   - On a locked/failing keychain write, `writeAndReport` already returns a wrapped "store password in keychain … try unlocking your Keychain" error — keep that path (Failure Modes: "Keychain locked during login").

2. **Add the `config` command group** (new file `pkg/cli/config.go`), added to the root in `NewRootCommand`:
   - Parent: `&cobra.Command{Use: "config", Short: "Configuration templating commands"}` with no `RunE` of its own.
   - **`config parse`** via `createConfigParseCommand(ctx, sf *sharedFlags) *cobra.Command`:
     - `Use: "parse"`, `Args: cobra.NoArgs`.
     - `RunE`: `conn, err := sf.buildConnector(ctx)`; read all of `cmd.InOrStdin()` with `io.ReadAll`; `output, err := teamvault.NewConfigParser(conn).Parse(ctx, content)`; write `output` to `cmd.OutOrStdout()` unchanged. Wrap errors with `errors.Wrapf(ctx, err, "parse config failed")` etc. Use `cmd.InOrStdin()`/`cmd.OutOrStdout()` (not `os.Stdin`/`os.Stdout`) so the round-trip test can inject buffers.
   - **`config generate`** via `createConfigGenerateCommand(ctx, sf *sharedFlags) *cobra.Command`:
     - `Use: "generate"`, `Args: cobra.NoArgs`.
     - Two LOCAL required flags:
       ```go
       var sourceDir, targetDir string
       cmd.Flags().StringVar(&sourceDir, "source-dir", "", "source directory")
       cmd.Flags().StringVar(&targetDir, "target-dir", "", "target directory")
       _ = cmd.MarkFlagRequired("source-dir")
       _ = cmd.MarkFlagRequired("target-dir")
       ```
     - `RunE`: `conn, err := sf.buildConnector(ctx)`; `gen := teamvault.NewConfigGenerator(teamvault.NewConfigParser(conn))`; `gen.Generate(ctx, teamvault.SourceDirectory(sourceDir), teamvault.TargetDirectory(targetDir))`; wrap error with `errors.Wrapf(ctx, err, "generate failed")`.
     - Decision: `--source-dir`/`--target-dir` are required flags via `MarkFlagRequired`, with NO env seeding; the old `SOURCE_DIR`/`TARGET_DIR` env fallback is intentionally dropped. Rationale: the spec's AC and Failure Modes mandate these two be required flags, and `SOURCE_DIR`/`TARGET_DIR` are not part of the seven-flag shared env contract (which covers only url/user/pass/config/staging/timeout/cache). Do not add env defaulting for these two flags.

3. **Delete the `cmd/` directory entirely.** `rm -rf cmd`. After this, `test ! -d cmd` must succeed. All login/parser/generator business logic now lives in `pkg/cli`.

4. **Rewrite the Makefile `install` target** to build the single binary. Replace the seven `go build -o $(GOPATH)/bin/teamvault-… cmd/teamvault-…/*` lines with exactly one:
   ```
   .PHONY: install
   install:
   	go build -o $(GOPATH)/bin/teamvault .
   ```
   (Keep the tab indentation Make requires.) There must be no remaining `cmd/teamvault-` reference anywhere in the Makefile.

5. **Remove the now-unused `libservice` dependency.** No code references `libservice.MainCmd` or imports `github.com/bborbe/service` anymore (it lived only in the deleted `cmd/` files). Run `go mod tidy` so `github.com/bborbe/service` drops out of `go.mod`/`go.sum`. `grep -rn 'libservice.MainCmd' --include='*.go' .` and `grep -rn 'bborbe/service' go.mod` must both return 0.

6. **Tests** (extend `pkg/cli` tests; Ginkgo/Gomega):
   - **Port the `loginFlow` tests** from `cmd/teamvault-login/main_test.go` into `pkg/cli` (e.g. `pkg/cli/login_test.go`), adjusting the import paths to `.../v5`. These exercise `loginFlow` directly with `mocks.Connector` + `mocks.Keychain` and a `connectorFactory` closure — the login regression net.
     - DELETE the `It("Compiles")` spec that calls `gexec.Build("github.com/bborbe/teamvault-utils/v4/cmd/teamvault-login")`. That path is `/v4` inside a `/v5` module AND names the `cmd/teamvault-login` directory this prompt deletes, so it would break `make precommit`. Port ONLY the `loginFlow` specs (and their `BeforeEach` setup); drop the compile smoke-test spec entirely.
     - Make the ported file an INTERNAL test in `package cli` (NOT `package cli_test`), because `loginFlow`, `writeAndReport`, `isAuthError`, `termReader`, and `connectorFactory` are unexported and only reachable from inside the package.
   - **Login writes to stderr, not stdout.** Add a test that runs the `login` command path (or `loginFlow` with a `bytes.Buffer` as `errOut` and asserts stdout stays empty) confirming the success message lands on stderr and stdout is empty. If exercising the full `login` cobra command, capture via `cmd.SetOut(outBuf)` / `cmd.SetErr(errBuf)` and assert `outBuf.Len() == 0` while `errBuf` contains the success message.
   - **`config parse` round-trip.** Build a `config parse` command wired to a `sharedFlags` whose `buildConnector` yields a `mocks.Connector` (inject the mock — e.g. pipe a template with no `[[…]]` substitutions so the connector is not called, OR stub the connector's methods), set `cmd.SetIn(bytes.NewBufferString(template))` and `cmd.SetOut(outBuf)`, execute, and assert `outBuf.String()` equals the expected rendered output. A template containing no TeamVault placeholders round-trips unchanged, which is sufficient to prove stdin→stdout wiring without a live connector. If `buildConnector` cannot be injected cleanly, add a minimal seam (e.g. a package-private connector-builder function variable defaulting to the real one) rather than reaching the network in a unit test.

</requirements>

<constraints>
- The root-package public API and `factory.*` are UNCHANGED — only `pkg/cli` grows and `cmd/`/Makefile/go.mod shrink.
- `login`: no `--teamvault-key`; resolution precedence flag → config file → keychain preserved exactly; the same validation error messages as the old `Run`; ALL user-facing output on stderr; the two 10s per-probe `context.WithTimeout` deadlines unchanged.
- `config generate`: `--source-dir` and `--target-dir` are required (see requirement 2 open question re env).
- Flag/env names for the shared flags are preserved exactly (unchanged from prompt 2).
- Errors wrapped via `github.com/bborbe/errors`; never `fmt.Errorf`, never `errors.Wrapf(ctx, nil, …)`.
- cobra + pflag only; NO stdlib `flag`, NO glog, NO `libservice` in `pkg/cli`. `slog` to stderr for any logging.
- Subcommand run logic returns `error` and is testable without `os.Exit`.
- All exported items keep GoDoc comments per `docs/dod.md`.
- Tests use Ginkgo/Gomega; mocks via Counterfeiter under `mocks/`.
- Do NOT commit — dark-factory handles git.
</constraints>

<verification>
- `make precommit` exits 0.
- `test ! -d cmd` succeeds.
- `go run . --help 2>&1` lists `login`, `config` (in addition to `password`/`username`/`url`/`file`); `go run . config --help 2>&1` lists `parse` and `generate`.
- `go run . login --help 2>&1` does NOT list `--teamvault-key`.
- `go run . config generate --help 2>&1` shows `--source-dir` and `--target-dir`; running `go run . config generate` with neither exits non-zero naming the missing flag(s).
- `grep -rn 'cobra.Command' pkg/cli/ | wc -l` returns ≥8 (root + 7 subcommands + config parent).
- `grep -rn 'libservice.MainCmd' --include='*.go' .` returns 0; `grep -rn 'bborbe/service' go.mod` returns 0.
- `grep -c 'cmd/teamvault-' Makefile` returns 0; `grep -c 'go build -o $(GOPATH)/bin/teamvault ' Makefile` returns ≥1.
- `go test ./pkg/cli/...` passes (login-to-stderr, ported loginFlow tests, config-parse round-trip).
- `make install` produces a `teamvault` binary: `test -x $(go env GOPATH)/bin/teamvault`.
</verification>
