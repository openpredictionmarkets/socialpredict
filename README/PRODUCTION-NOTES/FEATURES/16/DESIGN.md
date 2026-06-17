---
title: Grouped Discovery Cache Invalidation Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Define structural invalidation events for grouped discovery snapshots."
status: draft
---

# Grouped Discovery Cache Invalidation Design

## Design Posture

Read models may be stale, but structural changes should not be hidden behind a soft freshness window when the visible market set, answer count, or resolution state changed.

## Structural Group Events

| Event | Structural reason |
| --- | --- |
| Group approved | New parent row appears in discovery. |
| Group rejected | Proposed group should not appear as active discovery. |
| Answer added | Answer count, child IDs, tags, volume/user projections, and search terms can change. |
| Group resolved | Status and resolution summary change. |
| Group stewardship/tag update | Public display metadata can change. |

## Freshness Rule

A structurally stale discovery snapshot should either be refreshed before serving or marked stale in a way that the handler refuses to treat it as fresh.

## Out Of Scope

- Replacing all discovery caching.
- Changing transaction path behavior.
