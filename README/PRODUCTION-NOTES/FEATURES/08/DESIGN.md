---
title: Market Yank And Cancellation Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-04T00:00:00Z
updated_at_display: "Thursday, June 4, 2026"
update_reason: "Define the design boundary for admin market yanking before implementation."
status: draft
---

# Market Yank And Cancellation Design

## Design Position

Market yanking is a market-governance use case with accounting side effects. It belongs primarily in the prediction market domain, with transaction-scoped collaboration from the betting/position ledger and participant account contexts.

It must not be implemented as:

- frontend-only hiding
- hard deletion
- direct database edits
- ordinary `resolved = N/A` without explicit cancellation state
- handler-owned refund math

## Design Inputs

Primary inputs:

- [08-market-yank-cancellation.md](./08-market-yank-cancellation.md)
- [../01/01-moderators.md](../01/01-moderators.md)
- [../07/07-market-stewardship.md](../07/07-market-stewardship.md)
- Canonical design plan: `/Users/patrick/Projects/spec-socialpredict-tasks/lib/design/design-plan.json`

## Boundary Alignment

| Boundary | Responsibility |
| --- | --- |
| Prediction Market Context | Owns lifecycle transition to `cancelled`, eligibility rules, visibility mode, public cancellation state, and admin yank use case. |
| Betting And Position Ledger Context | Owns market-specific exposure calculation and compensating cancellation accounting. |
| Participant Account Context | Owns user balance updates through existing transaction semantics, not ad hoc market code. |
| API And Auth Boundary | Owns admin route, failure reasons, OpenAPI contract, and public response shape. |
| Frontend Experience Context | Owns admin confirmation UI and cancelled market presentation, consuming backend state. |
| Repository And Migration Boundary | Owns additive fields/tables and transaction-scoped persistence. |
| Auditability Boundary | Owns reconstructable cancellation history, actor, reason, visibility, and accounting summary. |

## State Model

Cancelled should be represented as a first-class lifecycle state.

Candidate model fields:

- `lifecycle_status = cancelled`
- `status = cancelled` for direct lifecycle surfaces
- `resolution_result = N/A` only as compatibility/display support
- `cancelled_by`
- `cancelled_at`
- `cancellation_reason_private`
- `cancellation_reason_public`
- `cancellation_obfuscates_public_content`
- `cancellation_accounting_status`

The migration should be additive and timestamped. Existing lifecycle constants already include `cancelled`, but behavior should be audited before assuming full support.

## Visibility Model

The market domain should expose a public-safe projection for cancelled markets.

Public projection rules:

- If not cancelled, preserve existing behavior.
- If cancelled and not obfuscated, show original title/description/labels plus cancellation banner.
- If cancelled and obfuscated, replace title/description/labels with neutral cancellation copy.
- Always hide trade/sell/resolve controls for cancelled markets.
- Preserve creator/steward attribution unless there is a separate safety reason to suppress user identity.

Admin projection rules:

- Admin governance pages can see original content, visibility mode, private reason, public reason, actor, timestamp, and accounting status.

## Accounting Design Direction

Cancellation accounting should use compensating transactions, not historical mutation.

Design target:

1. Load the market, participants, bets, sales, fees, subsidies, and current market-specific positions inside a transaction.
2. Compute each participant's net market effect under a named cancellation policy.
3. Write explicit cancellation compensation transactions to neutralize market-specific net effect.
4. Update market lifecycle and cancellation audit in the same transaction.
5. Return an accounting summary to the admin caller.

Do not zero out old bet rows as the primary mechanism. Old rows are historical facts needed for audit and debugging. If UI should display zero active exposure, derive that from cancelled lifecycle plus compensation state.

## Accounting Open Questions

Before implementation, resolve these questions with tests:

- Does refund include fees?
- Does refund include proposal cost?
- Are subsidies reversed or treated as platform cost?
- What is the exact formula for users who bought and later sold some shares?
- What happens if clawback would push a user below maximum debt?
- Is the system allowed to create negative compensation transactions?
- How should market-level aggregate metrics display cancelled markets?
- Does cancellation affect leaderboards and financial statements immediately or through derived accounting views?

## Audit Model

A cancellation audit row should record:

- market ID
- actor username
- previous status/lifecycle
- new status/lifecycle
- private reason
- public reason
- obfuscation flag
- accounting policy version
- users affected
- total refunded
- total clawed back
- created timestamp

The audit row should be immutable after creation. Later corrections should create a new audit/correction fact rather than editing the original.

## API Design

Candidate route:

```http
PATCH /v0/admin/markets/{id}/yank
```

Design constraints:

- admin-only
- requires `confirm=true`
- requires non-empty private reason
- validates public reason length and unsafe content
- returns failure reason for invalid state
- returns cancellation/accounting summary on success
- appears in OpenAPI before frontend consumes it

## Frontend Design

Admin UI should place yank controls in market governance, near stewardship, not ordinary public market tabs.

The UI must make risk explicit:

- irreversible action warning
- cancellation status preview
- obfuscation choice
- private reason field
- public notice field
- accounting preview/summary
- confirmation checkbox or typed confirmation

Cancelled public market pages should be calm and explicit: cancelled, not tradable, no ordinary resolution, with public reason if present.

## Risk Register

| Risk | Mitigation |
| --- | --- |
| Incorrect refund math | Backend-first accounting design, unit tests, transaction tests, no frontend implementation until API is stable. |
| Audit loss | No hard delete; append-only cancellation audit. |
| Unsafe content remains visible | Obfuscation flag and public-safe projection. |
| Confusing cancelled with resolved N/A | First-class `cancelled` lifecycle and explicit cancelled UI. |
| Balance mutation outside ledger semantics | Use participant account transaction APIs or a ledger-owned unit of work. |
| Metrics drift | Add analytics rules for cancelled markets before relying on production dashboards. |

## Exit Criteria For Design

Implementation should not start until:

- cancellation lifecycle semantics are agreed
- accounting formula is documented
- migration fields/tables are listed
- API request/response and failure reasons are drafted
- OpenAPI update path is known
- backend tests are planned before UI work
