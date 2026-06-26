---
title: Market Bet-History Query Boundaries And Replay Efficiency Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-23T00:00:00Z
updated_at_display: "Tuesday, June 23, 2026"
update_reason: "Align repository/query optimization design with domain-first language, evolutionary architecture, clean dependency boundaries, and transaction-safe read-model rules."
status: proposed
---

# Market Bet-History Query Boundaries And Replay Efficiency Design

## Design Posture

The design goal is not to avoid canonical bet history. The design goal is to avoid asking Postgres for more canonical history than the caller needs.

This follows existing SocialPredict boundaries:

- transaction math must remain canonical and non-stale
- display/read models may be cached or snapshotted
- repositories own database access details
- domain services should ask for intent-specific data rather than loading tables and filtering in Go

## Policy Versus Mechanism

Policies this feature must preserve:

- market replay uses canonical chronological bet history in deterministic order;
- transaction-critical paths use fresh authoritative state;
- single-market use cases must not request platform-wide bet history;
- display snapshots must declare freshness and must not drive execution;
- SQL aggregation must not redefine WPAM, DBPM, dust, payout, refund, or work-profit rules.

Mechanisms this feature may change:

- GORM chain shape;
- raw SQL where clearer;
- projection DTOs;
- composite indexes;
- CTE/window-query experiments;
- read-model refresh and invalidation strategy;
- query-plan and logger-based verification.

## Design-Plan Compatibility

This design intentionally maps the external design-plan postures into a narrow backend performance feature:

| Design posture | Applied here |
| --- | --- |
| Evans/domain-first | Preserve SocialPredict accounting language; optimize access to domain facts without redefining the facts. |
| Fowler/evolutionary | Prefer audit, measurement, scoped methods, indexes, and reversible projection changes before bigger caching or SQL-math rewrites. |
| Martin/clean architecture | Keep use-case policy in domain services and persistence mechanics in repository adapters. Repository interfaces express intent; GORM does not define the domain. |

The design is not a general database-performance campaign. It is specifically about aligning query scope with domain use cases while keeping transaction math authoritative.

## Bounded Contexts And Ownership

| Boundary | Owns | Must not own |
| --- | --- | --- |
| Prediction Market Core | canonical bet history, WPAM/DBPM policy, sale/dust/resolution/refund correctness | GORM query mechanics, index names, display freshness policy |
| Transaction Use Cases | authoritative buy/sell/quote/resolve/refund/payout flows | stale read-model data, async snapshots, UI pagination shortcuts |
| Display Read Models | cached/snapshotted market cards, charts, leaderboards, stats where freshness is acceptable | order execution, balance mutation, payout decisions |
| Repository Interfaces | use-case-shaped data access contracts | generic table dumps or framework types leaking inward |
| GORM/Postgres Adapters | SQL predicates, indexes, projections, aggregate queries, transaction mechanics | business policy decisions or alternative market-math semantics |

The desired dependency direction is:

```text
handlers -> domain/application service -> repository interface -> GORM/Postgres adapter
```

Use-case/application packages own repository ports. GORM repositories are outer adapters implementing those ports.

Domain and use-case code must not import:

- `gorm`;
- SQL builders or concrete query helpers;
- migration packages;
- concrete repository packages;
- ORM row types unless they have been intentionally promoted into domain entities.

The adapter can use projection structs, raw SQL, or GORM chains as implementation details.

## Repository Boundary

Repository methods should expose intent, not generic table access.

Preferred methods:

```go
ListChronologicalBetEventsForMarket(ctx, marketID)
ListRecentVisibleBetRowsForMarket(ctx, marketID, page)
LoadPositionReplayInputs(ctx, username, marketIDs)
CountFirstPositiveParticipationsByMarket(ctx, marketIDs)
UserHasBet(ctx, marketID, username)
```

Avoid new generic methods like:

```go
ListBets(ctx)
AllBets(ctx)
```

Those names hide whether the caller truly needs a system-wide read.

When a global read is legitimate, the method name should say so:

```go
ComputeGlobalParticipationCounts(ctx)
LoadSystemMetricSnapshot(ctx)
ListGlobalLeaderboardCandidates(ctx, page)
```

This gives reviewers a language-level way to distinguish intentional platform-wide aggregation from accidental broad scans.

When reasons to change differ, prefer narrow ports instead of one broad `BetRepository`:

| Port/interface | Reason to change |
| --- | --- |
| `MarketReplayReader` | WPAM/DBPM/resolution/refund replay input requirements. |
| `MarketDisplayReader` | visible bet rows, display pagination, chart/read-model behavior. |
| `UserPositionReader` | user-scoped portfolio and position inputs. |
| `PlatformAggregateReader` | system metrics, global leaderboard candidates, participation aggregates. |

