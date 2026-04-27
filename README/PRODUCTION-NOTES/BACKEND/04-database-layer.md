---
title: Database Layer
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-26T04:03:07Z
updated_at_display: "Sunday, April 26, 2026 at 4:03 AM UTC"
update_reason: "Replace the older greenfield DAL-oriented plan with current-state-first database architecture guidance grounded in the live backend."
status: active
---

# Database Layer

## Update Summary

This note was updated on Sunday, April 26, 2026 to replace an older greenfield `database/` plus `repository/` plus `dal/` plan with guidance that matches the current SocialPredict backend, the active design-plan posture, and the high-availability or fault-tolerance objective.

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

This means the backend is not starting from a blank slate.

Important current limitations:

- DB config is environment-driven but narrow
- `SSLMode` defaults to `disable`
- pool sizing and connection lifetime are not configured yet
- the runtime seam still stores a shared fallback handle through `SetDB` and `GetDB`

### Startup ownership is still too broad

The main startup path in [main.go](/workspace/socialpredict/backend/main.go) currently does all of the following in every process:

- load DB config
- open DB
- wait for DB ping success
- run migrations
- load config service
- seed admin user
- seed homepage
- start serving

That is simple for a single local process, but it is too broad for HA replica startup because every instance is performing shared-state startup work rather than only local process startup work.

### Readiness exists at startup but not as a serving contract

The backend already does startup DB pinging in [seed.go](/workspace/socialpredict/backend/seed/seed.go):

- `EnsureDBReady`

But the live infrastructure health route in [server.go](/workspace/socialpredict/backend/server/server.go) currently returns a hard-coded `200 ok` from `/health`.

So the backend currently has:

- startup DB reachability gating
- but no strong post-startup readiness contract tied to DB condition

That is not enough for orchestrated HA operation.

### Migrations already exist, but HA discipline is weak

The backend already has a migration registry and schema history model in [migrate.go](/workspace/socialpredict/backend/migration/migrate.go):

- `registry`
- `SchemaMigration`
- ordered application via `sortedRegistryIDs`

The current migration runner is therefore real. However, the HA posture is weak because:

- every replica currently calls `MigrateDB`
- there is no explicit cross-replica writer discipline shown here
- `MigrateDB` falls back to `AutoMigrate` if no migrations are registered
- `main.go` logs migration failure as a warning instead of failing closed

For production HA, that is too permissive.

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

A practical near-term target is:

1. one startup writer performs migrations
2. one startup writer performs one-time seed operations if they remain startup-owned
3. request-serving replicas wait for safe startup state or remain unready

This note does not lock the mechanism. That mechanism could later be:

- a dedicated migration job
- a leader-only startup mode
- an explicit DB-backed lock or equivalent serialization strategy

But the architectural requirement is clear: shared startup writes must not be treated as every-replica default behavior.

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

### Concurrency control

Concurrency control here is about protecting writes against the same primary system of record.

Typical mechanisms might later include:

- row-level locking
- optimistic locking
- conditional atomic SQL updates

The exact mechanism is a later implementation choice. The architectural requirement is to choose one deliberately for balance-sensitive and position-sensitive flows.

## Runtime DB Hardening Direction

The live runtime seam needs hardening, not replacement.

That hardening should include:

- max open connection ownership
- max idle connection ownership
- connection lifetime ownership
- idle connection lifetime ownership
- explicit SSL or TLS mode ownership
- startup ping and readiness policy
- eventual graceful shutdown of the underlying `sql.DB`

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
