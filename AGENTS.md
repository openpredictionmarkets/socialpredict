# AGENTS instructions for this repo

Scope: these instructions apply to the entire repository unless a more specific `AGENTS.md` is added in a subdirectory.

## General guidelines

- Keep changes small, focused, and consistent with the existing code in the same directory.
- Prefer using the existing tooling and conventions (e.g., package managers, linters, formatters) instead of introducing new ones.
- Do not add new external dependencies unless clearly necessary for the task; if you do, update the appropriate lockfiles and configs.
- Avoid large-scale refactors unless explicitly requested.
- Do not add license or copyright headers.
- Follow best practices where feasible, including SOLID and especially Single Responsibility.
- Prefer smaller, more focused functions.
- Prefer shared helper functions.
- Strive for a cyclomatic complexity score of 4 or lower.

## Languages and tooling

- **Go backend (`backend/`)**
  - Follow existing patterns in handlers, models, and server wiring.
  - Run `go fmt` on any Go files you modify (or match the surrounding style if you cannot run tools).
  - Keep HTTP contracts in sync with `backend/docs/openapi.yaml` when changing public APIs.

- **TypeScript / JavaScript (`frontend/`, `sdk/`, `bin/`)**
  - Prefer TypeScript where possible and match existing compiler targets and module systems.
  - Keep import paths and project structure consistent with nearby files.
  - Use existing test frameworks (e.g., Jest, Vitest, or similar) if you need to add tests; do not introduce a new framework.

## Testing and validation

- When modifying behavior rather than just comments/docs, run the narrowest relevant tests you can (unit tests for the touched package, then broader suites if appropriate).
- If tests are failing for reasons unrelated to your change, do not attempt to fix them unless explicitly asked; instead, note the failures in your summary.
- For CLI scripts under `bin/`, provide a short usage note or `--help` output when adding new commands.

## Documentation

- Update README files, OpenAPI docs, or inline docstrings when you change public-facing behavior or APIs.
- Keep documentation concise and colocated with the relevant code when possible.
