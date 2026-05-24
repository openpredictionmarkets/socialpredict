---
title: Moderator Mode Implementation Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Clarify backend-first sequencing, API contract validation, service ownership, and migration expectations."
status: draft
---

# Moderator Mode Implementation Plan

## Purpose

This plan turns [DESIGN.md](./DESIGN.md) and [01-moderators.md](./01-moderators.md) into an implementation sequence.

The plan is intentionally split into reviewable slices. Moderator mode touches configuration, accounts, markets, ledger behavior, APIs, frontend dashboards, and tests; it should not land as one large branch.

Agents implementing this feature should mark checklist items as they complete them and leave unchecked items in place when intentionally deferred.

## Planning Principles

- Preserve open-mode behavior by default.
- Complete backend design and backend contract work before frontend implementation.
- Put backend domain policy ahead of frontend affordances.
- Keep user functionality in the users domain/service and market functionality in the markets domain/service.
- Touch the bets domain only for trade eligibility guards and cancellation/refund accounting that actually crosses buy/sell paths.
- Add explicit seams before broad migrations.
- Keep accounting-sensitive behavior behind transaction-scoped use cases.
- Add Postgres-backed tests only where SQLite cannot prove the behavior.
- Keep every PR independently buildable and reviewable.

## Backend-First Gate

Frontend work must not begin until these backend-facing items are stable enough to consume:

- [x] Setup/config exposes game-mode policy with default `open` behavior.
- [x] Users domain owns moderator role/status/suspension semantics.
- [ ] Markets domain owns proposal, approval, rejection, publication, amendment, resolution, and cancellation lifecycle semantics.
- [ ] Bets domain has only the necessary buy/sell guards for lifecycle/self-trade and any cancellation/refund accounting hooks that are actually required.
- [ ] `backend/docs/openapi.yaml` describes the new or changed routes and reason values.
- [ ] Go OpenAPI contract tests pass.
- [x] Migration and backend tests pass for the implemented backend slices.

## Progress Ledger

Backend design and contract baseline:

- [x] 01. Feature artifact and design alignment
- [x] 02. Game-mode configuration policy
- [x] 03. Participant role and moderator status baseline
- [x] 04. Market lifecycle and proposal creation
- [x] 05. Admin approval and rejection backend API
- [ ] 06. Moderator backend API and proposal views
- [ ] 07. Trade eligibility guards and suspension enforcement
- [ ] 08. Market contract immutability and backend amendments
- [ ] 09. Admin yank and cancellation refund unit of work
- [ ] 10. Postgres cancellation refund truth tests
- [ ] 11. Backend API contract completion gate

Frontend after backend contract:

- [ ] 12. Moderator frontend proposal tracking
- [ ] 13. Admin approval dashboard baseline
- [ ] 14. Admin moderator management dashboard
- [ ] 15. End-to-end feature verification

Frontend smoke-test exception:

- [x] Add a temporary admin market-review view that can approve or reject a known proposed market ID after the approval backend API exists.
- [x] Update market creation UI so moderator-mode `proposed` responses display the proposal ID instead of redirecting into a possibly non-public market detail page.
- [x] Keep this separate from the final proposal queue, because the moderator proposal list API is still planned in item 06.

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

Service ownership: Configuration Service Slice.

Checklist:

- [x] Extend setup/application policy with default `open` game mode.
- [x] Add moderation config fields for approval-required and moderator trading policy.
- [x] Add typed moderation config in the Configuration Service Slice.
- [x] Expose game-mode policy to domain services through narrow interfaces.
- [x] Add tests proving missing config defaults to open mode.
- [x] Add tests proving moderator-mode config parses and is visible through typed policy.
- [x] Update related docs if the config shape changes from [01-moderators.md](./01-moderators.md).

Exit criteria:

- [x] Existing installs behave as open mode without setup changes.
- [x] Moderator-mode config can be parsed and read from typed config.
- [x] No frontend-only flag controls market creation policy.

Validation:

- [x] `cd backend && go test ./...`
- [x] `git diff --check`

### 03. Participant Role And Moderator Status Baseline

Service ownership: users domain/service and users repository.

Checklist:

