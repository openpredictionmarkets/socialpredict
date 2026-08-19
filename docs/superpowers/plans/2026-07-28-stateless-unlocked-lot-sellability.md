# Stateless Unlocked-Lot Sellability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `$subagent-driven-development` (recommended, if installed) or `$executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make backend sell quote and settlement execute against stateless sequence-derived unlocked inventory so older unlocked value remains sellable after later same-user buys.

**Architecture:** Keep DBPM as the source of current position and per-row payouts. Replace the current look-ahead sellable-share summation with an `O(n)` lot replay using `unlockedEnd` and `consumeHead`, then replace the full-position post-sale projection guard with a narrower check against derived sellable inventory. Quote and settlement continue to use the same backend repository seams; settlement remains inside `SellBetTransaction`.

**Tech Stack:** Go backend domain/repository tests, existing Node HTTP integration runner, no schema migration, no frontend math.

## Global Constraints

- Preserve backend quote/sell as the only authority for Sale Order executability.
- Preserve the public API contract: no request DTO changes and no success response DTO changes.
- Keep unsafe rejected sales on the existing `422 INSUFFICIENT_SHARES` path.
- Keep all valuation and payout math routed through DBPM.
- Use sequence-based locking only. Do not add time-based locks.
- Avoid durable schema changes and new persisted lot tables.
- Let later same-user buys remain locked without relocking older value that was already unlocked.
- Rejected sales must not create a sell ledger row, credit the user balance, increase market dust, or update market volume.
- The derived lot replay must be `O(n)` over market rows, not `O(n^2)`.

---

### Task 1: Add Red Regression Coverage

**Files:**
- Modify: `backend/internal/domain/math/positions/positionsmath_test.go`
- Modify: `backend/internal/domain/bets/service_test.go`

**Interfaces:**
- Consumes: `CalculateUnlockedSellablePosition_WPAM_DBPM(snapshot MarketSnapshot, bets []boundary.Bet, username string, outcome string) (UserMarketPosition, error)`
- Produces: failing tests proving the expected sequence behavior before production code changes.

- [ ] **Step 1: Add math regression cases**

Add cases to `TestCalculateUnlockedSellablePosition_WPAM_DBPM`:

```go
{
    name: "prior sale consumes unlocked lot before later own buy",
    bets: []struct {
        Amount   int64
        Outcome  string
        Username string
        Offset   time.Duration
    }{
        {Amount: 50, Outcome: "NO", Username: "alice", Offset: 0},
        {Amount: 25, Outcome: "NO", Username: "bob", Offset: time.Minute},
        {Amount: -75, Outcome: "NO", Username: "alice", Offset: 2 * time.Minute},
        {Amount: 75, Outcome: "NO", Username: "alice", Offset: 3 * time.Minute},
        {Amount: 10, Outcome: "NO", Username: "bob", Offset: 4 * time.Minute},
        {Amount: 20, Outcome: "NO", Username: "alice", Offset: 5 * time.Minute},
    },
    username: "alice",
    outcome:  "NO",
    wantNo:   17,
    wantVal:  17,
},
{
    name: "counterparty unlocked lots remain sellable in mixed sequence",
    bets: []struct {
        Amount   int64
        Outcome  string
        Username string
        Offset   time.Duration
    }{
        {Amount: 50, Outcome: "NO", Username: "alice", Offset: 0},
        {Amount: 25, Outcome: "NO", Username: "bob", Offset: time.Minute},
        {Amount: -75, Outcome: "NO", Username: "alice", Offset: 2 * time.Minute},
        {Amount: 75, Outcome: "NO", Username: "alice", Offset: 3 * time.Minute},
        {Amount: 10, Outcome: "NO", Username: "bob", Offset: 4 * time.Minute},
        {Amount: 20, Outcome: "NO", Username: "alice", Offset: 5 * time.Minute},
    },
    username: "bob",
    outcome:  "NO",
    wantNo:   35,
    wantVal:  35,
},
```

Keep the existing service-level red test `TestServiceQuoteSell_SequenceBasedUnlockedValueRemainsExecutableAfterLaterOwnBuy`.

- [ ] **Step 2: Run red tests**

Run:

```bash
cd backend
go test ./internal/domain/math/positions -run TestCalculateUnlockedSellablePosition_WPAM_DBPM -count=1
go test ./internal/domain/bets -run TestServiceQuoteSell_SequenceBasedUnlockedValueRemainsExecutableAfterLaterOwnBuy -count=1
```

Expected: at least one test fails because current sellable math does not consume prior sale rows and the service still rejects through `validateProjectedSale`.

---

### Task 2: Implement `O(n)` Derived Unlocked Inventory

**Files:**
- Modify: `backend/internal/domain/math/positions/sellable.go`
- Test: `backend/internal/domain/math/positions/positionsmath_test.go`

