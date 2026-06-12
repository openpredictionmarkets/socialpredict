---
title: Multiple Choice Binary Markets Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-12T00:00:00Z
updated_at_display: "Friday, June 12, 2026"
update_reason: "Track implementation slices for multiple-choice binary market groups."
status: draft
---

# Multiple Choice Binary Markets Plan

## Purpose

This plan turns [13-multiple-choice-binary-markets.md](./13-multiple-choice-binary-markets.md) and [DESIGN.md](./DESIGN.md) into implementation slices.

## 01. Feature Artifact And Design Alignment

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/13/`.
- [x] Add feature overview.
- [x] Add design artifact.
- [x] Add implementation plan.
- [x] Keep external platform research out of the production notes unless it becomes an explicit SocialPredict product decision.
- [x] Decide baseline probability policy: independent binary child markets, not sum-to-one.
- [x] Align with existing design-plan rules for backend-owned market truth, frontend experience boundaries, and timestamped additive migrations.

## 02. Domain Model And Migration

Checklist:

- [x] Add domain types for `MarketGroup` and `MarketGroupMember`.
- [x] Add persistence models for parent groups and group members.
- [x] Add timestamped Go migration for `market_groups`.
- [x] Add timestamped Go migration for `market_group_members`.
- [x] Register compact migration IDs in backend migration registry.
- [x] Add migration tests proving schema creation, indexes, and default values.
- [x] Add repository tests for group create/read/list and member ordering.
- [x] Keep existing `markets` rows as canonical child trading entities.

Implementation note:

- `20260612090000` creates both `market_groups` and `market_group_members` as one additive schema slice.
- The `MarketGroupRepository` interface is intentionally separate from the existing binary market `Repository` so buy/sell/resolve transaction paths do not accidentally depend on grouped display state.

## 03. Creation And Validation

Checklist:

- [ ] Add request DTO for multiple-choice binary group creation.
- [ ] Validate parent title and description using existing market rules.
- [ ] Validate answer labels: minimum count, maximum count, length, uniqueness, and sanitization.
- [ ] Create parent group and child binary markets transactionally.
- [ ] Charge one group proposal cost in baseline.
- [ ] Link child markets through `market_group_members`.
- [ ] Add service tests for successful group creation.
- [ ] Add service tests for invalid labels, duplicate labels, too few answers, and too many answers.

## 04. Admin Review And Governance

Checklist:

- [ ] Add grouped proposal review API response.
- [ ] Add approve group action that publishes parent and children together.
- [ ] Add reject group action that rejects parent and children together and refunds proposal cost.
- [ ] Add stewardship reassignment for parent and child markets together.
- [ ] Add audit records for group approval/rejection/stewardship decisions.
- [ ] Add admin UI to review parent question, description, tags, generated child markets, and answer labels.

## 05. Trading And Transaction Boundary

Checklist:

- [ ] Keep buy/sell endpoints child-market-scoped.
- [ ] Ensure group display never drives transaction decisions.
- [ ] Ensure sale quote and sale execution still use child market canonical state.
- [ ] Ensure child market read models and parent group read models are display-only.
- [ ] Add boundary tests proving transaction interfaces do not depend on market group read models.

## 06. Group Read Models And Discovery

Checklist:

- [ ] Add read-model endpoint for group summary and ordered answers.
- [ ] Add compact answer card payloads with child market probability, volume, users, and chart snapshot.
- [ ] Add freshness metadata to group read responses.
- [ ] Prefer parent group cards in `/markets` and topic pages.
- [ ] Decide whether child markets are hidden from normal discovery or shown with group context.
- [ ] Add search behavior for parent title and answer labels.
- [ ] Add tag projection or inheritance behavior for group discovery.

## 07. Frontend Creation And Display

Checklist:

- [ ] Add market type selector to `/create`.
- [ ] Add multiple-choice answer editor with add/remove/reorder behavior before submit.
- [ ] Add copy explaining independent binary answer markets.
- [ ] Add parent group page route.
- [ ] Render answer cards with child probability, chart, volume, and trade affordance.
- [ ] Preserve direct child market pages.
- [ ] Add responsive/mobile layout for answer cards.
- [ ] Add accessibility labels and keyboard support for answer editor and answer cards.

## 08. Resolution

Checklist:

- [ ] Add independent child resolution flow from group page/admin page.
- [ ] Add optional exclusive helper that resolves one child YES and remaining children NO.
- [ ] Add group N/A helper that resolves every child N/A.
- [ ] Ensure each helper calls existing child-market resolution/refund/payout logic.
- [ ] Add tests proving group helper payouts match individually resolving child markets.
- [ ] Add tests proving resolved children cannot be traded afterward.

## 09. Testing And Verification

Checklist:

- [ ] Add domain tests for creation, approval, rejection, stewardship, resolution, and group read models.
- [ ] Add repository tests for group/member persistence.
- [ ] Add handler tests for create/review/read endpoints.
- [ ] Add OpenAPI schema entries and contract tests.
- [ ] Add frontend smoke path for creating, reviewing, viewing, and trading a group child market.
- [ ] Add migration tests.
- [ ] Run relevant Go package tests.
- [ ] Run frontend build.

## 10. Future Decisions

Checklist:

- [ ] Decide whether a true `SUM_TO_ONE_EXCLUSIVE` coupled market type is needed.
- [ ] Decide whether answer additions after approval are ever allowed.
- [ ] Decide whether group proposal cost should scale with answer count.
- [ ] Decide whether child markets should support answer-specific amendments.
- [ ] Decide whether group-level work profits should aggregate child market participation fees.
- [ ] Decide whether group pages need a normalized illustrative display separate from tradable child probabilities.
