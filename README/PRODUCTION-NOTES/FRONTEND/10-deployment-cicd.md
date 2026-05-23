---
title: Frontend Deployment and CI Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Keep active deployment guidance focused on frontend PR feedback and move broader platform deployment ideas into FUTURE."
status: draft
---

# Frontend Deployment and CI Baseline

## Purpose

This active note covers the frontend CI/deployment baseline.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The repository already has release/manual Docker image publishing that builds the frontend image. The missing near-term feedback loop is a dedicated frontend PR check.

Long-term maintenance and deployment-platform work lives in [FUTURE/10-long-term-maintenance-automation.md](./FUTURE/10-long-term-maintenance-automation.md).

## Current Baseline

- `.github/workflows/docker.yml` builds/publishes frontend images on release/manual paths.
- [`.github/workflows/frontend.yml`](../../../.github/workflows/frontend.yml) now runs the first dedicated frontend PR install/build check.
- Frontend test tooling is not yet declared.
- Browser performance and bundle evidence are not yet CI artifacts.

## Active Direction

1. Add a frontend PR job or extend an existing workflow with a frontend job.
2. Use Node 22 unless the project deliberately chooses another supported runtime.
3. Run `npm ci` from `frontend/`.
4. Run `npm run build` from `frontend/`.
5. Add tests, accessibility checks, and bundle budgets only after tooling and thresholds are explicit.

## Design Plan Alignment

The canonical design plan tracks this as:

- `Frontend Verification Baseline Seam`
- `W08 Frontend Experience Boundary and CI Baseline`
- `ADR-033: Treat frontend dependency and performance maintenance as CI evidence`

## Active Acceptance Criteria

- Frontend PRs have a visible GitHub Actions check.
- Broken frontend builds fail before merge.
- First job is small and stable.
- Any later checks have declared dependencies and documented thresholds.

## Explicitly Deferred

- Playwright/Lighthouse CI in the first frontend workflow.
- Terraform or frontend infrastructure redesign.
- Rollback workflow redesign.
- Custom dependency-management platform.
- Browser monitoring dashboard.

See [FUTURE/10-long-term-maintenance-automation.md](./FUTURE/10-long-term-maintenance-automation.md).
