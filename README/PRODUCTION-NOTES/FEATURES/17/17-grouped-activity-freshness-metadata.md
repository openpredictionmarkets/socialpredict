---
title: Grouped Activity Freshness Metadata
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record grouped activity freshness metadata implementation."
status: implemented
---

# Grouped Activity Freshness Metadata

## Purpose

Grouped positions and grouped leaderboard responses expose freshness fields, and the frontend expects freshness metadata. Handlers currently do not populate that field consistently.

This feature makes grouped activity freshness explicit: either populate it correctly or remove it until a grouped activity read model exists.

## Source Finding

Design review finding: P2 group activity freshness is exposed in DTO/frontend expectations but not populated by handlers.

Relevant refs:

- `backend/handlers/markets/dto/responses.go:157`
- `backend/handlers/markets/market_groups.go:668`
- `backend/handlers/markets/market_groups.go:713`

## Desired Outcome

The API contract and UI agree about freshness. Users should not see stale-data messaging based on absent or misleading metadata.

## Acceptance Criteria

- [x] Grouped bets/positions/leaderboard responses either include real freshness or do not advertise freshness.
- [x] Frontend only renders freshness copy when metadata is meaningful.
- [x] Handler tests cover freshness response behavior.
- [x] OpenAPI matches implementation.

## Implementation Notes

Grouped activity endpoints are still live-computed. They now return `source = live` freshness metadata so frontend warnings are explicit and honest. Grouped activity snapshots are deferred until profiling shows the live path is too expensive.
