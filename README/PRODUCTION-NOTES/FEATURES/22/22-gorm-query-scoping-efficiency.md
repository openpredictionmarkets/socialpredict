---
title: Market Bet-History Query Boundaries And Replay Efficiency
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-25T00:00:00Z
updated_at_display: "Thursday, June 25, 2026"
update_reason: "Start implementation with deterministic market bet replay ordering, composite bet indexes, and scoped-read adapter tests."
status: in_progress
---

# Market Bet-History Query Boundaries And Replay Efficiency

GORM is the current persistence mechanism this feature audits. The domain subject is broader: which market, user, participant, and system-wide facts each use case is allowed to read, and how much canonical bet history it should ask Postgres to return.

## Purpose

SocialPredict market math often depends on chronological bet history. That is correct for transaction-time order math, sale math, probability history, resolution, refunds, and auditability. It does not mean every display or analytics path should repeatedly materialize broad tables or more rows than it needs.

For this feature, **chronological bet history** means the authoritative bet event stream ordered deterministically by:

```text
placed_at ASC, id ASC
```

Any optimized replay input must preserve that event order unless a separate domain decision names a different authoritative sequence.

As the platform grows to many markets and many bets, inefficient GORM reads can become a dominant cost. The most important rule is:

```text
When a calculation is for one market, the repository query should fetch only that market's rows unless the caller explicitly asks for a system-wide aggregate.
```

This feature starts a repository/query audit and implementation plan for tightening bet-history reads, adding the right database indexes, and using display read models where exact transaction-time state is not required.

## Implementation Slice 1

The first implementation slice keeps market math unchanged and improves the persistence adapter boundary:

- canonical per-market bet replay now orders by `placed_at ASC, id ASC` where market history is loaded for market, analytics, sale, and user-position paths;
- a timestamped migration adds composite indexes for market chronology, market/user existence checks, and user/market chronological reads;
- repository tests seed unrelated market rows between same-timestamp target rows to prove scoped reads exclude cross-market rows and preserve deterministic tie ordering.

## Design-Plan Alignment

This feature follows the design posture captured in `spec-socialpredict-tasks/lib/design/design-plan.json`:

- **Domain-first language:** optimization work must preserve the canonical meaning of bets, market volume, probability, participant fees, work profits, dust, and payouts.
- **Evolutionary architecture:** start with measurement, scoped repository methods, projections, and indexes before introducing broader caching or SQL rewrites.
- **Clean architecture:** handlers and domain services should ask for use-case-shaped data through repository interfaces; GORM, SQL, indexes, and projection structs are persistence details.
- **Policy over mechanism:** transaction paths use canonical fresh state; display paths may use projections, aggregates, or snapshots only when their freshness contract is explicit.
- **Migration discipline:** index/schema changes use timestamped Go migrations and focused tests where practical.

## Ubiquitous Language

| Term | Meaning in this feature | Avoid confusing with |
| --- | --- | --- |
| Bet event | One canonical row in the market event stream used for replay and audit. | Visible bet-table row. |
| Buy exposure | A positive amount row that can indicate first participation and increased market exposure. | Sale proceeds, refunds, or zero/negative rows. |
| Sell exposure | A negative amount row produced by a sale path. | First participation for fee/work-profit policy. |
| Participant identifier | The stable identity used for participation counting, currently `username`. | Display name, email, or mutable profile fields. |
| Canonical bet history | The append-only source rows needed to reconstruct market math and audit financial changes. | Display snapshots or cached widgets. |
| Market replay | Reconstructing probability, positions, payouts, dust, or refund state from the ordered bet event stream. | Aggregated platform statistics. |
| Market-scoped read | A repository query whose SQL predicate limits rows to the relevant `market_id` or explicit market-ID set. | Loading all bets and filtering in Go. |
| Projection row | A narrow data shape containing only fields needed by a use case. | Full GORM model hydration. |
| First participation | The first positive buy exposure for a participant in a market under the current fee/work-profit policy. | Repeated buys, sells, refunds, or cancelled-market payout state. |
| Active volume | Display/accounting volume whose exact semantics must be specified by the caller: gross buy volume, net exposure, or snapshot total including dust. | Assuming `SUM(amount)` always means the same thing. |
| Aggregate SQL | Database-side count/sum/grouping that returns a reduced result. | Reimplementing WPAM/DBPM math in SQL. |
| Read model snapshot | Durable display-oriented data with a freshness contract. | Authoritative transaction state. |
| Transaction-critical path | Buy, sell, quote execution, dust, refund, resolution, payout, and balance mutation paths. | Public/read-only display paths. |
| System aggregate | A platform-wide metric that is global by definition and must be named/documented as such. | Accidental table-wide scan. |

