---
status: completed
summary: Added `teamvault-cli search <query>` subcommand that calls Connector.Search and prints keys one per line or as JSON array with `--json`
execution_id: sm-teamvault-cli-exec-014-add-search-command
dark-factory-version: dev
created: "2026-07-14T00:00:00Z"
queued: "2026-07-14T17:18:28Z"
started: "2026-07-14T17:19:04Z"
completed: "2026-07-14T17:21:25Z"
---

<summary>
- Adds a new `teamvault-cli search <query>` subcommand that lists the keys of secrets matching a search term.
- Reuses the already-existing `Connector.Search` API (currently only reachable from the Go library, not the CLI).
- Default output: one matching key (hashid) per line, plain text.
- `--json` flag: prints a JSON array of keys, e.g. `["ABC123","DEF456"]`.
- Zero matches is a success, not an error: prints nothing (or `[]` with `--json`) and exits 0.
- A missing query argument produces a usage error.
- Follows the house command style already used by `password`/`info`/`create` (shared connector, `--json`, cobra, Ginkgo tests, counterfeiter mocks).
- Updates the README command reference + examples and adds a CHANGELOG entry.
</summary>

<objective>
Expose the existing `Connector.Search` capability as a first-class CLI command `teamvault-cli search <query>` so users can discover secret keys by name, with plain-line output by default and a `--json` array under `--json`.
</objective>

<context>
Read `CLAUDE.md` at the repo root for project conventions before making changes.

Read these files to match the existing house style and verify signatures — do NOT paraphrase from memory:
- `pkg/connector.go` — the `Connector` interface. `Search(ctx context.Context, name string) ([]Key, error)` is at line 17.
- `pkg/cli/cli.go` — `NewRootCommand` (subcommand registration via `rootCmd.AddCommand(...)`), `SharedFlags`, `(sf *SharedFlags) buildConnector(ctx) (teamvault.Connector, error)`, and the `createInfoCommand` / `createSecretCommand` / `resolveKey` helpers that show the positional-arg + `--json` + shared-connector house style.
- `pkg/cli/writer.go` — the `newWriter` test seam plus `SetNewWriterForTest` / `ResetNewWriterForTest`. You will add a **connector** seam mirroring this exactly so tests can inject a fake connector.
- `pkg/cli/create.go` — house style for a non-secret command (cobra `RunE`, `errors.Wrap`, `writeKey` output helper).
- `pkg/cli/create_test.go` and `pkg/cli/cli_suite_test.go` — external `cli_test` package, Ginkgo/Gomega layout, `mocks.Connector` / `mocks.Writer` usage, command-registration test, `--json` output test.
- `pkg/mocks/connector.go` — the counterfeiter fake. Relevant methods: `SearchReturns([]teamvault.Key, error)`, `SearchArgsForCall(i) (context.Context, string)`, `SearchCallCount() int`.
- `pkg/key.go` — `type Key string` with `func (k Key) String() string`.
- `README.md` — the `## Command reference` table (around line 158) and the usage examples (around line 100-118).
- `CHANGELOG.md` — top of file; versioned sections like `## v5.8.0` with `- feat(cli): …` bullets.

