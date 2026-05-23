---
title: Frontend State Management and API/Auth Boundary
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Keep active state-management work focused on auth/API adapter seams before any broad Redux or RTK Query platform migration."
status: draft
---

# Frontend State Management and API/Auth Boundary

## Purpose

This active note covers the near-term frontend state-management work that is actually next for SocialPredict.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The immediate problem is not the absence of Redux. The immediate problem is that auth state, browser storage, API transport, backend envelope parsing, and route access decisions are mixed across React context, routes, hooks, pages, and components.

Broad state-platform ideas now live in [FUTURE/01-long-term-state-platform.md](./FUTURE/01-long-term-state-platform.md).

## Current Baseline

- `AuthContent` owns React auth context, login transport, JWT decoding, browser persistence, and backend envelope parsing.
- Some frontend code reads `localStorage` directly for tokens.
- Components and hooks still call `fetch(API_URL...)` directly.
- `frontend/src/utils/apiResponse.js` exists, but duplicated response-unwrapping helpers remain.
- `AppRoutes` embeds auth/access decisions directly in route JSX.

## Active Direction

Create small seams before considering a global state library:

1. Inventory token reads/writes and direct `fetch(API_URL...)` call sites.
2. Define a frontend API/auth adapter for authenticated requests.
3. Centralize auth header construction.
4. Centralize backend envelope parsing.
5. Keep pages and presentational components dependent on hooks/use-case helpers, not raw transport or storage details.
6. Preserve current runtime behavior while reducing coupling.

## Design Plan Alignment

The canonical design plan tracks this as:

- `Frontend Experience Context`
- `Frontend API Auth Adapter Seam`
- `Frontend Published Language and API Contract Seam`
- `ADR-032: Defer browser platform capabilities until baseline seams are stable`

## Active Acceptance Criteria

- Token reads are isolated behind an auth storage/API adapter boundary, except documented temporary compatibility call sites.
- Direct authenticated `fetch` calls are reduced or inventoried for incremental migration.
- Backend envelope parsing has one preferred helper path.
- Login, logout, and must-change-password behavior remain unchanged from the user's perspective.
- Server-managed sessions or memory-only token handling are treated as separate backend/session design work, not slipped into this slice.

## Explicitly Deferred

- Full Redux, Zustand, or RTK Query migration.
- Persisted global store design.
- Offline data synchronization.
- Optimistic write framework.
- Large route/data-loader redesign.

See [FUTURE/01-long-term-state-platform.md](./FUTURE/01-long-term-state-platform.md).
