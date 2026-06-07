---
title: Moderator Work Profits
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-07T00:00:00Z
updated_at_display: "Sunday, June 7, 2026"
update_reason: "Define stateless moderator work-profit payout and financial reporting for resolved markets."
status: draft
---

# Moderator Work Profits

## Purpose

Moderators pay to create markets. If those markets attract participants, the current steward who governs and resolves the market should later receive income from the first-participation fees collected on that market. This feature treats that income as moderator work profit while preserving transaction correctness and existing canonical tables.

## Rule Summary

- The market creator pays the existing market creation/proposal cost when creating the market.
- Each unique participant pays the existing initial market entry fee the first time they place a positive bet on that market.
- Selling out and buying back in does not create another initial entry fee for the same user on the same market.
- Collected entry fees remain server-side until market resolution.
- After a market resolves to `YES` or `NO`, ordinary winner payouts run first.
- After ordinary payout, the current steward/resolver receives the derived collected entry-fee income as a `WORK_PROFIT` balance credit.
- `N/A` refund resolutions do not pay moderator work profit in this baseline.
- User financial display derives net work profit for resolved markets stewarded by that user. If the steward was also the creator, display subtracts the market creation/proposal fee; if the steward took over later, display does not subtract someone else's creation cost.

## Stateless Accounting

This feature does not add new database columns or tables. Work-profit income is derived from existing market and bet records:

```text
uniquePositiveParticipants = count(distinct username where market_id = M and bet.amount > 0)
marketFeeIncome = uniquePositiveParticipants * initialBetFee
marketCreationFee = market.proposalCost if steward is creator, otherwise 0
netWorkProfit = marketFeeIncome - marketCreationFee
```

Resolution-time payout credits only `marketFeeIncome`; the creation fee has already been deducted from the original creator when the market was created. Financial display subtracts the creation fee only when the work-profit beneficiary is also the market creator.

## Boundary Notes

- Payout decisions use canonical bet history, not cached/read-model snapshots.
- Work-profit display may appear in user financial read models, but those read models remain display-only.
- System metrics should treat participation fees on resolved `YES`/`NO` markets as redistributed steward work income, not retained system fees, to avoid double-counting the same money.
- The payout goes to the current steward who governs resolution. A reassigned steward receives the work income; the original creator does not receive it unless they remain steward.
- Future historical policy changes may require durable per-market fee policy capture. The current implementation intentionally uses existing state and current economics config to avoid adding new state for this baseline.

## Acceptance Criteria

- Resolving a non-`N/A` market pays ordinary winners first.
- Resolving a non-`N/A` market then credits the current steward/resolver for one initial fee per unique positive-bet participant.
- Resolving `N/A` refunds participant bet amounts and does not credit work profit.
- User financials show `workProfits` for resolved markets stewarded by that user, subtracting market creation/proposal cost only when that user was also the creator.
- Tests prove repeated participation by the same user is counted once.
- Tests prove sale/negative bet rows do not create extra work-profit income.
