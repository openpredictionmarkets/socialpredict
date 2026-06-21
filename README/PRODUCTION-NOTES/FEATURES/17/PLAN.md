---
title: Grouped Activity Freshness Metadata Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record grouped activity freshness metadata implementation."
status: implemented
---

# Grouped Activity Freshness Metadata Plan

## Checklist

- [x] Decide short-term live freshness vs. remove freshness fields.
- [x] Update grouped activity handlers accordingly.
- [x] Update DTOs and OpenAPI.
- [x] Update frontend rendering to match real contract.
- [x] Add handler tests.
- [x] Add frontend smoke/manual verification notes.

## Implementation Notes

- Short-term grouped activity freshness is `source = live`, `transactionSafeRead = false`, and `targetFreshnessSeconds = 0`.
- Grouped bets, grouped positions, and grouped leaderboard responses now carry explicit freshness metadata.
- Frontend copy now says grouped activity is computed live and keeps trade confirmations authoritative.
- Longer-term grouped activity snapshots remain a separate optimization if these endpoints become hot.
