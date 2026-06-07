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
Order calculations and settlement = canonical, non-stale, transaction-safe reads
Display, discovery, statistics, and leaderboards = cacheable read models
```

Caching should reduce load, not hide correctness problems.

## Non-Negotiable Correctness Rule

Do not use stale cache values to decide whether a trade, sale, refund, resolution, or payout is valid.

Order-critical paths must continue to use canonical tables and transaction-safe services.

Examples:

- placing a bet
- confirming a sale
- calculating transaction-time dust
- debiting or crediting user balances
- resolving/yanking/cancelling a market
- final payout/refund execution
- any logic that can create or destroy user credits

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
| System financial metrics display | No | Cacheable snapshot with freshness metadata. | Informational dashboard; does not execute orders. |
| Global leaderboard display | No | Cacheable snapshot, paginate. | Ranking is display-oriented and can tolerate short staleness. |
| Market leaderboard display | No | Cacheable snapshot, paginate. | Informational market widget; does not execute orders. |
| `/markets` discovery page | No | Cacheable card/read-model payload. | Browsing page can tolerate short staleness. |
| `/markets/topic/:slug` page | No | Cacheable topic/card/read-model payload. | Browsing page can tolerate short staleness. |
| Pinned market chart cards | No | Cacheable compact chart snapshot. | Discovery display, not transaction execution. |
| Individual market probability display | No, unless used for order execution | Cache briefly for display; order confirmation must use canonical path. | Users need reasonably fresh UI, but the displayed value should not settle trades. |
| Individual market volume/user count display | No | Cache briefly with freshness metadata. | Informational display only. |
| Market bet table display | No | Paginate and optionally cache first page briefly. | Users are reading history, not executing from the table. |
| Market comments display | No | Paginate and optionally cache briefly. | Not settlement-critical. |

## Cacheable Display Candidates

These are good early candidates because they are read-heavy and not directly responsible for executing trades.

| Area | Candidate data | Suggested freshness |
|---|---|---:|
| System statistics | financial metrics, active volume, user counts, market counts | 30s-5m |
| Global leaderboard | user ranking, profit summaries, resolved market counts | 1m-15m |
| Market leaderboard | per-market participant ranking | 30s-5m |
| Market cards | title, status, probability, volume, users, tags, steward, close time | 10s-60s |
| `/markets` page | discovery cards, pinned market summaries, topic navigation payloads | 10s-60s |
| `/markets/topic/:slug` page | filtered topic cards and pinned topic markets | 10s-60s |
| Pinned charts | compact probability history snapshots | 10s-60s |
| Bet table display | paginated recent bets, first page only by default | 5s-30s |
| Comment table display | paginated comments | 5s-60s |

## Less Cacheable Or Non-Cacheable Paths

| Area | Position |
|---|---|
| Confirm buy/sell order | Do not cache for decision-making. |
| Sale quote | Can be preview-only but should be treated as informational; final sale recalculates. |
| User balance | Avoid stale values for transaction decisions; display can poll or show last refreshed timestamp. |
| Final payout | Do not use display cache as payout truth. |
| Admin mutation actions | Read fresh before mutating. |

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
