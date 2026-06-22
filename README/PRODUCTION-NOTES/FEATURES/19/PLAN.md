---
title: Grouped Market N/A And Documentation Alignment Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record grouped N/A and documentation alignment implementation."
status: implemented
---

# Grouped Market N/A And Documentation Alignment Plan

## Checklist

- [x] Decide whether group `N/A` ships in this branch.
- [x] If shipping, add group `N/A` API mode and tests.
- [x] If deferred, update FEATURE/13 design and overview to say deferred.
- [x] Update answer-addition scope in FEATURE/13 overview.
- [x] Update OpenAPI text to match final grouped resolution behavior.
- [x] Update frontend copy where needed.

## Implementation Notes

- Group `N/A` ships in this branch.
- `mode = na` resolves every child answer market through the existing child-market `N/A` refund path.
- Group `N/A` does not pay grouped work-profit income.
- Dynamic answer additions are implemented baseline behavior, so FEATURE/13 no longer lists them as out of scope.
