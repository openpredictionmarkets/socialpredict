# Stateless Unlocked-Lot Sellability Design

Status: draft for user review
Date: 2026-07-27
Branch: `pwdel/debug-sell-sequence-20260722`

## Context

SocialPredict keeps trading, accounting, market lifecycle, canonical Outcome semantics, and authorization truth in the backend. The frontend is an experience boundary: it may present backend-owned values and recovery copy, but it must not calculate prediction-market math or decide whether a Sale Order is executable.

The current sell flow already improved user-facing errors for projection-inexecutable sales. That was useful, but it exposed a deeper domain issue: the current projection guard can block all sales once locked and unlocked value coexist in the same Position. We need a backend-owned sellability model that can separate value created by already-unlocked prior orders from value created by the user's latest still-locked order.

## Problem

The reported market history is:

```text
testuser02 NO  50
testuser03 NO  25
testuser02 NO -75
testuser02 NO  75
testuser03 NO  10
testuser02 NO  20
```

After this sequence, backend DBPM position math reports:

- `testuser02`: current `NO` shares/value `70`, old nominal unlocked value `35`, derived remaining unlocked executable value `17`
- `testuser03`: current `NO` shares/value `35`, nominal unlocked value `35`

Both users should have some executable unlocked value. Instead, `/v0/sell/quote` rejects even a small `NO` sale for each user.

The root cause is `validateProjectedSale`. It appends the proposed negative Sale Order row to the whole market history and requires the seller's aggregate projected DBPM Position to decrease by at least the sale value. In this sequence, appending a small sale row can make the user's full projected Position go up under DBPM normalization, even though a separate prior lot is already unlocked. The guard is therefore too broad: it protects against immediate share harvesting, but it also blocks legitimate sales from previously unlocked lots.

## Goals

- Preserve backend quote/sell as the only authority for Sale Order executability.
- Preserve the public API contract: no request DTO changes and no success response DTO changes.
- Keep unsafe rejected sales on the existing `422 INSUFFICIENT_SHARES` path.
- Keep all valuation and payout math routed through DBPM.
- Use sequence-based locking only. Do not add time-based locks.
- Avoid durable schema changes and new persisted lot tables.
- Let later same-user buys remain locked without relocking older value that was already unlocked.
- Add regression coverage for the exact reported sequence and for preventing executed sales beyond derived unlocked inventory.

## Non-Goals

- Do not change DBPM settlement economics.
- Do not let the frontend compute unlocked value, projected position, dust, sale proceeds, or executable sale limits.
- Do not add a new public endpoint solely for this fix.
- Do not add database migrations for a persisted lot ledger.
- Do not perform historical production repair in this PR.
- Do not introduce broader frontend platform work such as Redux, RTK Query, browser telemetry, CSP changes, offline/PWA behavior, or server-managed sessions.

## Chosen Approach

Replace the full-position post-sale projection invariant with a stateless derived unlocked-lot ledger.

The ledger is not persisted. It is derived on demand from the market's existing ordered bet rows inside the same transaction-scoped repository used by settlement. It treats positive buy rows as potential lots and negative sell rows as consumption of previously unlocked lots. Quote and settlement use the same derived inventory so a quote cannot say "allowed" for a sale that settlement later rejects under the same locked market state.

This keeps the public API stable and keeps DBPM as the source of valuation. The new layer only answers: "How much of this user's current DBPM-valued position is currently executable under the sequence-based unlock rule?"

## Domain Terms

- **Buy lot**: a positive bet row for a Participant, Market, and Outcome.
- **Sale row**: a negative bet row. Today this records sold shares as `Amount = -sharesSold`, not the historical credited sale value.
- **Unlocked lot**: a buy lot that has a later positive buy row from another Participant in the same Market. The later buy can be for either Outcome. A Participant's own later buy does not unlock that same Participant's previous lot.
- **Locked lot**: a buy lot with no later positive buy from another Participant.
- **Remaining unlocked inventory**: unlocked buy-lot shares minus prior sale-row consumption, evaluated in chronological order.
- **Executable sale value**: the current DBPM value represented by remaining unlocked inventory, rounded with the existing sell calculator semantics and capped by the user's current DBPM Position Value.

## Pre-Implementation Code References

These are the concrete seams this design changes. Line numbers refer to the branch state when the design was written, before the implementation patch replaced the projection guard.

