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

- [ ] Inventory discovery/search/topic/profile endpoints that return market rows.
- [ ] Add backend DTO for grouped discovery row.
- [ ] Add repository/read-model query that groups before limit/offset.
- [ ] Include parent title and child answer labels in search matching.
- [ ] Return grouped total counts.
- [ ] Update `/markets` and topic frontend to consume grouped rows.
- [ ] Update profile owned/lifecycle tabs to consume grouped rows instead of all child rows.
- [ ] Keep frontend `groupMarketRows` as fallback only.
- [ ] Add tests for child rows straddling page boundaries.
- [ ] Add tests for search matching one child answer but returning one parent group row.
