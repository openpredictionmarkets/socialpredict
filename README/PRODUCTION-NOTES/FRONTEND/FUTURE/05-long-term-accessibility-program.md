---
title: Long-Term Frontend Accessibility Program
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Keep program-level accessibility deferred after the first form/navigation pass until broader workflow findings are known."
status: future
---

# Long-Term Frontend Accessibility Program

## Purpose

This note holds accessibility-program work that should follow the broader current-workflow accessibility audit.

The active accessibility note is [../05-accessibility.md](../05-accessibility.md).

## Deferred Topics

- Full WCAG 2.1 AA certification-style audit.
- Assistive-technology/browser compatibility matrix.
- Automated accessibility CI across many routes.
- Design-system-wide accessibility contract tests.
- Screen-reader acceptance checklist for every component.
- Accessibility regression platform.

## Why Deferred

The first accessibility pass improved create-market, change-password, navigation controls, and status/error announcements. A program-level rollout needs the remaining market, trade, resolve, profile/account, and admin workflows to be audited first.

## Entry Criteria

Reconsider this when:

- Frontend CI and declared frontend test tooling are stable enough to host accessibility checks.
- Core workflow accessibility issues have been fixed or documented beyond the first form/navigation pass.
- Shared component contracts exist for common controls.
- Automated checks can be scoped narrowly enough to avoid noisy failures.

## Guardrail

Do not use a future accessibility program as a reason to defer obvious current labels, focus behavior, keyboard access, or color-only state problems.
