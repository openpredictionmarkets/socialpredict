---
title: Runtime Performance Tuning
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-01T11:50:49Z
updated_at_display: "Friday, May 01, 2026 at 11:50 AM UTC"
update_reason: "Refresh the performance note against runtime DB pool ownership landed on upstream main at 051aac6."
status: active
---

# Runtime Performance Tuning

## Update Summary

This note was updated on Monday, April 27, 2026 to replace an older performance-platform plan with guidance that matches the live backend, the current design-plan posture, and the three-agent recommendation to treat optimization as later than runtime correctness and observability.

On Thursday, April 30, 2026, the earlier runtime-readiness blocker called out by this note was finished for the serving path: `/health` now has liveness semantics, `/readyz` checks the readiness gate and database availability, and Docker black-box checks confirmed both endpoints on `http://localhost:8080`. As of upstream `main` at `051aac6b2fefa5634b8c98cc38caf52acf0043a9`, runtime DB ownership also includes explicit `sql.DB` pool and lifetime knobs. That completion does not make broad optimization work current; it narrows the active performance note to tuning the now-landed pool lifecycle controls and measuring real bottlenecks.

| Topic | Prior to April 27, 2026 | After April 27, 2026 |
| --- | --- | --- |
| Core framing | Treated performance as a broad subsystem buildout | Treats performance as evidence-driven hardening of the live backend |
| Current-state accuracy | Assumed query optimizers, cache layers, compression middleware, and monitoring were all the next move | Recognizes that the live backend still needs stronger runtime ownership, readiness, and observability before most optimization work is credible |
| Main proposal | Build query-optimizer packages, caching layers, compression middleware, and generalized pool managers | Focus on `sql.DB` pool and lifecycle tuning, targeted query and index changes, and measured bottlenecks only after earlier runtime notes land |
| Architecture posture | Proposed new performance packages and middleware as the center of the work | Keeps performance work inside the existing runtime, migration, repository, and proxy seams |
| Cache posture | Mixed caching directly into the active performance note | Defers caching to [12-database-caching.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/12-database-caching.md) |
| HA posture | Optimized for speed features first | Optimizes for making a correct and observable system faster only after it is safer to operate |
| Future ideas | Mixed larger performance-platform ideas into the active note | Defers longer-term ideas to [FUTURE/06-long-term-performance-optimization.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/06-long-term-performance-optimization.md) |

## Executive Direction

SocialPredict should treat runtime performance tuning as measured hardening of the backend that already exists, not as a greenfield optimization platform.

The active direction is:

1. Keep correctness, startup discipline, readiness, and observability ahead of speed work.
2. Treat the runtime DB seam's existing `sql.DB` pool and connection-lifecycle knobs as the main near-term performance lever in this note, and tune them only against measured bottlenecks.
3. Make query and index changes only when a real hotspot is visible, and keep those changes owned by repository or migration seams rather than inventing a runtime "query optimizer" package.
4. Prefer proxy-edge features that already exist, such as nginx gzip, over adding speculative in-app compression middleware.
5. Keep business/accounting metrics separate from latency and operational performance signals.
6. Defer caching to [12-database-caching.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/12-database-caching.md) and defer queue or worker ideas to [13-background-jobs.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/13-background-jobs.md).
7. Defer load-testing programs, profiling programs, advanced cache hierarchies, and broader performance infrastructure until earlier production-hardening notes are materially landed.

For a high-availability and fault-tolerant backend, performance work should prefer:

- a correct system before a fast system
- explicit DB lifecycle tuning before speculative caching
- measured bottlenecks over architecture-by-guessing
- migration-owned index changes over ad hoc runtime SQL patches
- minimal new failure domains until the core runtime is safer

This note explicitly rejects building a new `performance/` subsystem as the main move for the active slice.

## Why This Matters

Performance work done too early often speeds up the wrong things.

For SocialPredict, the live backend still has runtime concerns that should stay ahead of speculative speed work:

- startup ownership is too broad in [main.go](/workspace/socialpredict/backend/main.go)
- real liveness and readiness semantics landed on April 30, 2026, but those probes are a correctness baseline rather than a performance strategy
- the DB pool and connection-lifecycle seam now exists in [db.go](/workspace/socialpredict/backend/internal/app/runtime/db.go), but its operating values and real saturation points still need measurement
- accounting-sensitive flows still need clearer atomic boundaries before any optimization layer should obscure behavior

