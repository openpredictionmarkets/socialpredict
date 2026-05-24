---
title: Moderator Mode Implementation Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Add a sequenced implementation plan for moderator mode from the feature design."
status: draft
---

# Moderator Mode Implementation Plan

## Purpose

This plan turns [DESIGN.md](./DESIGN.md) and [01-moderators.md](./01-moderators.md) into an implementation sequence.

The plan is intentionally split into reviewable slices. Moderator mode touches configuration, accounts, markets, ledger behavior, APIs, frontend dashboards, and tests; it should not land as one large branch.

## Planning Principles

- Preserve open-mode behavior by default.
- Put backend domain policy ahead of frontend affordances.
- Add explicit seams before broad migrations.
- Keep accounting-sensitive behavior behind transaction-scoped use cases.
- Add Postgres-backed tests only where SQLite cannot prove the behavior.
- Keep every PR independently buildable and reviewable.

## Suggested PR Stack

### 01. Feature Artifact And Design Alignment

Scope:

- Keep the overview, design, and plan together under `README/PRODUCTION-NOTES/FEATURES/01/`.
- Update production-note index links.
- Confirm the feature design aligns with the canonical design plan boundaries.

Exit criteria:

- Documentation describes product behavior, domain design, and implementation sequence separately.
- No runtime behavior changes.

### 02. Game-Mode Configuration Policy

Scope:

- Extend setup/application policy with default `open` game mode.
- Add typed moderation config in the Configuration Service Slice.
- Expose game-mode policy to domain services through narrow interfaces.
- Add tests proving missing config defaults to open mode.

Exit criteria:

- Existing installs behave as open mode without setup changes.
- Moderator-mode config can be parsed and read from typed config.
- No frontend-only flag controls market creation policy.

### 03. Participant Role And Moderator Status Baseline

Scope:

- Introduce stable role constants or typed values for `ADMIN`, `REGULAR`, and `MODERATOR`.
- Add moderator status fields or table with `active` and `suspended` semantics.
- Add role/suspension audit records or an explicit audit seam.
- Add admin-domain use cases for promote, suspend, and unsuspend without broad dashboard work.

Exit criteria:

- Moderator status is represented in backend policy, not only UI copy.
- Suspended moderators are distinguishable from demoted or deleted users.
- Role/suspension changes are auditable.

### 04. Market Lifecycle And Proposal Creation

Scope:

- Add market lifecycle or approval state support for `proposed`, `rejected`, `published`, `closed`, `resolved`, and `cancelled` behavior.
- In moderator mode, `POST /v0/markets` creates `proposed` markets for active moderators.
- In open mode, existing create-market behavior remains compatible.
- Prevent proposed/rejected/cancelled markets from appearing as tradable public markets.

Exit criteria:

- Proposed markets are not tradable.
- Existing open-mode tests continue to pass.
- Market lifecycle terminology is consistent with DESIGN.md.

### 05. Admin Approval And Rejection Use Cases

Scope:

- Add admin use cases and repository methods for approving and rejecting proposed markets.
- Require confirmation semantics at the API/application boundary for approval.
- Store approval/rejection actor, reason, and timestamp.
- Add OpenAPI entries and public reason responses for invalid state or unauthorized approval attempts.

Exit criteria:

- Admin can approve a proposed market into published/tradable state.
- Admin can reject a proposal with reason.
- Non-admins cannot approve or reject.
- Approval/rejection history is preserved.

### 06. Moderator Views And Frontend Proposal Tracking

Scope:

- Add moderator APIs for their own proposed, rejected, published, closed, resolved, and cancelled markets.
- Add frontend routes/components for moderator proposal tracking.
- Keep UI language conforming to backend lifecycle terms.
- Use existing frontend API/auth adapter patterns.

Exit criteria:

- Moderators can see proposal status without admin dashboard access.
- Frontend does not invent a separate lifecycle model.
- Existing frontend CI/build remains green.

### 07. Admin Approval Dashboard Baseline

Scope:

- Add admin UI for proposed-market queue.
- Show market details, creator/moderator status, creation time, and prior review history.
- Add approve confirmation prompt and reject reason flow.
- Keep visual implementation aligned with Tailwind/styleguide direction.

Exit criteria:

