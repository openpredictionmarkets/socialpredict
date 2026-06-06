---
title: Market Stewardship Design
document_type: design
domain: features
author: Patrick Delaney
updated_at: 2026-06-03T00:00:00Z
status: draft
---

# Market Stewardship Design

## Boundary Decisions

Creator and steward are intentionally separate:

- Creator is attribution and history.
- Steward is current operational authority.

The market domain owns the policy that steward authority controls market-governance actions. The account domain owns whether a moderator is active or suspended. Persistence owns additive storage and audit rows.

## Invariants

- A market can have only one current steward.
- Empty steward values fall back to creator for legacy compatibility.
- Steward reassignment must not mutate `creator_username`.
- Steward reassignment must write an audit fact.
- A suspended moderator cannot be assigned as steward.
- A suspended steward cannot resolve a market.

## Tradeoffs

This is more schema and API surface than mutating `creator_username`, but it preserves attribution and supports future governance workflows such as overdue resolution, voluntary handoff, or conflict-of-interest reassignment.