These can share one concrete GORM adapter internally, but the inner contracts should stay use-case-shaped.

## Per-Market Calculations

Per-market calculations may need every bet for that market. That is acceptable. They should not need every bet for every market.

Examples:

- market probability history
- market detail chart
- market resolution payout
- N/A refund calculation
- market-level leaderboard
- market volume/dust display

Required repository behavior:

```text
market_id predicate is applied in SQL
chronological ordering is done in SQL
only needed columns are selected where practical
```

## Column Projection

Every repository method should choose the narrowest correct projection for its caller.

| Caller type | Preferred projection |
| --- | --- |
| WPAM replay | chronological amount/outcome/time rows; include `id` for deterministic tie-breaks |
| DBPM positions | username, market, amount, outcome, time rows; include `id` where ordering matters |
| Volume/dust display | amount rows or aggregate volume snapshot |
| First participation fee counts | market/user positive participation rows or SQL aggregate counts |
| Recent bet table | visible row fields plus display probability source |
| Existence checks | `EXISTS`/`COUNT`; do not materialize bet rows |

Full model reads remain acceptable where the caller truly needs the whole model, but they should not be the default for high-frequency or high-volume paths.

Projection-row rules:

- use simple scalar DTOs;
- include deterministic ordering fields such as `id` and `placed_at` where replay order matters;
- avoid GORM hooks, lifecycle methods, or business methods on projection rows;
- avoid framework-specific types crossing inward;
- name display DTOs distinctly from transaction replay inputs;
- do not return `models.Bet` to WPAM/DBPM unless it is intentionally treated as a domain entity rather than an ORM row.

## Query Consolidation

Some display endpoints may currently assemble a page by calling several repositories that each perform their own table scan or replay. The implementation audit should identify those call chains and decide whether one of these is better:

- one aggregate SQL query;
- one scoped query returning a compact DTO;
- one read-model refresh that stores the display payload;
- one explicit transaction-safe repository call for canonical execution paths.

Consolidation should not merge unrelated domain decisions. It should only reduce redundant database work for one coherent display or metric payload.

## Critical Decisions Versus Mechanisms

| Concern | Critical policy decision? | Mechanism choice? | Rule |
| --- | --- | --- | --- |
| Whether transaction paths use canonical fresh state | Yes | No | Must not be changed by this feature. |
| Whether a display widget may use a stale snapshot | Yes | Partly | Requires an explicit freshness contract and UI/API behavior. |
| Whether a query uses GORM chains or raw SQL | No | Yes | Choose the clearer/testable adapter implementation. |
| Which indexes support common predicates | No | Yes | Add through migrations and verify with tests or query plans. |
| Whether WPAM/DBPM math moves into SQL | Yes | No for this feature | Out of scope unless a separate equivalence feature proves it. |
| Whether global metrics are cached hourly/minutely | Yes | Partly | Belongs with read-model/cache policy, not hidden inside repository code. |

## Pagination Caveat

The market bet table can show a paginated page of rows, but the current implementation may still need full market history to attach probability-after-bet values without creating a second math path.

Implementation options:

| Option | Tradeoff | Migration cost | Operational burden | Reversibility | When to choose |
| --- | --- | --- | --- | --- | --- |
| Keep full market-scoped history for bet table | Correct and simple, but expensive for very large single markets. | Low | Low | High | First slice and moderate market history sizes. |
| Add probability-history read model | Faster display; must remain display-only and invalidated after transactions. | Medium | Medium | Medium | After profiling proves full market replay dominates display latency. |
| Use window/CTE query for recent rows plus probability snapshot | More complex; may be useful after profiling. | Medium/high | Medium | Medium | When recent visible rows are hot but full probability history is already snapshotted. |

The first implementation should focus on avoiding platform-wide scans and adding indexes. Later slices can optimize very large single-market history display.

## Global Aggregates

Some queries are global by definition. Examples:

- system financial metrics
- global leaderboard
- platform participation fees

These should not be confused with accidental broad scans. However, they are still risky as the platform grows.

Preferred future shapes:

- aggregate SQL for counts/sums where possible
- durable read-model snapshots refreshed asynchronously or after mutations
- paginated/ranked materialized views for leaderboards
- explicit repository method names that include `Global`, `Aggregate`, or `Snapshot`

Tradeoff notes:

