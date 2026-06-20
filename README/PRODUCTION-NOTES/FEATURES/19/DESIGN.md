---
title: Grouped Market N/A And Documentation Alignment Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record alignment rules for grouped N/A and answer-addition language."
status: implemented
---

# Grouped Market N/A And Documentation Alignment Design

## Design Posture

Documentation is part of the published language boundary. If docs describe a resolution mode or governance flow that the system does not actually support, users and reviewers get a false contract.

## Decision Needed

| Capability | Implemented state | Decision |
| --- | --- | --- |
| Group `N/A` | `mode = na` resolves every child answer market N/A through existing child refund logic. | Shipped in this branch. |
| Later answer additions | Active moderators can propose/add answers through the implemented governance flow. | Baseline behavior, not out of scope. |

## Current Direction

Docs and implementation now match the shipped behavior: group `N/A` exists, and dynamic answer additions are part of the baseline governance flow.

## Implemented Behavior

Group `N/A` is implemented as a group helper, not a new payout formula. The helper maps every child answer market to `N/A`, calls the ordinary child-market resolution/refund path, marks the parent group resolved, and skips grouped work-profit payout.

Answer additions are documented as an implemented governance flow. They remain separate from initial answer creation economics: initial answers are included in one group proposal cost, while later approved answers use the setup-configured add-answer cost.
