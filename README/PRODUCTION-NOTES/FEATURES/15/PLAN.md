---
title: Grouped Discovery Read Model Pagination Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Plan backend-grouped discovery pagination."
status: draft
---

# Grouped Discovery Read Model Pagination Plan

## Checklist

- [x] Inventory discovery/search/topic/profile endpoints that return market rows.
- [x] Add backend DTO for grouped discovery row.
- [x] Add repository/read-model query that groups before limit/offset.
- [x] Include parent title and child answer labels in search matching.
- [x] Return grouped total counts.
- [x] Update `/markets` and topic frontend to consume grouped rows.
- [x] Update profile owned/lifecycle tabs to consume grouped rows instead of all child rows.
- [x] Keep frontend `groupMarketRows` as fallback only.
- [x] Add tests for child rows straddling page boundaries.
- [x] Add tests for search matching one child answer but returning one parent group row.

## Implementation Notes

- Added `MarketDiscoveryRow` and `MarketDiscoveryPage` as display/read-model domain shapes.
- Added GORM discovery queries that filter matching child markets, hydrate parent groups and all child answers, collapse to one grouped row, then apply `limit` and `offset`.
- Added grouped lifecycle discovery for profile owned/lifecycle tabs so the browser no longer has to fetch every child market page before showing 20 grouped rows.
- Preserved frontend grouping helpers as compatibility fallbacks, including a pass-through for backend aggregate rows.
