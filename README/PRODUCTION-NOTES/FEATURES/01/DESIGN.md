---
title: Moderator Mode Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Add the domain and architecture design artifact for moderator mode before implementation planning."
status: draft
---

# Moderator Mode Design

## Purpose

This document translates [01-moderators.md](./01-moderators.md) into a domain and architecture design artifact before implementation begins.

It is not the implementation plan. The implementation sequence lives in [PLAN.md](./PLAN.md).

## Design Inputs

Primary inputs:

- [01-moderators.md](./01-moderators.md)
- Canonical design plan: `spec-socialpredict-tasks-auto/lib/design/design-plan.json`
- Designer-agent postures from `spec-socialpredict-tasks-auto/.codex/agents/`

The canonical design plan remains the repository-level design source of truth. This feature design must conform to it rather than create a competing architecture.

## Designer Lens Review

Evans/domain lens:

- Moderator mode is a game-mode and governance model, not just an admin screen.
- The feature needs stable language for `open mode`, `moderator mode`, `moderator`, `admin`, `proposed market`, `published market`, `cancelled market`, `contract amendment`, and `contract version reference`.
- Market contract language must not drift between backend, frontend, admin UI, and API responses.

Fowler/evolutionary lens:

- Keep explicit `open` behavior backward compatible while making `moderator` the default project mode.
- Add seams in small increments: config policy, role/status policy, market lifecycle, approval use cases, then cancellation/refund behavior.
- Complete backend domain/API contract work before frontend work consumes the feature.
- Avoid a workflow-engine or RBAC-platform rewrite until the feature proves it needs those abstractions.

Martin/clean-architecture lens:

- Domain rules belong in domain/application services, not in frontend routing or handler-only checks.
- HTTP handlers and repositories are adapters around use cases; they should not own moderator policy.
- Accounting-sensitive cancellation/refund behavior needs an explicit unit-of-work boundary and Postgres-backed verification.

## Problem Framing

SocialPredict currently has one game mode: authenticated participants can create markets, trade, and resolve markets they created.

Moderator mode introduces governance before market publication. It allows selected moderators to propose markets while admins retain approval, suspension, and cancellation authority. The feature must protect participant funds, preserve market-contract audit history, and keep open-mode behavior unchanged unless moderator mode is enabled.

## Business Outcomes

Moderator mode should enable:

- Controlled market supply when open market creation is too risky.
- Admin review before moderator-created markets become tradable.
- Auditable governance around moderator promotion, suspension, approval, rejection, amendment, and cancellation.
- Participant confidence that published market contracts cannot be silently rewritten.
- Economically correct refunds when an admin yanks a bad market.

## Out Of Scope

This design does not introduce:

- A generic workflow engine.
- AI moderation.
- A separate authorization platform.
- A microservice split.
- Frontend-only authorization.
- Hard deletion for yanked markets.

## Ubiquitous Language

| Term | Meaning | Avoid confusing with |
| --- | --- | --- |
| Open mode | Existing behavior where regular authenticated users may create markets. | Moderator mode. |
| Moderator mode | Game mode where only active moderators may propose markets and admin approval is required before trading. | Admin mode or a general RBAC platform. |
| Regular participant | Authenticated non-admin, non-moderator user. | Anonymous user. |
| Moderator | Elevated participant allowed to propose markets and resolve their own published markets. | Admin. |
| Suspended moderator | Moderator blocked from moderator-only actions. | Deleted user or demoted user. |
| Admin | Platform governance actor who can promote, suspend, approve, reject, and cancel. | Moderator. |
| Proposed market | Moderator-created market awaiting admin decision and not tradable. | Active/published market. |
| Published market | Admin-approved market visible and tradable. | Proposed market. |
| Rejected market | Proposal rejected by admin and retained for audit/history. | Deleted market. |
| Cancelled market | Published market yanked by admin with cancellation audit and refund handling. | Creator resolution to `N/A`. |
| Market contract | Original title, original description, approved additive clarifications, and ordered change record. | Display copy only. |
| Contract amendment | Additive title or description clarification that produces a change record. | In-place update. |
| Contract version reference | Stable hash/reference for the effective contract version. | Database row ID alone. |
| Net unrecovered exposure | Participant money still at risk after accounting for buys and sell recoveries. | Gross buys or current shares only. |

## Bounded Context Alignment

