---
title: Database Caching
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-26T04:03:07Z
updated_at_display: "Sunday, April 26, 2026 at 4:03 AM UTC"
update_reason: "Create a lower-priority starter draft that explicitly defers caching until runtime DB ownership and correctness are stronger."
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

## Executive Direction

If SocialPredict later introduces caching, it should do so as a supporting optimization layer rather than as a replacement for the system of record.

That means:

1. Postgres remains the authoritative system of record for balances, bets, markets, and resolutions
2. Caching is introduced only after DB ownership, startup safety, and accounting-sensitive correctness are explicit
3. Redis, if introduced, should be treated as a supporting runtime system for selected concerns rather than as a financial source of truth
4. Cache policy should be selective and use-case-specific, not a blanket "cache everything" rule

## Why This Is Deferred

The live backend still has more urgent problems:

- startup ownership is too broad
- migration safety is too permissive
- readiness semantics are too weak
- some request-boundary DB access still bypasses consistent ownership
- accounting-sensitive transaction boundaries still need clearer architecture

Adding caching before those concerns are stabilized would risk making an incorrect or operationally unsafe system faster instead of making it safer.

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

- clearer runtime DB ownership
- stronger startup-writer semantics
- stronger readiness versus liveness behavior
- clearer atomic boundaries for accounting-sensitive writes
- basic performance evidence showing where real read hotspots exist

## Open Questions For Later

- Which read paths are actually expensive enough to benefit from caching
- Whether Redis is the right first caching or coordination tool for this backend
- What cache invalidation strategy would be acceptable for market and leaderboard-style reads
- Which cache surfaces are safe to tolerate as eventually consistent
- Whether coordination use cases such as rate limiting or distributed locks should be considered before read caching

## Explicit Do-Not-Do List

- Do not treat caching as a substitute for transaction correctness
- Do not treat Redis as the source of truth for balances or bets
- Do not add caching before the runtime/bootstrap DB seam is clearer
- Do not hide weak ownership or weak atomicity behind a cache
