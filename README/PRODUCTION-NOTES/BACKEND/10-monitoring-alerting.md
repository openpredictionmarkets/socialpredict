---
title: Monitoring and Alerting
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-27T02:03:51Z
updated_at_display: "Monday, April 27, 2026 at 2:03 AM UTC"
update_reason: "Replace the older monitoring-platform plan with a current-state-first note that separates app-owned operational signals from deferred external monitoring stack work."
status: active
---

# Monitoring and Alerting

## Update Summary

This note was updated on Monday, April 27, 2026 to replace an older Prometheus-and-dashboards-first plan with guidance that matches the live backend and the current design-plan recommendation to harden app-owned signals before adopting a larger monitoring platform.

| Topic | Prior to April 27, 2026 | After April 27, 2026 |
| --- | --- | --- |
| Core framing | Treated monitoring as a full external platform rollout | Treats monitoring as clarifying the operational signals the live backend should expose and the small first alert surface that should consume them |
| Current-state accuracy | Assumed there was no meaningful observability surface yet | Recognizes the live logging packages, `/health`, `/v0/system/metrics`, and deployment topology already present in the repo |
| Main proposal | Build Prometheus, Alertmanager, dashboards, and broad metric families first | Focus on truthful health and readiness, log and trace correlation direction from the logging note, and a minimal operational metrics seam only when the app can own it cleanly |
| Architecture posture | Proposed a new `monitoring/` package and platform stack as the main move | Keeps monitoring concerns tied to existing runtime, logging, server, and deployment seams |
| Metrics model | Mixed operational telemetry and business metrics together | Separates economics metrics from operational health and alert signals |
| Relationship to logging | Duplicated logging and observability concerns from the logging note | Explicitly treats [02-logging-observability.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/02-logging-observability.md) as the app-facing telemetry-model note and this note as the operational-consumption note |
| Future ideas | Mixed external platform choices directly into the active note | Defers platform-heavy work to [FUTURE/05-long-term-monitoring-platform.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/05-long-term-monitoring-platform.md) |

## Executive Direction

SocialPredict should treat monitoring and alerting as an operational contract for the existing backend, not as a greenfield monitoring stack project.

The active direction is:

1. Keep [02-logging-observability.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/02-logging-observability.md) as the owner of app-facing logging, OpenTelemetry direction, and correlation vocabulary.
2. Use this note to describe what operations should be able to observe from the running backend and proxy stack.
3. Treat truthful liveness, readiness, startup-failure, and request-failure signals as the first monitoring concern.
4. Keep operational metrics separate from business and economics metrics such as `/v0/system/metrics`.
5. Add minimal app-owned operational metrics only where the backend can own them clearly and without inventing a second monitoring architecture.
6. Defer Prometheus, Grafana, Alertmanager, ELK, larger dashboard programs, and on-call platform design until the operational signal contract is materially clearer.

For a high-availability and fault-tolerant backend, monitoring should prefer:

- trustworthy readiness over decorative dashboards
- a small first alert surface over speculative metric sprawl
- correlated logs and traces over duplicated ad hoc logging
- clear distinction between business outputs and runtime health
- signals the app truly owns before external platform expansion

This note explicitly rejects treating a broad external monitoring stack rollout as the immediate main move.

## Why This Matters

The current backend already emits some signals, but the operational contract is incomplete.

That matters because:

- `/health` in [server.go](/workspace/socialpredict/backend/server/server.go) is still a placeholder
- logging is still mostly stdlib-wrapped in [simplelogging.go](/workspace/socialpredict/backend/logger/simplelogging.go) and [loggingutils.go](/workspace/socialpredict/backend/logging/loggingutils.go)
- there is no `/metrics` route or Prometheus client usage in the active backend runtime
- the existing `/v0/system/metrics` route is an economics and accounting surface, not a runtime availability or latency surface

So the active problem is not “we lack a dashboard stack.” The active problem is that the runtime and deployment signals still need to become trustworthy enough to alert on.

## Current Code Snapshot

### Logging already exists, but it is still transitional

The backend already has:

- [logger/simplelogging.go](/workspace/socialpredict/backend/logger/simplelogging.go)
- [logging/loggingutils.go](/workspace/socialpredict/backend/logging/loggingutils.go)

That means the repo is not blank on logging, but it is also not yet at a mature operational-monitoring posture. This note should describe the operational outcome expected from that logging direction rather than duplicate the app-internal logging design.

### Health exists, but readiness does not

The backend currently serves only:

- `GET /health`

And that route currently returns plain-text `ok` in [server.go](/workspace/socialpredict/backend/server/server.go). There is no live:

- readiness endpoint
- liveness endpoint distinct from readiness
- startup probe endpoint

That is the first operational signal gap.

### Business metrics already exist

The backend already exposes economics and accounting metrics through:

- [GetSystemMetricsHandler](/workspace/socialpredict/backend/handlers/metrics/getsystemmetrics.go)
- [analytics/system_metrics.go](/workspace/socialpredict/backend/internal/domain/analytics/system_metrics.go)

Those metrics are useful and should remain documented as business or accounting outputs. They should not be mislabeled as generic operational monitoring.

### The repo does not yet ship a monitoring stack

The active repo does not currently include:

- Prometheus config
- Grafana dashboards
- Alertmanager config
- ELK or Loki wiring
- an OTel collector deployment
- a `/metrics` endpoint in the backend server

That means the note should not promise those as if they already exist.

### Deploy workflows already exist, but they are not yet backed by strong operational signals

The repo already has:

- CI smoke and tests in [backend.yml](/workspace/socialpredict/.github/workflows/backend.yml)
- image publish in [docker.yml](/workspace/socialpredict/.github/workflows/docker.yml)
- staging and production dispatch workflows

So the monitoring question is not “do we have deployment automation at all.” The question is “what should those environments monitor once the runtime exposes safer signals.”

## What Monitoring and Alerting Should Own

This note should own:

- the distinction between business metrics and operational signals
- the expectation for liveness and readiness signals
- the small first alert surface the runtime should support
- the operational outcome expected from logging and correlation work already described in note `02`
- the distinction between app-owned signals and external monitoring-stack choices

## What This Note Should Not Own

This note should not become the home for every future observability-platform ambition.

It should explicitly defer:

- Prometheus rollout
- Grafana dashboard buildout
- Alertmanager or pager routing design
- ELK or Loki stack design
- fleet-wide log retention policy
- larger SLO and error-budget programs

Those topics now belong in [FUTURE/05-long-term-monitoring-platform.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/05-long-term-monitoring-platform.md), not in the active production note.

## Near-Term Sequencing

The design-plan-aligned monitoring direction is:

1. Make liveness and readiness truthful through the runtime and deployment notes.
2. Keep logging and OpenTelemetry alignment in note `02`.
3. Add a minimal operational metrics seam only once the runtime owns what it is actually reporting.
4. Define a small first alert set around:
   - backend down
   - backend unready
   - startup incompatibility or migration failure
   - severe request-failure spikes once the shared failure surface is more consistent
5. Defer dashboards and external stack selection until the signal contract is materially stronger.

## Open Questions

- Which operational metrics should be app-owned versus inferred from reverse-proxy or host-level systems
- Whether `/health` should later split into root liveness plus a deeper readiness surface
- What the smallest useful first alert set is once readiness exists
- Which operational signals belong at the backend layer versus nginx or Traefik