**Interfaces:**
- Produces: `deriveRemainingUnlockedShares(payouts []BetPayout, username string, outcome string) int64`
- Consumes: existing `CalculateBetPayouts_WPAM_DBPM` and `CalculateMarketPositionForUser_WPAM_DBPM`

- [ ] **Step 1: Replace look-ahead summation**

In `CalculateUnlockedSellablePosition_WPAM_DBPM`, replace:

```go
unlockedShares := int64(0)
for i, payout := range payouts {
    if isUnlockedBuy(payouts, i, username, outcome) && payout.Payout > 0 {
        unlockedShares += payout.Payout
    }
}
```

with:

```go
unlockedShares := deriveRemainingUnlockedShares(payouts, username, outcome)
```

- [ ] **Step 2: Add the linear replay helper**

Add this helper shape in `sellable.go`:

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

    return sumRemainingUnlockedLots(lots, consumeHead, unlockedEnd)
}
```

Add `consumeUnlockedLots` and `sumRemainingUnlockedLots` with bounds checks:

```go
func consumeUnlockedLots(lots []unlockedLot, consumeHead int, unlockedEnd int, shares int64) int {
    if unlockedEnd > len(lots) {
        unlockedEnd = len(lots)
    }
    for shares > 0 && consumeHead < unlockedEnd {
        consumed := lots[consumeHead].remainingShares
        if consumed > shares {
            consumed = shares
        }
        lots[consumeHead].remainingShares -= consumed
        shares -= consumed
        if lots[consumeHead].remainingShares == 0 {
            consumeHead++
        }
    }
    return consumeHead
}

