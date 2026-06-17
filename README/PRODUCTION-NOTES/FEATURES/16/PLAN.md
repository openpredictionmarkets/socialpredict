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

- [ ] Inventory all read-model invalidation reasons emitted by grouped-market routes.
- [ ] Mark answer-added and group-resolved reasons as structural.
- [ ] Add approval/rejection invalidation where missing.
- [ ] Add tests for structural stale snapshots.
- [ ] Add tests that soft-stale non-structural snapshots still follow existing behavior.
- [ ] Document the event vocabulary in the read-model feature notes.
