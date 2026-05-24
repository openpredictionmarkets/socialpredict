---
title: Frontend Performance Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Keep active performance work measurement-first and move broader optimization platform ideas into FUTURE."
status: draft
---

# Frontend Performance Baseline

## Purpose

This active note covers the first performance slice: establish evidence before optimizing.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The frontend uses Vite and several charting/rendering dependencies, but there is not yet a recorded production build-size baseline or CI-visible bundle evidence.

Broader performance-platform ideas now live in [FUTURE/02-long-term-performance-platform.md](./FUTURE/02-long-term-performance-platform.md).

## Current Baseline

- Vite production builds exist and frontend PRs now have a dedicated install/build signal.
- `npm run build:report` runs the production build and prints an informational size table from `frontend/build`.
- The current observed production build report totals roughly `1201 kB` raw and `326 kB` gzip, with the main JavaScript chunk at roughly `1101 kB` raw and `253 kB` gzip.
- Bundle budgets are not defined and should not be strict until this baseline is reviewed against real product goals.
- Route-level code splitting and dependency replacement have not been justified with current measurements.

## Active Direction

The first performance pass should be low-risk:

1. Run a clean frontend production build.
2. Record the current build output and major bundle contributors using `npm run build:report`.
3. Identify obvious duplicate or heavyweight dependencies, especially charting libraries.
4. Decide whether a permissive build-size or bundle-report artifact belongs in CI.
5. Avoid blocking unrelated work with strict budgets until the baseline is understood.

## Design Plan Alignment

The canonical design plan tracks this under:

- `Frontend Verification Baseline Seam`
- `ADR-033: Treat frontend dependency and performance maintenance as CI evidence`
- `W08 Frontend Experience Boundary and CI Baseline`

## Active Acceptance Criteria

- Current production build size is recorded.
- Frontend CI prints the build-size report after the production build.
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
