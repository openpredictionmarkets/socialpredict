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

- [x] Add request DTO for multiple-choice binary group creation.
- [x] Validate parent title and description using existing market rules.
- [x] Validate answer labels: minimum count, maximum count, length, uniqueness, and sanitization.
- [ ] Create parent group and child binary markets transactionally.
- [x] Charge one group proposal cost in baseline.
- [x] Link child markets through `market_group_members`.
- [x] Add service tests for successful group creation.
- [x] Add service tests for invalid labels, duplicate labels, too few answers, and too many answers.

Implementation note:

- The current vertical slice creates child markets and then links the parent group. A fully atomic repository/service transaction across child-market creation plus group creation remains pending.

## 04. Admin Review And Governance

Checklist:

- [ ] Add grouped proposal review API response.
- [ ] Add approve group action that publishes parent and children together.
- [ ] Add reject group action that rejects parent and children together and refunds proposal cost.
- [ ] Add stewardship reassignment for parent and child markets together.
- [ ] Add audit records for group approval/rejection/stewardship decisions.
- [ ] Add admin UI to review parent question, description, tags, generated child markets, and answer labels.
- [ ] Collapse Proposed, Published, and Rejected admin queue rows by parent group.
- [ ] Collapse Proposed, Published, and Rejected moderator profile rows by parent group.
- [ ] Ensure grouped queue actions operate on the parent group rather than forcing manual per-child approval/rejection.

## 05. Trading And Transaction Boundary

Checklist:

- [ ] Keep buy/sell endpoints child-market-scoped.
- [ ] Ensure group display never drives transaction decisions.
- [ ] Ensure sale quote and sale execution still use child market canonical state.
- [ ] Ensure child market read models and parent group read models are display-only.
- [ ] Add boundary tests proving transaction interfaces do not depend on market group read models.

## 06. Group Read Models And Discovery

Checklist:

- [x] Add read endpoint for group summary and ordered answers.
- [x] Add compact answer card payloads with child market probability, volume, and users.
- [ ] Add freshness metadata to group read responses.
- [x] Prefer parent group rows in `/markets`, search, and topic pages when child markets carry group metadata.
- [x] Decide whether child markets are hidden from normal discovery or shown with group context.
- [ ] Add search behavior for parent title and answer labels.
- [x] Add tag projection or inheritance behavior for group discovery.
- [x] Add structural discovery invalidation so newly created groups do not wait behind soft-stale cache windows.

Decision:

- Discovery and search collapse grouped children into one row.
- The row links to a representative child market URL.
- The child market page renders a consolidated grouped view when `marketGroup` metadata is present.

## 07. Frontend Creation And Display

Checklist:

- [x] Add market type selector to `/create`.
- [x] Add multiple-choice answer editor with add/remove behavior before submit.
- [ ] Add answer reorder behavior before submit.
- [x] Add copy explaining independent binary answer markets.
- [x] Add compatibility parent group route.
- [x] Render answer cards with child probability, volume, and trade affordance.
- [x] Add consolidated comparison chart across child answers.
- [x] Preserve direct child market pages.
- [x] Make direct child market pages render the consolidated grouped view by default.
- [x] Add per-answer trade controls on the consolidated grouped view.
- [x] Add responsive/mobile layout for answer cards.
- [ ] Add accessibility labels and keyboard support for answer editor and answer cards.

## 08. Resolution

Checklist:

- [ ] Add independent child resolution flow from group page/admin page.
- [ ] Add optional exclusive helper that resolves one child YES and remaining children NO.
- [ ] Add group N/A helper that resolves every child N/A.
- [ ] Ensure each helper calls existing child-market resolution/refund/payout logic.
- [ ] Add tests proving group helper payouts match individually resolving child markets.
- [ ] Add tests proving resolved children cannot be traded afterward.

## 08A. Group Amendments

Checklist:

- [ ] Decide parent-owned group amendment persistence vs. child-specific amendment projection.
- [ ] Add group amendment proposal API.
- [ ] Add group amendment admin review API.
- [ ] Add group amendment moderator status API.
- [ ] Show group amendments on parent group pages.
- [ ] Tie group amendment review rows to the parent group and ordered child answers.
- [ ] Preserve child-market transaction boundaries; amendments remain display/governance state.

## 08B. Answer Addition Policy

Checklist:

- [ ] Add explicit answer addition policy enum: `NO_ONE`, `CREATOR_ONLY`, `ANY_ACTIVE_MODERATOR`.
- [ ] Keep answer additions disabled by default.
- [ ] Add admin control for enabling answer additions.
- [ ] Add moderator proposal flow for new answers when policy allows.
- [ ] Create added answers as new normal child binary markets.
- [ ] Add audit trail for who proposed/approved added answers.
- [ ] Ensure answer additions never rewrite existing child market bet or payout history.

## 09. Testing And Verification

Checklist:

- [x] Add domain tests for creation and group read models.
- [ ] Add domain tests for approval, rejection, stewardship, and resolution.
- [x] Add repository tests for group/member persistence.
- [ ] Add handler tests for create/review/read endpoints.
- [x] Add OpenAPI schema entries and contract tests.
- [ ] Add frontend smoke path for creating, reviewing, viewing, and trading a group child market.
- [x] Add migration tests.
- [x] Run relevant Go package tests.
- [x] Run frontend build.

## 10. Future Decisions

Checklist:

- [ ] Decide whether a true `SUM_TO_ONE_EXCLUSIVE` coupled market type is needed.
- [ ] Decide whether answer additions after approval are ever allowed.
- [ ] Decide whether group proposal cost should scale with answer count.
- [ ] Decide whether child markets should support answer-specific amendments.
- [ ] Decide whether group-level work profits should aggregate child market participation fees.
- [ ] Decide whether group pages need a normalized illustrative display separate from tradable child probabilities.
