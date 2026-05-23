---
title: Frontend Security Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Keep active frontend security focused on auth/API/session-adjacent seams and move broader platform security into FUTURE."
status: draft
---

# Frontend Security Baseline

## Purpose

This active note covers the near-term frontend security work.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The immediate issue is not only that tokens have used browser storage. The broader issue is the scattered auth/API boundary: token reads, authenticated headers, transport, backend envelope parsing, route access checks, and public failure copy are split across frontend layers.

Broader security-platform work now lives in [FUTURE/04-long-term-security-platform.md](./FUTURE/04-long-term-security-platform.md).

## Current Baseline

- Login state and token persistence are handled in `AuthContent`.
- Some code reads tokens directly from browser storage.
- Components and hooks can construct authenticated requests directly.
- `mustChangePassword` persistence was removed from localStorage in a prior security fix.
- CSP, server-managed sessions, and advanced auth flows are not current frontend-only changes.

## Active Direction

1. Inventory browser token reads/writes.
2. Inventory authenticated request construction.
3. Introduce or document the auth/API adapter seam.
4. Keep route access behavior equivalent while extracting helper decisions.
5. Separate user-safe failure copy from raw diagnostic detail.
6. Coordinate server-managed session ideas with backend/API design instead of treating them as a frontend-only refactor.

## Design Plan Alignment

The canonical design plan tracks this under:

- `Frontend API Auth Adapter Seam`
- `Frontend Failure Presentation Seam`
- `Request-Boundary Security Surface`
- `API and Auth Contract Boundary`

## Active Acceptance Criteria

- Token/storage access is isolated or inventoried.
- Auth header construction has one preferred path.
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
