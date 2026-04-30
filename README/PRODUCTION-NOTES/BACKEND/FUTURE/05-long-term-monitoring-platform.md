---
title: Long-Term Monitoring Platform
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-30T11:55:00Z
updated_at_display: "Thursday, April 30, 2026 at 11:55 AM UTC"
update_reason: "Record that the serving-path liveness/readiness signal prerequisite finished on April 30 while keeping monitoring-platform rollout deferred."
status: draft
---

# Long-Term Monitoring Platform

## Purpose

This note holds longer-term monitoring-platform ideas that should not drive the active runtime and observability sequence.

The active operational-monitoring work remains in [09-monitoring-alerting.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/09-monitoring-alerting.md) and the app-facing telemetry model remains in [02-logging-observability.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/02-logging-observability.md).

## Completed Prerequisite

Real serving-path liveness and readiness signals finished on April 30, 2026: `/health` now reports liveness, and `/readyz` reports readiness plus database availability. That closes the first signal gap, but broader monitoring-platform work remains deferred until the backend has operational metrics, stable correlation fields, and deployment environments ready to consume them.

## Deferred Topics

Deferred monitoring-platform ideas include:

- Prometheus rollout
- Grafana dashboards
- Alertmanager or pager routing
- ELK, Loki, or other centralized log platforms
- OTel collector topology choices
- SLO and error-budget programs
- alert runbook formalization
- long-term retention and search platform choices

## Preconditions

These ideas should stay deferred until the backend has:

- deployment environments that consume the April 30, 2026 liveness/readiness signals reliably
- a clearer shared failure surface
- stable request correlation fields
- a small first operational metrics seam
- deployment environments that can consume those signals reliably

## Guardrail

This document is non-binding on the active design plan and task queue until the active signal contract is materially stronger.