## Current Grounding

A quick audit shows that some important paths are already market-scoped:

- `markets.GormRepository.ListBetsForMarket` uses `WHERE market_id = ?`.
- `analytics.GormRepository.ListBetsForMarket` uses `WHERE market_id = ?`.
- market bet history, market volume, probability projection, market leaderboard, market overview, and resolution paths call market-scoped bet loaders.
- user financial positions already narrow from a user's bets to `market_id IN ?` for only the relevant markets.

The remaining performance concern is not simply "every market path loads the whole bets table." The more precise concern is:

- some analytics/system paths are intentionally global and may need aggregate SQL or read-model snapshots rather than row-by-row replay;
- some display pagination paths still load the full market history before slicing a page;
- some probability/chart paths may require full market history but can read a narrower column projection;
- some repository methods may fetch more columns than a calculation needs, for example `Select("bets.*")` when only amount/outcome/time are required;
- some routes may issue multiple related queries that can be consolidated into one purpose-built query or read-model refresh;
- indexes may not yet match the most common market-scoped and user-scoped read patterns;
- future GORM additions need tests or guardrails to prevent accidental table-wide reads in per-market code.

## Prior PR #352 Finding

PR #352 explored using a raw SQL query to count first-time market participation by grouping `bets` on `(market_id, username)`. The direction was correct: platform-wide aggregate questions should often be answered by SQL aggregates instead of loading every row and counting in Go.

The improved version should be more precise:

- GORM can express multi-column grouping with `Group("market_id, username")`; raw SQL is optional, not required just because there are two group columns.
- If the caller only needs a count, the database should return a count, not all grouped pairs for Go to count.
- The query should count only positive first participation rows when matching the current fee/work-profit policy.
- For Postgres, a raw query may still be appropriate when it is clearer or faster, for example `COUNT(*) FROM (SELECT 1 FROM bets WHERE amount > 0 GROUP BY market_id, username)`.
- The repository method should be named for business intent, such as `CountUniquePositiveParticipantsByMarket` or `CountPlatformFirstParticipations`, rather than exposing generic raw bet grouping.

This gives us the useful part of PR #352 without assuming raw SQL is always faster than GORM. GORM-generated SQL and raw SQL both execute inside Postgres; the optimization comes from pushing aggregation/filtering into the database and supporting it with the right indexes.

## Column Projection And Query Consolidation

Row filtering is only one part of query efficiency. Column projection matters too.

For example, a market-scoped query can still be wasteful if it uses:

```go
Select("bets.*")
```

when the caller only needs a narrow event stream:

```text
id, username, market_id, amount, outcome, placed_at, created_at
```

Some use cases need even less:

| Use case | Likely needed columns |
| --- | --- |
| WPAM probability replay | `id`, `amount`, `outcome`, `placed_at` |
| DBPM/user position calculation | `id`, `username`, `market_id`, `amount`, `outcome`, `placed_at` |
| Market volume | `amount` |
| First participation fees | `market_id`, `username`, `amount` |
| Recent bet table | `username`, `amount`, `outcome`, `placed_at`, plus probability source |
| User-has-bet check | no row materialization; use `COUNT`, `EXISTS`, or equivalent |

The goal is to define purpose-specific row structs and repository methods so each path receives the smallest correct row shape.

