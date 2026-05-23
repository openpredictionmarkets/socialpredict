---
title: Database Layer
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-11T22:15:00Z
updated_at_display: "Monday, May 11, 2026 at 10:15 PM UTC"
update_reason: "Ground database topology, DB TLS posture, and startup-writer ownership in the v3.0.1 production baseline."
status: active
---

# Database Layer

## Update Summary

This note was updated on Sunday, April 26, 2026 to replace an older greenfield `database/` plus `repository/` plus `dal/` plan with guidance that matches the current SocialPredict backend, the active design-plan posture, and the high-availability or fault-tolerance objective.

On Monday, May 11, 2026, this note was corrected against the v3.0.1 production
baseline. The packaged Docker production topology now has one
`backend-startup-writer` service, request-serving backend containers run with
`STARTUP_WRITER=false`, `/readyz` checks database availability after the
readiness gate opens, and `./SocialPredict install -e production` writes
`DB_REQUIRE_TLS=false` for the local Docker Postgres topology. External
production databases still require an explicit operator decision for
`DB_REQUIRE_TLS` and `DB_SSLMODE`.

| Topic | Prior to April 26, 2026 | After April 26, 2026 |
| --- | --- | --- |
| Core framing | Treated the database layer as a new technical stack to build from scratch | Treats database architecture as runtime/bootstrap ownership, repository edge ownership, and explicit accounting-sensitive transaction boundaries |
| Organizing principle | Proposed top-level `database/`, `repository/`, and `dal/` trees | Extends the live `internal/app/runtime`, `internal/app/container`, `internal/repository`, and `internal/domain` shape |
| Current-state accuracy | Assumed bootstrap, repository, health, and migration structure were still mostly missing | Recognizes that DB bootstrap, repositories, startup pinging, migration registry, startup-writer mode, and serving-path readiness now exist, while custom topology and transaction work remain |
| DAL posture | Assumed a unified data-access facade should become central | Rejects a new central DAL and keeps repositories as edge translators rather than turning data access into the architecture center |
| Startup posture | Treated migrations and startup writes as implementation details | Treats startup ownership, migration serialization, and seed ownership as first-class HA concerns |
| Consistency posture | Discussed generic transaction helpers and retry language | Focuses on explicit atomic boundaries for accounting-sensitive flows and rejects generic retry for money-moving writes |
| Health posture | Claimed health checks were absent | Recognizes the landed `/health` liveness and `/readyz` database-readiness split |
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
5. Preserve fail-closed startup or unready startup on schema incompatibility rather than warning-only continuation
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

- non-packaged topologies still need to prove equivalent startup-writer
  enforcement
- migration failure and schema incompatibility must remain fail-closed startup
  concerns
- readiness must stay tied to the runtime database check and startup-writer
  posture
- some handlers still bypass service and repository seams
- accounting-sensitive write flows still need clearer atomicity and concurrency rules

If those issues are not fixed, then the system can become fast or feature-rich while still being operationally unsafe.

## Current Code Snapshot

As of the v3.0.1 production baseline, the backend already has meaningful
database-layer structure, but it still has targeted cleanup and transaction
work remaining.

### Runtime DB bootstrap already exists

DB bootstrap is already owned by [internal/app/runtime/db.go](/workspace/socialpredict/backend/internal/app/runtime/db.go).

The live seam already includes:

- `LoadDBConfigFromEnv`
- `BuildPostgresDSN`
- `PostgresFactory`
- `InitDB`

This means the backend is not starting from a blank slate.

Important current posture:

- DB config is environment-driven but narrow
- `SSLMode` defaults to `disable`
- production runtime requires TLS by default unless `DB_REQUIRE_TLS=false` is
  explicitly supplied
- packaged Docker production installs set `DB_REQUIRE_TLS=false` because their
  Postgres service is local to the compose network and uses `sslmode=disable`
- external production databases must explicitly set `DB_REQUIRE_TLS` and
  `DB_SSLMODE` for their topology