- [backend/internal/domain/math/positions/sellable.go](../../../backend/internal/domain/math/positions/sellable.go): `CalculateUnlockedSellablePosition_WPAM_DBPM` currently sums every unlocked positive buy payout at lines 58-64, then only caps by current shares at lines 65-81. It does not replay negative sale rows as lot consumption.
- [backend/internal/domain/math/positions/sellable.go](../../../backend/internal/domain/math/positions/sellable.go): `isUnlockedBuy` at lines 99-115 already encodes the market-sequence unlock trigger: a matching user/outcome buy is unlocked when there is any later positive row from another user.
- [backend/internal/domain/bets/bet_sell.go](../../../backend/internal/domain/bets/bet_sell.go): `QuoteSell` loads current and sellable positions at lines 47-55, calculates a quote, then calls `validateProjectedSale` at lines 64-68.
- [backend/internal/domain/bets/bet_sell.go](../../../backend/internal/domain/bets/bet_sell.go): `sellInTransaction` repeats the same path inside the sell transaction at lines 82-100 before `CreditSale` mutates ledger/user balance at line 103.
- [backend/internal/domain/bets/bet_sell.go](../../../backend/internal/domain/bets/bet_sell.go): `validateSaleProjection` at lines 189-229 is the too-broad invariant. It requires the whole DBPM-projected Position to lose enough shares/value after appending the sale row.
- [backend/internal/domain/bets/bet_sell.go](../../../backend/internal/domain/bets/bet_sell.go): `saleCalculator.Quote` at lines 250-288 defines the sell quote math we should keep.
- [backend/internal/domain/bets/service.go](../../../backend/internal/domain/bets/service.go): `PositionReader` and `PositionProjector` are internal service seams at lines 49-58. We should keep quote/sell on internal backend seams and avoid public API changes.
- [backend/internal/repository/bets/sell_transaction.go](../../../backend/internal/repository/bets/sell_transaction.go): settlement uses `SellBetTransaction` at lines 20-26 and transaction-scoped market/bet reads. `loadMarketData` orders rows by `placed_at ASC` and locks rows on Postgres at lines 129-162.
- [backend/internal/domain/bets/models.go](../../../backend/internal/domain/bets/models.go): `SellRequest`, `SellResult`, and `SellQuoteResult` at lines 79-140 are the public domain DTO shape that should not change for this fix.
- [backend/internal/domain/bets/errors.go](../../../backend/internal/domain/bets/errors.go): existing user-facing copy and requester-only error detail fields are at lines 55-75.

## Mathematical Model

Let a market history be ordered by the existing repository sort:

```text
B = [b_1, b_2, ..., b_n], ordered by placed_at ASC
```

For each row:

```text
b_i = (u_i, o_i, a_i, t_i)
u_i = username
o_i = outcome in {YES, NO}
a_i = amount; a_i > 0 means buy credits, a_i < 0 means sold shares
t_i = placed time
```

For the target Participant and Outcome:

```text
u = requested username
x = requested outcome
```

DBPM remains the valuation source. The existing helper calculates per-row payout shares:

```text
p_i = DBPM_PAYOUT_i(B)
```

where `p_i` comes from `CalculateBetPayouts_WPAM_DBPM`. Only positive matching target rows create sellability lots:

```text
lot_i exists iff u_i = u and o_i = x and a_i > 0
lot_i.initialShares = max(p_i, 0)
lot_i.remainingShares starts as lot_i.initialShares
lot_i is locked while its index is >= unlocked_end
```

The sequence unlock indicator for a target buy row is:

```text
unlocked_i(k) = exists j such that i < j <= k, a_j > 0, and u_j != u
```

This intentionally does not require `o_j = x`. The current code's `isUnlockedBuy` already unlocks from any later positive row by another user, regardless of Outcome.

Historical sale rows are replayed as consumption, not as new negative lots:

```text
sale_k exists iff u_k = u and o_k = x and a_k < 0
sale_k.sharesToConsume = abs(a_k)
```

At each row `k`, process events in order. The efficient implementation should not mark every existing lot on every outside buy. Instead, keep:

```text
lots = ordered target-user target-outcome buy lots
unlocked_end = count of lots currently unlocked
consume_head = first unlocked lot that may still have remaining shares
```

Then replay rows with a moving unlock boundary:

```text
if a_k > 0 and u_k = u and o_k = x:
    append lot_k

if a_k > 0 and u_k != u:
    unlocked_end = len(lots)

if a_k < 0 and u_k = u and o_k = x:
    q = abs(a_k)
    while q > 0 and consume_head < unlocked_end:
        c = min(lots[consume_head].remainingShares, q)
        lots[consume_head].remainingShares -= c
        q -= c
        if lots[consume_head].remainingShares == 0:
            consume_head += 1
```

