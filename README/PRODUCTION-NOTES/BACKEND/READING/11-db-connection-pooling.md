---
title: DB Connection Pooling Reading
document_type: reading-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-05T04:25:00Z
updated_at_display: "Tuesday, May 5, 2026 at 4:25 AM UTC"
update_reason: "Add book-first reading recommendations for understanding WAVE11 DB connection pooling."
status: active
---

# DB Connection Pooling Reading

This reading note supports
[11-runtime-performance-tuning.md](../11-runtime-performance-tuning.md). It is
not a general database bookshelf. These references are selected because they
explain the production patterns behind WAVE11: app-local SQL connection pools,
database connection budgets, wait pressure, connection lifetime policy,
PostgreSQL capacity limits, and the query behavior that can keep connections
checked out too long.

A DB pool is a managed set of database connections owned by the application
process. Instead of opening a new PostgreSQL connection for every query, the app
borrows a connection from the pool, runs work, and returns it. This reduces
connection setup cost and gives each backend process a hard concurrency budget.
If all allowed connections are busy, new DB work waits; in Go that wait is
visible in `sql.DBStats.WaitCount` and `WaitDuration`.

The selection bias is toward direct topical fit, established publishers or
primary authors, operator credibility, and community reception. Marketplace
ratings drift over time, so this note treats ratings as a secondary signal
rather than the deciding factor.

## Best First Reads

1. [**Database Reliability Engineering**](https://www.oreilly.com/library/view/database-reliability-engineering/9781491925935/)
   by Laine Campbell and Charity Majors, O'Reilly Media, 2017.
   Start here for the production-operations frame. WAVE11 is not just a code
   knob change; it turns database connection usage into an observable runtime
   contract. This book is the best first read for thinking about database
   safety, service levels, capacity, ownership, toil, and why database
   reliability has to be treated as part of application reliability.

2. [**Release It!, 2nd Edition**](https://www.oreilly.com/library/view/release-it-2nd/9781680504552/)
   by Michael T. Nygard, Pragmatic Bookshelf, 2018.
   Read this for the production-systems mindset behind pools as a bounded
   resource. The most useful idea for WAVE11 is that concurrency, back pressure,
   timeouts, and resource saturation need explicit limits and operational
   signals. A DB connection pool is one of those limits.

3. [**Designing Data-Intensive Applications, 2nd Edition**](https://www.oreilly.com/library/view/designing-data-intensive-applications/9781098119058/)
   by Martin Kleppmann and Chris Riccomini, O'Reilly Media, 2026.
   Read this for the broader data-systems foundation: reliability, scalability,
   load, latency, transactions, replication, and the tradeoffs behind database
   architecture. It will not teach SocialPredict's exact Go pool settings, but
   it gives the mental model for why connection pressure, query behavior, and
   database capacity cannot be reviewed in isolation.

4. [**PostgreSQL 16 Administration Cookbook**](https://www.oreilly.com/library/view/postgresql-16-administration/9781835460580/)
   by Gianni Ciolli, Boriss Mejias, Jimmy Angelakos, Vibhor Kumar, Simon Riggs,
   and contributors, Packt Publishing, 2023.
   Read this when you want practical PostgreSQL administration context. The
   useful WAVE11 bridge is server-side configuration, connection access,
   monitoring, troubleshooting, and the DBA view of what an application pool is
   allowed to consume from the database.

5. [**SQL Performance Explained**](https://use-the-index-luke.com/sql/table-of-contents)
   by Markus Winand, self-published, 2012.
   Read this for the query and indexing bridge. Pool waits tell you requests are
   waiting for database connections, but not why those connections are busy. The
   next move is often to find which query, index shape, sort, join, pagination
   pattern, or ORM-generated SQL is keeping connections checked out too long.

6. [**High Performance PostgreSQL for Rails**](https://www.oreilly.com/library/view/high-performance-postgresql/9798888651070/)
   by Andrew Atkinson, Pragmatic Bookshelf, 2024.
   Read this as a web-application PostgreSQL performance book, even though
   SocialPredict is Go rather than Rails. Treat the Rails sections as
   framework-specific and keep the PostgreSQL, observability, indexing,
   migration, and production workflow lessons.

## Primary References

- Go documentation:
  [**Managing connections**](https://go.dev/doc/database/manage-connections).
  This is the direct reference for SocialPredict's app-local pool. Go's
  `sql.DB` owns a pool of active connections, exposes `DB.Stats`, and supports
  `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, and
  `SetConnMaxIdleTime`.
- PostgreSQL documentation:
  [**Connections and Authentication**](https://www.postgresql.org/docs/current/runtime-config-connection.html).
  This is the database-side budget. App pool limits are only half the picture;
  PostgreSQL also has `max_connections`, reserved connection slots, and memory
  implications for high connection counts.
- PgBouncer documentation:
  [**Usage**](https://www.pgbouncer.org/usage).
  PgBouncer is an external PostgreSQL connection pooler. SocialPredict does not
  currently use it, but it is the natural next concept if many app processes or
  short-lived clients create too much connection pressure.

## How This Applies To WAVE11

WAVE11 gives SocialPredict an app-local pool because the backend uses Gorm on
top of Go's `database/sql`, and Go's `*sql.DB` is itself a connection pool. The
implementation proves that path in four places:

- `InitDB` calls `ConfigureDBPool` before returning the runtime DB handle.
- `ConfigureDBPool` obtains the underlying `*sql.DB` with `db.DB()` and applies
  `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, and
  `SetConnMaxIdleTime`.
- `SnapshotDBPool` reads `sqlDB.Stats()` and maps those values into the
  `/ops/status.dbPool` response.
- startup logs emit the `db_pool.configured` event with the effective pool
  posture, without logging DSNs or secrets.

The local WAVE11 measurement then showed the pool behaving like a real bounded
resource:

- a one-connection pool created repeatable waits when one connection was held
  and one request waited
- a two-connection pool removed waits in the same controlled scenario
- the active default, `DB_MAX_OPEN_CONNS=25`, is deliberately above that measured
  saturation case

That evidence justifies rejecting a one-connection pool. It does not prove that
SocialPredict needs a larger pool, a cache, a new query layer, or PgBouncer
today.

The first operational follow-up should be evidence-driven:

- watch `/ops/status.dbPool.waitCount`
- watch `/ops/status.dbPool.waitDurationNanoseconds`
- correlate rising waits with request latency and logs
- only then inspect the specific route, query, transaction, or index that is
  holding connections too long

## What Not To Infer

Do not infer that:

- `DB_MAX_OPEN_CONNS=25` is universally correct for every deployment
- increasing max-open always improves performance
- zero waits in local development proves production will have zero waits
- connection pooling replaces query tuning, indexes, migrations, or database
  capacity planning
- PgBouncer is required before there is evidence of app-level or database-level
  connection pressure

The current SocialPredict posture is intentionally conservative: make the
app-local pool visible, keep defaults explicit, and use observed wait pressure
before changing architecture.
