---
title: External Market Movement Sale Lock Design
document_type: feature-design
domain: features
author: Patrick Delaney
updated_at: 2026-06-25T00:00:00Z
updated_at_display: "Thursday, June 25, 2026"
update_reason: "Define the transaction-time sale seasoning rule and design boundaries for external market movement locks."
status: in_progress
---

# External Market Movement Sale Lock Design

## Design Posture

The sale lock is a transaction-time economic invariant:

```text
A user cannot extract appreciation created only by that user's own bet history.
```

The design should be conservative, auditable, and stateless where possible. It should derive sell eligibility from canonical bet history rather than adding mutable per-lot state unless profiling later proves that replay cost is too high.

## Boundary

This feature belongs to the betting/sale transaction path.

It must be enforced in backend domain/application code, not only in the frontend. The frontend may preview locked versus sellable shares, but the backend remains authoritative.

## Dependency Direction

Preferred dependency direction:

```text
handler -> bet sale service -> sale lock evaluator -> repository interface -> GORM/Postgres adapter
```

The sale lock evaluator should receive ordered market bet events or a purpose-specific projection. It should not know about GORM.

## Canonical Input

The lock evaluator needs canonical chronological buy/sell event context for one market:

```text
market_id
bet_id
username
outcome
amount
placed_at
```

The authoritative order should match the existing market replay order:

```text
placed_at ASC, id ASC
```

## Policy

A buy lot is sellable only if it has been seasoned by a qualifying later bet.

A qualifying later bet is:

```text
same market
AND amount > 0
AND later in canonical order
AND username != original_lot.username
```

Non-qualifying rows:

- same-user buys;
- sell rows;
- refunds;
- zero rows;
- administrative accounting rows;
- rows from a different child market in a grouped market.

## Stateless Derivation

The first implementation should derive lot seasoning from bet history during quote/sale execution.

High-level derivation:

```text
1. Replay market events in canonical order.
2. Track positive buy lots by user/outcome.
3. When a positive buy from user B appears, season prior unseasoned lots on that market owned by users other than B.
4. During sale calculation, expose only the seller's seasoned shares as sellable.
5. Reject or cap sale orders that exceed sellable shares.
```

This keeps the rule auditable and avoids adding new state that can drift from canonical history.

## Sale Behavior

Implementation options:

| Option | Behavior | First-version recommendation |
| --- | --- | --- |
| Reject oversize sale | If requested sale exceeds sellable shares, return a lock error. | Yes, simplest and clearest. |
| Auto-cap sale | Execute only the sellable portion and leave locked shares. | No, surprising unless UI explicitly opts in. |
| Sell at basis only | Allow unseasoned shares to be sold only without appreciation. | No, adds another math path. |

The initial behavior should reject sales that require locked shares.

The first implementation rejects the full sale when the seller's latest same-outcome positive buy is still unseasoned. This prevents the self-top-up exploit without introducing lot-selection UI or partial-sale auto-capping.

## Quote Behavior

Sale quotes should include enough information to explain the result:

```text
total_shares
sellable_shares
locked_shares
requested_sale_amount
max_sellable_amount
lock_reason
```

If the requested sale is blocked, the quote should say that another user must place a later bet on the same market before those shares can be sold.

## Grouped Markets

Multiple-choice binary markets are groups of independent binary child markets.

The sale lock applies to the child market where the bet was placed. A user buying answer A does not season another user's answer B lots unless those answers share the same child market, which they do not.

## Design-Plan Alignment

This design follows the SocialPredict design posture:

| Principle | Application |
| --- | --- |
| Domain-first language | Use explicit terms such as lot, seasoned, unseasoned, qualifying later bet, and sellable shares. |
| Clean architecture | Keep lock policy in sale/domain code; keep GORM and SQL as adapter details. |
| Transaction safety | Enforce the rule during backend quote/sale execution using canonical fresh history. |
| Evolutionary design | Start with stateless replay and targeted tests before adding durable per-lot state or caches. |
| Auditability | Make every unlock decision derivable from the market bet stream. |

## Risks

| Risk | Mitigation |
| --- | --- |
| Existing users expect all shares to be sellable. | Surface locked/sellable shares in quote UI and return a clear backend error. |
| Sale math becomes harder to reason about. | Keep lock eligibility separate from WPAM/DBPM valuation. |
| Performance cost on large hot markets. | Use market-scoped chronological projections and indexes from Feature 22. |
| Partial frontend-only enforcement. | Add backend tests proving the exploit fails even without frontend help. |
| Ambiguous external movement policy. | First version uses any later positive buy by another user on the same market. Record stricter favorable-movement policy as future work. |

## Future Work

A later feature can evaluate whether profitable sale unlocks should require favorable external price movement rather than any external same-market buy. That would better match the philosophy that profit should require another participant moving the market further in the seller's favor, but it introduces additional edge cases around loss-taking exits and probability history.
