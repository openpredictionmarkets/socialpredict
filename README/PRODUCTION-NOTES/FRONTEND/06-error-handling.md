---
title: Frontend Failure Presentation Baseline
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Record the safe global fallback and keep component-level failure translation as the next active work."
status: draft
---

# Frontend Failure Presentation Baseline

## Purpose

This active note covers frontend error-handling work after the first global fallback slice.

Start with [00-TRIAGE.md](./00-TRIAGE.md). The app now has a safe global error boundary fallback. Production UI should continue to avoid raw exception messages or arbitrary backend/runtime text to participants.

Browser observability and reporting platform ideas now live in [FUTURE/09-long-term-browser-observability.md](./FUTURE/09-long-term-browser-observability.md).

## Current Baseline

- `App.jsx` has a global error boundary fallback with `role="alert"` and a user-safe recovery message.
- Raw `error.message` detail is limited to development builds inside a collapsed details region.
- Some components still render or alert raw `err.message` values.
- Backend envelopes and public reasons are not consistently mapped into frontend recovery copy.

## Active Direction

1. Keep global fallback copy user-safe by default.
2. Keep detailed diagnostics available for development, not production UI.
3. Distinguish public failure reasons from telemetry/error details.
4. Map backend envelopes/reasons into stable recovery messages where practical.
5. Remove or translate raw component-level `err.message` presentation where it reaches users.
6. Add focused tests after frontend test tooling is declared.

## Design Plan Alignment

The canonical design plan tracks this as:

- `Frontend Failure Presentation Seam`
- `User-Safe Recovery Message`
- `Failure Translation and Recovery Boundary`
- `Frontend Published Language and API Contract Seam`

## Active Acceptance Criteria

- Production UI does not render raw exception messages from the global fallback.
- Expected business rejections do not read like runtime crashes.
- Users receive a stable recovery path for unexpected failures.
- Developers can still diagnose errors locally through development-only details.
- Remaining raw component-level errors are inventoried before broader failure-translation work.

## Explicitly Deferred

- Sentry or browser APM SDK rollout.
- Session replay.
- Central browser log shipping.
- Broad retry framework.
- Full incident dashboard.

See [FUTURE/09-long-term-browser-observability.md](./FUTURE/09-long-term-browser-observability.md).
