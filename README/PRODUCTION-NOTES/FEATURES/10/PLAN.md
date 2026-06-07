---
title: Market Description Amendments Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-06T00:00:00Z
updated_at_display: "Saturday, June 6, 2026"
update_reason: "Track backend-first implementation slices for immutable titles, append-only description amendments, markdown-lite validation, admin review, and frontend amendment display."
status: draft
---

# Market Description Amendments Plan

## Purpose

This plan turns [10-market-description-amendments.md](./10-market-description-amendments.md) and [DESIGN.md](./DESIGN.md) into a backend-first implementation sequence.

Agents implementing this feature should mark checklist items as they complete them and leave deferred work unchecked.

## Planning Principles

- Titles are immutable.
- Original descriptions are not overwritten.
- Amendments are append-only and server-versioned.
- Published-market amendments require admin governance.
- Markdown-lite is constrained and backend-validated.
- Public users see only approved contract text.
- Admins and appropriate moderators can see pending/rejected governance history.
- Every PR should be independently reviewable.

## 01. Feature Artifact And Design Alignment

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/10/`.
- [x] Add feature overview.
- [x] Add design artifact.
- [x] Add implementation plan.
- [ ] Review final terminology with product/user-facing language.
- [ ] Confirm how this feature should interact with yanked/cancelled market rules once Feature 08 is implemented.

## 02. Persistence And Migration

Service ownership: repository and migration boundary.

Checklist:

- [x] Add timestamped migration for `market_description_amendments`.
- [x] Add unique `(market_id, version)` constraint.
- [x] Add admin queue indexes for status and created time.
- [x] Add market/version lookup indexes.
- [x] Add model struct for amendment records.
- [x] Add migration tests following existing migration conventions.
- [x] Verify migration does not mutate existing market descriptions.

## 03. Domain Model And Service Policy

Service ownership: prediction market context and moderator governance context.

Checklist:

- [x] Add domain model for description amendment.
- [x] Add domain read model for original description plus approved amendments.
- [x] Add service method for proposing an amendment.
- [x] Assign next version server-side in a transaction.
- [x] Enforce title immutability in amendment use cases.
- [x] Enforce append-only behavior.
- [x] Enforce current-steward authority for moderator proposals.
- [x] Block suspended moderators from proposing amendments.
- [x] Add service tests for version ordering and pending-review protection.

## 04. Markdown-Lite Validation And Sanitization

Service ownership: security boundary.

Checklist:

- [x] Define allowed markdown-lite syntax.
- [x] Add backend validation for amendment body length.
- [x] Reject or neutralize raw HTML.
- [x] Reject unsafe links and suspicious script patterns.
- [x] Preserve safe markdown source for rendering.
- [x] Add sanitizer tests for allowed markdown.
- [x] Add sanitizer tests for disallowed HTML/script/image/embed content.
- [ ] Decide whether original market descriptions remain plain text or are migrated to `plain_text` version 1 read model.

## 05. API And OpenAPI

Service ownership: API and auth boundary.

Checklist:

- [x] Add moderator/steward endpoint to propose amendment.
- [x] Add admin endpoint to list pending amendments.
- [x] Add admin endpoint to approve/reject amendment.
- [x] Include approved amendments in public market details or add a public amendment endpoint.
- [ ] Include pending/rejected amendments in admin/moderator governance payloads.
- [x] Return stable public error reasons for invalid state, unauthorized steward, suspended moderator, and unsafe markdown.
- [x] Update `backend/docs/openapi.yaml`.
- [ ] Run go-kin OpenAPI validation.
- [ ] Run Schemathesis validation for the new/changed endpoints.

## 06. Admin Review UX

Service ownership: frontend admin workflow.

Checklist:

- [x] Add amendment review tab/panel to Market Review or Market Stewardship area.
- [ ] Show market title, original description, existing approved amendments, and proposed amendment.
- [x] Show submitter, timestamp, and reason.
- [x] Add approve/reject actions with required decision reason.
- [x] Clearly identify pending, approved, and rejected states.
- [ ] Avoid placing amendment actions where admins might confuse them with tag/steward actions.

## 07. Moderator UX

Service ownership: frontend moderator workflow.

Checklist:

- [x] Add amendment proposal entry point for current steward.
- [x] Explain title immutability and append-only amendment behavior.
- [x] Add markdown-lite help text.
- [x] Add markdown-lite preview.
- [x] Submit amendment through authenticated API.
- [ ] Show pending/approved/rejected amendment status in moderator profile or market governance view.
- [x] Hide amendment controls for suspended moderators and non-stewards.

## 08. Public Market Detail UX

Service ownership: frontend market detail experience.

Checklist:

- [x] Render original description separately from approved amendments.
- [x] Render approved amendments in chronological version order.
- [x] Display version, author, and timestamp for each amendment.
- [x] Make amendment text visible enough that users understand it is part of the contract text.
- [x] Use safe markdown-lite rendering with raw HTML disabled.
- [x] Preserve readable collapsed/expanded behavior for long descriptions and amendments.

## 09. Tests And Regression Coverage

Checklist:

- [x] Test no route can mutate market title through amendment flow.
- [x] Test original description remains unchanged after amendments.
- [x] Test amendment versions increment correctly.
- [x] Test public detail returns only approved amendments.
- [ ] Test admin/moderator views can see pending/rejected records as intended.
- [x] Test unauthorized moderator cannot propose amendment.
- [x] Test suspended moderator cannot propose amendment.
- [x] Test admin approval/rejection records actor, time, and reason.
- [x] Test markdown-lite sanitization rejects unsafe content.
- [x] Run frontend build.
- [x] Run targeted backend tests.

## 10. Deferred Enhancements

Checklist:

- [ ] User notifications when published-market amendments are approved.
- [ ] Market page banner for recently amended markets.
- [ ] Diff-like display between original description and amendment text.
- [ ] Admin-direct emergency amendment path with stricter audit language.
- [ ] Amendment inclusion in social share/Open Graph copy.
- [ ] Amendment lockout after market close or resolution.
- [ ] User-visible policy explaining whether amendments apply prospectively or to the full market history.

## Exit Criteria

- Market titles remain immutable.
- Original descriptions are preserved.
- Description amendments are append-only and versioned.
- Public market detail shows original description plus approved amendments.
- Pending/rejected amendments are governed and auditable.
- Markdown-lite is safe and consistently rendered.
- Moderator authority respects stewardship and suspension rules.
- Admin review has clear approval/rejection controls.
- OpenAPI and tests cover the new behavior.