func sumRemainingUnlockedLots(lots []unlockedLot, consumeHead int, unlockedEnd int) int64 {
    if consumeHead < 0 {
        consumeHead = 0
    }
    if unlockedEnd > len(lots) {
        unlockedEnd = len(lots)
    }
    total := int64(0)
    for i := consumeHead; i < unlockedEnd; i++ {
        total += lots[i].remainingShares
    }
    return total
}
```

- [ ] **Step 3: Run math tests**

Run:

```bash
cd backend
go test ./internal/domain/math/positions -run TestCalculateUnlockedSellablePosition_WPAM_DBPM -count=1
```

Expected: math tests pass.

---

### Task 3: Replace Full-Position Projection Guard

**Files:**
- Modify: `backend/internal/domain/bets/bet_sell.go`
- Modify: `backend/internal/domain/bets/service_test.go`
- Test: `backend/internal/domain/bets/service_test.go`

**Interfaces:**
- Consumes: derived sellable position returned by `GetUserSellablePositionInMarket`
- Produces: `validateSaleWithinSellableInventory(req SellRequest, current *dmarkets.UserPosition, sellable *dmarkets.UserPosition, sellableShares int64, sale SaleQuote, outcome string) error`

- [ ] **Step 1: Replace guard call in quote**

In `QuoteSell`, keep dust behavior but replace the projection call with:

```go
if allowed {
    if err := validateSaleWithinSellableInventory(req, currentPosition, sellablePosition, sellableShares, sale, outcome); err != nil {
        return nil, err
    }
}
```

- [ ] **Step 2: Replace guard call in settlement**

In `sellInTransaction`, keep sale bet creation before ledger write but validate inventory before crediting:

```go
now := s.clock.Now()
bet := req.NewSaleBet(outcome, sale.SharesToSell, now)
if err := validateSaleWithinSellableInventory(req, currentPosition, sellablePosition, sellableShares, sale, outcome); err != nil {
    return err
}
if err := (betLedger{repo: repo, users: users}).CreditSale(txCtx, bet, netSaleProceeds(sale)); err != nil {
    return err
}
```

- [ ] **Step 3: Add inventory validation helper**

Add:

```go
func validateSaleWithinSellableInventory(req SellRequest, current *dmarkets.UserPosition, sellable *dmarkets.UserPosition, sellableShares int64, sale SaleQuote, outcome string) error {
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

Remove `validateProjectedSale` and `validateSaleProjection` if unused after this change. Clean unused imports such as `socialpredict/internal/domain/boundary`.

- [ ] **Step 4: Run service tests**

Run:

```bash
cd backend
go test ./internal/domain/bets -run 'TestService(QuoteSell_SequenceBasedUnlockedValueRemainsExecutableAfterLaterOwnBuy|Sell_UsesUnlockedSellableCap|Sell_AttachmentSequenceCapsOvercashoutBeforeTinyTail|Sell_RejectsCalculatorResultBeyondSellableInventoryBeforeMutatingLedger)' -count=1
```

Expected: tests pass after replacing obsolete projection-inexecutable expectations with inventory-based assertions.

---

### Task 4: Update Integration Scenario

**Files:**
- Modify: `integrationtest/scripts/sell-shares-overcashout.mjs`
- Modify: `integrationtest/cases/sell-shares-overcashout.md`
- Modify: `integrationtest/README.md`

**Interfaces:**
- Consumes: `/v0/sell/quote`, `/v0/sell`, `/v0/userposition/{marketId}`, `/v0/markets/{marketId}`
- Produces: HTTP scenario proving the reported sequence has valid executable sell value and still blocks invalid over-execution.

- [ ] **Step 1: Change projection sequence from rejection to executable regression**

Rename the old projection sequence to `sequenceBasedUnlockedSequence` and include the final same-user buy:

```js
{ seq: 6, user: 'bettor', type: 'buy', outcome: 'NO', amount: 20 },
```

Replace `assertProjectionInexecutableRejected` with `assertSequenceBasedUnlockedSaleAllowed`:

```js
async function assertSequenceBasedUnlockedSaleAllowed({ token, username, marketId, amount }) {
  const beforeFinancial = await financial(username);
  const beforePosition = await position(token, marketId);
  const quote = (await quoteRaw(token, marketId, 'NO', amount, [200])).result;
  check(`${username} sequence quote allowed`, quote.allowed === true);
  check(`${username} sequence quote executable shares`, Number(quote.sharesSold) > 0, JSON.stringify(quote));
  const sell = (await sellRaw(token, marketId, 'NO', amount, [201])).result;
  sameInt(`${username} sequence sell shares match quote`, sell.sharesSold, quote.sharesSold);
  sameInt(`${username} sequence sell value match quote`, sell.saleValue, quote.saleValue);
  const afterFinancial = await financial(username);
  const afterPosition = await position(token, marketId);
  sameInt(`${username} sequence balance net proceeds`, Number(afterFinancial.accountBalance || 0) - Number(beforeFinancial.accountBalance || 0), sell.netProceeds);
  check(`${username} sequence position does not gain value`, positionValue(afterPosition) <= positionValue(beforePosition), `before=${positionValue(beforePosition)}, after=${positionValue(afterPosition)}`);
}
```

Use fresh markets for each sequence-based sell assertion so one successful sell does not disturb the other user's expected state. Update the attachment over-request assertion to prove the request caps safely instead of producing a tiny-share large-proceeds cashout.

- [ ] **Step 2: Update docs**

Update the case file and README wording from projection-inexecutable rejection to sequence-based unlocked-lot executable regression and capped over-cashout requests.

- [ ] **Step 3: Static script validation**

Run:

```bash
node --check integrationtest/scripts/sell-shares-overcashout.mjs
```

Expected: no syntax errors.

---

### Task 5: Final Verification, Commit, Push, PR

**Files:**
- Verify all modified files.
- Update PR description with design, math, tests, and validation.

**Interfaces:**
- Produces: pushed branch and GitHub PR targeting `main`.

- [ ] **Step 1: Focused backend verification**

Run:

```bash
cd backend
go test ./internal/domain/bets ./internal/domain/math/positions ./internal/repository/bets ./handlers/bets/selling -count=1
```

Expected: pass.

- [ ] **Step 2: Broader backend verification**

Run:

```bash
cd backend
JWT_SIGNING_KEY=test-secret-key-for-testing go test ./...
```

Expected: pass or report exact pre-existing failures if any.

- [ ] **Step 3: Integration script syntax**

Run:

```bash
node --check integrationtest/scripts/sell-shares-overcashout.mjs
```

Expected: pass. Run the HTTP script against a local stack if available.

- [ ] **Step 4: Commit implementation**

Commit in logical commits:

```bash
git add docs/superpowers/plans/2026-07-28-stateless-unlocked-lot-sellability.md
git commit -m "docs: plan stateless unlocked sellability"
git add backend/internal/domain/math/positions/sellable.go backend/internal/domain/math/positions/positionsmath_test.go
git commit -m "fix: derive sellable shares from unlocked lots"
git add backend/internal/domain/bets/bet_sell.go backend/internal/domain/bets/service_test.go
git commit -m "fix: validate sells against unlocked inventory"
git add integrationtest/scripts/sell-shares-overcashout.mjs integrationtest/cases/sell-shares-overcashout.md integrationtest/README.md
git commit -m "test: cover sequence-based unlocked sellability"
```

- [ ] **Step 5: Push and open PR**

Run:

```bash
git push -u origin pwdel/debug-sell-sequence-20260722
gh pr create --base main --head pwdel/debug-sell-sequence-20260722 --title "Fix sequence-based sellability for unlocked lots" --body-file .context/pr-sequence-sellability.md
```

If a PR already exists for the branch, use `gh pr edit` with the same title/body instead.
