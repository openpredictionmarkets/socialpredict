---
title: Grouped Market N/A And Documentation Alignment
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record grouped N/A implementation and answer-addition doc alignment."
status: implemented
---

# Grouped Market N/A And Documentation Alignment

## Purpose

The grouped-market documentation and implementation currently disagree in two areas:

- whether group-level `N/A` resolution is supported now or deferred
- whether dynamic answer additions after trading starts are out of scope or implemented baseline behavior

This feature reconciles documentation and behavior before merge/release.

## Source Finding

Design review finding: P3 feature docs have drift.

Relevant refs:

- `README/PRODUCTION-NOTES/FEATURES/13/DESIGN.md:269`
- `README/PRODUCTION-NOTES/FEATURES/13/DESIGN.md:271`
- `README/PRODUCTION-NOTES/FEATURES/13/13-multiple-choice-binary-markets.md:126`

## Desired Outcome

Feature docs, OpenAPI, frontend copy, and backend behavior all describe the same grouped-market capabilities.

## Acceptance Criteria

- [x] Group `N/A` implementation includes endpoint behavior and payout/refund tests.
- [x] Docs and UI describe group `N/A` as a supported resolution helper.
- [x] Answer additions are described consistently as implemented baseline behavior when enabled by governance policy.
- [x] `FEATURES/13` checklists are updated to reflect completed vs deferred work.

## Implementation Notes

Group `N/A` is supported through the grouped resolution endpoint with `mode = na`. It resolves each child answer market N/A through the existing child-market refund logic, marks the parent group resolved, and skips grouped work-profit payout.

Answer additions after initial approval are implemented baseline governance behavior. They do not rewrite existing child history.