| Shape | Benefit | Cost | Use when |
| --- | --- | --- | --- |
| Aggregate SQL | Reduces row materialization and app CPU. | Couples repository adapter to Postgres query shape. | Metric is a count/sum/group and does not require chronological market math. |
| Durable read-model snapshot | Stable under repeated reads and restart-inspectable. | Needs refresh/invalidation ownership and migration/storage. | Display route is hot and can tolerate explicit staleness. |
| Materialized/ranked view | Efficient for leaderboards and rankings. | More operational complexity and Postgres-specific behavior. | Repeated global ranking pressure is measured. |

For first-participation fee accounting, prefer pushing the distinct `(market_id, username)` grouping into SQL rather than materializing all `bets` rows. A repository method could return either a platform total or per-market counts depending on the caller:

```sql
SELECT market_id, COUNT(*) AS participant_count
FROM (
  SELECT market_id, username
  FROM bets
  WHERE amount > 0
  GROUP BY market_id, username
) first_participations
GROUP BY market_id;
```

This is the refined version of the idea explored in PR #352: aggregate in the database, but return only the aggregate shape the domain needs.

## Evolution Path

The implementation should proceed in reversible slices:

1. **Inventory and classify** current GORM reads and callers.
2. **Add tests around existing scoped behavior** before changing queries.
3. **Add missing indexes** using timestamped migrations.
4. **Introduce projection methods** behind existing repository interfaces where call sites can stay stable.
5. **Move simple global counts/sums to aggregate SQL** only where domain semantics are not chronology-sensitive.
6. **Defer probability-history/read-model rewrites** until profiling proves hot single-market history is the bottleneck.

Each slice should be independently reviewable and should not require a frontend or market-math rewrite.

Fallback rules:

- Keep old repository methods while introducing projection methods until call sites are migrated and tested.
- Index migrations should inspect for duplicates first and should consider write-path cost and lock behavior.
- If a projection or aggregate changes behavior, revert to the older full canonical input and keep the failed optimization as evidence.

## SQL Versus Go Math Boundary

Postgres should reduce and shape data. Go should preserve SocialPredict's domain math.

| Work | Belongs in Postgres? | Belongs in Go? | Notes |
| --- | --- | --- | --- |
| Filtering a market's bets | Yes | No | `WHERE market_id = ?` should happen in SQL. |
| Selecting only needed bet columns | Yes | No | Use projection row structs. |
| Counting users/markets | Yes | No | Pure aggregate. |
| Counting first positive market participants | Yes | Usually no | Aggregate in SQL; apply policy thresholds in Go if clearer. |
| Summing plain active volume | Yes | Maybe | SQL or read-model snapshot is appropriate for display. |
| WPAM chronological probability replay | No, unless separately proven | Yes | Chronology-sensitive domain math. |
| DBPM position valuation | No, unless separately proven | Yes | Keep audited math path in Go. |
| Sale order execution/dust calculation | No stale aggregate | Yes | Transaction path must use canonical current state. |
| Chart/card display probability | Snapshot preferred | Maybe | Display can use read models; transaction paths cannot. |

This boundary prevents the optimization work from accidentally creating a second market-math implementation inside SQL.

## Candidate Function/Metric Audit Seeds

Initial candidates to inspect during implementation:

| Function/path | Owner boundary | Current shape to verify | Candidate optimization |
| --- | --- | --- | --- |
| `markets.GormRepository.ListBetsForMarket` | Prediction Market Core / Persistence Adapter | market-scoped, currently full bet model projection | narrow projection for callers that do not need full `models.Bet` |
| `analytics.GormRepository.ListBetsForMarket` | Analytics Read Model / Persistence Adapter | market-scoped, already selected columns | verify ordering includes deterministic `id`; add index support |
| `markets.Service.GetMarketBetsPage` | Market Display Read Model | market-scoped full history, then in-memory page | keep correctness; later use probability snapshot plus recent-row projection |
| `analytics.Service.ComputeSystemMetrics` | Platform Aggregate | global metrics assembled from users/markets/bets | aggregate SQL/read-model snapshot for counts, active volume, participation fees |
| `analytics.Service.ComputeGlobalLeaderboardSnapshot` | Platform Aggregate / Display Read Model | all users, all markets, per-market bet loads | cached leaderboard snapshot / pre-aggregated candidate rows |
| user financial summary paths | Participant Account / User Display | user-scoped then affected market IDs | ensure narrow projections and use user financial read model for display |
| `bets.GormRepository.UserHasBet` | Betting/Position Ledger | count with `market_id`/`username` predicate | consider `EXISTS` and supporting composite index |

## ADR Candidates

