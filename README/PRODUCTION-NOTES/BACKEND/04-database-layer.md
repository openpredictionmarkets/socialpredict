---
title: Database Layer
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-14T10:23:25Z
updated_at_display: "Thursday, May 14, 2026 at 10:23 AM UTC"
update_reason: "Document the implemented runtime DB contract, packaged DB TLS boundary, and startup-writer guarantees."
status: active
---

# Database Layer

## Update Summary

This note was updated on Sunday, April 26, 2026 to replace an older greenfield `database/` plus `repository/` plus `dal/` plan with guidance that matches the current SocialPredict backend, the active design-plan posture, and the high-availability or fault-tolerance objective.

On Thursday, May 14, 2026, this note was updated from the future baseline
triage to make the current production topology contract explicit. Runtime DB
configuration, TLS/SSL validation, pool lifecycle, readiness pinging, close
behavior, readiness drain, and startup-writer posture are documented here as
implemented deployment boundaries, not generic database preferences.

| Topic | Prior to April 26, 2026 | After April 26, 2026 |
| --- | --- | --- |
| Core framing | Treated the database layer as a new technical stack to build from scratch | Treats database architecture as runtime/bootstrap ownership, repository edge ownership, and explicit accounting-sensitive transaction boundaries |
| Organizing principle | Proposed top-level `database/`, `repository/`, and `dal/` trees | Extends the live `internal/app/runtime`, `internal/app/container`, `internal/repository`, and `internal/domain` shape |
| Current-state accuracy | Assumed bootstrap, repository, health, and migration structure were still mostly missing | Recognizes that DB bootstrap, repositories, startup pinging, and migration registry already exist, but they are operationally inconsistent |
| DAL posture | Assumed a unified data-access facade should become central | Rejects a new central DAL and keeps repositories as edge translators rather than turning data access into the architecture center |
| Startup posture | Treated migrations and startup writes as implementation details | Treats startup ownership, migration serialization, and seed ownership as first-class HA concerns |
| Consistency posture | Discussed generic transaction helpers and retry language | Focuses on explicit atomic boundaries for accounting-sensitive flows and rejects generic retry for money-moving writes |
| Health posture | Claimed health checks were absent | Recognizes existing startup DB pinging and `/health`, but requires a stronger readiness/liveness split |
| Caching posture | Left optimization and persistence concerns blurred together | Defers caching and Redis-related work until after correctness, runtime ownership, and transaction boundaries are clear |

## Executive Direction

SocialPredict should treat the database layer as a combination of:

1. runtime/bootstrap ownership of DB configuration, connection lifecycle, readiness semantics, migration invocation, and startup sequencing
2. repository ownership of persistence translation at the boundary between Postgres and domain or application workflows
3. explicit transaction and concurrency policy for accounting-sensitive use cases such as placing bets, selling positions, and resolving markets

The backend direction is:

1. Keep DB bootstrap under `internal/app/runtime` and startup orchestration under `main.go`
2. Keep the application container as the composition root that wires repositories and services
3. Keep repositories as edge translators and adapters, not as generic CRUD facades over `models.*`
4. Treat shared startup mutation such as migrations and one-time seed writes as single-writer work, not as per-replica default behavior
5. Move toward fail-closed startup or unready startup on schema incompatibility rather than warning-only continuation
6. Introduce explicit atomic transaction boundaries and concurrency control for accounting-sensitive flows
7. Remove remaining handler-level raw DB access and residual process-global DB fallback over time
8. Defer caching and Redis-related work until correctness and ownership are explicit

For a high-availability, fault-tolerant, enterprise-ready system, the backend should prefer:

- stateless app replicas over one primary relational system of record
- explicit runtime ownership of DB lifecycle and health semantics
- fail-fast or unready behavior for unsafe startup states
- one startup writer for migrations and other shared startup mutations
- atomic write behavior for economically sensitive flows where practical
- observability around DB availability and readiness after ownership is clear
- caching only after the correctness model is trusted

