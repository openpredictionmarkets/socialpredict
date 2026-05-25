---
title: Future Frontend Baseline Triage
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Refresh future re-entry criteria after the first frontend CI, API/auth, failure, accessibility, and build-size baseline stack."
status: draft
---

# Future Frontend Baseline Triage

## Purpose

This note indexes frontend work that is intentionally not first.

The active frontend sequence lives one directory up in [../00-TRIAGE.md](../00-TRIAGE.md). The first baseline now exists for frontend CI/build feedback, API/auth boundaries, safe failure presentation, accessibility on current workflows, and measured build-size evidence. The immediate queue should now focus on finishing incremental migrations and adding declared test tooling before promoting larger platform programs.

## Current Active Workstreams

The canonical design plan currently keeps only two active frontend workstreams:

1. `W08 Frontend Experience Boundary and CI Baseline`
2. `W09 Frontend Visual System Baseline`

The design plan also records guardrails in:

- `ADR-032: Defer browser platform capabilities until baseline seams are stable`
- `ADR-033: Treat frontend dependency and performance maintenance as CI evidence`

## Future Notes

| Note | Topic | Why future |
| --- | --- | --- |
| [01-long-term-state-platform.md](./01-long-term-state-platform.md) | Redux, RTK Query, global store, offline sync | First API/auth seam exists, but remaining call-site migration should come first. |
| [02-long-term-performance-platform.md](./02-long-term-performance-platform.md) | Code splitting, Web Vitals, strict budgets, caching platform | Build-size evidence exists; strict budgets still need product thresholds. |
| [03-long-term-test-infrastructure.md](./03-long-term-test-infrastructure.md) | Playwright, coverage, visual regression, full accessibility automation | Frontend CI exists, but declared unit/component test tooling should come first. |
| [04-long-term-security-platform.md](./04-long-term-security-platform.md) | CSP, server-managed sessions, MFA, security monitoring | Requires backend/infra/API coordination. |
| [05-long-term-accessibility-program.md](./05-long-term-accessibility-program.md) | Full WCAG program and assistive-tech matrix | First pass has started; broader market/trade/profile workflow audit should come next. |
| [06-long-term-i18n-localization.md](./06-long-term-i18n-localization.md) | i18n, locale formatting, RTL | Needs product/localization requirements and canonical-language decisions. |
| [07-long-term-pwa-offline-platform.md](./07-long-term-pwa-offline-platform.md) | Service workers, offline, push, installability | Needs freshness and stale/read-only semantics. |
| [08-long-term-analytics-experimentation.md](./08-long-term-analytics-experimentation.md) | Product analytics, funnels, A/B tests | Needs event taxonomy and privacy posture. |
| [09-long-term-browser-observability.md](./09-long-term-browser-observability.md) | Browser APM, replay, dashboards, log shipping | Safe fallback exists; telemetry vocabulary and privacy posture still need design. |
| [10-long-term-maintenance-automation.md](./10-long-term-maintenance-automation.md) | Dependency automation, regression platform, frontend recovery | CI/build-size evidence exists; custom automation needs repeated maintenance pain first. |

## Decision Framework

Bring a future item back into the active queue only when at least one of these is true:

- The current active baseline exposes a concrete production risk.
- Product requirements make the capability necessary now.
- The canonical design plan has an explicit seam, rule, or ADR for the capability.
- CI or release feedback shows that manual review is no longer enough.
- Backend/infra/API dependencies are ready when the frontend cannot own the work alone.

## Guardrail

Do not promote a future platform because a package or vendor makes it easy. Promote it only when SocialPredict has a current failure mode, product requirement, or operational evidence that justifies the extra architecture surface.