Query consolidation is the related rule: avoid issuing several broad queries when one narrower query or one read-model refresh can produce the needed display payload. This is especially relevant for stats, leaderboards, and market-card widgets.

## Query Classes

| Query class | Examples | Required behavior |
| --- | --- | --- |
| Transaction-critical market query | buy, sell, sale quote execution, resolution, refund, payout, work-profit payout, dust calculation | Use canonical state, scoped by market/user as tightly as correctness allows. Do not use stale display snapshots. |
| Market display query | market bet table, positions, leaderboard, chart, volume widget | Prefer market-scoped query or read model. Avoid full platform bet scans. |
| User display query | portfolio, user financials, user positions | Start from username and then load only affected market IDs. |
| System aggregate query | system metrics, global leaderboard, platform totals | May be global by definition, but should use aggregate SQL or durable read models rather than repeatedly replaying all bets on every request. |
| Admin review/search query | market review, stewardship, tags, user queue | Use filters, search predicates, limit/offset, and supporting indexes. |

## Domain Invariants

- A per-market use case must not perform a platform-wide bet scan unless the method name and documentation explicitly justify why it is global.
- Transaction-critical paths must not depend on stale display snapshots, cached leaderboard rows, or asynchronously refreshed read models.
- SQL may reduce rows, filter rows, order rows, count rows, and sum simple values; it must not silently become a second WPAM/DBPM implementation.
- Repository interfaces should express business intent such as first positive participation, market replay, or user-in-market existence, not generic table access.
- Any new index or durable read-model storage must be introduced through the established migration convention.
- Performance work must not change market economics, payout math, dust handling, participant-fee rules, or moderator work-profit policy without a separate feature decision.
- Display snapshots must expose freshness metadata, for example `generated_at`, `as_of_event_id`, or an equivalent field that lets callers and UI copy distinguish cached display state from authoritative transaction state.

## Architecture Drivers

- Scale `bets` reads as the number of markets and bet events grows.
- Preserve correctness and auditability for WPAM, DBPM, resolution, refund, dust, balance, and work-profit calculations.
- Improve changeability by moving from generic table reads toward intent-specific repository contracts.
- Prefer operationally simple improvements before Redis, materialized views, or SQL math rewrites.
- Keep migration risk visible through additive timestamped migrations, duplicate-index inspection, and tests.
- Require performance evidence before introducing broader read-model or caching infrastructure.

## Candidate Metrics For SQL Aggregation

The stats page and other display-only dashboards should push simple reductions into Postgres where possible. Go should keep domain math that depends on WPAM/DBPM semantics, transaction state, or audited chronological replay.

| Metric or display area | Good Postgres aggregation candidate? | Suggested SQL/GORM shape | Keep out of Postgres |
| --- | --- | --- | --- |
| Number of users | Yes | `COUNT(*) FROM users` | none |
| Number of markets | Yes | `COUNT(*) FROM markets`, optionally grouped by lifecycle/resolution state | none |
| Active market volume | Yes, when the volume meaning is explicit | `SUM(amount)` for a defined net-exposure metric, a gross-buy aggregate, or market accounting snapshot | final transaction settlement logic; ambiguous volume labels |
| Market creation fees | Yes | `COUNT(markets) * createMarketCost` or `SUM(proposal_cost)` when stored | fee-policy decisions that depend on historical config until policy capture exists |
| First participation fees | Yes | grouped distinct positive `(market_id, username)` counts | work-profit payout execution still needs canonical transaction path |
| Global leaderboard | Partially | pre-aggregate candidate rows/counts, or read-model snapshot | final DBPM/WPAM profitability semantics unless proven equivalent |
| Top/active markets | Yes | filtered/sorted/paginated market rows plus snapshot columns | live transaction probability math |
| Current probability per market | Prefer snapshot | read from market accounting/probability snapshot | recomputing WPAM in SQL |
| User financial summaries | Partially | user-scoped bets and affected market IDs only, or user financial snapshot | final position valuation math if it depends on DBPM/WPAM replay |

Rule of thumb:

```text
SQL reduces data.
Go applies domain math.
Snapshots serve repeated display reads.
Transaction paths stay canonical and fresh.
```

## DBPM/WPAM Boundary

DBPM and WPAM can benefit from narrower inputs, but their math should remain in Go unless a separate math-optimization feature proves an equivalent SQL implementation.

Safe optimizations:

- fetch only one market's chronological event stream;
- fetch only columns needed by the probability/position calculator;
- order rows in SQL by `placed_at ASC, id ASC`;
- add indexes that support that ordered market replay;
- use display snapshots for charts/cards/leaderboards when freshness policy allows.

Unsafe without proof:

- grouping bets by user/outcome before WPAM/DBPM replay;
- collapsing chronological events when intermediate probability changes matter;
- using cached display snapshots for buy, sell, resolve, refund, or dust execution.

## Product/Engineering Rule

For any route or service method named for a single market, a table-wide bet scan is a bug unless explicitly justified in the code and docs.

Examples of acceptable single-market query shape:

```sql
SELECT id, username, market_id, amount, outcome, placed_at, created_at
FROM bets
WHERE market_id = $1
ORDER BY placed_at ASC, id ASC;
```

Examples of suspect single-market query shape:

```sql
SELECT * FROM bets ORDER BY placed_at ASC;
```

```go
db.Find(&bets)
// then filter market IDs in Go
```

## Database Index Direction

The first implementation should verify and add missing indexes for the observed access patterns.

Recommended baseline indexes:

| Table | Index | Why |
| --- | --- | --- |
| `bets` | `(market_id, placed_at, id)` | Chronological market replay, charts, resolution, per-market display. |
| `bets` | `(market_id, username, placed_at, id)` | First participation checks, user-in-market checks, sale/position paths. |
| `bets` | `(username, market_id, placed_at, id)` | User portfolio/financial paths that start from a username. |
| `markets` | `(lifecycle_status, created_at)` | Admin/public lifecycle queues and status lists. |
| `market_tag_assignments` | `(tag_slug, market_id)` or equivalent actual schema columns | Topic pages and tag-filtered discovery. |

Index names and exact columns should be verified against current models and migrations before implementation.

## Relationship To Read Model Caching

This feature complements Feature 11, Read Model Caching And Performance.

- Transaction paths still use canonical tables.
- Display paths can use snapshots/read models when freshness is acceptable.
- If a display route uses raw bets, the repository should still scope the query tightly.
- System/global displays should prefer aggregate read models or aggregate SQL over full replay at request time.

## Acceptance Criteria

- Audit all production GORM reads against the query classes above.
- Identify every route/service that reads `bets` and mark it as transaction-critical, market display, user display, or system aggregate.
- Identify every caller that depends on canonical chronological replay and record why projection/aggregation is safe or unsafe.
- Define edge cases for first participation: repeated buys, sells after buy, refunds, cancelled/N/A markets, zero/negative rows, and participant identity changes.
- Identify over-wide column projections such as `Select("bets.*")` or full model loads where a narrow row struct is sufficient.
- Identify related query groups that can be consolidated into a single aggregate query or read-model refresh.
- Add missing indexes through timestamped migrations.
- Add tests proving per-market repository reads include `WHERE market_id = ?` behavior and exclude other market rows.
- Add tests proving narrowed projections still provide all required WPAM/DBPM inputs.
- Add domain-equivalence tests proving optimized/reduced query inputs produce identical WPAM, DBPM, resolution, refund, and dust outcomes for representative replay histories.
- Add at least one guardrail test proving transaction paths do not call display snapshot/read-model repositories for authoritative execution.
- Replace any per-market table-wide bet scans with scoped repository methods.
- Replace or defer global row replay paths with aggregate SQL/read-model snapshots where appropriate.
- Document intentionally global queries so future reviewers can distinguish valid aggregate work from accidental broad scans.

## Numbering And Retargeting Note

This PR was originally authored while `FEATURES/22` was available. If this branch is retargeted to the current `main`, confirm the feature number before merge because newer grouped-market feature docs may already occupy this slot.