| Design-plan boundary | Moderator-mode ownership |
| --- | --- |
| Configuration Service Slice | Owns typed game-mode policy loaded from setup assets, with default `moderator` behavior. |
| Participant Account Context | Owns user role constants, moderator status, suspension state, and participant-facing account policy. User functionality belongs in the users domain/service. |
| Prediction Market Context | Owns market lifecycle, approval state, publication eligibility, contract immutability, and resolution eligibility. Market functionality belongs in the markets domain/service. |
| Betting and Position Ledger Context | Owns only the buy/sell guard integration and accounting behavior that must occur on trade paths. Moderator mode should not cause a broad bets redesign unless cancellation/refund accounting proves it is needed. |
| API and Auth Contract Boundary | Owns route-visible authorization failures, reason vocabulary, and OpenAPI alignment. |
| Repository and Legacy Model Adapter Boundary | Owns persistence translation, migrations, and transaction-scoped data access. |
| Frontend Experience Context | Owns participant/admin/moderator presentation while conforming to backend-owned domain language and policy. |
| Failure Translation and Recovery Boundary | Owns safe public errors for unauthorized, invalid-state, and cancellation-related failures. |

## Context Map

Moderator mode cuts across existing contexts but should not collapse them into one package.

- Configuration Service Slice supplies immutable game-mode policy to business services.
- Participant Account Context supplies actor role and moderator status facts through the users domain/service.
- Prediction Market Context decides whether a market can be proposed, published, amended, resolved, or cancelled through the markets domain/service.
- Betting and Position Ledger Context checks market/user eligibility at buy/sell boundaries and owns only the accounting portions that genuinely require ledger data.
- API handlers adapt HTTP requests to use cases and translate use-case outcomes into public responses.
- Frontend screens present allowed actions and status, but backend policy remains authoritative.

## Backend-First Design Rule

Moderator mode must be designed and implemented backend-first.

Backend-first means:

- user role, moderator status, and suspension behavior are owned by users domain/service code before admin or moderator frontend screens are added
- market proposal, approval, rejection, publication, amendment, resolution, and cancellation behavior are owned by markets domain/service code before frontend status displays are added
- buy/sell restrictions are enforced through backend buy/sell paths when policy requires them
- API routes, response schemas, and public failure reasons are documented in `backend/docs/openapi.yaml` before frontend code depends on them
- frontend work consumes the backend contract instead of inventing its own feature state machine

## Core Domain Rules

Game-mode rules:

- Default game mode is `moderator`.
- Moderator mode is enabled through setup/application policy, not frontend flags.
- Explicit open mode preserves existing market creation behavior unless changed by later design.

Role and suspension rules:

- `ADMIN`, `REGULAR`, and `MODERATOR` must be stable role constants or equivalent typed values.
- A moderator can be `active` or `suspended`.
- Suspended moderators cannot create, amend, or resolve moderator-owned markets.
- Admin actions that change role or suspension state must be auditable.

Market lifecycle rules:

- Moderator-created markets in moderator mode start as `proposed`.
- Proposed and rejected markets are not tradable.
- Published markets are tradable unless closed, resolved, or cancelled.
- Cancelled markets are terminal and cannot be traded or resolved.
- Hard delete is not valid for admin yanks.

Contract rules:

- Original market title is immutable.
- Original market description is not overwritten in place.
- Clarifications are append-only amendments.
- Backend creates change records; clients do not supply audit history.
- Published moderator-market amendments require admin approval unless performed by an admin.
- Approved amendments produce a new contract version reference.

Accounting rules:

- Moderator self-trade restrictions must be enforced on buy and sell paths, not only UI controls.
- Rejected moderator proposals refund the market proposal creation cost to the creator.
- Admin cancellation refunds net unrecovered exposure, not simply gross buys.
- Cancellation state update and refund ledger entries must commit atomically.
- Cancellation math that depends on buy/sell history requires Postgres-backed tests.

## Use Cases

Admin use cases:

- Promote participant to moderator.
- Suspend or unsuspend moderator.
- Review proposed market.
- Approve proposed market with confirmation.
- Reject proposed market with reason.
- Review contract amendment where required.
- Cancel/yank published market with reason and refund handling.

Moderator use cases:

- Create proposed market.
- View proposal status.
- View rejected, published, closed, resolved, and cancelled markets they created.
- Land on their private Proposed Markets tab after creating a proposal, rather than manually handing market IDs to admins.
- Resolve their own published market when eligible.
- Propose additive contract amendment where allowed.

Participant use cases:

- Browse published markets.
- Trade on published markets they are allowed to trade on.
- See cancellation/refund outcomes when a market is yanked.
- View relevant contract history without being overwhelmed by audit detail.

## Aggregate And Entity Sketch

This is a design sketch, not a required package layout.

Market aggregate:

- Original immutable title and description.
- Lifecycle state.
- Creator/proposer identity.
- Current contract version reference.
- Approval/rejection/cancellation metadata.
- Domain methods for propose, approve, reject, publish, close, resolve, cancel, and amend.

Market contract amendment entity:

- Market ID.
- Amendment sequence.
- Changed field.
- Previous and next contract version references.
- Actor, reason, timestamp.
- Approval state and approver when required.

