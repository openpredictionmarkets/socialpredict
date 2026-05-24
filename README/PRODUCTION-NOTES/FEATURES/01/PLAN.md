---
title: Moderator Mode Implementation Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Convert the moderator-mode plan into an agent-usable implementation checklist."
status: draft
---

# Moderator Mode Implementation Plan

## Purpose

This plan turns [DESIGN.md](./DESIGN.md) and [01-moderators.md](./01-moderators.md) into an implementation sequence.

The plan is intentionally split into reviewable slices. Moderator mode touches configuration, accounts, markets, ledger behavior, APIs, frontend dashboards, and tests; it should not land as one large branch.

Agents implementing this feature should mark checklist items as they complete them and leave unchecked items in place when intentionally deferred.

## Planning Principles

- Preserve open-mode behavior by default.
- Put backend domain policy ahead of frontend affordances.
- Add explicit seams before broad migrations.
- Keep accounting-sensitive behavior behind transaction-scoped use cases.
- Add Postgres-backed tests only where SQLite cannot prove the behavior.
- Keep every PR independently buildable and reviewable.

## Progress Ledger

- [x] 01. Feature artifact and design alignment
- [ ] 02. Game-mode configuration policy
- [ ] 03. Participant role and moderator status baseline
- [ ] 04. Market lifecycle and proposal creation
- [ ] 05. Admin approval and rejection use cases
- [ ] 06. Moderator views and frontend proposal tracking
- [ ] 07. Admin approval dashboard baseline
- [ ] 08. Moderator self-trade and suspension enforcement
- [ ] 09. Market contract immutability and amendments
- [ ] 10. Admin yank and cancellation refund unit of work
- [ ] 11. Postgres cancellation refund truth tests
- [ ] 12. Admin moderator dashboard expansion
- [ ] 13. End-to-end feature verification

## Implementation Checklist

### 01. Feature Artifact And Design Alignment

Status: complete for this documentation PR.

Checklist:

- [x] Keep the overview, design, and plan together under `README/PRODUCTION-NOTES/FEATURES/01/`.
- [x] Move the moderator overview to `01/01-moderators.md`.
- [x] Add `DESIGN.md` aligned with the canonical design plan boundaries.
- [x] Add `PLAN.md` as an agent-usable implementation sequence.
- [x] Update production-note index links.
- [x] Keep the change documentation-only.

Exit criteria:

- [x] Documentation describes product behavior, domain design, and implementation sequence separately.
- [x] No runtime behavior changes.

Validation:

- [x] `git diff --check`
- [x] `git diff --cached --check`

### 02. Game-Mode Configuration Policy

Checklist:

- [ ] Extend setup/application policy with default `open` game mode.
- [ ] Add moderation config fields for approval-required and moderator trading policy.
- [ ] Add typed moderation config in the Configuration Service Slice.
- [ ] Expose game-mode policy to domain services through narrow interfaces.
- [ ] Add tests proving missing config defaults to open mode.
- [ ] Add tests proving moderator-mode config parses and is visible through typed policy.
- [ ] Update related docs if the config shape changes from [01-moderators.md](./01-moderators.md).

Exit criteria:

- [ ] Existing installs behave as open mode without setup changes.
- [ ] Moderator-mode config can be parsed and read from typed config.
- [ ] No frontend-only flag controls market creation policy.

Validation:

- [ ] `go test ./...`
- [ ] `git diff --check`

### 03. Participant Role And Moderator Status Baseline

Checklist:

- [ ] Introduce stable role constants or typed values for `ADMIN`, `REGULAR`, and `MODERATOR`.
- [ ] Add moderator status representation with at least `active` and `suspended` semantics.
- [ ] Add suspension reason, actor, and timestamp storage.
- [ ] Add role/suspension audit records or an explicit audit seam.
- [ ] Add admin-domain use cases for promote, suspend, and unsuspend without broad dashboard work.
- [ ] Add tests for role/status state transitions.
- [ ] Add tests that suspended moderators are distinguishable from demoted or deleted users.

