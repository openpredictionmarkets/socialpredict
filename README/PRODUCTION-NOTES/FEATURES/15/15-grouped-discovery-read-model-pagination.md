---
title: Grouped Discovery Read Model Pagination
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Capture design-agent finding that grouped discovery must paginate after backend grouping."
status: draft
---

# Grouped Discovery Read Model Pagination

## Purpose

Grouped markets should appear as one discovery row/card across `/markets`, topic pages, search, and profile market lists. Some paths still fetch child markets from the backend, paginate them, and then group them in the frontend. That can split children across pages and produce incomplete group rows.

This feature moves grouped discovery rows into the backend/read-model boundary before pagination.

## Source Finding

Design review finding: P1 grouped discovery pagination is conceptually wrong because grouping happens after backend pagination.

Relevant refs:

- `backend/handlers/markets/discovery_read_model.go:192`
- `backend/handlers/markets/discovery_read_model.go:202`
- `frontend/src/components/tables/MarketsByStatusTable.jsx:222`
- `frontend/src/helpers/marketGroups.js:85`

## Desired Outcome

The backend returns display-ready discovery rows where one market group is one row before limit/offset are applied.

## Acceptance Criteria

- `/markets` active/closed/resolved/all discovery uses backend-grouped rows.
- Topic pages use backend-grouped rows.
- Search returns parent group rows when child answers match.
- Group row total/count reflects grouped rows, not raw child rows.
- Frontend grouping helper remains only as compatibility/fallback.
- Profile owned/lifecycle market lists no longer need to fetch every child page to display 20 grouped rows.