- [x] Introduce stable role constants or typed values for `ADMIN`, `REGULAR`, and `MODERATOR` in the users domain.
- [x] Add moderator status representation with at least `active` and `suspended` semantics.
- [x] Add suspension reason, actor, and timestamp storage.
- [x] Add role/suspension audit records or an explicit audit seam.
- [x] Add a timestamped Go migration under `backend/migration/migrations` for any new user role/status/audit columns or tables.
- [x] Add a package-local migration test for the role/status schema change where practical.
- [x] Add user-domain use cases for promote, suspend, and unsuspend without broad dashboard work.
- [x] Add tests for role/status state transitions.
- [x] Add tests that suspended moderators are distinguishable from demoted or deleted users.

Exit criteria:

- [x] Moderator status is represented in backend users policy, not only UI copy.
- [x] Suspended moderators are distinguishable from demoted or deleted users.
- [x] Role/suspension changes are auditable.

Validation:

- [x] `cd backend && go test ./...`
- [x] `git diff --check`

### 04. Market Lifecycle And Proposal Creation

Service ownership: markets domain/service and markets repository.

Checklist:

- [x] Add market lifecycle or approval state support for `proposed`, `rejected`, `published`, `closed`, `resolved`, and `cancelled` behavior.
- [x] Add a timestamped Go migration under `backend/migration/migrations` for lifecycle or approval-state storage.
- [x] Add a package-local migration test for lifecycle/default backfill behavior where practical.
- [x] Preserve compatibility with existing public statuses where needed.
- [x] In moderator mode, make `POST /v0/markets` create `proposed` markets for active moderators.
- [x] In open mode, preserve existing create-market behavior.
- [x] Prevent proposed markets from appearing as tradable public markets.
- [x] Prevent rejected markets from appearing as tradable public markets.
- [x] Prevent cancelled markets from appearing as tradable public markets.
- [x] Add domain tests for lifecycle transitions.
- [x] Add handler/API tests for open-mode and moderator-mode creation behavior.

Exit criteria:

- [x] Proposed markets are not tradable.
- [x] Existing open-mode tests continue to pass.
- [x] Market lifecycle terminology is consistent with [DESIGN.md](./DESIGN.md).

Validation:

- [x] `cd backend && go test ./...`
- [x] `git diff --check`

### 05. Admin Approval And Rejection Backend API

Service ownership: markets domain/service for market state transitions; users domain/service for actor authorization facts; handlers adapt HTTP only.

Checklist:

- [x] Add markets-domain use case for approving proposed markets.
- [x] Add markets-domain use case for rejecting proposed markets.
- [x] Add repository methods required by approval/rejection use cases.
- [x] Require confirmation semantics at the API/application boundary for approval.
- [x] Store approval actor and timestamp.
- [x] Store rejection actor, timestamp, and reason.
- [x] Add a timestamped Go migration under `backend/migration/migrations` if approval/rejection metadata needs new columns or tables.
- [x] Add a package-local migration test for approval/rejection schema defaults where practical.
- [x] Add authorization checks so non-admins cannot approve or reject.
- [x] Add or update admin/markets handlers for approval and rejection.
- [x] Update `backend/docs/openapi.yaml` for approval and rejection endpoints.
- [x] Add public reason responses for invalid state and unauthorized approval/rejection attempts.
- [x] Add tests for approve, reject, unauthorized, and wrong-state cases.

Exit criteria:

- [x] Admin can approve a proposed market into published/tradable state.
- [x] Admin can reject a proposal with reason.
- [x] Non-admins cannot approve or reject.
- [x] Approval/rejection history is preserved.
- [x] OpenAPI matches live routes and public reason values.

Validation:

- [x] `cd backend && go test ./...`
- [x] `cd backend && go test . -run 'TestOpenAPI|TestRouteFamily|TestReasonResponse|TestEmbedded|TestDocsPublishing'`
- [x] `git diff --check`

### 06. Moderator Backend API And Proposal Views

Service ownership: markets domain/service for moderator-created market queries; users domain/service for moderator identity/status facts.

Checklist:

- [ ] Add backend API for markets created by the current moderator.
- [ ] Include proposed, rejected, published, closed, resolved, and cancelled markets in the appropriate moderator view.
- [ ] Ensure backend response vocabulary uses market lifecycle terms from [DESIGN.md](./DESIGN.md).
- [ ] Ensure non-moderators cannot access moderator-only proposal views.
- [ ] Update `backend/docs/openapi.yaml` for moderator routes.
- [ ] Add public reason responses for non-moderator or suspended-moderator access where applicable.
- [ ] Add handler/domain tests for authorized and unauthorized moderator views.