Exit criteria:

- [ ] Moderator status is represented in backend policy, not only UI copy.
- [ ] Suspended moderators are distinguishable from demoted or deleted users.
- [ ] Role/suspension changes are auditable.

Validation:

- [ ] `go test ./...`
- [ ] `git diff --check`

### 04. Market Lifecycle And Proposal Creation

Checklist:

- [ ] Add market lifecycle or approval state support for `proposed`, `rejected`, `published`, `closed`, `resolved`, and `cancelled` behavior.
- [ ] Preserve compatibility with existing public statuses where needed.
- [ ] In moderator mode, make `POST /v0/markets` create `proposed` markets for active moderators.
- [ ] In open mode, preserve existing create-market behavior.
- [ ] Prevent proposed markets from appearing as tradable public markets.
- [ ] Prevent rejected markets from appearing as tradable public markets.
- [ ] Prevent cancelled markets from appearing as tradable public markets.
- [ ] Add domain tests for lifecycle transitions.
- [ ] Add handler/API tests for open-mode and moderator-mode creation behavior.

Exit criteria:

- [ ] Proposed markets are not tradable.
- [ ] Existing open-mode tests continue to pass.
- [ ] Market lifecycle terminology is consistent with [DESIGN.md](./DESIGN.md).

Validation:

- [ ] `go test ./...`
- [ ] `git diff --check`

### 05. Admin Approval And Rejection Use Cases

Checklist:

- [ ] Add admin use case for approving proposed markets.
- [ ] Add admin use case for rejecting proposed markets.
- [ ] Add repository methods required by approval/rejection use cases.
- [ ] Require confirmation semantics at the API/application boundary for approval.
- [ ] Store approval actor and timestamp.
- [ ] Store rejection actor, timestamp, and reason.
- [ ] Add authorization checks so non-admins cannot approve or reject.
- [ ] Add OpenAPI entries for approval and rejection endpoints.
- [ ] Add public reason responses for invalid state and unauthorized approval/rejection attempts.
- [ ] Add tests for approve, reject, unauthorized, and wrong-state cases.

Exit criteria:

- [ ] Admin can approve a proposed market into published/tradable state.
- [ ] Admin can reject a proposal with reason.
- [ ] Non-admins cannot approve or reject.
- [ ] Approval/rejection history is preserved.

Validation:

- [ ] `go test ./...`
- [ ] OpenAPI/contract checks if available for changed routes.
- [ ] `git diff --check`

### 06. Moderator Views And Frontend Proposal Tracking

Checklist:

- [ ] Add moderator API for markets created by the current moderator.
- [ ] Include proposed, rejected, published, closed, resolved, and cancelled markets in the appropriate moderator view.
- [ ] Add frontend route or dashboard surface for moderator proposal tracking.
- [ ] Use backend lifecycle terms in frontend copy.
- [ ] Use the existing frontend API/auth adapter patterns.
- [ ] Avoid direct token/API coupling in new frontend code.
- [ ] Add focused frontend tests if the test baseline exists by this point.

Exit criteria:

- [ ] Moderators can see proposal status without admin dashboard access.
- [ ] Frontend does not invent a separate lifecycle model.
- [ ] Existing frontend CI/build remains green.

Validation:

- [ ] `go test ./...` if backend routes change.
- [ ] Frontend build workflow or `npm run build:report` if frontend changes.
- [ ] `git diff --check`

### 07. Admin Approval Dashboard Baseline

Checklist:

- [ ] Add admin UI for proposed-market queue.
- [ ] Show market title, description, labels, outcome type, and resolution time.
- [ ] Show creator username, moderator status, and creation timestamp.
- [ ] Show prior review history where available.
- [ ] Add approve confirmation prompt.
- [ ] Add reject reason flow.
- [ ] Align the UI with Tailwind/styleguide direction.
- [ ] Keep backend authorization authoritative.

