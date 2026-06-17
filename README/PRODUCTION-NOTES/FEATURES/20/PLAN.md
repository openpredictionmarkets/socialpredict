---
title: Analytics Repository Boundary Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Plan analytics GORM adapter extraction."
status: draft
---

# Analytics Repository Boundary Plan

## Checklist

- [ ] Inventory GORM imports and SQL in `internal/domain/analytics`.
- [ ] Create `internal/repository/analytics` adapter package.
- [ ] Move GORM-backed repository implementation into adapter package.
- [ ] Keep domain-owned repository interfaces in `internal/domain/analytics`.
- [ ] Update container wiring.
- [ ] Move or add repository tests for SQL behavior.
- [ ] Run full backend tests.
