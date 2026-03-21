---
status: completed
summary: Enabled golangci-lint in Makefile check target, updated .golangci.yml to standard bborbe config, and fixed all lint violations including depguard, bodyclose, forcetypeassert, staticcheck, and unparam issues
container: teamvault-utils-002-enable-lint
dark-factory-version: v0.59.5-dirty
created: "2026-03-21T12:00:00Z"
queued: "2026-03-21T10:12:35Z"
started: "2026-03-21T10:23:56Z"
completed: "2026-03-21T10:34:19Z"
---

<summary>
- The `check` Makefile target includes `lint` as a dependency
- The TODO comment about enabling lint is removed from the Makefile
- The `.golangci.yml` config is updated to match the standard v2 config used across bborbe repos
- All golangci-lint violations are fixed in Go source files
- `make precommit` passes cleanly
</summary>

<objective>
Enable the golangci-lint linter in the `check` Makefile target, update `.golangci.yml` to the standard bborbe config, and fix all resulting lint violations so the project passes `make precommit`.
</objective>

<context>
Read CLAUDE.md for project conventions.

Read these files before making changes:
- `Makefile` — current check target with lint commented out
- `.golangci.yml` — current lint config (has typos, missing linters)

The current `.golangci.yml` has typos: `"github.com/pkg/erros"` and `"github.com/bborbe/erros"` (missing the second 'r' in "errors"). It also lacks several linters and settings present in the standard config.

Reference standard `.golangci.yml` config to adopt:
```yaml
version: "2"

run:
  timeout: 5m
  tests: true

linters:
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - revive
    - gosec
    - gocyclo
    - depguard
    - dupl
    - nestif
    - errname
    - unparam
    - bodyclose
    - forcetypeassert
    - asasalint
    - prealloc
  settings:
    depguard:
      rules:
        Main:
          deny:
            - pkg: "github.com/pkg/errors"
              desc: "use github.com/bborbe/errors instead"
            - pkg: "github.com/bborbe/argument"
              desc: "use github.com/bborbe/argument/v2 instead"
            - pkg: "golang.org/x/net/context"
              desc: "use context from standard library instead"
            - pkg: "golang.org/x/lint/golint"
              desc: "deprecated, use revive or staticcheck instead"
            - pkg: "io/ioutil"
              desc: "deprecated since Go 1.16, use io and os packages instead"
    funlen:
      lines: 80
      statements: 50
    gocognit:
      min-complexity: 20
    nestif:
      min-complexity: 4
    maintidx:
      min-maintainability-index: 20
  exclusions:
    presets:
      - comments
      - std-error-handling
      - common-false-positives
    rules:
      - linters:
          - staticcheck
        text: "SA1019"
      - linters:
          - revive
        path: "_test\\.go$"
        text: "dot-imports"
      - linters:
          - revive
        text: "unused-parameter"
      - linters:
          - revive
        text: "exported"
      - linters:
          - dupl
        path: "_test\\.go$"
      - linters:
          - unparam
        path: "_test\\.go$"
      - linters:
          - dupl
        path: "-test-suite\\.go$"
      - linters:
          - revive
        path: "-test-suite\\.go$"
        text: "dot-imports"

formatters:
  enable:
    - gofmt
    - goimports
```
</context>

<requirements>
1. Update the `Makefile`:
   - Change the `check:` target line from:
     ```
     check: vet errcheck vulncheck osv-scanner gosec trivy
     ```
     to:
     ```
     check: lint vet errcheck vulncheck osv-scanner gosec trivy
     ```
   - Remove the two lines above it (the TODO comment and the commented-out check line):
     ```
     # TODO: enable lint
     # check: lint vet errcheck vulncheck osv-scanner gosec trivy
     ```

2. Replace the contents of `.golangci.yml` with the standard config from `the reference standard config shown in the context section`, but adapt the exclusion rules for this project:
   - Keep the `errname` exclusion text pattern only if this project has custom error types matching that pattern; otherwise adjust to match actual error type names in this repo
   - Keep `revive` exclusions for `dot-imports`, `unused-parameter`, and `exported`
   - Keep `dupl` and `unparam` exclusions for test files
   - Keep `staticcheck` SA1019 exclusion
   - Fix the depguard typos: `"github.com/pkg/erros"` -> `"github.com/pkg/errors"` and `"github.com/bborbe/erros"` -> `"github.com/bborbe/errors"`

3. Run `go run -mod=mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config .golangci.yml ./...` to identify all lint violations.

4. Fix all lint violations in the Go source files. Common fixes include:
   - Adding error checks for unchecked return values
   - Removing unused parameters or variables
   - Fixing type assertion safety (use comma-ok pattern)
   - Reducing function complexity / nesting
   - Closing HTTP response bodies
   - Preallocating slices where flagged

5. If a lint violation cannot be reasonably fixed (e.g., it's a false positive or fixing it would require a major refactor), add a targeted exclusion rule to `.golangci.yml` rather than using `//nolint` comments. Only use `//nolint` as a last resort and always include the linter name: `//nolint:lintername // reason`.

6. Run `make format` after fixing violations to ensure formatting is consistent.

7. Run `make precommit` to verify everything passes end-to-end.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git
- Do NOT change functionality — only fix lint violations and update config
- Do NOT add `//nolint` comments without specifying the linter name and a reason
- Do NOT remove or weaken existing exclusion rules that are still needed
- Existing tests must still pass
- Minimize code changes — prefer the simplest fix for each violation
</constraints>

<verification>
Run `make lint` — must pass with exit code 0.
Run `make precommit` — must pass with exit code 0.
Verify the Makefile no longer contains `# TODO: enable lint` or the commented-out check line.
Verify `.golangci.yml` has correct spelling for `github.com/pkg/errors` and `github.com/bborbe/errors`.
</verification>