Exit criteria:

- [ ] Moderators can retrieve proposal status without admin dashboard access.
- [ ] Backend API owns the lifecycle model that frontend will consume later.
- [ ] OpenAPI matches live routes and public reason values.

Validation:

- [ ] `cd backend && go test ./...`
- [ ] `cd backend && go test . -run 'TestOpenAPI|TestRouteFamily|TestReasonResponse|TestEmbedded|TestDocsPublishing'`
- [ ] `git diff --check`

### 07. Trade Eligibility Guards And Suspension Enforcement

Service ownership: markets domain/service owns market lifecycle eligibility; users domain/service owns moderator status; bets domain/service owns buy/sell guard integration only.

Checklist:

- [ ] Enforce proposed/rejected/cancelled market restrictions before buy operations.
- [ ] Enforce proposed/rejected/cancelled market restrictions before sell operations where needed.
- [ ] Enforce self-trade guard for moderator-created markets on buy path.
- [ ] Enforce self-trade guard for moderator-created markets on sell path only if current sell semantics could create forbidden exposure or recover value in a way policy disallows.
- [ ] Enforce suspended-moderator restriction on market creation.
- [ ] Enforce suspended-moderator restriction on amendment creation.
- [ ] Enforce suspended-moderator restriction on resolution.
- [ ] Add domain tests for lifecycle buy/sell restrictions.
- [ ] Add domain tests for moderator self-trade restrictions.
- [ ] Add handler/API tests proving clients cannot bypass UI restrictions.
- [ ] Add public reason responses for forbidden self-trade and suspended moderator actions.
- [ ] Update `backend/docs/openapi.yaml` if buy/sell or market action failure contracts change.

Exit criteria:

- [ ] API clients cannot bypass UI restrictions.
- [ ] Buy and sell paths reject forbidden actions consistently where policy requires.
- [ ] Suspended moderators cannot perform moderator-only actions.
- [ ] Bets code changes are limited to eligibility guard integration and necessary accounting seams, not a broad bets redesign.

Validation:

- [ ] `cd backend && go test ./...`
- [ ] `cd backend && go test . -run 'TestOpenAPI|TestReasonResponse'` if API reason values change.
- [ ] `git diff --check`

### 08. Market Contract Immutability And Backend Amendments

Service ownership: markets domain/service and markets repository.

Checklist:

- [ ] Preserve original title without in-place overwrite.
- [ ] Preserve original description without in-place overwrite.
- [ ] Add append-only contract amendment records.
- [ ] Add a timestamped Go migration under `backend/migration/migrations` for contract amendment/change-record storage.
- [ ] Add a package-local migration test for amendment/change-record schema creation where practical.
- [ ] Generate backend-owned change records.
- [ ] Generate contract version references on approved amendments.
- [ ] Require admin approval for published moderator-market amendments unless actor is admin.
- [ ] Add API route or use case for creating amendments.
- [ ] Add API route or use case for reading market change records.
- [ ] Update `backend/docs/openapi.yaml` for amendment and change-record routes.
- [ ] Add tests for immutability, amendment approval, and contract version references.

Exit criteria:

- [ ] Original title remains immutable.
- [ ] Original description is not overwritten in place.
- [ ] Change records are backend-generated and ordered.
- [ ] Contract version references are stable enough for audit and tests.
- [ ] OpenAPI matches live routes and public reason values.

Validation:

- [ ] `cd backend && go test ./...`
- [ ] `cd backend && go test . -run 'TestOpenAPI|TestRouteFamily|TestReasonResponse|TestEmbedded'`
- [ ] `git diff --check`

### 09. Admin Yank And Cancellation Refund Unit Of Work

Service ownership: markets domain/service owns cancellation state; bets/ledger repository owns refund accounting where existing ledger data is required; users domain/service owns account-balance mutation interface.

Checklist:

- [ ] Add admin cancellation/yank use case.
- [ ] Store cancellation actor, reason, timestamp, and contract version reference.
- [ ] Add a timestamped Go migration under `backend/migration/migrations` for cancellation metadata, cancellation/refund event linkage, or related ledger fields.
- [ ] Add a package-local migration test for cancellation/refund schema changes where practical.
- [ ] Move market into terminal `cancelled` state.
- [ ] Generate cancellation refund ledger entries.
- [ ] Commit cancellation state update and refund ledger entries atomically.
- [ ] Prevent buy after cancellation.
- [ ] Prevent sell after cancellation.
- [ ] Prevent amendment after cancellation.
- [ ] Prevent resolution after cancellation.
- [ ] Add tests for cancellation authorization, wrong-state cancellation, and post-cancellation action rejection.
- [ ] Update `backend/docs/openapi.yaml` for cancellation/yank routes and failure reasons.

Exit criteria:

- [ ] Cancellation is not implemented as hard delete.
- [ ] Cancellation is distinguishable from ordinary creator `N/A` resolution.
- [ ] Cancellation state and refund ledger entries commit or roll back together.
- [ ] OpenAPI matches live routes and public reason values.

Validation:

- [ ] `cd backend && go test ./...`
- [ ] Targeted transaction/unit-of-work tests.
- [ ] `cd backend && go test . -run 'TestOpenAPI|TestReasonResponse'` if API reason values change.
- [ ] `git diff --check`

### 10. Postgres Cancellation Refund Truth Tests

Service ownership: repository/unit-of-work tests for accounting truth.

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

- [ ] Standard `cd backend && go test ./...` skips DSN-gated tests when DSN is absent.
- [ ] DSN-enabled Postgres test command passes where a local/test Postgres DSN is available.
- [ ] `git diff --check`

### 11. Backend API Contract Completion Gate

Service ownership: API and Auth Contract Boundary.

Checklist:

- [ ] Confirm every new backend route is registered in `backend/server/server.go`.
- [ ] Confirm every new backend route is documented in `backend/docs/openapi.yaml`.
- [ ] Confirm `ReasonResponse` enum includes any new public reason values.
- [ ] Confirm `x-route-family-migration-matrix` stays truthful for changed route families.
- [ ] Confirm `backend/docs/README.md` still describes the source-of-truth order accurately.
- [ ] Confirm generated or embedded OpenAPI assets still match the maintained YAML.
- [ ] Run existing Go OpenAPI tests.
- [ ] Decide separately whether to run external Schemathesis/Python runtime conformance from `spec-socialpredict-tasks-auto`; do not claim it is an in-repo requirement unless wired into this repo.

Exit criteria:

- [ ] Backend APIs are stable enough for frontend work to consume.
- [ ] OpenAPI and live server routes are aligned.
- [ ] Public failure reasons are documented and test-covered.

Validation:

- [ ] `cd backend && go test ./...`
- [ ] `cd backend && go test . -run 'TestOpenAPI|TestRouteFamily|TestReasonResponse|TestEmbedded|TestDocsPublishing'`
- [ ] `cd backend && go test ./server`
- [ ] Optional external runtime conformance if configured: Schemathesis/Python tooling from `spec-socialpredict-tasks-auto`.
- [ ] `git diff --check`

### 12. Moderator Frontend Proposal Tracking

Prerequisite: backend-first gate and backend API contract completion gate.

Checklist:

- [ ] Add frontend route or dashboard surface for moderator proposal tracking.
- [ ] Use backend lifecycle terms in frontend copy.
- [ ] Use the existing frontend API/auth adapter patterns.
- [ ] Avoid direct token/API coupling in new frontend code.
- [ ] Add frontend states for loading, empty, success, and failure.
- [ ] Add collapsed or summarized proposal history where available.
- [ ] Add focused frontend tests if the test baseline exists by this point.

Exit criteria:

- [ ] Moderators can see proposal status without admin dashboard access.
- [ ] Frontend does not invent a separate lifecycle model.
- [ ] Existing frontend CI/build remains green.

Validation:

- [ ] Frontend build workflow or `cd frontend && npm run build:report`.
- [ ] `git diff --check`

### 13. Admin Approval Dashboard Baseline

Prerequisite: backend approval/rejection APIs and OpenAPI completion.

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

- [ ] Frontend build workflow or `cd frontend && npm run build:report`.
- [ ] `git diff --check`

### 14. Admin Moderator Management Dashboard

Prerequisite: backend users-domain role/status APIs and OpenAPI completion.

Checklist:

- [x] Add users dashboard capabilities for promote, suspend, and unsuspend.
- [x] Add moderators dashboard or moderator-filtered admin surface.
- [x] Show moderator status.
- [x] Show suspension reason and timestamp.
- [ ] Show created-market counts by lifecycle state.
- [ ] Preserve audit history access.
- [x] Add frontend states for loading, success, and failure.
- [x] Keep backend authorization authoritative.