Exit criteria:

- [ ] Admins can review, approve, and reject from UI.
- [ ] Approval is two-step.
- [ ] Rejection captures reason.

Validation:

- [ ] Frontend build workflow or `npm run build:report`.
- [ ] `go test ./...` if backend route behavior changes.
- [ ] `git diff --check`

### 08. Moderator Self-Trade And Suspension Enforcement

Checklist:

- [ ] Enforce self-trade guard for moderator-created markets on buy path.
- [ ] Enforce self-trade guard for moderator-created markets on sell path.
- [ ] Enforce suspended-moderator restriction on market creation.
- [ ] Enforce suspended-moderator restriction on amendment creation.
- [ ] Enforce suspended-moderator restriction on resolution.
- [ ] Add domain tests for buy/sell self-trade restrictions.
- [ ] Add handler/API tests proving clients cannot bypass UI restrictions.
- [ ] Add public reason responses for forbidden self-trade and suspended moderator actions.

Exit criteria:

- [ ] API clients cannot bypass UI restrictions.
- [ ] Buy and sell paths reject forbidden self-trade consistently.
- [ ] Suspended moderators cannot perform moderator-only actions.

Validation:

- [ ] `go test ./...`
- [ ] `git diff --check`

### 09. Market Contract Immutability And Amendments

Checklist:

- [ ] Preserve original title without in-place overwrite.
- [ ] Preserve original description without in-place overwrite.
- [ ] Add append-only contract amendment records.
- [ ] Generate backend-owned change records.
- [ ] Generate contract version references on approved amendments.
- [ ] Require admin approval for published moderator-market amendments unless actor is admin.
- [ ] Add API route or use case for creating amendments.
- [ ] Add API route or use case for reading market change records.
- [ ] Add collapsed frontend presentation for change record on market detail.
- [ ] Add tests for immutability, amendment approval, and contract version references.

Exit criteria:

- [ ] Original title remains immutable.
- [ ] Original description is not overwritten in place.
- [ ] Change records are backend-generated and ordered.
- [ ] Contract version references are stable enough for audit and tests.

Validation:

- [ ] `go test ./...`
- [ ] Frontend build workflow or `npm run build:report` if frontend changes.
- [ ] OpenAPI/contract checks if routes change.
- [ ] `git diff --check`

### 10. Admin Yank And Cancellation Refund Unit Of Work

Checklist:

- [ ] Add admin cancellation/yank use case.
- [ ] Store cancellation actor, reason, timestamp, and contract version reference.
- [ ] Move market into terminal `cancelled` state.
- [ ] Generate cancellation refund ledger entries.
- [ ] Commit cancellation state update and refund ledger entries atomically.
- [ ] Prevent buy after cancellation.
- [ ] Prevent sell after cancellation.
- [ ] Prevent amendment after cancellation.
- [ ] Prevent resolution after cancellation.
- [ ] Add tests for cancellation authorization, wrong-state cancellation, and post-cancellation action rejection.

Exit criteria:

- [ ] Cancellation is not implemented as hard delete.
- [ ] Cancellation is distinguishable from ordinary creator `N/A` resolution.
- [ ] Cancellation state and refund ledger entries commit or roll back together.

Validation:

- [ ] `go test ./...`
- [ ] Targeted transaction/unit-of-work tests.
- [ ] `git diff --check`

### 11. Postgres Cancellation Refund Truth Tests

Checklist:

- [ ] Add DSN-gated Postgres test for one user buys and never sells.
- [ ] Add DSN-gated Postgres test for one user buys, partially sells, then market is yanked.
- [ ] Add DSN-gated Postgres test for one user buys, fully sells out, then market is yanked.
- [ ] Add DSN-gated Postgres test for multiple users on both outcomes.
- [ ] Verify cancellation is atomic with market state update.
- [ ] Verify final balances equal original balance minus net unrecovered exposure.
- [ ] Verify all cancelled-market positions have zero remaining claim value.
- [ ] Document the DSN environment variable required to run the tests if not already documented.

