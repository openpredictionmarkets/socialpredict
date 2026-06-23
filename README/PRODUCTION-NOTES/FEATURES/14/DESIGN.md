---
title: Grouped Market Transaction Boundaries Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Define transaction boundary requirements for grouped-market write paths."
status: draft
---

# Grouped Market Transaction Boundaries Design

## Design Posture

This feature follows the canonical design plan rule that transaction-time decisions and balance mutations must use backend-owned canonical state and must not be split across unverifiable partial writes.

Grouped markets are display/governance aggregates, but grouped creation, answer addition approval, and grouped resolution are economic/audit write paths. They need a single use-case transaction boundary or an explicit idempotent compensation strategy.

## Transaction Boundary Candidates

| Use case | Current risk | Required boundary |
| --- | --- | --- |
| Create grouped market | Proposal fee and child markets can be written before parent/member rows. | Fee, child markets, tags, parent group, and member links commit together. |
| Approve answer addition | Fee, child market, member row, review status, and amendments can diverge. | Approval command commits child, member, review, fee, tags, and amendments together. |
| Resolve grouped market | Child resolutions, parent resolved state, and work income can diverge. | Resolution command is atomic or idempotent and retry-safe. |
| Pay grouped work income | Retry can overpay if parent resolved/work income state is not guarded. | Payment is tied to a durable once-only resolution marker or ledger guard. |

## Design Direction

Preferred direction: add a grouped-market unit-of-work port owned by the market domain service and implemented by the GORM repository adapter.

The port should not move business rules into GORM. The domain service should still decide:

- who can create, approve, add, or resolve
- which child outcomes are valid
- when fees apply
- which generated amendments must exist
- when work income is due

The adapter should own only the database transaction mechanics.

## Testing Strategy

Add failure-injection tests around each boundary:

- fail after fee but before child creation
- fail after child creation but before group/member rows
- fail after member row but before answer-addition approval
- fail after answer-addition approval but before amendments
- fail after one child resolution but before remaining children
- retry grouped resolution after partial already-resolved matching children
- retry grouped resolution cannot duplicate work income

## Out Of Scope

- Replacing binary market transaction math.
- Parent-level pooled liquidity.
- Moving child buy/sell execution into parent group state.
