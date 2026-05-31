---
title: Frontend Verification Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Record the dedicated frontend install/build workflow and reset active verification work around declared test tooling."
status: draft
---

# Frontend Verification Baseline

## Purpose

This active note defines the next frontend verification target after the first frontend build workflow.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The repo now has a visible frontend PR signal. Do not claim test coverage until test tooling is declared and reproducible.

Long-term testing infrastructure ideas live in [FUTURE/03-long-term-test-infrastructure.md](./FUTURE/03-long-term-test-infrastructure.md).

## Current Baseline

- One visible Vitest-style test file exists.
- `frontend/package.json` does not declare a `test` script.
- `vitest` and `jsdom` are not declared as frontend dependencies.
- [`.github/workflows/frontend.yml`](../../../.github/workflows/frontend.yml) now provides the first dedicated frontend PR build check and runs `npm run build:report`.
- GitHub PR checks now fail broken frontend production builds before merge.

## Active Direction

The next verification slice should add declared test tooling only if it is kept small and reproducible:

1. Keep the Node 22 `npm ci` and `npm run build:report` workflow stable.
2. Add `vitest`, `jsdom`, and a `test` script before running tests in CI.
3. Start with a small component/hook test that proves the tooling path.
4. Add tests to CI only after the command is reproducible from a clean install.
5. Keep Playwright, coverage gates, visual regression, and broad accessibility automation out of the next slice unless separately justified.

## Design Plan Alignment

The canonical design plan tracks this as:

- `Frontend Verification Baseline Seam`
- `W08 Frontend Experience Boundary and CI Baseline`
- `ADR-033: Treat frontend dependency and performance maintenance as CI evidence`

## Active Acceptance Criteria

- Frontend PRs produce a visible GitHub Actions status.
- A broken Vite build fails before merge.
- The frontend build-size report runs as part of the same verification path.
- Any test command run by CI is declared in `frontend/package.json` and reproducible from a clean install.
- The CI job stays narrow enough to be stable.

## Explicitly Deferred

- Playwright E2E suite.
- Full component-testing framework rollout.
- Coverage thresholds.
- Visual regression testing.
- Performance/load test platform.

See [FUTURE/03-long-term-test-infrastructure.md](./FUTURE/03-long-term-test-infrastructure.md).
