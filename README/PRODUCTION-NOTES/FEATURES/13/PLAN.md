---
title: N/A Neutral Unwind Refunds Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-07T00:00:00Z
updated_at_display: "Sunday, June 7, 2026"
update_reason: "Track implementation slices for N/A neutral unwind refunds and zero-position handling."
status: proposed
---

# N/A Neutral Unwind Refunds Plan

## Purpose

This plan turns [13-na-neutral-unwind-refunds.md](./13-na-neutral-unwind-refunds.md) and [DESIGN.md](./DESIGN.md) into implementation slices.

## 01. Feature Artifact And Design Alignment

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/13/`.
- [x] Add feature overview.
- [x] Add design artifact.
- [x] Add implementation plan.
- [x] Define `N/A` as neutral unwind, not final `50%` outcome.
- [x] Define zero-position/effective-exposure invariant.
- [x] Define divide-by-zero safety requirements.

## 02. Refund Calculator

Checklist:

- [ ] Add a dedicated neutral unwind refund calculator or an explicit neutral-resolution adapter around the existing payout path.
- [ ] Load canonical chronological bet history.
- [ ] Derive current claim holders without using cached display snapshots.
- [ ] Use neutral `P = 0.5` for refund valuation only.
- [ ] Calculate `refundPool = VolumeWithDust`.
- [ ] Allocate primary refunds.
- [ ] Allocate residual credits deterministically to current claim holders by earliest positive participation, then username.
- [ ] Assert `sum(refunds) == refundPool` before balance mutation.
- [ ] Reject impossible negative refund-pool states.

## 03. Resolution Transaction Path

Checklist:

- [ ] Replace the current raw-bet `N/A` refund path with neutral unwind refund policy.
- [ ] Apply `REFUND` transactions from the computed allocation.
- [ ] Mark market resolved `N/A` only after refund allocation is valid.
- [ ] Do not pay `WORK_PROFIT` on `N/A`.
- [ ] Do not refund market creation/proposal cost.
- [ ] Do not claw back prior sale proceeds.
- [ ] Invalidate affected read models after successful `N/A` resolution.

## 04. Position And Display Semantics

Checklist:

- [ ] Confirm existing frontend resolve modal only exposes ordinary binary outcomes.
- [ ] Add moderator/admin `N/A` resolution option where authorized by governance policy.
- [ ] Add required confirmation disclosure before submitting `N/A`.
- [ ] Explain neutral `P = 0.5` refund valuation without presenting it as final probability.
- [ ] Explain no winner payout, no work-profit payout, no creation-cost refund, no sale-proceeds clawback, and zero effective positions.
- [ ] Ensure resolved `N/A` markets report zero effective YES shares.
- [ ] Ensure resolved `N/A` markets report zero effective NO shares.
- [ ] Ensure resolved `N/A` markets report zero effective position value.
- [ ] Ensure resolved `N/A` markets do not count as unresolved exposure.
- [ ] Ensure market detail charts show an `N/A` marker rather than a final directional probability.
- [ ] Ensure market cards show `N/A` / `?` state clearly.
- [ ] Ensure portfolio and financial views remove live exposure but preserve historical/refund facts.

## 05. Divide-By-Zero Safety

Checklist:

- [ ] Add guards in sell quote/sell paths for resolved `N/A` markets.
- [ ] Add guards in market detail probability/price display when terminal probability is not directional.
- [ ] Add guards in portfolio, financial, leaderboard, and market table calculations for zero denominators.
- [ ] Add tests for zero-volume `N/A` market detail.
- [ ] Add tests for zero-share `N/A` user positions.

## 06. Tests

Checklist:

- [ ] Unit test: only buys, `N/A` refund conserves pool.
- [ ] Unit test: buys and sells, prior sale proceeds are not clawed back.
- [ ] Unit test: retained dust is included in refund pool.
- [ ] Unit test: residual dust allocation is deterministic.
- [ ] Unit test: sold-out users do not receive residual dust.
- [ ] Unit test: `N/A` does not pay work profit.
- [ ] Unit test: market creation/proposal cost is not refunded.
- [ ] Integration test: Postgres `N/A` resolution leaves zero economy surplus/deficit.
- [ ] Integration test: portfolio and financial views show zero effective exposure after `N/A`.
- [ ] Regression test: no divide-by-zero on zero-volume or zero-position `N/A` markets.

## 07. Future Decisions

Checklist:

- [ ] Decide whether exact dust-payor attribution is worth new durable state.
- [ ] Decide whether initial participation fees should ever be refunded on voided markets.
- [ ] Decide whether `N/A` should require admin-only authority.
- [ ] Decide whether yanked/offensive markets share this refund policy or use a separate cancellation policy.