For historical replay only, if `q` remains positive after unlocked inventory is exhausted, discard the leftover consumption and continue. This avoids making a future corrective deploy reinterpret old questionable data as a reason to block newly unlocked later value. For new live sells, the service must reject before mutation whenever the sale calculator would execute more shares or value than the remaining unlocked inventory.

After replay:

```text
U_shares_raw = sum(lots[i].remainingShares for consume_head <= i < unlocked_end)
```

Calculate the current DBPM Position with the existing helper:

```text
C = CalculateMarketPositionForUser_WPAM_DBPM(B, u)
C_shares = sharesForOutcome(C, x)
C_value = C.Value
```

Then cap derived unlocked inventory by the live current Position:

```text
U_shares = min(U_shares_raw, C_shares)
```

Convert shares to executable value using the same integer style as `saleCalculator.Quote`:

```text
value_per_share = floor(C_value / C_shares)
U_value = min(C_value, U_shares * value_per_share)
```

If `C_shares <= 0`, `C_value <= 0`, `U_shares <= 0`, or `value_per_share <= 0`, the sellable Position is zero and the existing `NO_SELLABLE_SHARES`/`NO_POSITION` behavior applies.

Given a requested sale amount:

```text
R = requested credits
```

the existing sale calculator should remain the authority for quote fields:

```text
S = min(U_shares, floor(R / value_per_share))
sale_value = S * value_per_share
raw_dust = max(R - sale_value, 0)
dust = normalizeDust(raw_dust, MaxDustPerSale)
requestedCreditsReturned = sale_value + dust
netProceeds = sale_value - dust
```

These formulas are already implemented in [backend/internal/domain/bets/bet_sell.go](../../../backend/internal/domain/bets/bet_sell.go) lines 250-288 and should not be redefined in a second calculator.

The existing quote contract allows a request above the exact executable value to round down to the available share cap when dust rules allow it. That behavior should remain. The invariant is about execution, not raw text input:

```text
S <= U_shares
sale_value <= U_value
```

If the request cannot be normalized under those constraints and the dust cap, the existing insufficient-shares path applies.

## Complexity

Let:

```text
n = number of bet rows in the market
L = number of target Participant buy lots for the requested Outcome
S = number of target Participant sale rows for the requested Outcome
```

The naive version that loops over every existing lot on every outside buy is:

```text
O(n * L) worst case, which is O(n^2) when L grows with n
```

The boundary-and-head replay above is:

```text
O(n + L + S), which is O(n)
```

Reason:

- each market row is visited once;
- each target buy lot is appended once;
- `unlocked_end = len(lots)` unlocks existing lots by moving an integer boundary, not by walking the lots;
- sale consumption advances `consume_head`; across the whole replay, each fully consumed lot is advanced past once, and each sale row can touch at most one partially consumed head lot.

Space is:

```text
O(L), bounded by O(n)
```

This is asymptotically optimal for an exact stateless implementation because DBPM and sequence-based sellability both depend on the ordered market history. Reducing below `O(n)` would require durable derived state or a materialized lot ledger, which is intentionally out of scope for this PR.

## Worked Sequence

For the reported sequence:

```text
1. A = testuser02, NO,  +50
2. B = testuser03, NO,  +25
3. A = testuser02, NO,  -75
4. A = testuser02, NO,  +75
5. B = testuser03, NO,  +10
6. A = testuser02, NO,  +20
```

For `testuser02`:

- Row 1 creates an A/NO lot.
- Row 2 is a positive row by another Participant, so it unlocks the row 1 lot.
- Row 3 consumes unlocked A/NO shares from that unlocked row 1 lot.
- Row 4 creates a new A/NO lot, initially locked.
- Row 5 is a positive row by another Participant, so it unlocks the row 4 lot.
- Row 6 creates a new A/NO lot, initially locked. It does not relock rows 1 or 4.
- Current DBPM Position is `NO=70`, `Value=70`.
- Remaining unlocked executable value is `17`, so a small `NO` quote should succeed.

For `testuser03`:

- Row 2 creates a B/NO lot, initially locked.
- Row 3 is a sale by A, not an unlock event.
- Row 4 is a positive row by another Participant, so it unlocks B's row 2 lot.
- Row 5 creates a new B/NO lot, initially locked.
- Row 6 is a positive row by another Participant, so it unlocks B's row 5 lot.
- Current DBPM Position is `NO=35`, `Value=35`.
- Remaining unlocked executable value is `35`, so a small `NO` quote should succeed.

