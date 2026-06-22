---
title: Analytics Repository Boundary Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Record analytics GORM adapter extraction."
status: implemented
---

# Analytics Repository Boundary Plan

## Checklist

- [x] Inventory GORM imports and SQL in `internal/domain/analytics`.
- [x] Create `internal/repository/analytics` adapter package.
- [x] Move GORM-backed repository implementation into adapter package.
- [x] Keep domain-owned repository interfaces in `internal/domain/analytics`.
- [x] Update container wiring.
- [x] Move or add repository tests for SQL behavior.
- [x] Run full backend tests.

## Implementation Notes

- `internal/domain/analytics` now owns analytics interfaces, records, and service logic.
- `internal/repository/analytics` owns GORM-backed analytics persistence, SQL selects, snapshot repositories, and repository-focused tests.
- Application/container wiring now imports the repository adapter explicitly.
- Handler tests that need a concrete analytics repo use `internal/repository/analytics`.