- Admins can review, approve, and reject from UI.
- Approval is two-step.
- Rejection captures reason.

### 08. Moderator Self-Trade And Suspension Enforcement

Scope:

- Enforce self-trade guard for moderator-created markets on buy and sell paths.
- Enforce suspended-moderator restrictions on creation, amendment, and resolution paths.
- Add domain and handler tests for bypass attempts.

Exit criteria:

- API clients cannot bypass UI restrictions.
- Buy and sell paths reject forbidden self-trade consistently.
- Suspended moderators cannot perform moderator-only actions.

### 09. Market Contract Immutability And Amendments

Scope:

- Preserve original title and description without in-place overwrite.
- Add append-only contract amendment records.
- Generate contract version references on approved amendments.
- Require admin approval for published moderator-market amendments unless actor is admin.
- Add collapsed frontend presentation for change record on market detail.

Exit criteria:

- Original title remains immutable.
- Original description is not overwritten in place.
- Change records are backend-generated and ordered.
- Contract version references are stable enough for audit and tests.

### 10. Admin Yank And Cancellation Refund Unit Of Work

Scope:

- Add admin cancellation/yank use case.
- Store cancellation actor, reason, timestamp, and contract version reference.
- Move market into terminal `cancelled` state.
- Generate cancellation refund ledger entries atomically with market state update.
- Prevent further buy, sell, amendment, or resolution actions after cancellation.

Exit criteria:

- Cancellation is not implemented as hard delete.
- Cancellation is distinguishable from ordinary creator `N/A` resolution.
- Cancellation state and refund ledger entries commit or roll back together.

### 11. Postgres Cancellation Refund Truth Tests

Scope:

- Add DSN-gated Postgres tests for cancellation refund behavior.
- Cover no sell, partial sell, full sell-out, multi-user, and both-outcome scenarios.
- Verify final balances equal original balance minus net unrecovered exposure.
- Verify all cancelled-market positions have zero remaining claim value.

Exit criteria:

- Refund math is proven against Postgres behavior.
- Test is opt-in like existing Postgres truth tests.
- SQLite helpers are not treated as sufficient for this accounting behavior.

### 12. Admin Moderator Dashboard Expansion

Scope:

- Add users/moderators dashboard capabilities for promote, suspend, and unsuspend.
- Show moderator status, suspension reason, created-market counts, and lifecycle rollups.
- Preserve audit history access.

Exit criteria:

- Admin can manage moderators from UI.
- Moderator status changes are visible and auditable.
- Dashboard does not bypass backend authorization.

### 13. End-To-End Feature Verification

Scope:

- Add end-to-end or integration coverage for the full moderator mode journey after the stable APIs exist.
- Cover promote, propose, approve, trade restriction, resolve, amend, reject, suspend, and yank flows as appropriate.
- Keep tests focused on user-visible flows rather than duplicating all domain unit tests.

Exit criteria:

- A reviewer can validate the complete moderator-mode path.
- CI remains stable and not dependent on unavailable local secrets.

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

## Validation Expectations By Slice

Every code slice should run:

- `go test ./...` where backend code changes are involved, unless a known unrelated failure is documented.
- Frontend build workflow where frontend code changes are involved.
- `git diff --check`.

Additional targeted validation:

- Config slices should include defaulting tests.
- API slices should update `backend/docs/openapi.yaml` and any contract tests.
- Accounting slices should include DSN-gated Postgres tests.
- Frontend slices should preserve the frontend API/auth adapter pattern and avoid direct token/API coupling.

## Review Checklist

Before merging a moderator-mode implementation PR, check:

- Does open mode still behave as before?
- Is the new rule enforced in backend policy rather than only UI?
- Is public error/reason language stable and documented?
- Does the PR avoid hard deletion for audit-relevant state?
- Are financial writes transaction-scoped where required?
- Are migrations additive and reviewable?
- Does frontend terminology match backend/domain terminology?
- Does the PR avoid introducing generic RBAC/workflow/platform abstractions ahead of need?

## Deferred Until After Baseline

- AI moderation or scoring.
- General workflow engine.
- Multi-tenant policy configuration UI.
- Advanced analytics around moderator quality.
- Full audit-log search platform.
- Dedicated moderation microservice.
