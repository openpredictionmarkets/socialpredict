---
title: Monitoring and Alerting
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-02T18:20:00Z
updated_at_display: "Saturday, May 2, 2026 at 6:20 PM UTC"
update_reason: "Close WAVE09 with the first supported alert set and deferred monitoring-platform topics."
status: active
---

# Monitoring and Alerting

## Update Summary

This note was updated on Monday, April 27, 2026 to replace an older Prometheus-and-dashboards-first plan with guidance that matches the live backend and the current design-plan recommendation to harden app-owned signals before adopting a larger monitoring platform.

On Thursday, April 30, 2026, the first app-owned operational signal gap was closed for the serving path: `/health` now reports liveness, `/readyz` reports readiness and database availability, and Docker black-box checks confirmed both endpoints on `http://localhost:8080`. Broader metrics export, dashboarding, and alerting remain deferred.

On Saturday, May 2, 2026, this note was tightened into an explicit current-state signal inventory. The backend now documents the operator-facing contract for `/health`, `/readyz`, `/ops/status`, startup fatal failures, and shared request-boundary failure responses while keeping `/v0/system/metrics` classified as economics and accounting output.

The WAVE09 stop-and-review point is that the backend can now support a first alert set for backend down, backend unready, fatal startup failure before readiness opens, and severe server-side request-failure spikes. Broader dashboards or external monitoring platforms should not start until the next app-owned signal seam is chosen and landed.

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

- `/health`, `/readyz`, and `/ops/status` in [server.go](/workspace/socialpredict/backend/server/server.go) now provide the first liveness/readiness and request-failure counter contract, while broader latency and telemetry export remain deferred
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

### Liveness, readiness, and the first status export now exist

As of May 2, 2026, the backend serves:

- `GET /health`
- `GET /readyz`
- `GET /ops/status`

Those routes are intentionally small text responses in [server.go](/workspace/socialpredict/backend/server/server.go):

- `/health` reports liveness as `live`
- `/readyz` reports readiness as `ready` or `not ready` and checks database availability after the readiness gate opens
- `/ops/status` reports JSON `{ live, ready, requestFailuresTotal }` and uses `503` while the backend is live but unready

That first operational signal problem is intentionally narrow. The remaining monitoring gap is not basic liveness/readiness or a first request-failure counter; it is broader latency telemetry, external metrics export, and alert-consumption design.

### Current operator-facing signal inventory

The current backend signal contract is intentionally small:

| Signal | Live backend behavior | Operator meaning | Not a promise of |
| --- | --- | --- | --- |
| `GET /health` | `200 text/plain` body `live` while the HTTP process is serving; `503` body `not live` if the serving probe reports not live | Process liveness for restart or black-box availability checks | Database reachability, latency, request success rate, or telemetry pipeline health |
| `GET /readyz` | `200 text/plain` body `ready` only after the startup readiness gate opens and the primary database ping succeeds; otherwise `503` body `not ready` | Whether the instance should receive traffic | Business health, background job status, or external monitoring-stack health |
| `GET /ops/status` | `200` or `503 application/json` body `{ live, ready, requestFailuresTotal }`; non-probe 5xx responses increment the process-local request-failure counter | Minimal app-owned status export for backend down/unready and severe request-failure spike alerts | Prometheus format, fleet-wide aggregation, latency histograms, or business metrics |
| Startup fatal failures | Configuration load, DB initialization, DB readiness, security config, startup mutation mode, shutdown config, migration/verification, and seed failures call `logger.Fatal` before `readiness.MarkReady()` | Failed starts stay out of the ready pool and leave a fatal startup log | Automatic remediation, multi-replica coordination, or alert routing |
| Shared request-boundary failures | Router-owned `405`, security middleware `429`, and recovered unhandled panics use JSON `{ ok: false, reason }` responses with safe runtime reasons | A consistent first failure surface for operators to count from logs/proxy observations | A full error-budget system, Prometheus metric family, or handler-by-handler failure taxonomy |

This inventory reflects the live server wiring in [server.go](/workspace/socialpredict/backend/server/server.go), startup order in [main.go](/workspace/socialpredict/backend/main.go), and runtime failure helpers in [security/failures.go](/workspace/socialpredict/backend/security/failures.go) and [security/requestboundary.go](/workspace/socialpredict/backend/security/requestboundary.go).

### First supported alert set

The current backend can support only this first alert set:

| Alert | App-owned source | Supported interpretation | Current limit |
| --- | --- | --- | --- |
| Backend down | `GET /health` black-box failure or non-`200` response | The HTTP process is not serving the liveness contract | Does not distinguish process crash, host/network outage, or proxy routing failure |
| Backend unready | `GET /readyz` `503` or `/ops/status.ready=false` | The instance should not receive traffic because startup has not opened readiness or the database ping failed | Does not cover business-rule health, background work, or third-party dependencies |
| Fatal startup failure | Process exits before readiness opens with startup logger events such as `startup.incompatibility` or `startup.migration_failed` | Startup validation, mutation, migration, seed, security, config, or DB setup blocked service readiness | Requires the runtime supervisor or deployment environment to observe process exit and startup logs |
| Severe request-failure spike | Rising `/ops/status.requestFailuresTotal` for non-probe server-side `5xx` responses | The process is producing server-side failures after it starts serving traffic | Counter is process-local, monotonic until restart, and not yet split by route, reason, latency, or replica |

This is the end of the current wave's alert surface. Operators can wire these checks in their existing environment, but the backend is not yet promising fleet aggregation, dashboard panels, paging policy, latency histograms, or Prometheus-compatible exposition.

### Business metrics already exist

The backend already exposes economics and accounting metrics through:

- [GetSystemMetricsHandler](/workspace/socialpredict/backend/handlers/metrics/getsystemmetrics.go)
- [analytics/system_metrics.go](/workspace/socialpredict/backend/internal/domain/analytics/system_metrics.go)

Those metrics are useful and should remain documented as business or accounting outputs. They should not be mislabeled as generic operational monitoring.

In particular, `GET /v0/system/metrics` computes economy/accounting values such as money creation and utilization through the application reporting stack. It is not a replacement for `/health`, `/readyz`, request failure counting, latency telemetry, or a future runtime metrics exporter.

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
   - severe request-failure spikes using `/ops/status.requestFailuresTotal`
5. Defer dashboards and external stack selection until the signal contract is materially stronger.

## Remaining Signal Seams

The remaining gaps should stay narrow and tied to runtime signals the backend can own:

- Add request-boundary latency and duration fields before promising latency dashboards.
- Decide whether the next app-owned counter should classify failures by route family, stable reason, or status class before adding more counters.
- Add process identity or start-time metadata if operators need to distinguish counter resets from real recovery.
- Decide which traffic-volume and edge-failure signals belong in the backend versus nginx or Traefik before moving proxy observations into app docs.
- Keep `/health` and `/readyz` small unless a concrete readiness dependency needs a documented server-owned check.

These are signal-contract seams, not monitoring-platform backlog buckets.

## End-of-Wave Deferral

WAVE09 stops here. Prometheus exposition, Grafana dashboards, Alertmanager routing, centralized log-platform rollout, paging policy, formal SLOs, and error-budget programs remain deferred to future platform work after the owned runtime signal contract is stronger.

Until then, the backend monitoring contract is the first alert set above plus the logging and trace-correlation direction owned by [02-logging-observability.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/02-logging-observability.md).
