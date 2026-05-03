---
title: Database Caching
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-05-01T11:50:49Z
updated_at_display: "Friday, May 01, 2026 at 11:50 AM UTC"
update_reason: "Refresh the deferred caching draft against runtime DB ownership, readiness, and startup-writer seams landed on upstream main at 051aac6."
status: draft
---

# Database Caching

## Starter Draft Status

This is a lower-priority starter draft.

It exists to capture the intended direction for future caching work without confusing that work with the higher-priority database-layer concerns in [04-database-layer.md](/workspace/socialpredict/README/PRODUCTION-NOTES/BACKEND/04-database-layer.md).

The current posture is explicit:

- correctness comes before caching
- runtime/bootstrap DB ownership comes before caching
- accounting-sensitive transaction correctness comes before caching
- caching is not a prerequisite for rewriting the current database-layer note

This draft was refreshed on Friday, May 01, 2026 against upstream `main` at `051aac6b2fefa5634b8c98cc38caf52acf0043a9`. Runtime DB config, pool and lifetime policy, readiness checks, and startup-writer gating are now code reality. Caching remains deferred because those seams still need operational follow-through, broader transaction coverage, and evidence of real read hotspots.

## Executive Direction

If SocialPredict later introduces caching, it should do so as a supporting optimization layer rather than as a replacement for the system of record.

That means:

1. Postgres remains the authoritative system of record for balances, bets, markets, and resolutions
2. Caching is introduced only after DB ownership, readiness, and startup-writer seams are operationally trusted, and after accounting-sensitive correctness is explicit beyond the place-bet path
3. Redis, if introduced, should be treated as a supporting runtime system for selected concerns rather than as a financial source of truth
4. Cache policy should be selective and use-case-specific, not a blanket "cache everything" rule

## Why This Is Deferred

The live backend still has more urgent problems:

- runtime DB ownership, readiness probes, and startup-writer gating now exist, but deployment still needs to enforce them intentionally
- some persistence-placement and legacy compatibility seams still need cleanup before a cache should obscure them further
- accounting-sensitive transaction and concurrency policy still needs broader coverage outside the place-bet path
- the runtime still lacks the hotspot evidence and operational metrics that would justify a first cache target

Adding caching before those concerns are settled would risk making an incorrect or operationally unclear system faster instead of making it safer.

## Current Code Snapshot

### Runtime DB ownership is now explicit

The backend already owns runtime DB configuration in [db.go](/workspace/socialpredict/backend/internal/app/runtime/db.go), including:

- DSN normalization
- TLS and `sslmode` posture validation
- pool sizing and connection lifetime settings
- readiness ping behavior
- shutdown of the underlying `sql.DB`

That is important because caching should no longer be justified as a substitute for missing DB runtime ownership. That ownership already exists.

### Startup-writer and readiness behavior are now part of the baseline

The backend now has:

- `STARTUP_WRITER` mode in [runtime/startup_mutation.go](/workspace/socialpredict/backend/internal/app/runtime/startup_mutation.go)
- startup mutation and verification behavior in [startup_mutation.go](/workspace/socialpredict/backend/startup_mutation.go)
- liveness and readiness probes in [server.go](/workspace/socialpredict/backend/server/server.go)

Those changes reduce one class of uncertainty, but they do not make caching current. They mean caching should wait until deployment actually uses those runtime seams intentionally.

### There is still no live cache runtime

The active backend does not currently have:

- a Redis service in the production compose topology
- a backend cache package or cache-aside abstraction
- cache invalidation ownership in handlers, domain services, or repositories
- a metrics surface for cache hit, miss, or staleness behavior

That means the note should stay explicit that caching is still a future optimization layer, not current backend reality.

## What Redis Is And Is Not

Redis is an in-memory data store commonly used for:

- caching
- session storage
- rate limiting
- distributed locks
- pub/sub
- queues and short-lived coordination state

Redis is not the same thing as:

- Postgres
- DB connection pooling
- a source-of-truth accounting ledger

For SocialPredict, Redis should be thought of as a potential future support system, not as the replacement for primary relational storage.

## Candidate Future Cache Targets

If caching is introduced later, the most plausible early candidates are read-heavy and non-authoritative surfaces such as:

- market list or search responses
- leaderboard snapshots
- selected setup or frontend configuration reads
- public or semi-public derived views that are expensive to compute repeatedly

## Candidate Non-Targets For Early Caching

The following should not be early cache targets:

- balances
- bet placement correctness
- market resolution correctness
- payout correctness
- any source-of-truth accounting state

Those areas depend on stronger transaction and concurrency guarantees than caching can provide.

## Relationship To Connection Pooling

Caching is a different concern from DB connection pooling.

These concerns should remain separate:

- Postgres is the system of record
- runtime/bootstrap owns DB connection pooling and lifecycle
- Redis or another cache would later support read optimization or coordination

## Preconditions Before Caching Work

Before the backend should prioritize caching, it should first have:

- operationally verified runtime DB ownership
- an enforced startup-writer deployment posture
- proven readiness and liveness consumption in deployment health policy
- clearer atomic boundaries for accounting-sensitive writes
- basic performance evidence showing where real read hotspots exist
- clear invalidation ownership for whichever first read path gets cached

## Open Questions For Later

- Which read paths are actually expensive enough to benefit from caching
- Whether Redis is the right first caching or coordination tool for this backend
- What cache invalidation strategy would be acceptable for market and leaderboard-style reads
- Which cache surfaces are safe to tolerate as eventually consistent
- Whether coordination use cases such as rate limiting or distributed locks should be considered before read caching

## Explicit Do-Not-Do List

- Do not treat caching as a substitute for transaction correctness
- Do not treat Redis as the source of truth for balances or bets
- Do not add caching before the runtime/bootstrap DB seam is operationally proven
- Do not hide weak ownership or weak atomicity behind a cache
