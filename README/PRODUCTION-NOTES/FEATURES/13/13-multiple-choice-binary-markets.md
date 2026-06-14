---
title: Multiple Choice Binary Markets
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-14T00:00:00Z
updated_at_display: "Sunday, June 14, 2026"
update_reason: "Clarify grouped market economics: initial answers are included in one group proposal cost; later answer additions are controlled by setup policy."
status: draft
---

# Multiple Choice Binary Markets

## Purpose

SocialPredict currently supports one binary market per market page. This feature proposes a new participant-facing market class that lets a moderator create a single parent question with multiple answer choices, where each answer choice is traded as its own binary YES/NO market and displayed together on one page.

The purpose is to support questions like:

```text
Who will win the tournament?

- Team A
- Team B
- Team C
- Team D
```

without forcing moderators to manually create and maintain several disconnected markets.

## Core Direction

The baseline should be a **multiple-choice binary market group**, not a new coupled multi-outcome AMM.

That means:

- A parent market group owns the shared question, description, governance, tags, and display page.
- Each answer choice becomes a normal binary child market with its own market ID, bets, probability, chart, volume, positions, dust, and resolution state.
- Existing WPAM/DBPM binary math remains authoritative for each child market.
- Existing buy/sell/dust/accounting behavior remains scoped to one child market at a time.
- The parent page displays the child markets together and provides group-level governance controls.

## Probability Sum Decision

The probabilities should **not** be required to add to `1.0` in the baseline.

Reason:

- Current SocialPredict math is built around one binary YES/NO market.
- A group of binary child markets can preserve that math without inventing new accounting behavior.
- Sum-to-one behavior would require a different coupled liquidity model or a normalization layer that could mislead users if it does not match payout math.
- Independent child probabilities can naturally sum above or below `1.0` because each answer is its own binary market.

Product copy should make this explicit:

```text
Each answer is traded as its own YES/NO market. Probabilities are not normalized to add to 100%.
```

## Market Class Vocabulary

| Term | Meaning |
| --- | --- |
| Binary market | Current single-question YES/NO market. |
| Market group | New parent container for a shared multiple-choice question. |
| Binary child market | A normal SocialPredict binary market representing one answer choice inside a group. |
| Answer choice | Moderator-authored label such as `Team A`. |
| Probability policy | Group-level rule declaring whether child markets are independent or future sum-to-one/exclusive. |
| Independent binary policy | Baseline policy: every answer is a separate YES/NO market and probabilities do not need to sum to `1.0`. |
| Exclusive resolution helper | Future or optional workflow that resolves one child YES and remaining children NO, without changing the trading math. |

## Product Behavior

Moderator creation flow:

- Moderator selects `Binary market` or `Multiple-choice binary market`.
- For multiple-choice binary market, moderator enters a parent title and description.
- Moderator adds answer choices, each with a display label.
- Moderator adds tags at the parent/group level.
- Moderator submits the group for admin review using the existing market proposal pattern.
- Initial answer choices are included in the one group proposal cost; future answer additions, if enabled, use setup-configured answer-addition pricing.

Admin review flow:

- Admin reviews the parent group and all answer choices together.
- Admin can approve or reject the whole group.
- Admin can require edits before approval in a future slice.
- Admin can see each generated child market title, labels, and tags before approval.

Participant display flow:

- `/markets` and topic pages show the parent market group as one discovery card.
- The parent page displays all answer choices in a consistent list/grid.
- Each answer card shows current probability, volume, user count, chart preview, and trade action.
- Clicking or expanding an answer opens the normal binary trading controls for that child market.
- The child market can still have a stable direct URL for sharing and auditability.

Resolution flow:

- Baseline: steward/admin resolves each child market independently as YES, NO, or N/A.
- Optional helper: if a group is marked as an exclusive real-world outcome, steward/admin can pick one winner and the system resolves that child YES and remaining children NO.
- The helper must still execute normal child-market resolution paths for each child.
- Payouts remain child-market payouts; there is no parent-level pooled payout in the baseline.

## Acceptance Criteria

- Existing binary markets continue to work without behavior changes.
- A multiple-choice binary group can be created with at least two answer choices.
- Each answer choice maps to a normal binary child market.
- Child market probabilities are calculated with existing binary market math.
- Group probabilities are not normalized unless a future explicit sum-to-one market type is designed.
- Admin can approve/reject the group as a single proposal.
- Tags on the parent group are visible in market discovery and inherited or projected to children for search/filter behavior.
- The group page clearly states that each answer is a separate YES/NO market.
- Resolution does not use stale read models or parent display data for payouts.
- Group work profit is paid once to the current group steward after group resolution, using unique participants across the child answer markets.
- Migrations are additive timestamped Go migrations with package-local tests where practical.

## Out Of Scope For Baseline

- Coupled multi-outcome AMM.
- Sum-to-one probability enforcement.
- Parent-level pooled liquidity.
- Cross-answer arbitrage balancing.
- Allowing users to buy a bundle across all answers in one transaction.
- Dynamic answer additions after trading has started.
- Deleting answer choices after approval.
