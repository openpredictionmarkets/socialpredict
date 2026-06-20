---
title: Grouped Market N/A And Documentation Alignment
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Capture design-agent finding that grouped N/A and answer-addition scope docs drift from implementation."
status: draft
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

- If group `N/A` is supported, implementation includes endpoint behavior and payout/refund tests.
- If group `N/A` is deferred, docs and UI do not imply it exists.
- Answer additions are described consistently as implemented baseline behavior when enabled by governance policy.
- `FEATURES/13` checklists are updated to reflect completed vs deferred work.
