---
title: Market Description Amendments Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-06T00:00:00Z
updated_at_display: "Saturday, June 6, 2026"
update_reason: "Define backend-first market description amendment design, including markdown-lite policy, append-only persistence, versioning, authorization, and user-facing traceability."
status: draft
---

# Market Description Amendments Design

## Design Position

Market descriptions are contract-adjacent domain text. The system should not treat them like ordinary editable CMS copy.

The design should preserve the original market description and model every later clarification as an append-only domain event/record. This keeps the market understandable, auditable, and safer for users who traded under earlier text.

## Design Inputs

Primary inputs:

- [10-market-description-amendments.md](./10-market-description-amendments.md)
- Canonical design plan: `/Users/patrick/Projects/spec-socialpredict-tasks/lib/design/design-plan.json`
- Designer-agent postures from `/Users/patrick/Projects/spec-socialpredict-tasks/.codex/agents/`
- Existing market create flow and sanitizer posture
- Existing moderator stewardship and admin review flows
- Existing market detail rendering behavior

## Boundary Alignment

| Boundary | Responsibility |
| --- | --- |
| Prediction Market Context | Owns original market contract text, amendment versioning, and market-state rules. |
| Moderator Governance Context | Owns steward authority, suspended moderator restrictions, and amendment proposal permissions. |
| Admin Review Context | Owns approval/rejection workflow for amendments on published markets. |
| API And Auth Boundary | Owns amendment endpoints, OpenAPI schemas, auth checks, and public/private visibility. |
| Frontend Experience Context | Owns markdown-lite authoring, preview, public amendment display, and admin review UI. |
| Repository And Migration Boundary | Owns additive timestamped migration and append-only persistence model. |
| Security Boundary | Owns markdown-lite validation/sanitization and raw HTML rejection/neutralization. |

## Current Code Posture

Observed current behavior:

- `backend/models/market.go` has `QuestionTitle string` and `Description string`.
- `backend/handlers/markets/dto/requests.go` accepts `description` as a max 2000-character create field.
- `backend/security/sanitizer.go` sanitizes descriptions with a basic HTML policy.
- `frontend/src/components/marketDetails/MarketDetailsLayout.jsx` renders description as escaped text with `whitespace-pre-wrap`.
- No general market title/description update endpoint appears to exist.
- Homepage CMS has markdown/html rendering and sanitization, but that is a separate CMS path and should not be reused casually as market-contract editing.

Design implication: this is a new market-governance feature, not a small form edit.

## Domain Model

### Market Description Amendment

Candidate fields:

- `id`
- `market_id`
- `version`
- `body`
- `body_format`: `markdown_lite`
- `status`: `pending`, `approved`, `rejected`
- `created_by`
- `created_at`
- `approved_by`
- `approved_at`
- `rejected_by`
- `rejected_at`
- `rejection_reason`
- `review_reason` or `submit_reason`, optional

Rules:

- `version` is assigned server-side.
- Version 1 is the original market description, even if stored on `markets.description`.
- Amendment records start at version 2.
- Version assignment must be transaction-safe per market.
- Amendments are append-only records.
- Rejected amendments are not public contract text.
- Pending amendments are visible to admins and appropriate moderators, not public users.

### Description Contract Read Model

Public market detail should expose a clear contract read model.

Candidate shape:

```json
{
  "originalDescription": {
    "version": 1,
    "body": "Original market description",
    "bodyFormat": "plain_text",
    "createdAt": "2026-06-06T00:00:00Z",
    "createdBy": "moderator"
  },
  "approvedAmendments": [
    {
      "version": 2,
      "body": "Clarification text",
      "bodyFormat": "markdown_lite",
      "createdAt": "2026-06-06T01:00:00Z",
      "createdBy": "moderator",
      "approvedAt": "2026-06-06T02:00:00Z",
      "approvedBy": "admin"
    }
  ]
}
```

The existing `market.description` response can remain during migration, but the frontend should move toward this richer read model.

## Persistence Design

Add a timestamped Go migration under `backend/migration/migrations` following existing conventions.

