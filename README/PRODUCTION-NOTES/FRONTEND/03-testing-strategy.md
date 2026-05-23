---
title: Frontend Verification Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Focus active testing work on reproducible frontend build feedback before broader test infrastructure."
status: draft
---

# Frontend Verification Baseline

## Purpose

This active note defines the first frontend verification target.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The repo needs a visible frontend PR signal before larger auth, API, accessibility, or performance changes. Do not claim test coverage until test tooling is declared and reproducible.

Long-term testing infrastructure ideas live in [FUTURE/03-long-term-test-infrastructure.md](./FUTURE/03-long-term-test-infrastructure.md).

## Current Baseline

- One visible Vitest-style test file exists.
- `frontend/package.json` does not declare a `test` script.
- `vitest` and `jsdom` are not declared as frontend dependencies.
- [`.github/workflows/frontend.yml`](../../../.github/workflows/frontend.yml) now provides the first dedicated frontend PR build check.

## Active Direction

The first verification slice should add a small, reliable CI signal:

1. Use Node 22 unless the project deliberately chooses another supported runtime.
2. Run `npm ci` from `frontend/`.
3. Run `npm run build` from `frontend/`.
4. Add tests to CI only after adding the required test dependencies and an explicit script.
5. Keep Playwright, coverage gates, visual regression, and accessibility automation out of the first slice.

## Design Plan Alignment

The canonical design plan tracks this as:

- `Frontend Verification Baseline Seam`
- `W08 Frontend Experience Boundary and CI Baseline`
- `ADR-033: Treat frontend dependency and performance maintenance as CI evidence`

## Active Acceptance Criteria

- Frontend PRs produce a visible GitHub Actions status.
- A broken Vite build fails before merge.
- Any test command run by CI is declared in `frontend/package.json` and reproducible from a clean install.
- The CI job stays narrow enough to be stable.

## Explicitly Deferred

- Playwright E2E suite.
- Full component-testing framework rollout.
- Coverage thresholds.
- Visual regression testing.
- Performance/load test platform.

See [FUTURE/03-long-term-test-infrastructure.md](./FUTURE/03-long-term-test-infrastructure.md).
