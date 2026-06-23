---
title: Grouped Market Transaction Boundaries Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Plan grouped-market atomic write boundary implementation."
status: draft
---

# Grouped Market Transaction Boundaries Plan

## Checklist

- [x] Inventory grouped creation, answer-addition approval, and resolution write steps.
- [x] Define grouped-market unit-of-work port owned by the domain layer.
- [x] Implement the port in the GORM repository adapter.
- [x] Keep domain validation and policy outside the GORM adapter.
- [x] Make grouped resolution idempotent or fully transactional.
- [x] Guard grouped work income against duplicate payout.
- [x] Add failure-injection tests for grouped creation.
- [x] Add failure-injection tests for answer-addition approval.
- [x] Add failure-injection tests for grouped resolution retry.
- [x] Update FEATURE/13 plan once this is implemented.

## Implementation Notes

- Grouped creation, answer-addition approval, and grouped resolution now use a market-domain unit-of-work port.
- The GORM adapter owns the database transaction and supplies transaction-scoped market and user services.
- Business validation, lifecycle policy, answer policy, amendment generation, and payout decisions remain in the domain service.
- Failure-injection tests verify rollback for orphan child creation, charged-but-unlinked answer additions, and partial grouped resolution.
- Duplicate work-income payout is guarded by the same transaction and the resolved parent lifecycle marker: once committed, a second resolution attempt fails before another payout; if a later write fails, the payout rolls back with the transaction.
