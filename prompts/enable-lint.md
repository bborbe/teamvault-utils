---
status: inbox
---

<summary>
- Linting runs as part of `make precommit` via the check target
- All existing golangci-lint violations are resolved
- No linters are weakened or disabled in configuration
- Existing tests continue to pass
- `make precommit` passes with exit code 0
</summary>

<objective>
The codebase passes golangci-lint with all configured linters, and lint is enforced as part of `make precommit` on every change.
</objective>

<context>
Read CLAUDE.md for project conventions.
Read `Makefile` (~lines 33-36) and `.golangci.yml`.

The `check` target at Makefile line 36 currently excludes `lint`:
```
check: vet errcheck vulncheck osv-scanner gosec trivy
```
There is a commented-out version with lint at line 34 and a `# TODO: enable lint` at line 33.
The `.golangci.yml` is already in v2 format with linters configured.
</context>

<requirements>
1. In `Makefile`, change the `check` target (line 36) from:
   `check: vet errcheck vulncheck osv-scanner gosec trivy`
   to:
   `check: lint vet errcheck vulncheck osv-scanner gosec trivy`
   Remove the TODO comment (line 33) and commented-out line (line 34).
2. Fix all lint violations in the source code
3. If a violation cannot be reasonably fixed, add a targeted `//nolint:lintername` with a justification comment
4. Ensure all existing tests still pass
5. Run `make precommit` to confirm everything passes
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git
- Do NOT refactor code unrelated to fixing lint issues
- Do NOT add new features
- Do NOT change .golangci.yml to weaken or disable linters — fix the code instead
- Minimize changes — fix the root cause, not symptoms
</constraints>

<verification>
Run `make precommit` — must pass with exit code 0.
</verification>
