---
title: Long-Term Frontend Accessibility Program
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Move program-level accessibility work behind the active core-workflow accessibility baseline."
status: future
---

# Long-Term Frontend Accessibility Program

## Purpose

This note holds accessibility-program work that should follow fixes to current core workflows.

The active accessibility note is [../05-accessibility.md](../05-accessibility.md).

## Deferred Topics

- Full WCAG 2.1 AA certification-style audit.
- Assistive-technology/browser compatibility matrix.
- Automated accessibility CI across many routes.
- Design-system-wide accessibility contract tests.
- Screen-reader acceptance checklist for every component.
- Accessibility regression platform.

## Why Deferred

The first accessibility pass should make existing core market/account/auth workflows usable. A program-level rollout needs stable components, CI, and the visual-system baseline.

## Entry Criteria

Reconsider this when:

- Frontend CI exists.
- Core workflow accessibility issues have been fixed or documented.
- Shared component contracts exist for common controls.
- Automated checks can be scoped narrowly enough to avoid noisy failures.

## Guardrail

Do not use a future accessibility program as a reason to defer obvious current labels, focus behavior, keyboard access, or color-only state problems.