Candidate table:

```sql
market_description_amendments
```

Candidate indexes:

- unique `(market_id, version)`
- index `(market_id, status, version)`
- index `(created_by, created_at)`
- index `(status, created_at)` for admin queues

Important persistence rules:

- Do not mutate `markets.description` when adding amendments.
- Use a transaction to compute and insert the next version.
- Prefer database uniqueness as a guard against concurrent version collisions.
- Keep pending/rejected records for audit rather than deleting them.

## Markdown-Lite Design

Markdown-lite should be a constrained source format.

Recommended format value:

```text
markdown_lite
```

Allowed syntax:

- paragraphs
- line breaks
- bold
- italic
- bullet lists
- numbered lists
- blockquotes
- inline code
- safe links

Disallowed syntax:

- raw HTML
- script/style tags
- iframes
- images
- embedded media
- tables initially
- custom classes/styles

Implementation posture:

- Backend validates length and suspicious content before persistence.
- Backend rejects or neutralizes raw HTML.
- Frontend renders markdown-lite using a parser configured with raw HTML disabled.
- Sanitization remains backend-owned even if frontend renders the final markdown.
- Tests should include script injection, HTML tags, unsafe links, and allowed markdown examples.

## Authorization Design

### Proposed Markets

Allowed actors:

- creator/steward if they are an active approved moderator
- admin

Recommended behavior:

- amendments can be submitted while proposal is pending
- admin market review sees original description plus proposed amendments
- approving the market can publish approved proposal amendments as contract text

### Published Markets

Allowed actors:

- current steward can propose an amendment
- admin can approve/reject
- admin can create an auto-approved amendment if necessary, but must provide a reason

Suspended moderators:

- cannot propose amendments
- cannot approve/reject amendments
- prior amendments remain historically visible based on status

Rejected/resolved/yanked markets:

- should not accept moderator amendments by default
- admin-only correction/annotation may be a separate future path

## API Design Candidate

Moderator/steward proposal:

```text
POST /v0/markets/{id}/description-amendments
```

Admin review queue:

```text
GET /v0/admin/market-description-amendments?status=pending
```

Admin decision:

```text
PATCH /v0/admin/market-description-amendments/{id}
```

Public market detail:

```text
GET /v0/markets/{id}
```

Should include approved amendment read model or a linked amendment endpoint.

Request candidate:

```json
{
  "body": "Clarification text in markdown-lite",
  "bodyFormat": "markdown_lite",
  "reason": "Clarifies settlement source"
}
```

Decision candidate:

```json
{
  "status": "approved",
  "reason": "Clarifies criteria without changing outcome definition"
}
```

## Frontend Design

### Market Detail

Public display should show:

- immutable title
- original description
- approved amendments in version order
- amendment timestamps and authors
- compact explanation that amendments are appended clarifications

### Moderator Form

Moderators should see:

- title is immutable
- original description is preserved
- amendment will be appended and versioned
- markdown-lite help text
- preview before submission
- pending status after submission when approval is required

### Admin Review

Admin should see:

- market title
- original description
- existing approved amendments
- proposed amendment body
- submitted reason
- actor and timestamp
- approve/reject controls with reason

## Audit And Traceability

Traceability requirements:

- every amendment records actor and timestamp
- every admin decision records actor, timestamp, and reason
- market detail clearly separates original text from amendments
- admin review can reconstruct pending, approved, and rejected amendment history
- version order is deterministic

## Design-Agent Cross-Reference

Evans/domain posture:

- Treat market description changes as domain events/records, not as generic text edits.
- Use ubiquitous language around contract text, amendment, steward, and version.

Fowler/evolutionary posture:

- Start with append-only amendment records and simple approval flow before richer notification systems.
- Keep read models explicit so frontend behavior can evolve without rewriting persistence.

Martin/clean-boundary posture:

- Keep markdown parsing/sanitization outside React-only concerns.
- Keep amendment authorization in service/domain policy, not dashboard conditionals.
- Avoid mutating original market rows for behavior that is actually historical traceability.
