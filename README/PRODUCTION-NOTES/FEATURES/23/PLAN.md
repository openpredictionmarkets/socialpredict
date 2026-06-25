---
title: External Market Movement Sale Lock Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-06-25T00:00:00Z
updated_at_display: "Thursday, June 25, 2026"
update_reason: "Plan implementation and tests for requiring another user's later bet before sale value can be extracted."
status: in_progress
---

# External Market Movement Sale Lock Plan

## 01. Feature Spec

Checklist:

- [x] Document the lock as an economic sale rule, not a database lock.
- [x] Define the key policy: another user's later bet is required.
- [x] Define self-top-up as the exploit case to block.
- [x] Define same-market and grouped-market child boundaries.
- [x] Record the stricter favorable-movement rule as future work.

## 02. Code Audit

Checklist:

- [x] Locate every backend sell and sale-quote entrypoint.
- [x] Confirm whether any existing sale lock is implemented or only implied by position math.
- [x] Identify the canonical market bet-history reader used by sale execution.
- [x] Confirm canonical ordering is `placed_at ASC, id ASC`.
- [x] Identify whether sale rows and refund rows are distinguishable from positive buy exposure rows.
- [x] Confirm grouped child markets use independent market IDs for sale math.
- [ ] Document any frontend-only sale restrictions that need backend enforcement.

## 03. Domain Model

Checklist:

- [x] Add domain language for seasoned and locked sale exposure.
- [x] Add a sale-lock evaluator that consumes ordered market bet events.
- [x] Keep the evaluator free of GORM and HTTP dependencies.
- [x] Return a structured result with sellable shares, locked shares, and lock reason.
- [x] Decide whether the first version rejects oversize sale orders or caps them. Current recommendation: reject.

## 04. Backend Implementation

Checklist:

- [x] Wire sale quote through the lock evaluator.
- [x] Wire sale execution through the same lock evaluator.
- [x] Ensure sale execution rechecks lock eligibility inside the transaction.
- [x] Reject sale orders that require locked shares.
- [x] Add a stable error code such as `POSITION_LOCKED_AWAITING_EXTERNAL_MARKET_MOVEMENT`.
- [x] Ensure same-user later buys do not season earlier lots.
- [x] Ensure later positive buys by another user season prior lots on the same market.
- [x] Ensure sell rows do not season prior lots.
- [x] Ensure grouped-market answer children are isolated by child market ID.

## 05. Frontend Implementation

Checklist:

- [x] Display sellable and locked shares in sale terms/quote UI.
- [x] Explain that locked shares require a later bet from another user on the same market.
- [x] Show backend lock errors clearly.
- [ ] Avoid promising that a quote is guaranteed if high trade volume changes market state before submission.
- [ ] Keep the backend as authoritative; frontend checks are advisory.

## 06. Tests

Checklist:

- [x] Alice buys YES 50, Alice buys YES 1, Alice tries to sell: blocked.
- [x] Alice buys YES 50, Bob buys YES or NO 1, Alice sells: allowed under the initial same-market external-entry rule.
- [x] Alice buys YES 50, Bob buys NO 1, Alice sells: allowed under the initial same-market external-entry rule.
- [ ] Alice buys YES 50, Alice sells negative/sale row exists after it: still blocked unless another user bought.
- [ ] Alice buys YES on grouped child A, Bob buys YES on grouped child B: Alice child A sale remains blocked.
- [ ] Multiple same-timestamp rows use `placed_at ASC, id ASC` ordering.
- [x] Sale quote and sale execution return consistent sellable/locked results.
- [x] Execution rechecks inside transaction so a stale quote cannot bypass the lock.
- [x] Existing dust tests still pass.
- [x] Existing payout/resolution tests still pass.

## 07. Verification

Checklist:

- [x] Run targeted backend tests for sale lock scenarios.
- [x] Run full backend tests.
- [x] Run frontend build if sale quote UI changes.
- [ ] Run schemathesis/kin if API response contracts change.
- [ ] Add or update OpenAPI response schemas if sale quote/error payload changes.
- [ ] Manually verify the exploit case in local dev with two test users.

## 08. Rollout

Checklist:

- [ ] Keep the first version behind normal sale path behavior, not a separate opt-in route.
- [ ] Document the new error and UI language in release notes.
- [ ] Watch support/admin reports for users seeing locked shares after deployment.
- [ ] Consider a migration-free rollout because the first version is stateless.
- [ ] Revisit stricter favorable-movement unlocking only after the basic external-user lock is stable.
