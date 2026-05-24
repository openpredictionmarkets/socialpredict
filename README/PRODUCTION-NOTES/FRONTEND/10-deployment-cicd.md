---
title: Frontend Deployment and CI Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Record the first frontend PR workflow and non-blocking build-size report."
status: draft
---

# Frontend Deployment and CI Baseline

## Purpose

This active note covers the frontend CI/deployment baseline after the first frontend PR workflow.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The repository already has release/manual Docker image publishing that builds the frontend image. The missing near-term feedback loop was a dedicated frontend PR check; that baseline now exists.

Long-term maintenance and deployment-platform work lives in [FUTURE/10-long-term-maintenance-automation.md](./FUTURE/10-long-term-maintenance-automation.md).

## Current Baseline

- `.github/workflows/docker.yml` builds/publishes frontend images on release/manual paths.
- [`.github/workflows/frontend.yml`](../../../.github/workflows/frontend.yml) now runs the first dedicated frontend PR install/build check.
- Frontend test tooling is not yet declared.
- Browser performance is not yet a CI artifact.
- Bundle size evidence is now printed in the frontend workflow logs through `npm run build:report`.

## Active Direction

1. Keep the frontend PR job small and stable.
2. Use Node 22 unless the project deliberately chooses another supported runtime.
3. Run `npm ci` from `frontend/`.
4. Run `npm run build:report` from `frontend/` so CI shows the production build and informational size table.
5. Add tests, accessibility checks, and enforceable bundle budgets only after tooling and thresholds are explicit.
6. Decide later whether build-size output should become a stored artifact or remain workflow-log evidence.

## Design Plan Alignment

The canonical design plan tracks this as:

- `Frontend Verification Baseline Seam`
- `W08 Frontend Experience Boundary and CI Baseline`
- `ADR-033: Treat frontend dependency and performance maintenance as CI evidence`

## Active Acceptance Criteria

- Frontend PRs have a visible GitHub Actions check on `main` and stacked `frontend/**` PRs.
- Broken frontend builds fail before merge.
- Frontend workflow logs include a non-blocking build-size report.
- First job is small and stable.
- Any later checks have declared dependencies and documented thresholds.

## Explicitly Deferred

- Playwright/Lighthouse CI in the first frontend workflow.
- Terraform or frontend infrastructure redesign.
- Rollback workflow redesign.
- Custom dependency-management platform.
- Browser monitoring dashboard.

See [FUTURE/10-long-term-maintenance-automation.md](./FUTURE/10-long-term-maintenance-automation.md).