Exit criteria:

- [x] Admin can manage moderators from UI.
- [ ] Moderator status changes are visible and auditable.
- [x] Dashboard does not bypass backend authorization.

Validation:

- [x] `cd frontend && npm run build`
- [x] `cd backend && go test ./...`
- [ ] `git diff --check`

### 15. End-To-End Feature Verification

Prerequisite: backend and frontend baseline flows exist.

Checklist:

- [ ] Add end-to-end or integration coverage for admin promotes participant to moderator.
- [ ] Cover moderator proposes market.
- [ ] Cover admin approves proposed market.
- [ ] Cover published market becomes tradable.
- [ ] Cover moderator self-trade is rejected if the policy applies to that path.
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

- Backend domain and API work should precede frontend work.
- Role/status work should precede proposal creation enforcement.
- Proposal lifecycle should precede admin dashboard work.
- API routes and `backend/docs/openapi.yaml` updates should land with their backend behavior, not in a delayed frontend PR.
- Cancellation refund design should wait until lifecycle and ledger seams are explicit.
- Frontend dashboard work should wait for stable backend route/reason contracts.
- Contract amendments can be developed after basic approve/reject behavior, but before relying on admin yanks for amendment abuse cases.

## API Contract Testing Convention

The current in-repo API contract baseline is Go-based:

- `backend/docs/openapi.yaml` is the canonical OpenAPI file.
- `backend/docs/README.md` defines the source-of-truth order and update rules.
- `backend/openapi_test.go` uses `github.com/getkin/kin-openapi/openapi3` to validate the OpenAPI document, route/spec parity, embedded docs, public reason enums, and route-family migration matrix.
- `backend/server/server_contract_test.go` covers backend-served docs publishing such as `/openapi.yaml` and `/swagger/`.

Required command for API-affecting backend PRs:

```bash
cd backend && go test ./...
```

Focused API check when a narrower command is needed:

```bash
cd backend && go test . -run 'TestOpenAPI|TestRouteFamily|TestReasonResponse|TestEmbedded|TestDocsPublishing'
cd backend && go test ./server
```

External Python/Schemathesis runtime conformance exists in the task/agent repository, but it is not currently an in-repo `socialpredict` command. If a future PR wires that into this repo or CI, update this plan and the API docs to make it a required validation step.

## Migration Convention

Moderator mode changes persistent data, so implementation PRs that add or alter storage must use the existing backend migration system.

The repo convention is:

- create timestamped Go migration files under `backend/migration/migrations`
- name files with the readable timestamp pattern `YYYYMMDD_HHMMSS_description.go`
- register migrations with the compact timestamp ID used by `backend/migration`
- keep migrations additive and backward-compatible where possible
- include package-local migration tests where practical, following the current migration test style

Examples already in the repo:

- `backend/migration/migrations/20251013_080000_core_models.go`
- `backend/migration/migrations/20251020_140500_add_market_labels.go`

## Review Checklist

Before merging a moderator-mode implementation PR, check:

- [ ] Does open mode still behave as before?
- [ ] Is backend domain/API work complete before any dependent frontend change?
- [ ] Is users functionality owned by the users domain/service?
- [ ] Is market functionality owned by the markets domain/service?
- [ ] Are bets changes limited to necessary trade guards or accounting seams?
- [ ] Is public error/reason language stable and documented?
- [ ] Does the PR update `backend/docs/openapi.yaml` for any API behavior change?
- [ ] Do the Go OpenAPI tests pass for API-affecting changes?
- [ ] Does the PR avoid hard deletion for audit-relevant state?
- [ ] Are financial writes transaction-scoped where required?
- [ ] Are schema changes handled through additive timestamped Go migrations under `backend/migration/migrations`?
- [ ] Do migrations include package-local tests where practical?
- [ ] Does frontend terminology match backend/domain terminology?
- [ ] Does the PR avoid introducing generic RBAC/workflow/platform abstractions ahead of need?

## Deferred Until After Baseline

- AI moderation or scoring.
- General workflow engine.
- Multi-tenant policy configuration UI.
- Advanced analytics around moderator quality.
- Full audit-log search platform.
- Dedicated moderation microservice.
