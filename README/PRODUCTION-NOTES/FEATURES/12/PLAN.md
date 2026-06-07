---
title: Moderator Work Profits Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-07T00:00:00Z
updated_at_display: "Sunday, June 7, 2026"
update_reason: "Track implementation slices for moderator work-profit payout and financial reporting."
status: draft
---

# Moderator Work Profits Plan

## Purpose

This plan turns [12-moderator-work-profits.md](./12-moderator-work-profits.md) and [DESIGN.md](./DESIGN.md) into implementation slices.

## 01. Feature Artifact And Design Alignment

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/12/`.
- [x] Add feature overview.
- [x] Add design artifact.
- [x] Add implementation plan.
- [x] Align terminology with market creator, steward, work income, and work profit.

## 02. Resolution-Time Payout

Checklist:

- [x] Add explicit `WORK_PROFIT` balance transaction type.
- [x] Wire `InitialBetFee` into the market service config.
- [x] Derive unique positive-bet participants from canonical bet history.
- [x] Credit work income after ordinary winner payout.
- [x] Skip work-profit payout for `N/A` refund resolution.
- [x] Add service tests for payout ordering and unique participant counting.

## 03. Financial Display

Checklist:

- [x] Derive user `workProfits` from resolved markets stewarded by that user.
- [x] Use stored `ProposalCost` as the market creation fee when available.
- [x] Fall back to configured `CreateMarketCost` for legacy rows with no proposal cost.
- [x] Include work profits in direct financial calculations.
- [x] Include work profits in user financial read-model refresh.
- [x] Update retained participation-fee system metric semantics to avoid double-counting paid-out steward work income.
- [x] Add tests for net work-profit derivation.

## 04. API And Frontend Surface

Checklist:

- [x] Reuse existing user financial `workProfits` response field.
- [ ] Add copy or tooltip if users need an explanation of work profits in the financial UI.
- [ ] Consider market-level steward earnings display after product language is settled.

## 05. Future Decisions

Checklist:

- [ ] Decide whether future economics changes require durable per-market initial fee capture.
- [ ] Decide whether cancelled/yanked markets should ever pay work income.
- [ ] Decide whether future governance needs durable resolved-by attribution beyond current steward-at-resolution.
