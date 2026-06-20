---
title: Grouped Market N/A And Documentation Alignment Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Define alignment rules for grouped N/A and answer-addition language."
status: draft
---

# Grouped Market N/A And Documentation Alignment Design

## Design Posture

Documentation is part of the published language boundary. If docs describe a resolution mode or governance flow that the system does not actually support, users and reviewers get a false contract.

## Decision Needed

| Capability | Current state | Decision needed |
| --- | --- | --- |
| Group `N/A` | Docs describe it; implementation rejects manual `N/A` in grouped resolution. | Implement now or mark deferred. |
| Later answer additions | Overview says out of scope; implementation supports approved additions. | Update overview and plan to say in scope. |

## Preferred Direction

Short term: update docs to match current behavior unless group `N/A` is implemented in the same PR.

Long term: implement group `N/A` as a helper that calls existing child-market `N/A` refund logic for every child and proves money-in/money-out integrity.