The existing full-projection invariant fails this sequence because appending a new sale row can increase aggregate projected DBPM Position even when the sale is consuming a previously unlocked lot. The derived ledger avoids that false negative by checking the sellable sub-inventory instead of the whole Position.

## Sellability Algorithm

The backend should derive sellability from the current market history as follows:

1. Load the Market snapshot and ordered bet rows using the existing repository path. Settlement must use the transaction-bound repository so concurrent sells see locked market/bet state.
2. Calculate current user Position through existing DBPM position math.
3. Calculate DBPM payouts for source bet rows through `CalculateBetPayouts_WPAM_DBPM` or an equivalent internal helper. These payouts attach current DBPM-valued shares to each positive buy row.
4. Walk all market rows chronologically while tracking lots for the requested Participant and Outcome.
5. Add positive buy rows as lots with their DBPM payout shares.
6. When a later positive buy from another Participant appears in the same Market, mark earlier matching lots as unlocked. That outside order unlocks older value; it does not unlock the buyer's own latest order.
7. When a negative sale row for the requested Participant and Outcome appears, consume `abs(Amount)` shares from lots before the `unlocked_end` boundary. Use FIFO order so the oldest unlocked value is consumed before newer unlocked value.
8. Do not let a historical over-consumption consume future lots that were not unlocked when the sale occurred. Since this PR is not a historical repair, saturate the then-available inventory at zero and keep replaying later rows normally.
9. After replay, sum remaining unlocked shares for the requested Outcome.
10. Cap remaining unlocked shares by the user's current DBPM shares for that Outcome.
11. Convert remaining unlocked shares to executable value from the current DBPM Position. Use existing integer credit rounding semantics and cap by current Position Value.
12. Feed that derived sellable Position into the existing sale calculator for quote and settlement.

The key behavior change is step 7. Prior sells reduce future sellability, but a later own buy does not erase older unlocked inventory. The latest own buy can stay locked while older unlocked lots remain executable.

## Patch-Level Implementation Sketch

This is not intended to be applied verbatim. It is the expected code shape so review can focus on exactly what changes.

### 1. Rewrite sellability math in `sellable.go`

Current behavior:

```diff
 func CalculateUnlockedSellablePosition_WPAM_DBPM(...) (UserMarketPosition, error) {
     current := CalculateMarketPositionForUser_WPAM_DBPM(...)
     payouts := CalculateBetPayouts_WPAM_DBPM(snapshot, bets)
-    unlockedShares := int64(0)
-    for i, payout := range payouts {
-        if isUnlockedBuy(payouts, i, username, outcome) && payout.Payout > 0 {
-            unlockedShares += payout.Payout
-        }
-    }
+    unlockedShares := deriveRemainingUnlockedShares(payouts, username, outcome)
     unlockedShares = min(unlockedShares, currentShares)
     sellableValue := unlockedShares * (current.Value / currentShares)
 }
```

New helper shape:

```go
type unlockedLot struct {
    remainingShares int64
}

func deriveRemainingUnlockedShares(payouts []BetPayout, username string, outcome string) int64 {
    lots := make([]unlockedLot, 0)
    unlockedEnd := 0
    consumeHead := 0

    for _, payout := range payouts {
        bet := payout.Bet

        if bet.Amount > 0 && bet.Username == username && bet.Outcome == outcome && payout.Payout > 0 {
            lots = append(lots, unlockedLot{remainingShares: payout.Payout})
        }

        if bet.Amount > 0 && bet.Username != username {
            unlockedEnd = len(lots)
        }

        if bet.Amount < 0 && bet.Username == username && bet.Outcome == outcome {
            consumeHead = consumeUnlockedLots(lots, consumeHead, unlockedEnd, -bet.Amount)
        }
    }
    return sumRemainingUnlocked(lots, consumeHead, unlockedEnd)
}
```

This replaces the look-ahead-only `isUnlockedBuy` model with an event replay model. Delete `isUnlockedBuy` if no remaining test needs it directly; tests should primarily target `CalculateUnlockedSellablePosition_WPAM_DBPM`.

### 2. Replace the full-position projection guard in `bet_sell.go`

Current quote path:

```diff
 sale, err := s.saleCalculator.Quote(sellablePosition, sellableShares, req.Amount)
 if err != nil {
     return nil, err
 }
 ...
-bet := req.NewSaleBet(outcome, sale.SharesToSell, s.clock.Now())
-if err := validateProjectedSale(ctx, s.markets, req, outcome, currentPosition, sellablePosition, sale, *bet); err != nil {
+if err := validateSaleWithinSellableInventory(req, currentPosition, sellablePosition, sellableShares, sale, outcome); err != nil {
     return nil, err
 }
```

Current settlement path changes the same way before `CreditSale`:

```diff
 bet := req.NewSaleBet(outcome, sale.SharesToSell, now)
-if err := validateProjectedSale(txCtx, markets, req, outcome, currentPosition, sellablePosition, sale, *bet); err != nil {
+if err := validateSaleWithinSellableInventory(req, currentPosition, sellablePosition, sellableShares, sale, outcome); err != nil {
     return err
 }
 if err := (betLedger{repo: repo, users: users}).CreditSale(txCtx, bet, netSaleProceeds(sale)); err != nil {
     return err
 }
```

New invariant shape:

```go
func validateSaleWithinSellableInventory(
    req SellRequest,
    current *dmarkets.UserPosition,
    sellable *dmarkets.UserPosition,
    sellableShares int64,
    sale SaleQuote,
    outcome string,
) error {
    if current == nil {
        return ErrNoPosition
    }
    if sellable == nil || sellableShares <= 0 || sellable.Value <= 0 {
        return ErrNoSellableShares
    }
    if sale.SharesToSell <= 0 {
        return ErrInsufficientShares
    }
    if sale.SharesToSell > sellableShares || sale.SaleValue > sellable.Value {
        return newSaleProjectionNotExecutableError(req, current, sellable, sellable, outcome)
    }
    return nil
}
```

The name `newSaleProjectionNotExecutableError` can be renamed later, but it should keep matching `ErrInsufficientShares` and keep the same public `422 INSUFFICIENT_SHARES` mapping.

### 3. Leave the projector seam in place for this PR

Once `validateProjectedSale` is no longer called by quote/sell, `PositionProjector` in [backend/internal/domain/bets/service.go](../../../backend/internal/domain/bets/service.go) lines 55-65 is no longer needed by this sell path.

For this PR, prefer the smaller behavioral change:

- Stop calling `ProjectUserPositionAfterBet` from `QuoteSell` and `sellInTransaction`.
- Delete `validateProjectedSale` and `validateSaleProjection` from [backend/internal/domain/bets/bet_sell.go](../../../backend/internal/domain/bets/bet_sell.go) if that removes the only production caller.
- Leave `ProjectUserPositionAfterBet` on the market service and transaction repository for now. It is an internal projection utility and removing it is cleanup, not required for the sellability fix.

This preserves the public API and keeps the PR focused. The important behavior is that settlement still uses `SellBetTransaction` and transaction-scoped market/bet reads in [backend/internal/repository/bets/sell_transaction.go](../../../backend/internal/repository/bets/sell_transaction.go) lines 20-26 and 129-162.

## Projection Guard Replacement

`validateProjectedSale` currently asks whether the whole projected Position decreases enough after appending the Sale Order. That question is too broad for mixed locked/unlocked Positions.

The replacement invariant should ask narrower questions:

- Does the executed sale consume no more shares than the remaining unlocked inventory?
- Does the executed sale value stay within the executable value represented by that inventory?
- Does the sale calculator still produce a valid `sharesSold`, `saleValue`, `dust`, and `netProceeds` under the current dust rules?

If any answer fails, return the existing insufficient-shares path:

```text
422 INSUFFICIENT_SHARES
```

Rejected sales must not create a sell ledger row, credit the user balance, increase market dust, or update market volume.

The current projection error details can remain useful for the UI, but the field named `executableSaleValue` should be populated from the derived unlocked inventory. It should no longer default to `0` merely because the aggregate full-position projection fails.

## DBPM Preservation

This design does not replace DBPM or move market math to the frontend.

DBPM still calculates:

- current user Position,
- source-row payouts used to size buy lots,
- current value per Outcome share,
- market accounting and conservation behavior.

The unlocked-lot ledger only filters which DBPM-valued shares are eligible to be sold under platform rules. The sell calculator still determines `sharesSold`, `saleValue`, `dust`, and `netProceeds`. Dust remains retained by the market; users receive only `netProceeds`.

## Expected Behavior For Reported Sequence

For the sequence above:

