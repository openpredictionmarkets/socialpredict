---
title: Read Model Caching And Performance Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-07T00:00:00Z
updated_at_display: "Sunday, June 7, 2026"
update_reason: "Track implementation slices for display-safe caching, Postgres read models, optional Redis response caching, pagination, and correctness verification."
status: draft
---

# Read Model Caching And Performance Plan

## Purpose

This plan turns [11-read-model-caching-performance.md](./11-read-model-caching-performance.md) and [DESIGN.md](./DESIGN.md) into implementation slices.

## Planning Principles

- Do not use stale cache for order execution or settlement.
- Treat transaction-time anything as never cache-driven.
- Raw tables remain the audit source of truth.
- Read models must be testable against raw recomputation.
- Prefer pagination and simpler displays before broad caching.
- Make freshness explicit where users may care.
- Keep Redis optional until deployment posture is finalized.

## 01. Feature Artifact And Design Alignment

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/11/`.
- [x] Add feature overview.
- [x] Add design artifact.
- [x] Add implementation plan.
- [x] Review terminology against the canonical design plan.
- [x] Review implementation order with designer-agent postures.

## 02. Market Accounting Read Model Boundary

Service ownership: prediction market context and repository boundary.

Checklist:

- [x] Define domain type for market accounting snapshots.
- [x] Include net volume, market dust, volume with dust, probability, user count, bet count, and generated timestamp.
- [x] Add raw recomputation calculator for the snapshot.
- [x] Add tests proving snapshot fields match raw recomputation.
- [x] Keep display dust in the read model while preserving canonical transaction-time sale dust calculation.
- [x] Expose explicit market accounting freshness metadata.
- [x] Add boundary tests proving transaction repository interfaces do not expose snapshot methods.
- [x] Decide whether historical dust remains simple retained-dust convention or exact replay.
- [x] Ensure snapshot code does not affect order execution.

## 03. Display Snapshot Persistence

Service ownership: repository and migration boundary.

Checklist:

- [x] Add timestamped migration for durable read-model table(s) if needed.
- [x] Add model/repository methods for snapshot upsert/read.
- [x] Add generated-at and last-processed-bet tracking.
- [x] Add migration tests.
- [x] Add repository tests for snapshot writes/reads.
- [x] Add on-demand market accounting snapshot refresh service.

## 04. Market Discovery Cache

Service ownership: CMS/discovery and API boundary.

Checklist:

- [x] Add cached/read-model path for `/markets` card payloads.
- [x] Add cached/read-model path for `/markets/topic/:slug` card payloads.
- [x] Cache compact pinned chart payloads.
- [x] Use a target freshness of about 10 minutes for page-level discovery/pinned-card payloads.
- [x] Include freshness metadata where appropriate.
- [x] Invalidate discovery caches on tag/CMS layout changes.
- [x] Invalidate market card caches on bet/sale/status changes.

## 05. Statistics And Leaderboard Snapshots

Service ownership: analytics context.

Checklist:

- [x] Add CMS reporting visibility settings for public/login-required aggregate reporting.
- [x] Gate system metrics public access behind reporting visibility settings.
- [x] Gate global leaderboard public access behind reporting visibility settings.
- [x] Add admin CMS controls for reporting visibility settings.
- [x] Add system metrics snapshot read model.
- [x] Add global leaderboard snapshot read model.
- [x] Add market leaderboard snapshot read model.
- [x] Use a target freshness of about 1 hour for system financial metrics.
- [x] Use a target freshness of about 1 hour for global leaderboard snapshots.
- [x] Use a target freshness of about 10 minutes for market leaderboard snapshots.
- [x] Add scheduled or on-demand refresh service.
- [x] Add tests comparing snapshot outputs to raw recomputation.
- [x] Add pagination to global and market leaderboard responses.

## 05A. User Financial Metric Snapshots

Service ownership: user financial read-model boundary.

Checklist:

- [x] Identify individual user financial metrics that are computationally expensive.
- [x] Separate top-line transaction-sensitive balance/spend checks from display-only user financial summaries.
- [x] Define authenticated read-model shape for user financial snapshots.
- [x] Add durable snapshot persistence if the display path is expensive enough.
- [x] Add on-demand user financial snapshot refresh service.
- [x] Add freshness metadata to user financial read-model retrieval.
- [x] Add authenticated game-transparency user financial read-model endpoint with freshness metadata.
- [x] Keep user financial read-model endpoint unavailable to logged-out visitors.
- [x] Add stale-marker service and repository hooks for user financial snapshots after user bet/sale/market-resolution events.
- [ ] Add stale-marker calls for future cancellation/yank refund events when that feature exists.
- [x] Ensure user financial snapshots are never used for transaction decisions, spend checks, dust settlement, or payout/refund truth.
- [x] Add boundary tests proving buy/sell/user-balance transaction interfaces do not expose user financial snapshot services.
- [x] Add recomputation-vs-snapshot tests for user financial metrics.
- [ ] Consider Redis only after authenticated Postgres snapshots or read-model services prove correct and still become hot.

## 06. Market Detail Display Optimization

Service ownership: frontend/API boundary.

Checklist:

- [x] Keep transaction actions canonical and fresh.
- [x] Add pagination for market bets table, default latest 20.
- [x] Add pagination for market positions table, default top 20.
- [x] Keep market bets table uncached; optionally refresh/poll around 10 seconds after accepted transactions.
- [x] Cache or snapshot non-transactional market detail widgets around 1 minute.
- [x] Keep sale/buy confirmation responses authoritative.
- [x] Add UI freshness copy for cached widgets if useful.

## 07. Endpoint Boundary

Service ownership: API boundary.

Checklist:

- [x] Identify canonical transaction endpoints that must never read from display caches.
- [x] Identify cache-backed display/read-model endpoints.
- [x] Introduce explicit `/v0/read/...` route for user financial summaries.
- [x] Decide remaining `/v0/read/...` routes versus existing display handler rewrites.
- [x] Add shared read-model freshness metadata contract.
- [x] Add freshness metadata to user financial read-model display response.
- [x] Add domain boundary tests proving transaction interfaces do not expose read-model snapshot services.
- [x] Add API/domain tests proving transaction interfaces do not expose read-model cache services.

## 08. Optional Redis Layer

Service ownership: infrastructure and API boundary.

Checklist:

- [ ] Add Redis config/env posture behind feature flags.
- [ ] Define cache key conventions.
- [ ] Define TTL defaults by endpoint class.
- [ ] Add safe fallback when Redis is unavailable.
- [ ] Add integration tests with fake/in-memory cache where practical.
- [ ] Document production deployment requirements if Redis becomes required.

## 09. Verification And Load Testing

Service ownership: testing boundary.

Checklist:

- [x] Add recomputation-vs-snapshot tests for core read models.
- [x] Add handler/domain tests proving order endpoints do not read from display caches and only mark read models stale after success.
- [ ] Add load-test scenario for cached discovery pages.
- [x] Add market detail UI/API pagination for bets, positions, and leaderboard display.
- [x] Expose market leaderboard snapshot freshness in the API and UI.
- [x] Cache market positions as a display/read-model snapshot with freshness metadata.
- [ ] Capture before/after latency and CPU metrics.
- [ ] Update performance dossier with results.
