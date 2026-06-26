---
title: Market Bet-History Query Boundaries And Replay Efficiency Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-25T00:00:00Z
updated_at_display: "Thursday, June 25, 2026"
update_reason: "Start implementation with deterministic market bet replay ordering, composite bet indexes, and scoped-read adapter tests."
status: in_progress
---

# Market Bet-History Query Boundaries And Replay Efficiency Plan

## 01. Feature Spec

Checklist:

- [x] Create feature overview.
- [x] Create design document.
- [x] Create implementation plan.
- [x] Ground the spec in a first-pass code audit.

## 02. Query Audit

Checklist:

- [x] Define the domain term represented by each `bets` read: replay event stream, visible bet row, participation aggregate, position input, audit record, or display snapshot.
- [x] Inventory every production GORM read of `bets`.
- [x] Classify each read as transaction-critical, market display, user display, system aggregate, or admin/search.
- [x] Assign an owner boundary to each read: Prediction Market Core, Betting/Position Ledger, Participant Account, Analytics/Read Model, Admin/Search, or Persistence Adapter.
- [x] Document whether each query is already SQL-scoped, scoped through `IN`, or broad/global by design.
- [x] Identify accidental table-wide reads in per-market or per-user paths.
- [x] Identify over-wide projections such as `Select("bets.*")`, plain full-model `Find`, or unused model fields.
- [ ] Identify display paths that page in memory after loading full market history.
- [x] Identify related query groups that can be consolidated into one aggregate query, scoped DTO query, or read-model refresh.
- [ ] Identify global paths that should move to aggregate SQL or read-model snapshots.
- [ ] Revisit the PR #352 first-time participation query idea and convert it into aggregate repository methods rather than row materialization.
- [ ] Document first-participation edge cases: repeated buys, sells after buy, refunds, cancelled/N/A markets, zero/negative rows, and participant identity changes.

## 03. Baseline Evidence

Checklist:

- [ ] Capture current query count for at least one high-risk market route.
- [ ] Capture rows loaded and columns selected for at least one hot-market replay path.
- [ ] Capture current latency or p95 behavior for a representative display route where available.
- [ ] Record whether the bottleneck is platform-wide scanning, large single-market replay, over-wide projection, missing index, or repeated call-chain work.
- [ ] Commit or link the query-audit table in PR notes.

## 04. Index Migration

Checklist:

- [x] Inspect existing timestamped migrations for bet/market/tag indexes.
- [x] Confirm index naming and duplicate-index behavior before adding migrations.
- [x] Add missing timestamped migrations for market bet chronology.
- [x] Add missing timestamped migrations for market+user and user+market bet lookups.
- [ ] Add missing timestamped migrations for lifecycle/status/tag discovery if absent.
- [x] Add migration tests or Postgres verification where practical.
- [x] Record write-path/storage/lock-risk notes for each new composite index.

Implementation note:

- Added `idx_bets_market_id_placed_at_id` for canonical per-market replay reads ordered by `placed_at ASC, id ASC`.
- Added `idx_bets_market_id_username` for first-participation and market/user existence checks.
- Added `idx_bets_username_market_id_placed_at_id` for user-scoped position/portfolio reads that identify affected markets and then replay those markets.
- Added `idx_bets_username_placed_at_id` for user bet-history display ordered by `placed_at DESC, id DESC`.
- These indexes trade additional bet-insert write/storage cost for scoped read performance. They do not add business state or alter transaction math.

## 05. Repository Refactor

Checklist:

- [ ] Add or standardize narrow use-case repository ports, not one broad `BetRepository` for unrelated reasons to change.
- [ ] Split ports where appropriate: `MarketReplayReader`, `MarketDisplayReader`, `UserPositionReader`, and `PlatformAggregateReader`.
- [ ] Add aggregate methods for unique positive participants by market/platform where system metrics or work-profit reporting need them.
- [ ] Add purpose-specific row structs for narrow bet projections.
- [ ] Ensure projection rows are scalar DTOs with no GORM hooks/business methods.
- [ ] Ensure projection rows include deterministic ordering fields where needed.
- [ ] Keep framework-specific types and concrete repository packages out of domain/use-case packages.
- [x] Ensure per-market methods include SQL `market_id` predicates.
- [ ] Ensure high-volume reads select only needed columns.
- [x] Ensure user methods begin from `username` where appropriate.
- [ ] Avoid adding generic `ListBets`/`AllBets` methods without explicit global naming.
- [x] Add tests that seed multiple markets and prove per-market reads exclude unrelated rows.
- [ ] Add adapter-level query-capture, dry-run, or logger assertions for high-risk methods so correct results cannot hide broad SQL.
- [x] Add ordering tests for chronological market replay.
- [ ] Add tests proving narrow projections still satisfy WPAM/DBPM calculator inputs.

Implementation note:

- Updated market, analytics, sell-transaction, and user-position bet-history readers to use deterministic `placed_at ASC, id ASC` ordering.
- Added repository tests that seed unrelated market rows between same-timestamp target rows, then verify only the requested market's rows are returned and ties are ordered by ID.
- Updated user bet-history display to select only `market_id` and `placed_at` with deterministic reverse chronological ordering.
- Collapsed grouped work-profit member hydration from one query per group into one `WHERE group_id IN ?` query.

## 06. Display Path Optimization

Checklist:

- [ ] Keep transaction paths canonical and non-cached.
- [ ] Keep market bet table market-scoped even when full history is needed for probability display.
- [ ] Keep WPAM/DBPM math in Go unless a separate math-equivalence feature proves another implementation.
- [ ] Consider a probability-history display read model only after profiling proves very large single-market bet-table pressure.
- [ ] If a display read model is introduced, document owner, refresh trigger, invalidation triggers, max staleness, and exposed freshness metadata.
- [ ] Add guardrail tests proving display read models are not used by buy, sell, quote execution, resolve, refund, dust, payout, or balance mutation paths.
- [ ] Move system metrics and global leaderboard toward aggregate/read-model paths from Feature 11.
- [ ] Document intentionally global queries.
- [ ] Add entry criteria for future read models: measured p95 latency, rows loaded per request, repeated global replay, or single-market history size beyond an agreed threshold.

## 07. Verification

Checklist:

- [ ] Run backend unit tests.
- [x] Run targeted repository tests.
- [ ] Run Postgres integration tests for index-heavy paths when `POSTGRES_TEST_DSN` is available.
- [ ] Capture before/after query counts or query plans for at least one high-risk route.
- [ ] Capture before/after row count and column projection evidence for at least one hot market replay path.
- [ ] Run pure WPAM/DBPM tests with in-memory event DTOs for reduced replay inputs.
- [ ] Run use-case tests with fake repository ports.
- [ ] Run adapter contract tests with multiple markets/users.
- [x] Run migration/index existence tests where practical.
- [ ] Run import-guard/static checks preventing `gorm` imports from domain/use-case packages if such a guard exists or can be added cheaply.
- [ ] Prove optimized/reduced query inputs produce identical WPAM, DBPM, resolution, refund, and dust results for representative replay histories.

## 08. Rollout And Fallback

Checklist:

- [ ] Introduce projection/aggregate methods alongside old methods before deleting broad methods.
- [ ] Migrate the highest-risk call sites first.
- [ ] Leave a simple fallback path to the previous canonical full-input method until behavior and performance evidence are captured.
- [ ] Retire broad methods only after callers are migrated and tests prove no semantic drift.
- [ ] If a query optimization changes economic behavior, revert it and open a separate math-equivalence feature rather than silently accepting the change.
