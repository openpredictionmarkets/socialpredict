---
title: Analytics Repository Boundary
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Capture design-agent finding that analytics domain contains GORM persistence implementation."
status: draft
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

- Domain analytics package no longer imports GORM.
- GORM analytics implementation lives under `backend/internal/repository/analytics` or equivalent adapter package.
- Existing analytics service tests still pass.
- Repository tests cover moved query behavior.
- Public API behavior remains unchanged.
