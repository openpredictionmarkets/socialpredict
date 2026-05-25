---
title: Frontend Security Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Record the first auth/API storage and transport seam while keeping broader session and CSP work deferred."
status: draft
---

# Frontend Security Baseline

## Purpose

This active note covers the near-term frontend security work after the first auth/API adapter seam.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The immediate issue is not only that tokens use browser storage. The first adapter seam now reduces scattered login/auth transport behavior, but several authenticated calls still construct token headers and failure handling directly.

Broader security-platform work now lives in [FUTURE/04-long-term-security-platform.md](./FUTURE/04-long-term-security-platform.md).

## Current Baseline

- Login state now uses the first API/auth adapter seam from `AuthContent`.
- `authStorage.js` owns login/logout token storage for migrated flows.
- Some profile, admin, market-detail, and trading code still reads tokens directly from browser storage outside that seam.
- Create-market, change-password, homepage, and admin homepage flows now use the adapter; other components and hooks can still construct authenticated requests directly.
- `mustChangePassword` persistence was removed from localStorage in a prior security fix.
- The global error fallback now avoids raw runtime detail in production UI.
- CSP, server-managed sessions, and advanced auth flows are not current frontend-only changes.

## Active Direction

1. Inventory remaining browser token reads/writes.
2. Inventory remaining authenticated request construction.
3. Move additional authenticated requests through `authenticatedApiRequest` in small slices.
4. Keep route access behavior equivalent while extracting helper decisions.
5. Continue separating user-safe failure copy from raw diagnostic detail.
6. Coordinate server-managed session ideas with backend/API design instead of treating them as a frontend-only refactor.

## Design Plan Alignment

The canonical design plan tracks this under:

- `Frontend API Auth Adapter Seam`
- `Frontend Failure Presentation Seam`
- `Request-Boundary Security Surface`
- `API and Auth Contract Boundary`

## Active Acceptance Criteria

- Login/logout token storage access is isolated behind `authStorage`.
- Remaining token/storage access is inventoried before migration.
- Auth header construction has one preferred path for migrated callers.
- Login/logout/must-change-password behavior remains stable.
- Production UI does not expose raw runtime errors as user-facing security copy.
- Future server-managed session work is explicitly marked backend/API coordinated.

## Explicitly Deferred

- Full CSP/security-header program.
- Server-managed session migration.
- MFA or stronger auth features.
- Browser security monitoring platform.
- Client-side encryption/obfuscation wrappers for localStorage.

See [FUTURE/04-long-term-security-platform.md](./FUTURE/04-long-term-security-platform.md).