This note explicitly rejects creating a second architecture in the form of new top-level `database/`, `repository/`, or `dal/` subsystems.

## Why This Matters

The older database note was written like a greenfield platform plan. The live backend is no longer at that stage.

SocialPredict already has:

- explicit runtime DB bootstrap helpers
- an application container and composition root
- repository packages under `internal/repository`
- a migration registry and schema history table
- startup DB pinging

The real problems are now different:

- every replica still behaves like a startup coordinator
- migration failure handling is too permissive
- readiness after startup is too weak
- some handlers still bypass service and repository seams
- accounting-sensitive write flows still need clearer atomicity and concurrency rules

If those issues are not fixed, then the system can become fast or feature-rich while still being operationally unsafe.

## Current Code Snapshot

As of 2026-04-26, the backend already has meaningful database-layer structure, but it is split between good direction and risky transitional behavior.

### Runtime DB bootstrap already exists

DB bootstrap is already owned by [internal/app/runtime/db.go](/workspace/socialpredict/backend/internal/app/runtime/db.go).

The live seam already includes:

- `LoadDBConfigFromEnv`
- `BuildPostgresDSN`
- `PostgresFactory`
- `InitDB`
- `ConfigureDBPool`
- `CheckDBReadiness`
- `CloseDB`
- `SnapshotDBPool`

This means the backend is not starting from a blank slate.

The implemented runtime DB production contract is:

- config is loaded from environment variables and normalized before the DB is
  opened
- SSL mode is validated against the `DB_REQUIRE_TLS` or `POSTGRES_REQUIRE_TLS`
  posture before the Postgres DSN is built
- production-like runtimes default `RequireTLS` to `true` unless the topology
  explicitly overrides it
- `DB_MAX_OPEN_CONNS` defaults to `25`
- `DB_MAX_IDLE_CONNS` defaults to `5` and is clamped so it cannot exceed max
  open connections
- `DB_CONN_MAX_LIFETIME` defaults to `30m`
- `DB_CONN_MAX_IDLE_TIME` defaults to `5m`
- negative pool and duration values are normalized to zero
- the effective pool posture is logged without DSNs or secrets
- readiness uses a bounded SQL `PingContext` through the runtime serving probe
- `/ops/status.dbPool` exposes process-local SQL pool counters for operator
  pressure checks
- shutdown calls `CloseDB` on the underlying SQL pool after the server path
  leaves readiness and drains
- the legacy process-global `SetDB` and `GetDB` bridge remains for tests and
  narrow migration compatibility, not as the production composition model

Important current limitations:

- DB config is still environment-driven and should stay explicit in deployment
  docs
- `DB_REQUIRE_TLS=false` is valid for packaged local Docker Postgres, not a
  general external-production database rule
- runtime pool counters are process-local and should not be documented as
  fleet-wide metrics without a later aggregation design

### Startup ownership is explicit in packaged production compose

The main startup path in [main.go](/workspace/socialpredict/backend/main.go)
still initializes runtime prerequisites in every backend process:

- load DB config
- open DB
- wait for DB ping success
- load config service
- load security config
- load shutdown config
- load startup mutation mode
- start serving

Shared startup DB writes are now role-gated instead of being every-replica
defaults:

- writer mode runs migrations plus startup-owned user and homepage seeds
- non-writer mode verifies registered migrations before serving
- writer migration or seed failure is fatal before readiness opens
- non-writer schema incompatibility is fatal before readiness opens

The packaged production compose file makes that boundary concrete with exactly
one `backend-startup-writer` service using `STARTUP_WRITER=true` and a separate
request-serving `backend` service using `STARTUP_WRITER=false`.

### Readiness now exists as a serving contract

The backend already does startup DB pinging in [seed.go](/workspace/socialpredict/backend/seed/seed.go):

- `EnsureDBReady`

