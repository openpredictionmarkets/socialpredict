---
title: Read Model Caching And Performance Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-07T00:00:00Z
updated_at_display: "Sunday, June 7, 2026"
update_reason: "Define a cache-safe design for Postgres read models, optional Redis response caching, and transaction-safe separation from order calculations."
status: draft
---

# Read Model Caching And Performance Design

## Design Position

Caching should be applied to read models, not to order truth.

The design should preserve raw tables as the source of truth while creating explicit read-model seams for expensive displays. Cached values can be stale for browsing and dashboards, but must not decide user credits, order execution, or final settlement.

Transaction-time anything is canonical. Display-time anything can be considered
for caching if it cannot create, mutate, or settle user credits.

Market accounting snapshots are therefore outside the transaction path. The
snapshot repository is separate from the transaction repository interfaces used
by buy, sell, resolution, payout, and refund flows. If a transaction interface
starts exposing snapshot read/write methods, boundary tests should fail before
the change reaches production.

## Design Inputs

Primary inputs:

- [11-read-model-caching-performance.md](./11-read-model-caching-performance.md)
- [Market Position, Dust, And Volume Flows](../../../MATH/README-MATH-MARKET-FLOWS.md)
- Canonical design plan: `/Users/patrick/Projects/spec-socialpredict-tasks/lib/design/design-plan.json`
- Designer-agent postures from `/Users/patrick/Projects/spec-socialpredict-tasks/.codex/agents/`
- Existing load-test findings and performance dossiers
- Existing market discovery, leaderboard, stats, and market detail code paths

## Boundary Alignment

| Boundary | Responsibility |
|---|---|
| Prediction Market Context | Own canonical market, bet, position, probability, dust, and payout math. |
| Market Accounting Read Model | Own derived market accounting/display snapshots. |
| Analytics Context | Own system metrics and global leaderboard snapshots. |
| User Financial Read Model | Own authenticated display snapshots for individual user financial summaries and position-heavy profile views. |
| CMS/Discovery Context | Own cached market discovery cards, topic pages, and pinned market summaries. |
| API Boundary | Own freshness metadata, pagination, and cache-safe response schemas. |
| Infrastructure Boundary | Own optional Redis deployment/configuration and operational cache settings. |
| Repository Boundary | Own durable Postgres read-model tables/materialized views and migrations. |
| Testing Boundary | Own recomputation-vs-snapshot verification tests. |

## Cache Safety Classification

| Classification | Examples | Cache policy |
|---|---|---|
| Transaction-critical | buy, sell, dust settlement, user balance mutation, resolution payout, cancellation/yank refund, admin mutation | Never use stale cache for decision-making. |
| Market-page display | market detail probability, volume, user count, compact widgets | Cache around 1 minute with freshness metadata. |
| Market dust display | retained dust, net volume, volume with dust | Cache only as display/read-model dust; never use for transaction-time sale dust. |
| Market bet table display | recent bets | Paginate; do not cache, or refresh/poll around 10 seconds so accepted bets appear quickly. |
| Discovery display | `/markets`, topic pages, pinned cards, compact charts | Cache around 10 minutes. |
| Market leaderboard display | participant ranking for one market | Cache around 10 minutes and paginate. |
| Dashboard analytics | system metrics, global leaderboard | Cache around 1 hour or scheduled refresh; paginate leaderboard. |
| User financial display | individual user P/L, exposure, financial summaries, historical position views | Cache authenticated display snapshots around 1-10 minutes; never use for spend checks or settlement. |
| Audit/reconciliation | raw bets, migrations, payout verification | Recompute from source of truth. |

## Postgres Read Models

Postgres should be used for durable read models where repeatability, inspection, or restart persistence matters.

Candidate read-model shapes:

```text
market_accounting_snapshots
  market_id
  last_processed_bet_id
  probability
  net_bet_volume
  market_dust
  volume_with_dust
  user_count
  bet_count
  generated_at
```

```text
market_display_snapshots
  market_id
  title
  status
  probability
  volume_with_dust
  market_dust
  user_count
  close_time
  tags_json
  compact_probability_points_json
  generated_at
```

```text
global_leaderboard_snapshots
  snapshot_id
  rank
  username
  score/profit fields
  generated_at
```

```text
user_financial_metric_snapshots
  username
  summary_json
  position_count
  unresolved_exposure
  realized_profit_loss
  generated_at
```

Implementation may choose tables, materialized views, or service-managed snapshots. The key requirement is one explicit read-model boundary rather than ad hoc duplicate math.

User financial snapshots require authenticated cache boundaries. SocialPredict
is a game, so logged-in users may view each other's game-financial status for
transparency. These summaries still must not be served to logged-out/public
visitors. Private identity/security fields such as email, API keys, password
flags, and admin-only notes remain outside the financial read model.