Exit criteria:

- [ ] Refund math is proven against Postgres behavior.
- [ ] Test is opt-in like existing Postgres truth tests.
- [ ] SQLite helpers are not treated as sufficient for this accounting behavior.

Validation:

- [ ] Standard `go test ./...` skips DSN-gated tests when DSN is absent.
- [ ] DSN-enabled Postgres test command passes where a local/test Postgres DSN is available.
- [ ] `git diff --check`

### 12. Admin Moderator Dashboard Expansion

Checklist:

- [ ] Add users dashboard capabilities for promote, suspend, and unsuspend.
- [ ] Add moderators dashboard or moderator-filtered admin surface.
- [ ] Show moderator status.
- [ ] Show suspension reason and timestamp.
- [ ] Show created-market counts by lifecycle state.
- [ ] Preserve audit history access.
- [ ] Add frontend states for loading, success, and failure.
- [ ] Keep backend authorization authoritative.

Exit criteria:

- [ ] Admin can manage moderators from UI.
- [ ] Moderator status changes are visible and auditable.
- [ ] Dashboard does not bypass backend authorization.

Validation:

- [ ] Frontend build workflow or `npm run build:report`.
- [ ] `go test ./...` if backend APIs change.
- [ ] `git diff --check`

### 13. End-To-End Feature Verification

Checklist:

- [ ] Add end-to-end or integration coverage for admin promotes participant to moderator.
- [ ] Cover moderator proposes market.
- [ ] Cover admin approves proposed market.
- [ ] Cover published market becomes tradable.
- [ ] Cover moderator self-trade is rejected.
- [ ] Cover admin rejects proposed market with reason.
- [ ] Cover suspended moderator cannot create or resolve.
- [ ] Cover admin yanks market and affected participants see cancellation/refund state.
- [ ] Keep tests focused on user-visible flows rather than duplicating all domain unit tests.

Exit criteria:

- [ ] A reviewer can validate the complete moderator-mode path.
- [ ] CI remains stable and not dependent on unavailable local secrets.

Validation:

- [ ] Relevant backend integration command.
- [ ] Relevant frontend or E2E command if tooling exists.
- [ ] `git diff --check`

## First Implementation Target

The first code PR should be `02. Game-Mode Configuration Policy`.

Reason:

- It preserves backward compatibility.
- It creates the policy seam every later slice needs.
- It aligns with the canonical Configuration Service Slice instead of scattering hidden flags.

## Dependency Notes

- Role/status work should precede proposal creation enforcement.
- Proposal lifecycle should precede admin dashboard work.
- Cancellation refund design should wait until lifecycle and ledger seams are explicit.
- Frontend dashboard work should wait for stable backend route/reason contracts.
- Contract amendments can be developed after basic approve/reject behavior, but before relying on admin yanks for amendment abuse cases.

## Review Checklist

Before merging a moderator-mode implementation PR, check:

- [ ] Does open mode still behave as before?
- [ ] Is the new rule enforced in backend policy rather than only UI?
- [ ] Is public error/reason language stable and documented?
- [ ] Does the PR avoid hard deletion for audit-relevant state?
- [ ] Are financial writes transaction-scoped where required?
- [ ] Are migrations additive and reviewable?
- [ ] Does frontend terminology match backend/domain terminology?
- [ ] Does the PR avoid introducing generic RBAC/workflow/platform abstractions ahead of need?

## Deferred Until After Baseline

- AI moderation or scoring.
- General workflow engine.
- Multi-tenant policy configuration UI.
- Advanced analytics around moderator quality.
- Full audit-log search platform.
- Dedicated moderation microservice.