That means the current job is not to add cache tiers, response-compression middleware, or speculative query frameworks. The current job is to tune and observe the performance seams that already landed so later optimization is grounded in evidence.

## Current Code Snapshot

### Liveness and readiness are no longer future performance prerequisites

As of April 30, 2026, [server.go](/workspace/socialpredict/backend/server/server.go) exposes:

- `/health` as a liveness response with `Cache-Control: no-store`
- `/readyz` as a readiness response that checks both the runtime readiness gate and database availability

That problem is finished for the active runtime slice. Performance work should no longer cite "static health check only" as a blocker. The remaining performance-relevant runtime work is tuning the already-landed DB pool and connection-lifecycle seam, followed by measurement.

### DB pool and connection lifecycle tuning is now runtime-owned

The current runtime DB seam in [db.go](/workspace/socialpredict/backend/internal/app/runtime/db.go) normalizes environment variables, builds the Postgres DSN, opens GORM, and applies explicit `sql.DB` tuning through `ConfigureDBPool`, including:

- max open connections
- max idle connections
- connection lifetime
- idle timeout

Those knobs are already wired through env-backed config such as `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, and `DB_CONN_MAX_IDLE_TIME`.

That means the missing piece is not ownership. The missing piece is choosing per-environment values from evidence, understanding saturation behavior, and deciding whether any later hotspot needs query or index work.

### Proxy-edge compression already exists

The production nginx template in [default.conf.template](/workspace/socialpredict/data/nginx/vhosts/prod/default.conf.template) already enables gzip.

That matters because the current backend does not need to rush into app-owned compression middleware to claim it has a performance strategy. The edge already owns basic compression behavior.

### Business metrics already exist, but they are not latency telemetry

The backend already exposes [GetSystemMetricsHandler](/workspace/socialpredict/backend/handlers/metrics/getsystemmetrics.go) and computes accounting-oriented system metrics in [system_metrics.go](/workspace/socialpredict/backend/internal/domain/analytics/system_metrics.go).

Those metrics are valuable, but they are not the same thing as:

- request latency
- error rate
- DB wait time
- pool exhaustion
- queue depth

This note should not conflate business or economics metrics with operational performance telemetry.

### The repo does not yet have a live caching or profiling stack

The active backend does not currently have:

- a Redis runtime
- a cache layer in the backend
- a `/metrics` exporter
- a `pprof` surface
- a profiling or benchmark program wired into production notes

That makes it even more important to keep the note evidence-driven instead of speculative.

### Query and index work should be migration-owned

The repo already has an explicit [migration package](/workspace/socialpredict/backend/migration/migrate.go) and timestamped migrations under [migration/migrations](/workspace/socialpredict/backend/migration/migrations).

So if query or index work becomes necessary, the correct direction is:

- identify the hotspot
- validate it against real traffic or credible repro data
- land schema or index changes through migrations

The correct direction is not to invent a long-lived "query optimizer" service layer.

## What Runtime Performance Tuning Should Own

This note should own:

- evidence-driven bottleneck identification
- DB pool and connection-lifecycle tuning at the runtime seam
- targeted query and index follow-up when a real hotspot is known
- proxy-versus-app ownership of compression behavior
- explicit separation between business metrics and operational performance signals

## What This Note Should Not Own

This note should not become the home for every future speed idea.

It should explicitly defer:

- generalized cache layers
- Redis rollout
- background jobs or worker pools
- a top-level `performance/` package
- speculative response-compression middleware
- broad profiling or benchmark programs
- platform-wide autoscaling plans

Those topics now belong either in [12-database-caching.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/12-database-caching.md), [13-background-jobs.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/13-background-jobs.md), or [FUTURE/06-long-term-performance-optimization.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/06-long-term-performance-optimization.md).

## Near-Term Sequencing

The design-plan-aligned performance direction is:

1. Treat the April 30, 2026 liveness/readiness work as complete for the serving path.
2. Make operational signals and failure posture clearer through the logging and monitoring notes.
3. Tune the existing env-backed `sql.DB` pool and connection-lifecycle settings against measured behavior.
4. Measure real hotspots.
5. Make targeted query or index changes only where the evidence justifies them.
6. Consider caching later, and only after correctness and runtime safety are materially stronger.

## Open Questions

- Which live queries actually become hotspots once runtime readiness and observability are stronger
- Whether pool tuning alone resolves the first real production bottlenecks
- Which read paths, if any, later justify caching rather than pure DB/runtime tuning
- Whether response-size pressure is already materially addressed by nginx gzip at the edge