The live infrastructure routes in [server.go](/workspace/socialpredict/backend/server/server.go) now separate liveness and readiness:

- `/health` reports process liveness with body `live`
- `/readyz` reports `ready` only after the readiness gate is open and the
  runtime DB ping succeeds
- `/readyz` returns `503` body `not ready` when the readiness gate is closed or
  the DB ping fails
- shutdown closes readiness before the HTTP server drain window

That is enough for the current packaged compose and public deploy verification
baseline. It is not a complete monitoring platform or a proof of business
correctness.

### Migrations already exist, but HA discipline is weak

The backend already has a migration registry and schema history model in [migrate.go](/workspace/socialpredict/backend/migration/migrate.go):

- `registry`
- `SchemaMigration`
- ordered application via `sortedRegistryIDs`

The current migration runner is therefore real. The packaged HA posture is now
stronger because:

- only the explicit startup writer runs `MigrateDB` in production compose
- request-serving backends call migration verification before serving
- migration or verification failure is fatal before readiness opens
- frontend and nginx startup are health-gated on the non-writer backend's
  readiness

The remaining non-packaged topology question is the coordination mechanism. A
dedicated migration job or DB-backed advisory lock may replace compose
sequencing later, but only after a design-plan update names that mechanism.

### Repository architecture already exists

The backend already has repositories under:

- [internal/repository/users](/workspace/socialpredict/backend/internal/repository/users/repository.go)
- [internal/repository/markets](/workspace/socialpredict/backend/internal/repository/markets/repository.go)
- [internal/repository/bets](/workspace/socialpredict/backend/internal/repository/bets/repository.go)

Those repositories are already more valuable than the old note assumed. They are not blank CRUD wrappers waiting to be invented from scratch. They are part of the live architectural direction.

That means the next work is not "implement repository pattern." The next work is:

- keep using repositories as edge translators
- finish removing direct DB bypasses
- tighten repository and transaction behavior for correctness-sensitive workflows

### Remaining handler-level DB leaks still exist

The live backend is not fully service-backed yet.

Examples of remaining raw `*gorm.DB` request-boundary usage include:

- [StatsHandler](/workspace/socialpredict/backend/handlers/stats/statshandler.go)
- [AddUserHandler](/workspace/socialpredict/backend/handlers/admin/adduser.go)
- route wiring in [server.go](/workspace/socialpredict/backend/server/server.go) that still passes raw DB handles into some handler families

This matters because the goal is not only to "have repositories somewhere." The goal is to converge on consistent ownership.

### Accounting-sensitive write consistency remains under-specified

The backend still has money-moving or economically sensitive workflows that are not yet documented as explicit atomic units of work.

The clearest live example is the bet flow in [bet_support.go](/workspace/socialpredict/backend/internal/domain/bets/bet_support.go), which currently follows a compensation-style pattern:

- change user balance
- write the bet
- if later work fails, try to undo earlier work

That is weaker than one atomic DB transaction because:

- the undo can also fail
- overlapping requests can interleave
- partial commit can leave economic state inconsistent

This note does not fully redesign those flows, but it makes the architectural direction explicit: accounting-sensitive workflows need clearer atomic boundaries and concurrency control.

## What The Database Layer Should Own

### Runtime/bootstrap ownership

Runtime/bootstrap should own:

- DB configuration loading
- DSN construction
- pool setup
- SSL or TLS mode ownership
- connection lifecycle and shutdown
- startup DB reachability checks
- readiness and liveness semantics
- migration invocation and startup sequencing
- the rule for whether a process may become ready

### Repository edge ownership

Repositories should own:

- persistence translation
- query shaping at the persistence edge
- mapping between legacy persistence models and domain or boundary-friendly types
- transaction-scoped persistence behavior when a use case explicitly requires it

Repositories should not become the center of the architecture. They are supporting adapters.

### Application/use-case ownership

Application or domain workflows should own:

- which use cases require atomic commit
- which writes must succeed or fail together
- where transaction boundaries begin and end
- how concurrency is controlled for balance and position state

