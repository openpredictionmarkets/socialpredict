---
title: Frontend State Management and API/Auth Boundary
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Record the first auth/API adapter seam and representative call-site migrations while keeping broad state platforms deferred."
status: draft
---

# Frontend State Management and API/Auth Boundary

## Purpose

This active note covers the near-term frontend state-management work that is actually next for SocialPredict after the first auth/API adapter seam.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The immediate problem is not the absence of Redux. The first seam now exists, but route access decisions and many non-migrated transport/envelope call sites are still mixed across React context, routes, hooks, pages, and components.

Broad state-platform ideas now live in [FUTURE/01-long-term-state-platform.md](./FUTURE/01-long-term-state-platform.md).

## Current Baseline

- `AuthContent` now uses the first API/auth adapter seam for login transport, browser persistence, and backend envelope parsing.
- Browser token compatibility still uses `localStorage`, but token reads/writes should go through `frontend/src/api/authStorage.js` where migrated.
- Representative create-market, change-password, homepage, and admin homepage flows now use the API/auth adapter; other components and hooks still call `fetch(API_URL...)` directly.
- `frontend/src/api/httpClient.js` now wraps API URL construction, auth header injection, response parsing, and envelope unwrapping for migrated callers.
- `frontend/src/utils/apiResponse.js` remains the shared response helper, but duplicated response-unwrapping helpers remain.
- `AppRoutes` embeds auth/access decisions directly in route JSX.

## Active Direction

Continue small seam migration before considering a global state library:

1. Inventory remaining token reads/writes and direct `fetch(API_URL...)` call sites.
2. Prefer `apiRequest` or `authenticatedApiRequest` for new or migrated call sites.
3. Move remaining authenticated profile, trade, resolve, and admin calls behind the adapter in small slices.
4. Centralize backend envelope parsing by removing local `unwrapApiResponse` copies where practical.
5. Keep pages and presentational components dependent on hooks/use-case helpers, not raw transport or storage details.
6. Preserve current runtime behavior while reducing coupling.

## Design Plan Alignment

The canonical design plan tracks this as:

- `Frontend Experience Context`
- `Frontend API Auth Adapter Seam`
- `Frontend Published Language and API Contract Seam`
- `ADR-032: Defer browser platform capabilities until baseline seams are stable`

## Active Acceptance Criteria

- Login/logout token reads are isolated behind an auth storage/API adapter boundary.
- Remaining direct token reads and direct authenticated `fetch` calls are inventoried for incremental migration.
- Backend envelope parsing has one preferred helper path for migrated callers.
- Login, logout, and must-change-password behavior remain unchanged from the user's perspective.
- Server-managed sessions or memory-only token handling are treated as separate backend/session design work, not slipped into this slice.

## Explicitly Deferred

- Full Redux, Zustand, or RTK Query migration.
- Persisted global store design.
- Offline data synchronization.
- Optimistic write framework.
- Large route/data-loader redesign.

See [FUTURE/01-long-term-state-platform.md](./FUTURE/01-long-term-state-platform.md).