Aggregate reporting visibility should be controlled explicitly by CMS/admin
settings:

- system statistics may be public or login-required
- global leaderboard may be public or login-required
- user financial read models remain login-required regardless of aggregate settings
- transaction endpoints are unaffected by reporting visibility settings

## Redis Cache

Redis is optional and should sit above durable read models or above cheap-to-recompute display responses.

Recommended uses:

- hot `/markets` discovery responses
- hot `/markets/topic/:slug` responses
- compact pinned market chart payloads
- system metrics API response bodies
- leaderboard pages
- authenticated user financial summary pages

Avoid Redis for:

- final order validation
- user balance mutation decisions
- spend checks or debt-limit checks
- final payout truth
- any write transaction that needs canonical data

## Endpoint Boundary

Prefer explicit separation between transaction endpoints and cache-backed display
endpoints.

Transaction endpoints should be canonical and never cache-driven:

```text
POST /v0/bet
POST /v0/sell
POST /v0/markets/{id}/resolve
admin mutation endpoints
```

Display/read-model endpoints may be cache-backed:

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

The exact URL shape can change during implementation. The architectural rule is
more important than the names: display read models must not be accidentally
reused by transaction code.

## Freshness Tiers

Baseline TTL/freshness classes:

| Tier | Data | Target freshness |
|---|---|---:|
| Transaction | buy/sell/dust/balance/resolution/refund/admin mutations | never cached |
| Fast display refresh | market bet table first page | not cached; refresh/poll around 10s |
| Market detail widgets | probability, volume, user count, compact summary | about 1m |
| Market detail dust display | net volume, retained dust, volume with dust | about 1m |
| User financial widgets | individual P/L, exposure, position-heavy summaries | about 1-10m; refresh after user transaction success |
| Page-level discovery | `/markets`, topic pages, pinned chart cards, market cards | about 10m |
| Leaderboard snapshots | market leaderboards | about 10m |
| Slow dashboard snapshots | system financial metrics, global leaderboard | about 1h |

## Invalidation And Freshness

Recommended baseline policies:

| Trigger | Cache behavior |
|---|---|
| New bet/sale | Invalidate or mark stale market detail/card/leaderboard caches for that market. |
| User bet/sale/resolution payout/refund | Invalidate or mark stale affected user financial snapshots. |
| Market approval/rejection/resolution | Invalidate market card, detail, topic, and admin queue caches. |
| Tag/CMS layout update | Invalidate discovery and topic page caches. |
| User profile or leaderboard-affecting event | Invalidate user/global leaderboard caches. |
| Scheduled refresh | Rebuild system metrics and global leaderboard snapshots periodically, roughly hourly. |

Use freshness metadata in responses where staleness may be visible:

```json
{
  "generatedAt": "2026-06-07T00:00:00Z",
  "source": "snapshot",
  "ttlSeconds": 60
}
```

Implementation note:

- Market accounting snapshots and user financial metric snapshots expose a shared freshness metadata contract.
- Existing public display handlers should only add this metadata through explicit display/read-model wiring.
- User financial read-model routes are login-required game transparency routes, not public anonymous routes.
- Aggregate reporting routes can be public or login-required based on CMS reporting visibility settings.
- Existing transaction endpoints and transaction service interfaces must not consume this metadata.

## Pagination And Display Simplification

Caching alone is not enough. Heavy lists should also be paginated or hidden behind explicit user actions.

Recommended display changes:

- market page bets table defaults to latest 10 rows
- positions and leaderboard widgets are paginated
- global leaderboard is paginated
- system financial metrics show a simplified summary first
- complex financial paths are behind expansion buttons
- user financial positions are loaded on demand
- user financial metrics use authenticated read models and expose freshness metadata

## Correctness Tests

Every read-model calculator should have tests comparing snapshot output to raw recomputation.

Required test pattern:

```text
seed raw market/bet history
compute canonical raw result
compute read-model snapshot
assert equality for snapshot fields
```

Also test cache invalidation boundaries:

- new bet invalidates market card snapshot
- sale invalidates market accounting snapshot
- user transaction invalidates user financial metric snapshot
- market resolution invalidates leaderboard/system metrics snapshots
- tag/CMS changes invalidate discovery snapshots

## Open Questions

- Should market accounting snapshots be updated synchronously inside bet transactions or asynchronously with tail replay?
- Should Redis be required for production or optional behind env flags?
- What freshness SLA should market detail pages use versus discovery pages?
- Should market dust be exact historical replay or simple retained-dust convention until the accounting snapshot work lands?
