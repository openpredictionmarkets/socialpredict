---
title: Grouped Discovery Read Model Pagination Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Define backend-owned grouped discovery rows before pagination."
status: draft
---

# Grouped Discovery Read Model Pagination Design

## Design Posture

Discovery is display/read-model state. It may be cached and stale with freshness metadata, but it should not invent market truth in the frontend.

The backend should own grouped discovery aggregation because pagination, total counts, sorting, and filtering must all be coherent before the browser sees the page.

## Target Shape

A grouped discovery row should include:

- parent group ID
- representative child market ID for route compatibility
- parent question/title
- parent description excerpt
- answer count
- ordered answer labels and child market IDs
- aggregate volume/user/dust display values
- resolution summary
- tags
- creator and steward
- freshness metadata where applicable

## Query Rule

Filter first by public search/topic/status rules, group child matches under their parent group, then sort and paginate grouped rows.

The current implementation performs this as a backend read-model grouping pass: it queries matching child markets, batch-loads parent group/member metadata and all child answer markets for matched groups, collapses each parent to one display row, and only then applies `limit` and `offset`. That fixes pagination correctness without moving transaction math into SQL.

## Compatibility Rule

Frontend helpers may still collapse child rows defensively, but that must not be the primary correctness path.

## Risks

- Aggregating all child markets on every request could recreate the load problem that read models were intended to solve.
- Search must match both parent group text and answer labels.
- Topic tag filtering must preserve parent-group tag semantics.
