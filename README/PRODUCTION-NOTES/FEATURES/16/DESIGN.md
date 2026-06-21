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
| Group created | A new parent row and grouped child set can appear in discovery. |
| Group approved | New parent row appears in discovery. |
| Group rejected | Proposed group should not appear as active discovery. |
| Answer added | Answer count, child IDs, tags, volume/user projections, and search terms can change. |
| Answer addition reviewed | Approval can add a child answer; rejection can change review-visible state. |
| Group resolved | Status and resolution summary change. |
| Group stewardship update | Public steward display metadata can change. |
| Tag catalog or market tag update | Topic routes and tag chips can change. |

## Event Vocabulary

| Invalidation reason | Discovery behavior |
| --- | --- |
| `market_group_created` | Structural stale. |
| `market_group_approved` | Structural stale. |
| `market_group_rejected` | Structural stale. |
| `market_group_answer_added` | Structural stale. |
| `market_group_answer_reviewed` | Structural stale. |
| `market_group_resolved` | Structural stale. |
| `market_steward_changed` | Structural stale. |
| `market_tags_changed` | Structural stale. |
| `tag_catalog_changed` | Structural stale. |
| `cms_page_changed` | Structural stale. |
| `cms_pins_changed` | Structural stale. |
| `bet_accepted` | Soft stale; trade confirmation remains authoritative, discovery can refresh on the normal cadence. |
| `market_group_answer_addition_policy` | Soft stale; this affects moderator workflow controls, not public discovery structure. |

## Freshness Rule

A structurally stale discovery snapshot should either be refreshed before serving or marked stale in a way that the handler refuses to treat it as fresh.

## Out Of Scope

- Replacing all discovery caching.
- Changing transaction path behavior.