Moderator status entity or participant-role extension:

- Username/user ID.
- Role.
- Moderator status.
- Suspension reason, actor, and timestamp.

Cancellation/refund event:

- Market ID.
- Cancelling admin.
- Reason.
- Contract version reference at cancellation.
- Generated refund transaction references.

## API Contract Posture

Route shape can follow existing conventions, but public contract behavior needs explicit reason vocabulary.

Examples of reason categories to define before implementation:

- `MODERATOR_MODE_REQUIRED`
- `MODERATOR_ROLE_REQUIRED`
- `MODERATOR_SUSPENDED`
- `MARKET_NOT_PROPOSED`
- `MARKET_NOT_PUBLISHED`
- `MARKET_CANCELLED`
- `MARKET_CONTRACT_IMMUTABLE`
- `SELF_TRADE_FORBIDDEN`
- `APPROVAL_REQUIRED`
- `CANCELLATION_FORBIDDEN`

These names are placeholders until aligned with the existing OpenAPI reason vocabulary. The durable requirement is that UI and API share stable public reasons rather than parsing raw error text.

API-affecting moderator-mode PRs must update `backend/docs/openapi.yaml` in the same change as route or response behavior. The backend API notes define the source-of-truth order as `backend/server/server.go`, touched handlers/DTOs, `backend/docs/openapi.yaml`, and `backend/docs/README.md`.

The current in-repo API contract tests are Go-based:

- `backend/openapi_test.go` validates the OpenAPI document with `github.com/getkin/kin-openapi/openapi3`, route/spec parity, public reason enum alignment, embedded docs, and route-family migration metadata.
- `backend/server/server_contract_test.go` covers backend-served docs publishing for `/openapi.yaml` and Swagger surfaces.

Required validation for API-affecting moderator-mode PRs:

```bash
cd backend && go test ./...
```

Python/Schemathesis runtime conformance tooling exists in `spec-socialpredict-tasks-auto`, but it is not currently wired as an in-repo `socialpredict` validation command. Treat it as optional external conformance tooling until a PR adds it to this repo or CI.

## Frontend Design Posture

Frontend work should follow the existing frontend design plan decisions:

- Use backend-owned status/reason language for market lifecycle and authorization failures.
- Keep admin/moderator UI behind backend-enforced permissions.
- Use Tailwind and shared style surfaces for new UI, especially dashboards and approval queues.
- Do not create a separate frontend-only moderator state model.
- Do not add analytics, PWA, or browser platform features as part of the moderator baseline.

## Data And Persistence Posture

Persistence should support audit and rollback-free history:

- Additive migrations, not destructive rewrites.
- Database schema changes must use the repo's timestamped Go migration convention under `backend/migration/migrations`, such as `YYYYMMDD_HHMMSS_description.go`.
- Each migration should register with the compact timestamp ID used by the migration package and include package-local migration tests where practical.
- Explicit role/status fields or tables.
- Market lifecycle fields that preserve compatibility where possible.
- Separate change-record storage rather than overwriting market title/description history.
- Cancellation/refund event linkage to ledger entries.
- Repository methods that express use-case needs without exposing GORM models as domain policy.

## Verification Strategy

Fast tests:

- Domain tests for role, suspension, lifecycle, amendment, and self-trade rules.
- Handler/API tests for public reason translation and authorization boundaries.
- Frontend build and targeted UI tests after frontend test tooling is available.

Postgres-backed tests:

- Cancellation refund math with no sells.
- Cancellation refund math after partial sell-out.
- Cancellation refund math after full sell-out.
- Multi-user, both-outcome refund scenarios.
- Atomic cancellation state plus ledger writes.

Contract tests:

- OpenAPI route/reason alignment for new admin and moderator endpoints.
- Backward-compatible open-mode creation behavior.

## Risks

- Treating moderator mode as only UI gating would create authorization bypass risk.
- Treating cancellation as ordinary `N/A` resolution would hide governance/audit semantics.
- Refunding gross buys instead of net unrecovered exposure would overpay users who sold out.
- In-place title or description edits would undermine market-contract integrity.
- Adding generic RBAC/workflow infrastructure too early would slow the first safe version.
- Combining approval, amendments, cancellation, dashboards, and refund accounting in one PR would make review too risky.

## Open Questions

- Should admins be allowed to trade in moderator mode?
- Should admin-created markets bypass proposal review?
- Should rejected proposals be editable/resubmittable or replaced by new proposals?
- Should moderator resolution require admin review for some market categories?
- Should demoted moderators retain access to historical moderator dashboards?
- Which cancellation reason fields are public, participant-visible, moderator-visible, or admin-only?
- Is `published` a new stored status, an approval state layered over existing `active`, or a compatibility projection?
- What exact algorithm should define contract version reference SHA inputs?