### What the database layer should not own

The database layer should not become:

- a second architecture beside the existing container and repository structure
- a giant DAL facade
- a generic retry platform for accounting-sensitive writes
- a caching-first optimization project
- a replacement for business policy design

## Startup Ownership and Replica Model

The intended operational model should be:

- multiple stateless app replicas
- one primary relational system of record
- explicit single-writer startup semantics for shared startup mutations

That means:

- app replicas should be interchangeable request-serving processes
- not every replica should own migrations and one-time seeds by default
- a schema incompatibility should prevent readiness
- warning-only migration failure is not a strong production posture

A practical near-term target is now implemented for packaged compose:

1. one startup writer performs migrations
2. one startup writer performs one-time seed operations if they remain startup-owned
3. request-serving replicas wait for safe startup state or remain unready

The current packaged startup-failure posture is fail-closed:

- writer mode runs migrations and startup-owned seeds
- writer startup mutation failure is fatal before readiness opens
- non-writer mode verifies registered migrations before serving
- non-writer verification failure is fatal before readiness opens
- request-serving capacity comes from non-writer backends, not by scaling the
  startup-writer service

The next non-packaged topology rule is deployment-only exactly-one-writer until
a stronger mechanism is deliberately selected. That means every deployed
topology must identify exactly one startup mutation actor before serving
traffic, and request-serving replicas must run with `STARTUP_WRITER=false`.

This note does not lock the future mechanism. That mechanism could later be:

- a dedicated migration job
- a leader-only startup mode
- an explicit DB-backed lock or equivalent serialization strategy

But the architectural requirement is clear: shared startup writes must not be treated as every-replica default behavior.

## DB TLS And Topology Boundary

The packaged staging and production compose topology uses a local Docker
Postgres service on the internal Docker network. For that topology,
`./SocialPredict install -e production` writes `DB_REQUIRE_TLS=false` so the
backend accepts the local in-container `sslmode=disable` connection.

That exception must not be copied into external database topologies by default.
If operators replace the packaged local Postgres service with an external
production database, they must make these values explicit:

- `DB_REQUIRE_TLS`
- `DB_SSLMODE`
- the equivalent `POSTGRES_REQUIRE_TLS` or `POSTGRES_SSLMODE` only if those
  names are the chosen operator interface

External production databases should normally use a TLS-satisfying SSL mode
such as `require`, `verify-ca`, or `verify-full` when `DB_REQUIRE_TLS=true`.
Any choice to use `disable`, `allow`, or `prefer` in an external production
database topology requires design-plan review because it changes the production
security posture.

HostOps may inspect or orchestrate hosts, but it must not become the owner of
app runtime DB semantics. Runtime DB policy belongs in `./SocialPredict`,
deployment documentation, and the design plan.

## Transaction and Concurrency Direction

SocialPredict has accounting-sensitive flows. That means the architecture must be explicit about transaction policy.

### What needs to improve

The backend should define clear units of work for use cases such as:

- place bet
- sell position
- resolve market

For each such flow, the backend should decide:

- which writes must commit together
- which writes may be eventual
- which repository operations need a shared transaction scope
- which rows or records need concurrency protection

### What this note rejects

This note rejects the idea that later compensation plus metrics is a sufficient correctness model for economically sensitive writes.

Metrics are useful for diagnosis. They are not a substitute for atomic correctness.

### What this note allows

This note allows narrow, explicit transaction-scoped repository binding or unit-of-work behavior where the use case truly needs it.

It does not require a global DAL or generic transaction wrapper for every workflow.

### WAVE04-DB-005 transaction inventory

The first migrated accounting-sensitive workflow is place bet.

For place bet, the atomic unit of work is:

- read the betting user's current balance through a transaction-bound user adapter
- check prior participation for the initial-bet fee rule
- debit the user's account balance for bet amount plus applicable fees
- insert the bet row

