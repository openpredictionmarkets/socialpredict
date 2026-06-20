---
title: Grouped Answer Limit Contract Alignment
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record answer limit contract alignment implementation."
status: implemented
---

# Grouped Answer Limit Contract Alignment

## Purpose

The grouped market answer-count contract should be consistent across OpenAPI, DTO validation, setup configuration, domain validation, and frontend UI. The current branch has a documented/API cap of `20` while setup/frontend/runtime can imply `50`.

## Source Finding

Design review finding: P2 initial grouped-answer limits are inconsistent.

Relevant refs:

- `backend/docs/openapi.yaml:5960`
- `backend/handlers/markets/dto/requests.go:18`
- `backend/setup/setup.yaml:11`

## Desired Outcome

Every layer communicates the same hard safety cap behavior.

## Acceptance Criteria

- [x] OpenAPI schema matches the maximum runtime-supported hard cap.
- [x] DTO validation matches the domain and setup policy.
- [x] Frontend copy reflects setup-provided cap.
- [x] Tests cover too-many-answer validation at API and domain boundaries.

## Implementation Notes

The supported initial answer cap is `50`. Deployment setup may still lower the effective user-facing cap, but the API schema and DTO validation now align with the domain hard cap.
