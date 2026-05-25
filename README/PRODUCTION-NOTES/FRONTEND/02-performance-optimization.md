---
title: Frontend Performance Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Record the first CI-visible build-size report and keep optimization work evidence-led."
status: draft
---

# Frontend Performance Baseline

## Purpose

This active note covers the first performance slice: preserve evidence before optimizing.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The frontend uses Vite and several charting/rendering dependencies. A first build-size baseline is now recorded, but there is not yet a bundle budget, artifact comparison, or dependency-replacement plan.

Broader performance-platform ideas now live in [FUTURE/02-long-term-performance-platform.md](./FUTURE/02-long-term-performance-platform.md).

## Current Baseline

- Vite production builds exist and frontend PRs now have a dedicated install/build signal through [frontend.yml](../../../.github/workflows/frontend.yml).
- `npm run build:report` runs the production build and prints an informational size table from `frontend/build`.
- The current observed production build report totals roughly `1201 kB` raw and `326 kB` gzip, with the main JavaScript chunk at roughly `1101 kB` raw and `254 kB` gzip.
- Vite warns that the main JavaScript chunk is larger than `1000 kB` after minification. This is informational for now, not a merge blocker.
- Bundle budgets are not defined and should not be strict until this baseline is reviewed against real product goals.
- Route-level code splitting and dependency replacement have not been justified with current measurements.

## Active Direction

The next performance pass should be low-risk:

1. Keep `npm run build:report` green in CI.
2. Track build-size totals when frontend dependency or charting changes land.
3. Identify obvious duplicate or heavyweight dependencies, especially charting libraries.
4. Decide whether a permissive build-size artifact or trend comparison belongs in CI.
5. Avoid blocking unrelated work with strict budgets until the baseline is understood.

## Design Plan Alignment

The canonical design plan tracks this under:

- `Frontend Verification Baseline Seam`
- `ADR-033: Treat frontend dependency and performance maintenance as CI evidence`
- `W08 Frontend Experience Boundary and CI Baseline`

## Active Acceptance Criteria

- Current production build size is recorded.
- Frontend CI prints the build-size report after the production build.
- The current large-chunk warning is acknowledged as informational.
- Any CI budget starts permissive and evidence-based.
- Future performance PRs have a baseline to compare against.
- No route-splitting or dependency-replacement program starts before measurement.

## Explicitly Deferred

- Route-wide code splitting campaign.
- PWA/service-worker caching.
- CDN or edge-caching redesign.
- Web Vitals dashboard or monitoring vendor rollout.
- Strict Lighthouse or bundle-size gates.

See [FUTURE/02-long-term-performance-platform.md](./FUTURE/02-long-term-performance-platform.md).
