---
title: Moderator Work Profits
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-16T00:00:00Z
updated_at_display: "Tuesday, June 16, 2026"
update_reason: "Clarify gross resolution-time work-income payout versus net financial work-profit reporting."
status: draft
---

# Moderator Work Profits

## Purpose

Moderators pay to create markets. If those markets attract participants, the current steward who governs and resolves the market later receives the collected first-participation fee income. Financial reporting then subtracts the proposal cost to show net work profit, which can be negative when a market underperforms.

## Rule Summary

- The market creator pays the existing market creation/proposal cost when creating the market.
- Each unique participant pays the existing initial market entry fee the first time they place a positive bet on that market.
- Selling out and buying back in does not create another initial entry fee for the same user on the same market.
- Collected entry fees remain server-side until market resolution.
- After a market resolves to `YES` or `NO`, ordinary winner payouts run first.
- After ordinary payout, the current steward/resolver receives collected entry-fee income as a gross `WORK_PROFIT` balance transaction.
- `N/A` refund resolutions do not pay moderator work profit in this baseline.
- User financial display derives net work profit for resolved markets stewarded by that user by subtracting the market creation/proposal cost from collected entry-fee income.
- User financial display also shows unrealized work income and unrealized work profit for unresolved markets. Income is projected participant-fee income for markets currently stewarded by the user; profit nets that income against unresolved proposal-cost exposure for markets the user created.

## Stateless Accounting

This feature does not add new database columns or tables. Work-profit income is derived from existing market and bet records:

```text
uniquePositiveParticipants = count(distinct username where market_id = M and bet.amount > 0)
marketFeeIncome = uniquePositiveParticipants * initialBetFee
marketCreationCost = market.proposalCost or configured createMarketCost fallback
resolutionWorkIncomePayout = marketFeeIncome
financialNetWorkProfit = marketFeeIncome - marketCreationCost
unrealizedWorkIncome = unresolvedStewardedMarketFeeIncomeSoFar
unrealizedWorkProfit = unresolvedStewardedMarketFeeIncomeSoFar - unresolvedCreatedMarketProposalCosts
```

Resolution-time payout credits gross entry-fee income. The creation fee has already been deducted from the creator when the market was created, so subtracting it again from the balance transaction would double-charge the steward/creator path. Financial views still show the lifecycle economics by reporting net work profit as fee income minus proposal cost.

## Boundary Notes

- Payout decisions use canonical bet history, not cached/read-model snapshots.
- Work-profit display may appear in user financial read models, but those read models remain display-only.
- System metrics should distinguish collected participant fees, gross steward work-income payouts, and net work profit so retained-fee reporting does not double-count.
- The payout goes to the current steward who governs resolution. A reassigned steward receives the work income; the original creator does not receive it unless they remain steward.
- Future historical policy changes may require durable per-market fee policy capture. The current implementation intentionally uses existing state and current economics config to avoid adding new state for this baseline.

## Acceptance Criteria

- Resolving a non-`N/A` market pays ordinary winners first.
- Resolving a non-`N/A` market then credits the current steward/resolver for gross first-participation fee income.
- Resolving `N/A` refunds participant bet amounts and does not credit work profit.
- User financials show net `workProfits` for resolved markets stewarded by that user, including negative values when fee income is below proposal cost.
- Tests prove repeated participation by the same user is counted once.
- Tests prove sale/negative bet rows do not create extra work-profit income.
