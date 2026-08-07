# Sellable Position UI Implementation Plan

> Execute inline with `$executing-plans`. Do not use subagent-driven development.

**Goal:** Return per-outcome, server-derived sellability with the authenticated position and prevent predictable invalid sale quotes in the Sell view.

**Architecture:** A markets-domain read loads market history once, derives aggregate and YES/NO sellable positions from it, and powers the existing `/v0/userposition/{marketId}` response. The Sell view stays behind the existing trade API/client seams and constrains the selected sale amount to the returned sellable value. Quote and settlement remain authoritative for stale state.

**Tech Stack:** Go, Gorilla handlers, React, Vitest, Testing Library.

## Global Constraints

- Do not add direct browser fetches, token paths, polling, retry loops, or performance budgets.
- Treat sellability copy as a domain contract and test it.
- Reuse one market/history load for aggregate plus both sellability calculations.
- Preserve existing quote and settlement validation.
- Capture `npm run build:report` output as evidence.

## Files

- `backend/internal/domain/markets/service.go`: trade-position type and service interface method.
- `backend/internal/domain/markets/market_positions.go`: shared-history position calculation.
- `backend/internal/domain/markets/market_positions_test.go`: domain regression coverage.
- `backend/handlers/users/userpositiononmarkethandler.go`: authenticated response.
- `backend/handlers/users/userpositiononmarkethandler_test.go`: HTTP response coverage.
- `frontend/src/components/layouts/trade/SellSharesLayout.jsx`: UI constraints and labels.
- `frontend/src/components/layouts/trade/SellSharesLayout.test.jsx`: locked/partially-sellable UI coverage.

### Task 1: Add a shared trade-position read

**Files:** modify `backend/internal/domain/markets/service.go`, `backend/internal/domain/markets/market_positions.go`; test `backend/internal/domain/markets/market_positions_test.go`.

**Produces:**

```go
type TradePosition struct {
    UserPosition
    YesSellableShares, NoSellableShares int64
    YesSellableValue, NoSellableValue   int64
}
GetUserTradePositionInMarket(context.Context, int64, string) (*TradePosition, error)
```

- [ ] Write a failing test with one user YES buy followed by another user NO buy. Assert aggregate position, positive YES sellability, and zero NO sellability. Add a single-buy fully locked case with four zero sellability fields.
- [ ] Run `GOCACHE=/private/tmp/socialpredict-go-build go test ./internal/domain/markets -run TestServiceGetUserTradePositionInMarket -count=1`; expect missing method/type failure.
- [ ] Implement the method: validate inputs, call `GetByID` and `ListBetsForMarket` once, convert bets once, calculate aggregate/YES/NO positions from that shared slice, and map absent positions to zero fields.
- [ ] Re-run the focused test; expect PASS.
- [ ] Commit: `git add backend/internal/domain/markets/service.go backend/internal/domain/markets/market_positions.go backend/internal/domain/markets/market_positions_test.go && git commit -m "feat: expose per-outcome sellable positions"`.

### Task 2: Extend the authenticated position response

**Files:** modify `backend/handlers/users/userpositiononmarkethandler.go`; test `backend/handlers/users/userpositiononmarkethandler_test.go`.

**Consumes:** `GetUserTradePositionInMarket` from Task 1. **Produces:** existing response envelope plus `yesSellableShares`, `noSellableShares`, `yesSellableValue`, and `noSellableValue`.

- [ ] Extend `TestUserMarketPositionHandlerReturnsUserPosition` to decode the four new integer fields and assert positive YES/zero NO values for the existing cross-user history.
- [ ] Run `GOCACHE=/private/tmp/socialpredict-go-build go test ./handlers/users -run TestUserMarketPositionHandlerReturnsUserPosition -count=1`; expect the response assertion to fail.
- [ ] Replace the handler call to `GetUserPositionInMarket` with the new one-call trade-position read; retain its route, envelope, auth, and error mapping.
- [ ] Re-run the focused handler test; expect PASS.
- [ ] Commit: `git add backend/handlers/users/userpositiononmarkethandler.go backend/handlers/users/userpositiononmarkethandler_test.go && git commit -m "feat: return sellability with user positions"`.

### Task 3: Constrain the Sell view

**Files:** modify/test `frontend/src/components/layouts/trade/SellSharesLayout.jsx` and `frontend/src/components/layouts/trade/SellSharesLayout.test.jsx`.

**Consumes:** Task 2 fields through the existing `fetchUserShares` → `fetchUserPosition` adapter path. **Produces:** only outcomes with positive sellable value can be sold; amount is bounded by selected sellable value.

- [ ] Write tests for a fully locked owned position (unlock explanation and no enabled sale action) and a partly sellable YES position (visible sellable value and clamp from a larger amount). Unit-test normalized fields are non-negative.
- [ ] Run `npm test -- --run frontend/src/components/layouts/trade/SellSharesLayout.test.jsx`; expect failures for absent sellability UI state.
- [ ] Extend `normalizeShares` and initial state with `yesSellableShares`, `noSellableShares`, `yesSellableValue`, and `noSellableValue`. Add selected-outcome helpers; pass selected sellable value to `SaleInputAmount.max`; clamp changes; display `Sellable Value`; display the established no-sellable message; disable/omit unsellable outcome actions. Continue using `fetchUserShares`, `fetchSaleQuote`, and `submitSale`.
- [ ] Re-run the focused UI test; expect PASS.
- [ ] Commit: `git add frontend/src/components/layouts/trade/SellSharesLayout.jsx frontend/src/components/layouts/trade/SellSharesLayout.test.jsx && git commit -m "fix: guide users to sellable position value"`.

### Task 4: Verify the PR

- [ ] Run `GOCACHE=/private/tmp/socialpredict-go-build go test ./internal/domain/markets ./handlers/users ./internal/domain/bets ./internal/domain/math/positions -count=1`; expect PASS.
- [ ] Run `npm test -- --run`; expect PASS.
- [ ] Run `npm run build:report`; expect exit code 0 and record the build report in the PR without a threshold.
- [ ] Run `git diff --check && git status --short && git log --oneline main..HEAD`; expect no whitespace errors and only planned changes.
