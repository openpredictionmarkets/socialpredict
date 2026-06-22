---
title: Grouped Discovery Cache Invalidation Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Plan grouped discovery structural invalidation fixes."
status: draft
---

# Grouped Discovery Cache Invalidation Plan

## Checklist

- [x] Inventory all read-model invalidation reasons emitted by grouped-market routes.
- [x] Mark answer-added and group-resolved reasons as structural.
- [x] Add approval/rejection invalidation where missing.
- [x] Add tests for structural stale snapshots.
- [x] Add tests that soft-stale non-structural snapshots still follow existing behavior.
- [x] Document the event vocabulary in the read-model feature notes.

## Implementation Notes

- Group approval and rejection now emit explicit `market_group_approved` and `market_group_rejected` invalidation reasons instead of relying on a generic market-status reason.
- Group answer addition, answer-review approval, group resolution, and steward reassignment are treated as structural discovery changes.
- Transaction-like events such as `bet_accepted` and option-policy toggles remain soft-stale events so they do not force immediate discovery snapshot expiry.
- Repository tests assert that structural events age discovery snapshots out of the freshness window while soft events preserve the existing generated timestamp behavior.