- pool sizing and connection lifetime are runtime-owned through
  `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, and
  `DB_CONN_MAX_IDLE_TIME`
- the runtime seam still stores a shared fallback handle through `SetDB` and `GetDB`

### Startup ownership now has an explicit writer mode

The main startup path in [main.go](/workspace/socialpredict/backend/main.go)
now loads explicit startup mutation mode before shared startup work:

- load DB config
- open DB
- wait for DB ping success
- run migrations only when startup writer mode is enabled
- verify applied migrations when startup writer mode is disabled
- load config service
- seed admin user and homepage only from the startup writer path
- start serving

The packaged production compose file makes that role split concrete:

- `backend-startup-writer` runs with `STARTUP_WRITER=true`
- request-serving `backend` runs with `STARTUP_WRITER=false`
- frontend and nginx wait for the request-serving backend `/readyz` healthcheck

That closes the first broad-startup gap for the packaged topology. Custom
topologies still need to preserve equivalent exactly-one-writer behavior.

### Readiness is now a serving contract

The backend still does startup DB pinging, but it also exposes a serving-path
readiness contract:

- `/health` reports process liveness as `live`
- `/readyz` reports `ready` only after the readiness gate is open and the
  primary database ping succeeds
- `/readyz` reports `not ready` with `503` when the readiness gate is closed or
  the database check fails

This means database availability is no longer only a startup concern. The
runtime readiness path is the deploy and compose traffic-readiness signal.

### Migrations already exist, and packaged startup discipline is explicit

The backend already has a migration registry and schema history model in [migrate.go](/workspace/socialpredict/backend/migration/migrate.go):

- `registry`
- `SchemaMigration`
- ordered application via `sortedRegistryIDs`

The current migration runner is therefore real. In the packaged production
topology, the startup writer owns migration application and request-serving
backends verify applied migrations before opening readiness.

The remaining migration and HA questions are narrower:

- whether non-compose deployments need a dedicated migration job, advisory lock,
  or another exactly-one-writer enforcement mechanism
- how to retire older compatibility behavior safely
- which additional Postgres-backed tests are worth adding when behavior depends
  on real Postgres semantics

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

The place-bet flow now has an explicit transaction baseline, but the backend
still has other money-moving or economically sensitive workflows that are not
yet documented as explicit atomic units of work.

The remaining concern is the old compensation-style pattern in flows such as
sell position: write one side effect, then try to undo it if later work fails.
That is weaker than one atomic DB transaction because the undo can also fail,
overlapping requests can interleave, and partial commit can leave economic state
inconsistent.

This note does not fully redesign those remaining flows, but it keeps the
architectural direction explicit: accounting-sensitive workflows need clear
atomic boundaries and concurrency control.

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

The current packaged production operational model is:

- multiple stateless app replicas
- one primary relational system of record
- explicit single-writer startup semantics for shared startup mutations

That means:

- app replicas should be interchangeable request-serving processes
- not every replica owns migrations and one-time seeds by default
- a schema incompatibility should prevent readiness
- migration failure must fail the writer path and prevent request-serving
  readiness

The packaged compose topology implements this as:

1. `backend-startup-writer` runs with `STARTUP_WRITER=true`
2. the startup writer performs migrations and startup-owned seed operations
3. request-serving `backend` containers run with `STARTUP_WRITER=false`
4. request-serving containers verify applied migrations and remain unready if
   the database is not ready
5. frontend and nginx wait on request-serving backend readiness

This note does not require every self-hosted topology to use Docker Compose, but
it does require equivalent startup ownership. Other mechanisms could be:

- a dedicated migration job
- a leader-only startup mode
- an explicit DB-backed lock or equivalent serialization strategy

The architectural requirement is clear: shared startup writes must not be
treated as every-replica default behavior.

## Packaged Production DB TLS Policy

The runtime default is conservative: production requires database TLS unless an
operator explicitly disables that requirement.

The packaged Docker production topology is the deliberate exception. It runs
Postgres as a local compose service on the internal Docker network, so
`./SocialPredict install -e production` writes:

```text
DB_REQUIRE_TLS=false
```

That allows the local in-container Postgres connection to use the existing
`sslmode=disable` default without being rejected by runtime validation.

Operators who replace packaged local Postgres with an external production
database must make the TLS posture explicit instead of inheriting the compose
default. At minimum, review:

```text
DB_REQUIRE_TLS=true
DB_SSLMODE=verify-full
```

The exact `DB_SSLMODE` value depends on the external database provider and
certificate setup, but `DB_REQUIRE_TLS=false` should be treated as a packaged
local-compose topology setting, not a generic production recommendation.

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

The live runtime seam has a first hardening baseline and needs continued
cleanup, not replacement.

That baseline now includes:

- max open connection ownership
- max idle connection ownership
- connection lifetime ownership
- idle connection lifetime ownership
- explicit SSL or TLS mode ownership
- startup ping and readiness policy
- serving-path readiness tied to database availability

Remaining cleanup includes:

- retiring process-global DB fallback behavior when callers no longer need it
- deciding whether custom non-compose topologies need stronger exactly-one-writer
  enforcement
- adding source-of-truth Postgres checks only for behaviors that SQLite cannot
  model credibly

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

- what exact mechanism should enforce single-writer startup semantics outside
  the packaged compose topology
- whether one-time seed writes should remain startup-owned or move to controlled administration or bootstrap jobs
- which exact workflows must be atomic in one transaction
- which concurrency-control mechanism fits SocialPredict's balance and market flows best
- what additional concrete readiness conditions, if any, should gate request
  traffic beyond startup readiness and database reachability
- whether `SetDB` and `GetDB` can be reduced to test-only compatibility and later removed

## Explicit Do-Not-Do List

- Do not create new top-level `database/`, `repository/`, or `dal/` subsystems
- Do not redefine repository contracts around generic `models.*` CRUD
- Do not treat warning-only migration failure as a production HA end state
- Do not rely on compensation and metrics as the primary correctness strategy for accounting-sensitive writes
- Do not add generic retry or circuit-breaker behavior around money-moving writes without separate idempotency and accounting design
- Do not make caching or Redis a prerequisite for fixing ownership and correctness
