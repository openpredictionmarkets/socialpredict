---
title: External Market Movement Sale Lock
document_type: feature-overview
domain: features
author: Patrick Delaney
updated_at: 2026-06-25T00:00:00Z
updated_at_display: "Thursday, June 25, 2026"
update_reason: "Document the economic sale lock that prevents self-created market movement from becoming immediately extractable value."
status: in_progress
---

# External Market Movement Sale Lock

## Purpose

SocialPredict market math should not allow a user to create mark-to-market value with their own bet and then immediately extract that value through a sale.

This feature defines the economic lock as:

```text
A bet lot cannot realize gained value until another user places a later qualifying bet on the same market.
```

The key policy change is the phrase **another user**.

A later bet from the same user must not unlock that user's earlier bet lot. A self-top-up such as:

```text
Alice buys YES for 50.
Alice buys YES for 1.
Alice tries to sell the original 50.
```

must remain locked because no external participant has entered after Alice's original 50-credit bet.

## Math Grounding

The prediction-market math already documents the intended anti-self-dealing posture: share value should not reward a user simply for moving the market themselves. Gained value should require later market movement by other participants.

This feature turns that posture into an explicit sale policy.

## Ubiquitous Language

| Term | Meaning |
| --- | --- |
| Economic lock | A transaction-time sale restriction that prevents self-created appreciation from being extracted. |
| Bet lot | One positive buy exposure row for a user, market, and outcome. |
| Seasoned lot | A buy lot that has been followed by a qualifying later bet from a different user. |
| Unseasoned lot | A buy lot that has not yet been followed by a qualifying later bet from a different user. |
| Qualifying later bet | A later positive buy exposure on the same market by a different user. |
| Self-top-up | A small later buy by the same user that attempts to unlock a larger earlier buy. |
| Sellable shares | Shares attributable to seasoned lots. |
| Locked shares | Shares attributable to unseasoned lots. |

## Rule

The first implementation keeps the rule intentionally narrow:

```text
A positive buy lot becomes sell-eligible only after a later positive buy on the same market by a different user.
```

The following must not unlock a lot:

- a later bet by the same user;
- a sale row;
- a refund row;
- a system/admin accounting row;
- a zero or negative amount row;
- activity on a different child market in a grouped market.

For multiple-choice binary markets, the lock applies independently to each child binary market. A bet on one answer does not unlock lots on another answer.

## Implementation Slice 1

The first landed slice enforces a position-level version of the lock in the backend sale path:

- sale quotes now return `allowed=false` when the seller's latest positive buy for that outcome has not been followed by a later positive bet from another user on the same market;
- sale execution rechecks the same rule inside the sell transaction and returns `POSITION_LOCKED_AWAITING_EXTERNAL_MARKET_MOVEMENT` before mutating balances or writing the sale row;
- same-user self-top-ups do not unlock the position;
- a later positive bet from another user unlocks the position;
- quote responses expose `positionLocked`, `sellableShares`, `lockedShares`, and `lockReason`;
- the frontend sale Terms panel displays locked quotes distinctly.

This first slice is deliberately stricter than full lot-level partial unlocking: if the seller's latest same-outcome positive buy is unseasoned, the sale is blocked. A later slice can allow partial sale of older seasoned lots once the UI and sale order model can clearly select or cap lot-level exposure.

## Non-Goals

This feature does not change:

- WPAM or DBPM probability math;
- dust accounting;
- market resolution payout math;
- N/A refund policy;
- moderator work-profit policy;
- read-model/cache behavior;
- database concurrency locking.

This is not a database lock. It is an economic sale-eligibility rule.

## Expected User-Facing Behavior

If a user attempts to sell only locked shares, the backend should reject the sale with a clear reason such as:

```text
POSITION_LOCKED_AWAITING_EXTERNAL_MARKET_MOVEMENT
```

The frontend sale quote should distinguish:

```text
Total shares
Sellable shares
Locked shares
```

The UI copy should explain that locked shares become sellable after another user places a later bet on the same market.

## Core Examples

| Scenario | Expected result |
| --- | --- |
| Alice buys YES 50, then Alice buys YES 1, then Alice sells | Sale remains locked. Same-user top-up does not season the first lot. |
| Alice buys YES 50, then Bob buys YES 1, then Alice sells | Sale can proceed against Alice's seasoned YES lot. |
| Alice buys YES 50, then Bob buys NO 1, then Alice sells | Sale can proceed under the initial simple rule because an external participant entered the same market. |
| Alice buys YES on child market A, then Bob buys YES on child market B | Alice's child market A lot remains locked. Grouped-market children are independent binary markets. |

## Open Decision

The first version treats any later positive buy by another user on the same market as a qualifying event.

A stricter future version could require favorable external movement before profitable sale extraction, for example:

```text
YES lots unlock profitable sale value only after another user moves YES further upward.
NO lots unlock profitable sale value only after another user moves NO further upward.
```

That stricter rule may be more philosophically precise, but it is more complex and should not be the first implementation unless tests show the simpler same-market external-entry rule is insufficient.
