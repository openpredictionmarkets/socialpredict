---
title: Market Bet-History Query Boundaries And Replay Efficiency Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-23T00:00:00Z
updated_at_display: "Tuesday, June 23, 2026"
update_reason: "Align implementation slices with design-plan review: domain-term classification, narrow repository ports, migration evidence, and query-scope proof."
status: proposed
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

- [ ] Define the domain term represented by each `bets` read: replay event stream, visible bet row, participation aggregate, position input, audit record, or display snapshot.
- [ ] Inventory every production GORM read of `bets`.
- [ ] Classify each read as transaction-critical, market display, user display, system aggregate, or admin/search.
- [ ] Assign an owner boundary to each read: Prediction Market Core, Betting/Position Ledger, Participant Account, Analytics/Read Model, Admin/Search, or Persistence Adapter.
- [ ] Document whether each query is already SQL-scoped, scoped through `IN`, or broad/global by design.
- [ ] Identify accidental table-wide reads in per-market or per-user paths.
- [ ] Identify over-wide projections such as `Select("bets.*")`, plain full-model `Find`, or unused model fields.
- [ ] Identify display paths that page in memory after loading full market history.
- [ ] Identify related query groups that can be consolidated into one aggregate query, scoped DTO query, or read-model refresh.
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

- [ ] Inspect existing timestamped migrations for bet/market/tag indexes.
- [ ] Confirm index naming and duplicate-index behavior before adding migrations.
- [ ] Add missing timestamped migrations for market bet chronology.
- [ ] Add missing timestamped migrations for market+user and user+market bet lookups.
- [ ] Add missing timestamped migrations for lifecycle/status/tag discovery if absent.
- [ ] Add migration tests or Postgres verification where practical.
- [ ] Record write-path/storage/lock-risk notes for each new composite index.

## 05. Repository Refactor

Checklist:

- [ ] Add or standardize narrow use-case repository ports, not one broad `BetRepository` for unrelated reasons to change.
- [ ] Split ports where appropriate: `MarketReplayReader`, `MarketDisplayReader`, `UserPositionReader`, and `PlatformAggregateReader`.
- [ ] Add aggregate methods for unique positive participants by market/platform where system metrics or work-profit reporting need them.
- [ ] Add purpose-specific row structs for narrow bet projections.
- [ ] Ensure projection rows are scalar DTOs with no GORM hooks/business methods.
- [ ] Ensure projection rows include deterministic ordering fields where needed.
- [ ] Keep framework-specific types and concrete repository packages out of domain/use-case packages.
- [ ] Ensure per-market methods include SQL `market_id` predicates.
- [ ] Ensure high-volume reads select only needed columns.
- [ ] Ensure user methods begin from `username` where appropriate.
- [ ] Avoid adding generic `ListBets`/`AllBets` methods without explicit global naming.
- [ ] Add tests that seed multiple markets and prove per-market reads exclude unrelated rows.
- [ ] Add adapter-level query-capture, dry-run, or logger assertions for high-risk methods so correct results cannot hide broad SQL.
- [ ] Add ordering tests for chronological market replay.
- [ ] Add tests proving narrow projections still satisfy WPAM/DBPM calculator inputs.

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
- [ ] Run targeted repository tests.
- [ ] Run Postgres integration tests for index-heavy paths when `POSTGRES_TEST_DSN` is available.
- [ ] Capture before/after query counts or query plans for at least one high-risk route.
- [ ] Capture before/after row count and column projection evidence for at least one hot market replay path.
- [ ] Run pure WPAM/DBPM tests with in-memory event DTOs for reduced replay inputs.
- [ ] Run use-case tests with fake repository ports.
- [ ] Run adapter contract tests with multiple markets/users.
- [ ] Run migration/index existence tests where practical.
- [ ] Run import-guard/static checks preventing `gorm` imports from domain/use-case packages if such a guard exists or can be added cheaply.
- [ ] Prove optimized/reduced query inputs produce identical WPAM, DBPM, resolution, refund, and dust results for representative replay histories.

## 08. Rollout And Fallback

Checklist:

- [ ] Introduce projection/aggregate methods alongside old methods before deleting broad methods.
- [ ] Migrate the highest-risk call sites first.
- [ ] Leave a simple fallback path to the previous canonical full-input method until behavior and performance evidence are captured.
- [ ] Retire broad methods only after callers are migrated and tests prove no semantic drift.
- [ ] If a query optimization changes economic behavior, revert it and open a separate math-equivalence feature rather than silently accepting the change.
