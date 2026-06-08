---
title: N/A Neutral Unwind Refunds
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-07T00:00:00Z
updated_at_display: "Sunday, June 7, 2026"
update_reason: "Define N/A market resolution as a neutral unwind that conserves retained market value without endorsing final market probability."
status: proposed
---

# N/A Neutral Unwind Refunds

## Purpose

Some markets should not resolve to `YES` or `NO`. The contract may be ambiguous, impossible to adjudicate, invalid, or otherwise not suitable for ordinary winner payout. In those cases, SocialPredict needs an `N/A` resolution path that voids the directional outcome while preserving economy integrity.

`N/A` must not endorse the market-implied probability at the time the market is voided. Instead, it should unwind remaining positions using a neutral binary valuation and refund the retained market pool back to current participants.

## Product Rule

When a market resolves `N/A`:

- The public outcome is `N/A` / `?`, not `YES`, `NO`, or `50%`.
- The chart should show the market as voided, not as resolved to a final directional probability.
- Refund valuation uses neutral binary probability `P = 0.5` as an accounting convention only.
- Remaining positions are invalidated after the refund.
- Users should no longer carry YES/NO exposure for that market in portfolio or financial views.
- The market remains visible for audit/history unless a separate yank/obfuscation feature hides unsafe content.

## Moderator/Admin Disclosure

The backend currently accepts `N/A` as a resolution value, but the frontend must not expose that action without a confirmation disclosure. Any moderator/admin control that resolves a market as `N/A` must clearly explain what the action does before submission.

Required disclosure points:

- `N/A` voids the directional market outcome.
- No `YES` or `NO` winner payout is made.
- Refund valuation uses neutral `P = 0.5` as an accounting convention only.
- `P = 0.5` is not shown as the market's final probability.
- Remaining retained market value, including retained dust, is refunded according to the neutral unwind policy.
- Market creation/proposal cost is not refunded.
- Moderator work profit is not paid.
- Prior seller proceeds are not clawed back.
- User positions become zero effective exposure after the refund.
- The action is irreversible and remains auditable.

## Accounting Rule

The refund must be derived from canonical market bet history, not cached display data.

```text
refundPool = retained market pool = net market volume + retained market dust
sum(refunds) == refundPool
```

The refund pool includes capital still retained by the market. It does not claw back credits from users who previously sold positions and received sale proceeds.

## Fee Rule

| Money Source | N/A Treatment |
| --- | --- |
| Remaining market volume | Refund to current claim holders. |
| Retained market dust | Refund as deterministic residual allocation. |
| Market creation/proposal fee | Not refunded. |
| Moderator work profit | Not paid. |
| Prior sale proceeds | Not clawed back. |
| Initial participation fee | Open decision; baseline should avoid refunding it unless explicitly included in retained market pool. |

The market creation/proposal fee should remain non-refundable. Otherwise, market makers could be incentivized to create markets, observe low participation or bad economics, and void the market to recover their creation cost.

## Neutral Valuation

The neutral unwind deliberately avoids using the market's final live probability.

```text
N/A refund valuation probability = 0.5
```

This says:

```text
The platform is voiding the contract. It is not adjudicating YES/NO and is not validating the final market price.
```

That creates a possible incentive for participants to price risky or ambiguous markets closer to 50%. This is acceptable and may be beneficial: markets that appear vulnerable to `N/A` should not attract extreme directional confidence unless participants trust the contract quality.

## Position Invalidation

After an `N/A` refund, the market should no longer contribute live position value.

Required invariant:

```text
for every user position on N/A market:
  yesSharesOwned = 0 for display/effective portfolio purposes
  noSharesOwned = 0 for display/effective portfolio purposes
  positionValue = 0
  unresolvedExposure = 0
```

The raw bet history remains untouched for audit. Position invalidation should be achieved by resolution-aware position calculation, not by deleting or rewriting bet rows.

This avoids historical data loss while ensuring the user's books no longer show value in a voided market.

## Divide-By-Zero Safety

`N/A` can produce zero effective market volume and zero effective share value after invalidation. Every display and calculation path that reads resolved market positions must avoid divide-by-zero behavior.

Required guards:

- If market is resolved `N/A`, do not calculate sell quotes.
- If market is resolved `N/A`, do not calculate live value-per-share for user sale paths.
- If neutral unwind produces zero eligible positions, refund nothing beyond explicit residual policy.
- If denominator is zero in charts, position tables, financials, or leaderboard views, return zero/empty state rather than dividing.
- If `VolumeWithDust == 0`, avoid payout normalization and mark the market as void with zero effective exposure.

## Acceptance Criteria

- Resolving a market `N/A` does not use final market-implied probability for refunds.
- Resolving a market `N/A` uses neutral `P = 0.5` for refund valuation only.
- `N/A` refunds are derived from canonical chronological market bet history.
- `N/A` refunds conserve the retained market pool, including retained dust.
- `N/A` does not pay moderator work profit.
- `N/A` does not refund market creation/proposal cost.
- Prior seller proceeds are not clawed back.
- After `N/A`, effective user positions for that market display as zero exposure/value.
- Moderator/admin UI requires a confirmation disclosure before submitting `N/A`.
- Tests prove no divide-by-zero paths in market detail, portfolio, financial, leaderboard, and payout/refund calculations.
