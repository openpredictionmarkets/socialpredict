---
title: Analytics Repository Boundary
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record analytics repository boundary implementation."
status: implemented
---

# Analytics Repository Boundary

## Purpose

The analytics domain package currently contains GORM persistence implementation. That predates some grouped-market work, but the branch extends analytics SQL paths. This violates the clean-architecture posture that domain policy should not depend on database mechanisms.

This feature moves analytics persistence details out of `internal/domain/analytics` and into a repository adapter package.

## Source Finding

Design review finding: P3 clean-architecture concern. Analytics domain still contains GORM persistence implementation.

Relevant ref:

- `backend/internal/domain/analytics/repository.go:12`

## Desired Outcome

The analytics domain owns interfaces, records, and policy. The repository adapter owns GORM queries and SQL details.

## Acceptance Criteria

- [x] Domain analytics package no longer imports GORM.
- [x] GORM analytics implementation lives under `backend/internal/repository/analytics` or equivalent adapter package.
- [x] Existing analytics service tests still pass.
- [x] Repository tests cover moved query behavior.
- [x] Public API behavior remains unchanged.

## Implementation Notes

The GORM-backed analytics repository, analytics read-model snapshot persistence, and user financial metric snapshot persistence now live under `backend/internal/repository/analytics`. The domain package keeps the interfaces and service logic; concrete database behavior is tested from the repository package.
