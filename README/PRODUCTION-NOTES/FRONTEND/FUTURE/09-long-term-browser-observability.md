---
title: Long-Term Browser Observability
document_type: production-notes
domain: frontend
future: true
author: Patrick Delaney
updated_at: 2026-05-23T00:00:00Z
updated_at_display: "Saturday, May 23, 2026"
update_reason: "Move browser APM, replay, and log-shipping work behind safe failure presentation and telemetry vocabulary."
status: future
---

# Long-Term Browser Observability

## Purpose

This note holds browser observability work that should follow safe failure presentation and event vocabulary design.

The active failure-presentation note is [../06-error-handling.md](../06-error-handling.md).

## Deferred Topics

- Sentry or browser APM SDK rollout.
- Session replay.
- Browser log shipping.
- Real-user monitoring dashboards.
- Alerting and incident workflow.
- Error aggregation and sampling policy.

## Why Deferred

Browser observability affects privacy, bundle size, user/session identifiers, sampling, incident workflow, and vendor lock-in. The frontend should first stop exposing raw errors to users and define what telemetry is safe to collect.

## Entry Criteria

Reconsider this when:

- User-safe recovery messages are implemented.
- The browser telemetry event vocabulary exists.
- Redaction and forbidden fields are explicit.
- Operators know what incidents browser telemetry should detect.
- Sampling and retention expectations are known.

## Guardrail

Do not add browser APM or replay tooling before deciding what data may leave the browser and who owns the resulting incidents.