Relevant coding-plugin guides (read the ones you touch; in-container paths):
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-cli-guide.md` — cobra command conventions.
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-testing-guide.md` — Ginkgo/Gomega + counterfeiter.
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md` — `github.com/bborbe/errors` wrapping (no `fmt.Errorf`).
- `/home/node/.claude/plugins/marketplaces/coding/docs/teamvault-conventions.md` — repo conventions.
- `/home/node/.claude/plugins/marketplaces/coding/docs/changelog-guide.md` — changelog entry style.
</context>

<requirements>
1. Add a test seam for the connector, mirroring the `newWriter` seam in `pkg/cli/writer.go`. Add to `pkg/cli/cli.go` (or a small new file `pkg/cli/connector.go` in package `cli`):
   ```go
   // newConnector is a seam that returns a teamvault.Connector. Defaults to the
   // SharedFlags builder but is overridden by tests via SetNewConnectorForTest.
   var newConnector = func(sf *SharedFlags) func(context.Context) (teamvault.Connector, error) {
       return sf.buildConnector
   }

   // SetNewConnectorForTest overrides the connector constructor for tests.
   // Returns a function to reset it.
   func SetNewConnectorForTest(
       f func(sf *SharedFlags) func(context.Context) (teamvault.Connector, error),
   ) func() {
       prev := newConnector
       newConnector = f
       return func() { newConnector = prev }
   }

   // ResetNewConnectorForTest resets the connector seam. (Mirrors ResetNewWriterForTest.)
   func ResetNewConnectorForTest() {}
   ```
   Do NOT change the existing read/config commands (`password`/`username`/`url`/`file`/`info`/`config`) — they keep calling `sf.buildConnector` directly. Only the new `search` command uses `newConnector`.

2. Create `pkg/cli/search.go` (package `cli`, copyright header matching `pkg/cli/create.go` line 1-3) with `createSearchCommand(ctx context.Context, sf *SharedFlags) *cobra.Command`:
   - `Use: "search <query>"`, a `Short` describing it (e.g. `"Search for secrets by name and print matching keys"`).
   - `Args: cobra.ExactArgs(1)` so a missing (or extra) query produces cobra's usage error automatically.
   - In `RunE`:
     - `query := args[0]`.
     - Resolve the connector via the seam: `conn, err := newConnector(sf)(ctx)`; on error `return err` (already wrapped inside `buildConnector`).
     - `keys, err := conn.Search(ctx, query)`; on error `return errors.Wrap(ctx, err, "search failed")`.
     - Read the `--json` flag: `asJSON, _ := cmd.Flags().GetBool("json")`.
     - Call a `writeSearch(ctx, cmd.OutOrStdout(), keys, asJSON)` helper (below) and return its error.
   - Register a `--json` bool flag: `cmd.Flags().Bool("json", false, "print output as a JSON array of keys")`.
   - Return the command.

3. Add the `writeSearch` output helper in `pkg/cli/search.go`:
   ```go
   func writeSearch(ctx context.Context, out io.Writer, keys []teamvault.Key, asJSON bool) error
   ```
   - Default (non-JSON): print one key per line via `Key.String()` with a trailing newline each, e.g. `fmt.Fprintf(out, "%s\n", key.String())`. On zero keys print nothing. Wrap any write error with `errors.Wrapf(ctx, err, "write search result failed")`.
   - `--json`: marshal a JSON array of the key strings. IMPORTANT: build a non-nil slice so an empty result marshals to `[]`, not `null` — `Search` returns a nil slice on zero matches:
     ```go
     ids := make([]string, 0, len(keys))
     for _, k := range keys {
         ids = append(ids, k.String())
     }
     encoded, err := json.Marshal(ids)
     ```
     Wrap marshal errors with `errors.Wrapf(ctx, err, "marshal json failed")`. Write with a trailing newline: `fmt.Fprintf(out, "%s\n", encoded)`; wrap write errors with `errors.Wrapf(ctx, err, "write search result failed")`.
   - Use `github.com/bborbe/errors` for all wrapping — never `fmt.Errorf`. Imports needed: `context`, `encoding/json`, `fmt`, `io`, `github.com/bborbe/errors`, `github.com/spf13/cobra`, and `teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"`.

4. Register the command in `NewRootCommand` (`pkg/cli/cli.go`) alongside the others: add `rootCmd.AddCommand(createSearchCommand(ctx, sf))` next to the existing `createInfoCommand` / `createCreateCommand` / `createUpdateCommand` registrations.

