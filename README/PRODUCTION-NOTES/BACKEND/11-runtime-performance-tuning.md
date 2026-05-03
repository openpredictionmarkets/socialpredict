---
title: Runtime Performance Tuning
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-03T00:08:00Z
updated_at_display: "Sunday, May 03, 2026 at 12:08 AM UTC"
update_reason: "Close WAVE11 with the measured DB pool tuning stop-and-review outcome and defer unproven performance programs."
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

Startup now also logs the effective pool posture with the `db_pool.configured` event. That log line includes only normalized pool and lifecycle values plus TLS posture; it must not include passwords, full DSNs, hosts, database names, or usernames.

The current owned knobs are:

| Runtime setting | Env vars | Default | Notes |
| --- | --- | --- | --- |
| Max open DB connections | `DB_MAX_OPEN_CONNS`, `POSTGRES_MAX_OPEN_CONNS` | `25` | Process-local ceiling passed to `sql.DB.SetMaxOpenConns`; tune with database capacity and replica count in mind. |
| Max idle DB connections | `DB_MAX_IDLE_CONNS`, `POSTGRES_MAX_IDLE_CONNS` | `5` | Process-local idle pool size passed to `sql.DB.SetMaxIdleConns`; should usually stay no higher than max open connections. |
| Connection max lifetime | `DB_CONN_MAX_LIFETIME`, `POSTGRES_CONN_MAX_LIFETIME` | `30m` | Go duration passed to `sql.DB.SetConnMaxLifetime`; use to recycle long-lived connections before infrastructure-side expiry. |
| Connection max idle time | `DB_CONN_MAX_IDLE_TIME`, `POSTGRES_CONN_MAX_IDLE_TIME` | `5m` | Go duration passed to `sql.DB.SetConnMaxIdleTime`; use to shed idle connections during quieter periods. |
| TLS requirement | `DB_REQUIRE_TLS`, `POSTGRES_REQUIRE_TLS` | production-like envs default to `true`; other envs default to `false` | Validated with `DB_SSLMODE` or `POSTGRES_SSLMODE`; logged as posture only, not as a connection string. |

Unparseable integer or Go-duration values fall back to the defaults above. Negative pool and lifetime values normalize to zero before they are applied to `sql.DB`; when max idle connections exceeds a positive max open connection value, the effective runtime posture is capped to match the `database/sql` pool behavior.

For a single local or development instance, the defaults are intentionally modest and explicit enough to avoid hidden driver behavior. For multiple app replicas, treat max open connections as a per-process value and budget total possible connections across all replicas before increasing it. For production-like deployments, prefer evidence from readiness failures, database wait pressure, and request latency before changing these values; do not use this note as a reason to add cache tiers, query rewriters, or a new performance subsystem.

### Initial deployment defaults from the measurement seam

WAVE11-PERF-002 added a runtime-local DB pool wait measurement seam and exposed process-local SQL pool counters on `/ops/status` under `dbPool`. The targeted benchmark compared one held connection plus one waiter with two pool postures:

| Measured posture | Result |
| --- | --- |
| `DB_MAX_OPEN_CONNS=1` | Saturated: `1.000 pool_waits/op` and about `1 ms` to `2 ms` of pool wait per operation in short local benchmark runs. |
| `DB_MAX_OPEN_CONNS=2` | Headroom: `0 pool_waits/op` and `0 pool_wait_ns/op` for the same held-connection scenario. |

That evidence is enough to reject a one-connection pool for the current backend runtime, but it is not evidence that the service needs a larger pool than the existing runtime defaults. The current initial posture is therefore to make the already-owned defaults explicit and reversible in the environment surfaces:

| Deployment shape | Encoded defaults | Why |
| --- | --- | --- |
| Local development through `backend/.env.dev` | `DB_MAX_OPEN_CONNS=25`, `DB_MAX_IDLE_CONNS=5`, `DB_CONN_MAX_LIFETIME=30m`, `DB_CONN_MAX_IDLE_TIME=5m` | Keeps local behavior aligned with runtime defaults and well above the measured one-connection saturation point without changing code defaults. |
| Production compose `backend-startup-writer` | `${DB_MAX_OPEN_CONNS:-25}`, `${DB_MAX_IDLE_CONNS:-5}`, `${DB_CONN_MAX_LIFETIME:-30m}`, `${DB_CONN_MAX_IDLE_TIME:-5m}` | Keeps the startup writer on the same conservative process-local posture while allowing operators to lower or raise it through `.env` without editing compose. |
| Production compose request-serving `backend` | `${DB_MAX_OPEN_CONNS:-25}`, `${DB_MAX_IDLE_CONNS:-5}`, `${DB_CONN_MAX_LIFETIME:-30m}`, `${DB_CONN_MAX_IDLE_TIME:-5m}` | Keeps the serving process explicit and overrideable; total possible DB connections remain the per-process max multiplied by active backend containers. |

The tradeoff is intentional: these values avoid the measured saturation failure mode, preserve enough headroom for normal request bursts, and avoid increasing total database connection pressure before production traffic shows a need. They are also easy to reverse because operators can override `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, or `DB_CONN_MAX_IDLE_TIME` in the deployment environment.

Revisit this posture when `/ops/status.dbPool.waitCount` increases during normal traffic, `waitDurationNanoseconds` grows alongside request latency or readiness failures, or database-side connection limits make the per-process budget too high for the number of backend containers. Raise `DB_MAX_OPEN_CONNS` only when wait pressure is visible and Postgres has capacity for the new replica-adjusted total; lower it when database connection pressure appears without corresponding pool waits. Revisit lifetimes when infrastructure closes long-lived connections, Postgres restart/failover behavior leaves stale connections visible, or `maxLifetimeClosedConnections` grows in a way that correlates with request errors.

That means the missing piece is not ownership or initial deployment defaults. The missing piece is observing whether these explicit per-process values create real wait pressure under production-like traffic and deciding whether any later hotspot needs query or index work.

### WAVE11 stop-and-review outcome

The first measured performance hotspot was the runtime DB pool wait seam, not an application query. The WAVE11-PERF-002 benchmark showed that a one-connection pool turned one held connection plus one waiter into repeatable pool contention, while a two-connection pool removed the waits for the same controlled workload. WAVE11-PERF-003 then encoded the existing `25` max-open, `5` max-idle, `30m` lifetime, and `5m` idle-time defaults into local development and production compose surfaces without changing the runtime code defaults.

That tuning moved the first bottleneck for this wave: the measured one-connection pool saturation case is no longer represented by the active deployment defaults. There is no current measurement showing that a query, index, cache, worker, compression middleware, or broader performance-platform change is the next bottleneck.

Do not queue a query or index migration from this wave alone. The next migration-owned performance seam should be named only when `/ops/status.dbPool.waitCount` or `waitDurationNanoseconds` rises under normal or production-like traffic and the correlated request path identifies a concrete slow query or missing index. At that point, the follow-up should be a specific repository query and timestamped migration, not a runtime query-optimizer package.

Caching and Redis work remain deferred to [12-database-caching.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/12-database-caching.md). Background jobs and worker topology remain deferred to [13-background-jobs.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/13-background-jobs.md). App-owned compression middleware, profiling programs, load-testing programs, autoscaling work, and broader performance-platform ideas remain deferred to [FUTURE/06-long-term-performance-optimization.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/FUTURE/06-long-term-performance-optimization.md).

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
