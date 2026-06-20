---
title: Grouped Answer Limit Contract Alignment Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record answer limit contract alignment implementation."
status: implemented
---

# Grouped Answer Limit Contract Alignment Plan

## Checklist

- [x] Confirm intended domain maximum answer count.
- [x] Align DTO validation with domain/setup maximum.
- [x] Align OpenAPI `maxItems` and description.
- [x] Ensure frontend uses setup-provided cap.
- [x] Add/adjust API validation tests.
- [x] Run OpenAPI validation.

## Implementation Notes

- Domain hard cap remains `50`.
- Create-group DTO validation now accepts up to `50` answers.
- OpenAPI `answerLabels.maxItems` now documents `50` and notes deployment setup may lower the effective cap.
- Domain tests cover exactly `50` answers accepted and `51` rejected.
