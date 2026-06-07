---
title: Read Model Caching And Performance
document_type: production-notes
domain: features
author: Patrick Delaney
updated_at: 2026-06-07T00:00:00Z
updated_at_display: "Sunday, June 7, 2026"
update_reason: "Start the feature spec for Postgres/Redis-backed read model caching, materialized market accounting snapshots, and performance-safe display paths."
status: draft
---

# Read Model Caching And Performance

## Purpose

SocialPredict has several expensive read paths that derive statistics, leaderboards, market cards, market charts, and position-like displays from market bet history. These paths are useful for users browsing the site, but they should not force every page view to recompute full market math from raw bets.

This feature defines a safe caching and read-model strategy for display and analytics paths while preserving correctness for order calculation, sale execution, and final settlement.

## Feature Artifact Map

- [11-read-model-caching-performance.md](./11-read-model-caching-performance.md): feature overview, product behavior, and acceptance criteria.
- [DESIGN.md](./DESIGN.md): domain boundaries, cache/read-model strategy, invalidation, Redis/Postgres roles, and correctness constraints.
- [PLAN.md](./PLAN.md): implementation checklist and PR sequencing.

## Problem Framing

Some data is order-critical and must be read from the canonical current state at transaction time. Other data is display-oriented and can tolerate short staleness if the UI communicates freshness appropriately.

The key product distinction is:

```text
Transaction-time decisions = canonical, non-stale, transaction-safe reads
Display, discovery, statistics, and leaderboards = cacheable read models
```

Caching should reduce load, not hide correctness problems.

## Non-Negotiable Correctness Rule

Do not use stale cache values to decide whether a trade, sale, refund, resolution, or payout is valid.

Transaction paths are never cache-driven. Any operation that can create, mutate,
or settle user credits must continue to use canonical tables and
transaction-safe services at transaction time.

Examples:

- placing a bet
- confirming a sale
- calculating transaction-time dust
- debiting or crediting user balances
- resolving/yanking/cancelling a market
- final payout/refund execution
- any logic that can create or destroy user credits

This applies even if a nearby display widget is cached. A cached probability,
volume, leaderboard, or market card can help users browse, but it cannot decide
whether a transaction succeeds.

## Transaction Boundary Guardrails

Market accounting snapshots are display/read-model artifacts only. They must not
be exposed through transaction repository interfaces.

Implementation guardrails:

- buy/sell transaction interfaces must not expose snapshot read or write methods
- market resolution/payout/refund interfaces must not expose snapshot read or write methods
- transaction-time dust is calculated from canonical market, bet, position, and user state
- display dust can be stored in snapshots, but confirmed sales must not read it
- durable snapshots may be refreshed before or after transactions, but cannot determine transaction outcomes

The codebase includes boundary tests that fail if transaction interfaces start
exposing market accounting snapshot methods.

## Critical Decision Matrix

The following table defines which decisions are transaction-critical and which are display/read-model candidates.

| Decision or display area | Critical decision? | Cache/read-model policy | Reason |
|---|---|---|---|
| Place buy order | Yes | Do not use stale cache. Use canonical transaction-safe reads. | Can debit user balance and change market state. |
| Confirm sale order | Yes | Do not use stale cache. Recalculate at execution time. | Can credit user balance, write a negative sale row, and create transaction-time dust. |
| Sale quote preview | No, informational only | May use fresh canonical calculation or short-lived preview data, but final sale must recalculate. | Quote is not settlement; high-volume trading can make previews stale. |
| Transaction-time dust calculation | Yes | Do not use stale cache. Calculate from current position at execution time. | Determines actual sale value, retained dust, and ledger writes. |
| User balance mutation | Yes | Do not use stale cache. Use canonical balance inside transaction. | Can create incorrect debt, credits, or payouts. |
| Market resolution payout | Yes | Do not use display cache. Use canonical payout/read model explicitly approved for settlement. | Final settlement moves credits and must conserve money. |
| Market cancellation/yank refund | Yes | Do not use display cache. Use canonical bet/user state. | Refunds can move credits across many users. |
| Admin mutation actions | Yes | Read fresh before mutating. | Approval/rejection/stewardship/tag changes alter governance state. |
| System financial metrics display | No | Cacheable snapshot, roughly hourly. | Informational dashboard; does not execute orders. |
| Individual user financial metrics display | No | Cacheable authenticated read model, roughly 1-10 minutes depending on widget. | User profile financial summaries, P/L, historical positions, and heavier derived views are display-only and can be refreshed separately from transactions. |
| Global leaderboard display | No | Cacheable snapshot, roughly hourly; paginate. | Ranking is display-oriented and changes less urgently. |
| Market leaderboard display | No | Cacheable snapshot, roughly 10 minutes; paginate. | Informational market widget; does not execute orders. |
| `/markets` discovery page | No | Cacheable page/card read model, roughly 10 minutes. | High-visibility browsing page can tolerate page-level staleness. |
| `/markets/topic/:slug` page | No | Cacheable topic/card read model, roughly 10 minutes. | High-visibility browsing page can tolerate page-level staleness. |
| Pinned market chart cards | No | Cacheable compact chart snapshot, roughly 10 minutes. | Discovery display, not transaction execution. |
| Individual market probability display | No, unless used for order execution | Cache briefly for display, roughly 1 minute; order confirmation must use canonical path. | Users need fresher UI on a market page, but displayed value should not settle trades. |
| Individual market volume/user count display | No | Cache briefly, roughly 1 minute, with freshness metadata. | Informational display only. |
| Market bet table display | No | Paginate; do not cache, or at most poll/refresh around 10 seconds. | Users should see their accepted bet appear quickly after transaction success. |
| Market dust display | No | Cache only as display/read-model dust; never use display dust for transaction-time sale dust. | Dust can be displayed as retained accounting value, but confirmed sales must calculate dust from canonical current position. |

