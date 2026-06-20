---
title: Grouped Discovery Cache Invalidation
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Capture design-agent finding that grouped market structural events must invalidate discovery snapshots."
status: draft
---

# Grouped Discovery Cache Invalidation

## Purpose

Grouped market changes can alter public discovery structure: added answers, group resolution, group approval, group rejection, and group stewardship/tag changes. If those events do not structurally invalidate cached discovery snapshots, users can see stale grouped market rows until the normal freshness window expires.

## Source Finding

Design review finding: P2 group discovery cache invalidation misses new structural group events.

Relevant refs:

- `backend/internal/repository/readmodels/snapshots.go:98`
- `backend/server/server.go:450`
- `backend/handlers/markets/market_groups.go:176`

## Desired Outcome

Discovery snapshots become stale immediately when grouped-market structure changes.

## Acceptance Criteria

- `market_group_answer_added` is structural.
- `market_group_resolved` is structural.
- group approval/rejection invalidates discovery.
- group answer auto-approval that creates a child market invalidates discovery.
- tests prove structural group events are not served as fresh snapshots.
