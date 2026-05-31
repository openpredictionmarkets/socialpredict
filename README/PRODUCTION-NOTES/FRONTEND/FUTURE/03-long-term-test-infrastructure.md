---
title: Long-Term Frontend Test Infrastructure
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Keep broad test infrastructure deferred after frontend CI until declared unit/component test tooling exists."
status: future
---

# Long-Term Frontend Test Infrastructure

## Purpose

This note holds frontend test-platform ideas that should follow declared unit/component test tooling, not merely the first CI/build baseline.

The active verification note is [../03-testing-strategy.md](../03-testing-strategy.md).

## Deferred Topics

- Playwright E2E suite.
- Component testing framework expansion.
- Coverage thresholds.
- Visual regression testing.
- Automated accessibility test matrix.
- Performance test suite.
- Cross-browser CI matrix.

## Why Deferred

The frontend now has a stable install/build check. It still lacks a declared `npm test` script and the dependencies needed by the visible Vitest-style test, so broad testing infrastructure remains premature.

## Entry Criteria

Reconsider this when:

- Frontend PR CI runs clean install and production build.
- Test tooling is declared in `frontend/package.json`.
- A small test command is reproducible from a clean install.
- Core selectors and route flows are stable enough for E2E work.
- Accessibility baseline fixes have identified stable checks worth automating.

## Guardrail

Do not add broad test gates that look impressive but are noisy, unreproducible, or detached from current production risks.
