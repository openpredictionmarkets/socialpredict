---
title: Grouped Market Transaction Boundaries
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-17T00:00:00Z
updated_at_display: "Wednesday, June 17, 2026"
update_reason: "Capture design-agent finding that grouped market write paths need atomic transaction boundaries."
status: draft
---

# Grouped Market Transaction Boundaries

## Purpose

Grouped market creation, grouped market resolution, and grouped answer-addition approval currently span multiple domain and repository calls. If a later call fails, the system can leave partially written state: charged users, orphan child markets, missing group links, partially resolved groups, or approved additions without generated amendments.

This feature creates a dedicated transaction boundary for grouped-market write use cases.

## Source Finding

Design review finding: P1 grouped-market write paths are not atomic enough for the design plan's transaction and audit posture.

Relevant refs:

- `backend/internal/domain/markets/market_group_creation.go:43`
- `backend/internal/domain/markets/market_group_creation.go:82`
- `backend/internal/domain/markets/market_group_creation.go:96`
- `backend/internal/domain/markets/market_group_answer_additions.go:188`
- `backend/internal/domain/markets/market_group_resolution.go:45`

## Desired Outcome

Grouped write operations either fully commit or leave no user-visible economic/governance side effect.

Critical write use cases:

- create grouped market
- approve grouped answer addition
- resolve grouped market
- pay grouped work income exactly once
- write generated answer-addition amendments with the approved addition

## Acceptance Criteria

- Group creation cannot leave proposal fee charged without a parent group and child links.
- Group creation cannot leave orphan child markets if parent group creation fails.
- Answer-addition approval cannot charge the proposer without creating the child answer and review state.
- Answer-addition approval cannot create a child answer without generated amendment/audit state.
- Group resolution cannot leave partial child resolutions without a safe retry/idempotency path.
- Group work income cannot be paid twice on retry.
- Failure-injection tests prove the behavior above.