- `testuser02` has current `NO` Position Value `70` and remaining unlocked executable value `17`.
- `testuser03` has current `NO` Position Value `35` and remaining unlocked executable value `35`.
- A small valid `/v0/sell/quote` for `NO` should succeed for both users.
- A matching `/v0/sell` should succeed inside the transaction and consume unlocked inventory.
- A request that cannot be rounded or normalized without exceeding remaining unlocked executable value should return `422 INSUFFICIENT_SHARES`.
- A later same-user buy can remain locked, but it must not relock prior value that had already been unlocked by another Participant's later buy.

## API And UI Contract

No request DTO changes are planned for `/v0/sell/quote` or `/v0/sell`.

The frontend should continue to treat quote/sell responses and backend error details as the authority. It may display:

- `Position Value`: aggregate current DBPM Position Value.
- `Nominal Unlocked Value`: backend-computed value before final sale-calculator constraints.
- `Currently Executable Sale Value`: backend-computed executable value from remaining unlocked inventory.

The frontend must not derive these values itself.

When remaining executable value is zero while Position Value still exists, the primary backend message should remain:

```text
This value is not sellable yet. Wait for more market activity, then try again.
```

The More Info copy should remain:

```text
Your position still has value but is not sellable yet. This protects the market from users immediately cashing out value created by their own order. More market activity can unlock additional sellable value.
```

## Test Plan

Backend domain tests should cover:

- A single buy has no sellable value until another Participant places a later positive buy.
- A later buy from another Participant unlocks the earlier Participant's prior lot.
- A later buy by the same Participant does not unlock that Participant's own previous lot.
- A later same-user buy does not relock older inventory that was already unlocked.
- Prior sale rows consume unlocked inventory.
- Repeated sells cannot exceed remaining unlocked inventory.
- The exact reported sequence leaves both `testuser02` and `testuser03` with executable unlocked `NO` value.

Service tests should cover:

- Valid partial sell with dust succeeds and credits only `netProceeds`.
- Full unlocked-position sell succeeds only when it consumes available unlocked inventory.
- A sale calculator result that exceeds derived unlocked inventory is rejected before ledger, balance, dust, or volume mutation.
- The exact reported sequence can quote and execute a small valid sale for both users.
- The attachment over-request sequence caps safely and does not recreate a tiny-share large-proceeds cashout.

Repository/transaction tests should cover:

- Quote reads and settlement reads derive sellability from the same helper.
- Settlement derives sellability inside the existing transaction-scoped sell repository.
- Concurrent or stale quote conditions are handled by settlement revalidation before mutation.

Integration tests should update the final desired regression under `integrationtest/`:

- Replay the exact reported mixed buy/sell sequence.
- Assert both `/v0/sell/quote` and `/v0/sell` succeed for a valid unlocked sale.
- Assert `/v0/sell/quote` and `/v0/sell` never execute more shares or value than the remaining unlocked inventory.
- Assert `/v0/sell/quote` and `/v0/sell` cap oversized requests when the sell calculator can normalize them under remaining inventory and dust rules.
- Assert impossible or inconsistent sale executions do not increase user balance, market dust, market volume, or sale ledger rows.
- Assert successful dust-generating sells retain dust in market accounting and credit only net proceeds.

Use separate markets or reset state within the integration scenario when a happy-path sale would otherwise alter the sad-path assertions.

## Validation Commands

Run focused backend tests:

```bash
go test ./internal/domain/bets ./internal/domain/math/positions ./internal/repository/bets ./handlers/bets/selling
```

Run broader backend tests when practical:

```bash
JWT_SIGNING_KEY=test-secret-key-for-testing go test ./...
```

Run integration coverage against a local seeded dev stack:

```bash
node integrationtest/scripts/sell-shares-overcashout.mjs --base-url http://localhost:8080 --api-prefix /v0
```

If the existing integration script becomes too broad, add a separate executable script and case file for this sequence-based unlocked-lot regression.

## Acceptance Criteria

- Backend quote/sell remain the only authorities for Sale Order executability.
- No frontend code computes domain sellability or DBPM math.
- No public request DTOs or success response DTOs change.
- Unsafe sales continue to return `422 INSUFFICIENT_SHARES`.
- The reported sequence leaves previously unlocked value executable for both affected Participants.
- A Participant's own latest buy can stay locked without relocking older unlocked lots.
- Prior sales reduce remaining unlocked inventory.
- Rejected sales do not mutate ledger, user balance, dust, or volume.
- Market dust behavior remains covered for both successful and rejected sells.