5. Add tests in a new external-package file `pkg/cli/search_test.go` (package `cli_test`; header + imports modeled on `pkg/cli/create_test.go`). Cover:
   - **command registration**: `NewRootCommand(ctx).Commands()` names contain `"search"` (mirror the create_test registration test).
   - **positional query passed to Search**: inject a fake via `cli.SetNewConnectorForTest`, run `search my-query`, assert `fakeConn.SearchCallCount() == 1` and `fakeConn.SearchArgsForCall(0)` second return equals `"my-query"`.
   - **table output**: `fakeConn.SearchReturns([]teamvault.Key{"ABC123","DEF456"}, nil)`; run `search foo`; assert stdout equals `"ABC123\nDEF456\n"`.
   - **`--json` output shape**: same keys, run `search foo --json`; assert `strings.TrimSpace(out)` equals `["ABC123","DEF456"]`.
   - **empty results exit 0, no error**: `fakeConn.SearchReturns(nil, nil)`; run `search foo`; assert `cmd.Execute()` returns nil error and stdout is empty. Add a second case: `search foo --json` on empty result → stdout trimmed equals `[]`.
   - **missing query → usage error**: run `search` with no positional arg; assert `cmd.Execute()` returns a non-nil error (cobra's `ExactArgs(1)` "accepts 1 arg(s), received 0").
   - Test seam usage pattern (mirror create_test.go):
     ```go
     cli.SetNewConnectorForTest(
         func(sf *cli.SharedFlags) func(context.Context) (teamvault.Connector, error) {
             return func(ctx context.Context) (teamvault.Connector, error) {
                 return fakeConn, nil
             }
         },
     )
     defer cli.ResetNewConnectorForTest()
     ```
     Note: the reset is a no-op mirroring the existing writer seam; every test sets the seam itself, so this is safe. Set `os.Setenv("STAGING", "true")` in `BeforeEach` like create_test does, to keep the wiring hermetic even though the seam bypasses the real builder.
   - Each `cmd.SetOut`/`cmd.SetErr` with `bytes.Buffer` and build the command via `cli.NewRootCommand(ctx)` + `cmd.SetArgs([]string{"search", ...})`.

6. Update `README.md`:
   - Add a row to the `## Command reference` table (after the `info` row): `` | `teamvault-cli search <QUERY>` | search secrets by name and print matching keys | ``.
   - Update the note under the table (currently `Add \`--json\` to \`password\`/\`username\`/\`url\`/\`file\`/\`info\` for JSON output.`) to include `search` (JSON array output).
   - Add a short usage example near the other examples (around the `info` example, ~line 107-118):
     ```bash
     teamvault-cli search database
     # AbC123
     # XyZ789

     teamvault-cli search database --json
     # ["AbC123","XyZ789"]
     ```

7. Update `CHANGELOG.md`: add a new `## Unreleased` section at the very top (immediately above `## v5.8.0`) with a bullet:
   `- feat(cli): add a \`search <query>\` subcommand that lists the keys of secrets matching a name search (\`GET /api/secrets/?search=…\`). Prints one key per line by default, or a JSON array of keys with \`--json\`. Zero matches exits 0 (empty output, or \`[]\` with \`--json\`).`
   Match the existing bullet phrasing/scope-prefix style (`feat(cli):`).
</requirements>

<constraints>
- Container-autonomous: only file edits + `make` targets. No `kubectl`, no deploy, no `gh`/PR steps, no cross-repo writes.
- Do NOT commit — dark-factory handles git.
- Do NOT run `go mod vendor` and never pass `-mod=vendor`. This repo does not vendor; `/vendor` is gitignored and the Makefile uses `-mod=mod`.
- Wrap all errors with `github.com/bborbe/errors` (`errors.Wrap`/`errors.Wrapf`/`errors.New`/`errors.Errorf`) — never `fmt.Errorf`.
- No `time.Now()` / `context.Background()` in business logic — the context flows in from `RunE` and the shared `ctx`.
- Tests live in the external `cli_test` package, use Ginkgo/Gomega, and inject the counterfeiter `mocks.Connector`.
- Do NOT modify the existing read/config command wiring; only add the connector seam and the new `search` command consuming it.
- Existing tests must still pass.
</constraints>

<verification>
- `make test` — must pass.
- `make precommit` — must pass (lint, format, changelog check).
- Coverage sanity for the new command (do NOT use `-mod=vendor`):
  ```
  go test -coverprofile=/tmp/cover.out ./pkg/cli/... && go tool cover -func=/tmp/cover.out | grep search
  ```
  Expect `createSearchCommand` and `writeSearch` to appear with non-zero coverage.
</verification>
