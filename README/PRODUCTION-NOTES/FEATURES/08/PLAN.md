---
title: Market Yank And Cancellation Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-04T00:00:00Z
updated_at_display: "Thursday, June 4, 2026"
update_reason: "Create the implementation checklist for admin market yanking and cancellation refunds."
status: draft
---

# Market Yank And Cancellation Plan

## Purpose

This plan turns [08-market-yank-cancellation.md](./08-market-yank-cancellation.md) and [DESIGN.md](./DESIGN.md) into a backend-first implementation sequence.

Agents implementing this feature should mark checklist items as complete as they land and leave deferred items unchecked.

## Planning Principles

- Backend domain and accounting rules come before frontend controls.
- No hard deletion.
- No handler-owned refund math.
- No public visibility change without backend-owned public-safe projection rules.
- Preserve historical bets and write compensating facts where possible.
- Keep every PR reviewable and testable.

## 01. Feature Artifact And Design Alignment

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/08/`.
- [x] Add feature overview.
- [x] Add design artifact.
- [x] Add implementation plan.
- [ ] Review against canonical design plan and designer-agent posture.
- [ ] Decide exact accounting formula before code implementation.

## 02. Domain State And Migration

Service ownership: prediction market context and persistence boundary.

Checklist:

- [ ] Confirm existing `cancelled` lifecycle constant behavior.
- [ ] Add additive timestamped migration for cancellation metadata if needed.
- [ ] Add cancellation audit table or extend existing governance audit shape.
- [ ] Add domain model fields for cancellation actor/reasons/visibility/accounting status.
- [ ] Add repository mapping tests.
- [ ] Add lifecycle tests proving cancelled markets are non-tradable and non-resolvable through ordinary flow.

## 03. Accounting Design And Unit Of Work

Service ownership: betting/position ledger and participant account context.

Checklist:

- [ ] Define cancellation refund formula.
- [ ] Define fee/proposal-cost/subsidy treatment.
- [ ] Define handling for partial sells and prior payouts.
- [ ] Define behavior if clawback would exceed debt limits.
- [ ] Implement transaction-scoped cancellation accounting unit of work.
- [ ] Write unit tests for no bets, only buys, buys plus sells, multiple users, fees, and creator proposal cost.
- [ ] Add Postgres-backed transaction test if SQLite cannot prove rollback semantics.

## 04. Admin API Contract

Service ownership: API/auth boundary and markets domain.

Checklist:

- [ ] Add `PATCH /v0/admin/markets/{id}/yank`.
- [ ] Require admin authorization.
- [ ] Require confirmation and reason.
- [ ] Validate public/private reason lengths.
- [ ] Return cancellation accounting summary.
- [ ] Add failure reasons for invalid state, missing market, validation failure, and accounting failure.
- [ ] Update `backend/docs/openapi.yaml`.
- [ ] Add handler tests.
- [ ] Add schemathesis/go-kin API validation once OpenAPI is updated.

## 05. Public And Admin Read Models

Service ownership: market read models and frontend experience context.

Checklist:

- [ ] Include cancelled markets in admin governance search.
- [ ] Exclude cancelled markets from active/tradable public lists.
- [ ] Decide whether cancelled markets appear in public `all` or only via direct URL/user history.
- [ ] Add public-safe obfuscation projection.
- [ ] Add cancelled market response fields.
- [ ] Add tests for obfuscated and non-obfuscated cancelled market projections.

## 06. Frontend Admin UX

Service ownership: frontend admin dashboard.

Checklist:

- [ ] Add yank action to admin market governance/stewardship view.
- [ ] Add confirmation modal.
- [ ] Add private reason field.
- [ ] Add optional public reason field.
- [ ] Add obfuscation toggle.
- [ ] Show backend refund/accounting summary after success.
- [ ] Refresh governance list after yank.
- [ ] Hide yank action where backend says the state is ineligible.

## 07. Frontend Public UX

Service ownership: frontend market display.

Checklist:

- [ ] Show cancelled banner on direct market page.
- [ ] Disable/hide trade, sell, and resolve controls.
- [ ] Show public cancellation reason if available.
- [ ] Use obfuscated public title/description/labels when backend says to obfuscate.
- [ ] Keep user portfolio/financial history coherent for cancelled markets.

## 08. Analytics And Dossier Implications

Service ownership: analytics, stats, release dossier.

Checklist:

- [ ] Decide how cancelled markets affect platform stats.
- [ ] Decide how cancelled markets affect leaderboards.
- [ ] Decide how cancelled markets affect financial statements.
- [ ] Add evidence notes once cancellation accounting is tested.

## Exit Criteria

- Admin yank is backend-enforced and auditable.
- Cancellation accounting is transaction-scoped and tested.
- Cancelled markets are not tradable or ordinarily resolvable.
- Public cancelled-state display is safe and clear.
- Obfuscation mode protects public pages from unsafe market content.
- API contract and docs are updated before broad frontend rollout.