## Cacheable Display Candidates

These are good early candidates because they are read-heavy and not directly responsible for executing trades.

| Area | Candidate data | Suggested freshness |
|---|---|---:|
| System statistics | financial metrics, active volume, user counts, market counts | about 1h |
| Individual user financial metrics | user P/L, resolved/unresolved exposure, historical positions, market-by-market financial summaries | about 1-10m; load on demand and refresh after user-initiated transaction success |
| Global leaderboard | user ranking, profit summaries, resolved market counts | about 1h |
| Market leaderboard | per-market participant ranking | about 10m |
| Market cards | title, status, probability, volume, users, tags, steward, close time | about 10m |
| `/markets` page | discovery cards, pinned market summaries, topic navigation payloads | about 10m |
| `/markets/topic/:slug` page | filtered topic cards and pinned topic markets | about 10m |
| Pinned charts | compact probability history snapshots | about 10m |
| Individual market widgets | probability, volume, user count, compact summary metrics | about 1m |
| Market dust display | retained dust, net volume, volume with dust | about 1m on market detail; about 10m on cards/discovery |
| Bet table display | paginated recent bets, first page only by default | not cached; refresh/poll around 10s if needed |

## Less Cacheable Or Non-Cacheable Paths

| Area | Position |
|---|---|
| Confirm buy/sell order | Do not cache for decision-making. |
| Sale quote | Preview-only and informational; final sale recalculates. |
| User balance | Avoid stale values for transaction decisions; display can poll or show last refreshed timestamp. Complex user financial metrics can be cached separately from spend/settlement checks. |
| Final payout | Do not use display cache as payout truth. |
| Admin mutation actions | Read fresh before mutating. |

## API Shape Recommendation

Prefer separate API boundaries for canonical transaction paths and cached display paths.

Transaction endpoints should remain canonical and never cache-driven:

```text
POST /v0/bet
POST /v0/sell
POST /v0/markets/{id}/resolve
admin mutation endpoints
```

Display/read-model endpoints can be cache-backed and should expose freshness
metadata:

```text
GET /v0/read/markets
GET /v0/read/markets/topic/{slug}
GET /v0/read/markets/{id}/summary
GET /v0/read/markets/{id}/leaderboard
GET /v0/read/system/metrics
GET /v0/read/users/{username}/financial-summary
GET /v0/read/users/{username}/positions
GET /v0/read/leaderboard
```

The exact route names are placeholders. The important design rule is that cached
read endpoints are visibly separate from transaction endpoints. If existing
routes are reused for display, handlers should still call explicit read-model
services rather than sharing transaction services implicitly.

Every cache-backed response should be able to include freshness metadata:

```json
{
  "generatedAt": "2026-06-07T00:00:00Z",
  "cacheTtlSeconds": 600,
  "source": "read_model"
}
```

## Redis And Postgres Roles

Recommended baseline:

| Tool | Role |
|---|---|
| Postgres | Source of truth and durable materialized read-model tables. |
| Redis | Optional short-lived cache for API responses or read-model fragments. |
| Raw bet tables | Audit source of truth for recomputation and validation. |

Postgres read models are useful when the data should survive process restarts and be inspectable. Redis is useful for hot, short-lived responses and avoiding repeated serialization/calculation under traffic.

## Example Read Models

Candidate durable Postgres tables or materialized views:

- `market_accounting_snapshots`
- `market_display_snapshots`
- `market_leaderboard_snapshots`
- `global_leaderboard_snapshots`
- `user_financial_metric_snapshots`
- `system_metrics_snapshots`
- `topic_market_snapshots`

These names are placeholders. The implementation should pick names after reviewing current repository conventions.

## Freshness And User Experience

Cached display values should be transparent enough to avoid confusing users.

Recommended UX conventions:

- show cached metrics without promising exact transaction-time values
- use copy such as `Updated moments ago` for dashboards when appropriate
- keep order confirmation and final transaction responses authoritative
- paginate heavy tables instead of rendering full histories
- prefer lightweight market cards on discovery pages
- keep individual market pages more current than global dashboards

## Initial Scope Recommendation

Start with display-only caches and pagination before touching settlement-adjacent math.

Recommended first slice:

1. Define a market accounting/read-model boundary.
2. Add documentation and tests comparing read-model outputs to full recomputation.
3. Add pagination-first display behavior for expensive lists.
4. Cache `/markets` and `/markets/topic/:slug` card payloads.
5. Cache system statistics and leaderboards.
6. Keep all order calculations and final settlement on canonical paths.

## Acceptance Criteria

- Order and settlement paths do not use stale display caches.
- Display-oriented expensive paths have explicit freshness rules.
- Market discovery pages can render cached card data.
- System metrics and leaderboards can be served from snapshots/caches.
- Tests prove cached/read-model outputs match raw recomputation at snapshot time.
- Documentation explains which data is safe to cache and which data is not.
