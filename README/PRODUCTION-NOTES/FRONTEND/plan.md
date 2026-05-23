---
title: Frontend Production Readiness Notes Index
document_type: production-notes
domain: frontend
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Replace the broad greenfield plan with an index that separates active frontend baseline notes from FUTURE platform notes."
status: draft
---

# Frontend Production Readiness Notes Index

## Start Here

Use [00-TRIAGE.md](./00-TRIAGE.md) for the current frontend execution sequence.

The active frontend queue is intentionally narrow. It follows the canonical design plan's `W08 Frontend Experience Boundary and CI Baseline` and `W09 Frontend Visual System Baseline` workstreams. Larger frontend platform programs are preserved under [FUTURE/](./FUTURE/).

## Active Baseline Notes

| Order | Note | Active focus |
| --- | --- | --- |
| 00 | [00-TRIAGE.md](./00-TRIAGE.md) | Current sequence and decision framework |
| 01 | [01-state-management.md](./01-state-management.md) | API/auth adapter and state boundary baseline |
| 02 | [02-performance-optimization.md](./02-performance-optimization.md) | Build-size and performance measurement baseline |
| 03 | [03-testing-strategy.md](./03-testing-strategy.md) | Frontend install/build verification baseline |
| 04 | [04-security.md](./04-security.md) | Auth/API/session-adjacent security baseline |
| 05 | [05-accessibility.md](./05-accessibility.md) | Core workflow accessibility baseline |
| 06 | [06-error-handling.md](./06-error-handling.md) | Safe public failure presentation baseline |
| 10 | [10-deployment-cicd.md](./10-deployment-cicd.md) | Frontend PR CI and deployment-feedback baseline |

## Future Platform Notes

| Note | Deferred topic |
| --- | --- |
| [FUTURE/00-baseline-triage.md](./FUTURE/00-baseline-triage.md) | Future backlog index and re-entry framework |
| [FUTURE/01-long-term-state-platform.md](./FUTURE/01-long-term-state-platform.md) | Redux, RTK Query, persisted store, offline sync |
| [FUTURE/02-long-term-performance-platform.md](./FUTURE/02-long-term-performance-platform.md) | Code splitting, Web Vitals, strict budgets, caching platform |
| [FUTURE/03-long-term-test-infrastructure.md](./FUTURE/03-long-term-test-infrastructure.md) | Playwright, coverage, visual regression, accessibility automation |
| [FUTURE/04-long-term-security-platform.md](./FUTURE/04-long-term-security-platform.md) | CSP, server sessions, stronger auth, browser security monitoring |
| [FUTURE/05-long-term-accessibility-program.md](./FUTURE/05-long-term-accessibility-program.md) | Full WCAG/accessibility program |
| [FUTURE/06-long-term-i18n-localization.md](./FUTURE/06-long-term-i18n-localization.md) | i18n, locale formatting, RTL |
| [FUTURE/07-long-term-pwa-offline-platform.md](./FUTURE/07-long-term-pwa-offline-platform.md) | Service workers, offline, push, installability |
| [FUTURE/08-long-term-analytics-experimentation.md](./FUTURE/08-long-term-analytics-experimentation.md) | Product analytics, funnels, A/B tests |
| [FUTURE/09-long-term-browser-observability.md](./FUTURE/09-long-term-browser-observability.md) | Browser APM, replay, dashboards, log shipping |
| [FUTURE/10-long-term-maintenance-automation.md](./FUTURE/10-long-term-maintenance-automation.md) | Dependency automation, regression platform, frontend recovery |

## Current First Five PRs

1. Document frontend published-language and API-contract alignment.
2. Add frontend CI build check with Node 22, `npm ci`, and `npm run build`.
3. Sanitize the global error-boundary fallback and define safe public failure copy.
4. Add auth/API boundary discovery plus a small API client/auth-header seam.
5. Migrate token/API call sites incrementally, then run an accessibility pass on core workflows.

## Guardrail

Do not promote a future platform topic into the active queue until the active baseline exposes a concrete production risk, product requirement, or CI/release evidence that justifies the extra architecture surface.