Those steps now commit or roll back together. The transaction starts in the bets repository adapter and ends when the callback returns. There is no generic retry or circuit breaker around this money-moving path.

Primary-DB concurrency protection for place bet is the betting user's row. The transaction-bound user adapter applies `FOR UPDATE` when running on Postgres before computing and updating the balance. SQLite-backed tests skip that dialect-specific lock clause but still verify rollback behavior.

Remaining accounting-sensitive transaction surface:

- Sell position still uses the older compensation-style `CreditSale` path. It needs one transaction covering position/bet-history reads, balance credit, sale bet insert, and position-sensitive concurrency protection.
- Market resolution still needs an explicit transaction boundary for resolution state, payout/accounting writes, and prevention of overlapping resolution or betting writes.
- Market-resolution interaction with place bet is still a surface to harden because the current place-bet transaction protects the user balance row, not the market row or a resolution gate.

Caching and Redis remain deferred. They are not part of the source-of-truth accounting model for these workflows.

### Concurrency control

Concurrency control here is about protecting writes against the same primary system of record.

Typical mechanisms might later include:

- row-level locking
- optimistic locking
- conditional atomic SQL updates

The exact mechanism is a later implementation choice. The architectural requirement is to choose one deliberately for balance-sensitive and position-sensitive flows.

## Runtime DB Hardening Direction

The live runtime seam has been hardened enough to be the current baseline, and
future work should continue improving it rather than replacing it.

The current baseline includes:

- max open connection ownership
- max idle connection ownership
- connection lifetime ownership
- idle connection lifetime ownership
- explicit SSL or TLS mode ownership
- startup ping and readiness policy
- graceful shutdown of the underlying `sql.DB`

The key point is ownership:

- these concerns belong in runtime/bootstrap
- they do not belong in handlers
- they do not belong in app-policy config
- they do not require inventing a separate standalone "database service" inside the monolith

## Caching and Redis Are Later

Caching is a later problem.

This database-layer note explicitly treats caching and Redis-related work as lower priority than:

- startup ownership
- migration safety
- readiness semantics
- handler and repository ownership consistency
- accounting-sensitive transaction correctness

Redis is not a replacement for Postgres or for source-of-truth accounting state.

If Redis is introduced later, it should be treated as a supporting system for concerns such as:

- cacheable public reads
- short-lived coordination state
- sessions or rate-limiting if needed

It should not become the first solution to unresolved correctness or ownership problems.

See the later starter draft note [12-database-caching.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/12-database-caching.md) for the intentionally deferred caching direction.

## Near-Term Sequencing

The practical sequence for the database layer is:

1. clarify runtime/bootstrap ownership in the production notes and design plan
2. tighten migration failure posture and startup-writer semantics
3. define readiness versus liveness around actual DB condition
4. remove remaining raw request-boundary DB access where practical
5. define explicit transaction and concurrency policy for accounting-sensitive workflows
6. only then consider caching and Redis-related optimization work

## Open Questions

This note leaves several questions open for later implementation and design-plan updates:

- what exact mechanism should enforce single-writer startup semantics
- whether one-time seed writes should remain startup-owned or move to controlled administration or bootstrap jobs
- which exact workflows must be atomic in one transaction
- which concurrency-control mechanism fits SocialPredict's balance and market flows best
- what concrete readiness conditions should gate request traffic
- whether `SetDB` and `GetDB` can be reduced to test-only compatibility and later removed

## Explicit Do-Not-Do List

- Do not create new top-level `database/`, `repository/`, or `dal/` subsystems
- Do not redefine repository contracts around generic `models.*` CRUD
- Do not treat warning-only migration failure as a production HA end state
- Do not rely on compensation and metrics as the primary correctness strategy for accounting-sensitive writes
- Do not add generic retry or circuit-breaker behavior around money-moving writes without separate idempotency and accounting design
- Do not make caching or Redis a prerequisite for fixing ownership and correctness