- ADR: Transaction paths must use canonical bet history and never stale display snapshots.
- ADR: Repository methods are named by use case/intent rather than table access.
- ADR: Postgres may aggregate and project data, but WPAM/DBPM policy remains in Go unless separately proven equivalent.
- ADR: High-volume display paths prefer read models or aggregate SQL when freshness and domain rules allow it.
- ADR: Index additions follow timestamped migration convention and are backed by focused tests.

## Query Audit Output

The implementation audit should produce a table like:

| Path | Current query shape | Scope | Risk | Proposed change |
| --- | --- | --- | --- | --- |
| `GetMarketBetsPage` | `ListBetsForMarket` then in-memory page | market-scoped | single hot market history can grow large | keep scoped baseline; consider probability read model later |
| `ComputeSystemMetrics` | `ListMarkets` plus per-market bet loads and global ordered bets for fees | global | expensive at platform scale | aggregate SQL/read-model snapshot |
| `ComputeGlobalLeaderboardSnapshot` | all users, all markets, per-market bet loads | global | expensive at platform scale | cached leaderboard snapshot / aggregate strategy |

## Migration Boundary

Indexes must be added through the existing timestamped migration convention. The implementation should inspect current migrations first to avoid duplicate indexes.

Migration convention:

- file name: readable `YYYYMMDD_HHMMSS_description.go`;
- registration ID: compact timestamp format already used by `backend/migration`;
- behavior: additive/backward-compatible where practical;
- tests: package-local migration tests for created indexes/default behavior where practical;
- evidence: note duplicate-index inspection and expected lock/write-path impact in the PR.

## Testing Boundary

Tests should verify behavior first and query shape where necessary:

- seed bets for multiple markets
- call per-market repository method
- assert rows from other markets are excluded
- assert ordering by `placed_at`, then `id` when needed
- for Postgres integration tests, optionally use `EXPLAIN` to verify indexes are eligible after indexes exist
- add a fake repository or call-counter test proving transaction services do not route through display snapshot repositories for execution
- add projection tests that fail if a required WPAM/DBPM input is accidentally removed from the row shape
- add aggregate tests that seed multiple markets/users and prove returned counts/sums match business-intent names
- add pure WPAM/DBPM tests with in-memory event DTOs to prove reduced inputs are equivalent
- add use-case tests with fake repository ports to verify dependency direction
- add adapter contract tests with multiple markets/users to prove SQL-scoped behavior
- add migration/index existence tests where the migration framework supports it
- add import-guard tests or static checks preventing `gorm` imports from entering domain/use-case packages
- for high-risk methods, use query capture, GORM dry-run, or logger assertions to prove the generated SQL carries the required scope predicate

Result-only tests are not sufficient for query-scope proof because an implementation could load all rows and filter in Go while returning the correct answer.

## Read-Model Freshness Rules

Any display snapshot or read model introduced by this feature must document:

- owner;
- refresh trigger;
- invalidation after buy, sell, resolve, refund, or equivalent mutation;
- max acceptable staleness;
- freshness field exposed to callers, such as `generated_at` or `as_of_event_id`;
- explicit prohibition from transaction-critical execution paths.

## Risks And Open Questions

| Risk or ambiguity | Why it matters | Resolution path |
| --- | --- | --- |
| Feature number conflicts with newer docs on `main` | This branch was authored before later grouped-market feature docs occupied adjacent numbers. | Renumber before retargeting to current `main` if needed. |
| Narrow projections accidentally omit fields used by chronology-sensitive math | Could change probability, positions, dust, or payout behavior. | Projection tests must cover WPAM/DBPM required inputs and deterministic ordering. |
| Aggregate SQL creates a hidden second interpretation of market economics | Could make stats disagree with transaction truth. | Keep SQL to counts/sums/grouping; apply domain policy in Go unless separately proven. |
| Index additions improve reads but slow writes | Betting is write-heavy during hot events. | Measure write-path impact and keep indexes tied to demonstrated access patterns. |
| In-memory pagination remains expensive for very large single markets | Market-scoped is necessary but may not be sufficient at high event volumes. | Defer probability-history read model or recent-row strategy until profiling justifies it. |
| Display read models leak into transaction paths | Could make orders, balances, payouts, refunds, or dust depend on stale state. | Use repository naming, fake-port tests, and import/call guards. |

## Non-Goals

- Do not rewrite market math.
- Do not put stale snapshots on transaction paths.
- Do not remove canonical bet-history replay where correctness requires it.
- Do not implement Redis just to hide inefficient SQL.
- Do not introduce framework/database types into domain policy to make a query easier.
