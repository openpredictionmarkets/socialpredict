# Projection-Aware Sell Quote Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `$subagent-driven-development` (recommended, if installed) or `$executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make backend quote/sell the sole authority for executable Sale Orders, return requester-only projection details for rejected unsafe sales, and update the frontend to display backend-owned explanations without computing trading math.

**Architecture:** The backend keeps ownership of DBPM accounting and projection safety. Rejected projection-inexecutable sales keep the existing `422 INSUFFICIENT_SHARES` public contract while adding backend-computed `details` for explanation. The frontend routes trade calls through an API adapter, preserves backend error details, and only formats values returned by the backend.

**Tech Stack:** Go backend domain/HTTP handlers, existing SocialPredict API envelopes, Node HTTP integration scripts, React/Vite frontend, focused Vitest tests for adapter behavior if frontend test dependency setup is added during execution.

## Global Constraints

- Backend quote and sell simulation are the only source of truth for Sale Order executability.
- Preserve public request DTOs and success response DTOs for `/v0/sell/quote` and `/v0/sell`.
- Keep rejected unsafe sales on the existing `422 INSUFFICIENT_SHARES` path.
- Include requester-only diagnostic values in the existing error-envelope `details` object.
- The frontend must not calculate unlocked value, projected position reduction, dust, sale proceeds, or domain sellability.
- Frontend trade calls should pass through adapter seams instead of direct `fetch`, `localStorage`, or raw API-envelope handling.
- Do not change DBPM settlement economics.
- Do not add durable backend state or database migrations.
- Do not introduce Redux, RTK Query, server-managed sessions, browser telemetry, CSP work, offline behavior, or broad frontend platform changes.
- Public trading failure copy should originate from the backend when it describes domain/accounting state.

---

## File Structure

- Modify `backend/internal/domain/bets/errors.go`
  - Owns exported projection-inexecutable error type and requester-only detail struct.
- Modify `backend/internal/domain/bets/bet_sell.go`
  - Populates projection details from current, nominal sellable, and projected backend positions.
- Modify `backend/internal/domain/bets/service_test.go`
  - Adds red/green domain tests for quote/sell projection details and no mutation.
- Modify `backend/handlers/bets/selling/sellpositionhandler.go`
  - Maps the new internal error to existing `INSUFFICIENT_SHARES` failure envelope with details.
- Modify `backend/handlers/bets/selling/sellpositionhandler_test.go`
  - Adds HTTP envelope regression tests.
- Modify `integrationtest/scripts/sell-shares-overcashout.mjs`
  - Replays the two-user reported sequence and asserts backend detail fields.
- Modify `integrationtest/cases/sell-shares-overcashout.md`
  - Documents projection-inexecutable failure coverage.
- Modify `integrationtest/README.md`
  - Updates current coverage notes.
- Modify `frontend/src/api/httpClient.js`
  - Preserves `details` and raw backend payload on thrown API errors.
- Create `frontend/src/api/tradeApi.js`
  - Adapter for buy, position, quote, sell, and setup-fee calls.
- Create `frontend/src/api/tradeApi.test.jsx`
  - Focused adapter test for preserving projection details.
- Modify `frontend/package.json` and `frontend/package-lock.json` only if Vitest is not locally runnable during execution.
  - Adds a local `test` script and missing dev dependencies needed by the repo's existing frontend tests.
- Modify `frontend/src/components/layouts/trade/TradeUtils.jsx`
  - Removes direct transport logic and delegates to `tradeApi`.
- Modify `frontend/src/components/layouts/trade/BuySharesLayout.jsx`
  - Loads fee data through the adapter.
- Modify `frontend/src/components/layouts/trade/SellSharesLayout.jsx`
  - Treats Position Value as informational, removes aggregate-value sale clamp, and displays backend projection details.

---

### Task 1: Backend Domain Projection Details

**Files:**
- Modify: `backend/internal/domain/bets/errors.go`
- Modify: `backend/internal/domain/bets/bet_sell.go`
- Test: `backend/internal/domain/bets/service_test.go`

**Interfaces:**
- Consumes: existing `ErrInsufficientShares`, `SellRequest`, `SaleQuote`, `dmarkets.UserPosition`.
- Produces:
  - `const ProjectionInexecutableSaleMessage string`
  - `type SaleProjectionDetails struct`
  - `type SaleProjectionNotExecutableError struct`
  - `func (e SaleProjectionNotExecutableError) Error() string`
  - `func (e SaleProjectionNotExecutableError) Message() string`
  - `func (e SaleProjectionNotExecutableError) Unwrap() error`

- [ ] **Step 1: Write failing service tests**

Append these tests near the existing unsafe projected sale tests in `backend/internal/domain/bets/service_test.go`:

```go
func TestServiceQuoteSell_ProjectionInexecutableSaleIncludesRequesterOnlyDetails(t *testing.T) {
	now := serviceTestTime()
	current := &dmarkets.UserPosition{Username: "testuser03", MarketID: 1, NoSharesOwned: 34, Value: 34}
	sellable := &dmarkets.UserPosition{Username: "testuser03", MarketID: 1, NoSharesOwned: 17, Value: 17}
	projected := &dmarkets.UserPosition{Username: "testuser03", MarketID: 1, NoSharesOwned: 34, Value: 34}
	_, svc := newServiceFixture(
		now,
		withFixtureMaxDust(1),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(current),
		func(f *serviceFixture) {
			f.markets.positions.getUserSellablePositionInMarketFunc = func(context.Context, int64, string, string) (*dmarkets.UserPosition, error) {
				return sellable, nil
			}
		},
		withFixtureProjection(projected),
		withFixtureUser(&dusers.User{Username: "testuser03"}),
	)

	_, err := svc.QuoteSell(context.Background(), bets.SellRequest{Username: "testuser03", MarketID: 1, Amount: 17, Outcome: "NO"})
	if !errors.Is(err, bets.ErrInsufficientShares) {
		t.Fatalf("expected ErrInsufficientShares, got %v", err)
	}
	var projectionErr bets.SaleProjectionNotExecutableError
	if !errors.As(err, &projectionErr) {
		t.Fatalf("expected SaleProjectionNotExecutableError, got %T %v", err, err)
	}
	details := projectionErr.Details
	if details.Outcome != "NO" || details.RequestedCredits != 17 {
		t.Fatalf("unexpected request details: %+v", details)
	}
	if details.PositionValue != 34 || details.PositionOutcomeShares != 34 {
		t.Fatalf("unexpected current position details: %+v", details)
	}
	if details.NominalUnlockedValue != 17 || details.NominalUnlockedOutcomeShares != 17 {
		t.Fatalf("unexpected nominal unlocked details: %+v", details)
	}
	if details.ProjectedPositionValue != 34 || details.ProjectedOutcomeShares != 34 {
		t.Fatalf("unexpected projection details: %+v", details)
	}
	if details.ExecutableSaleValue != 0 {
		t.Fatalf("expected executable sale value 0, got %+v", details)
	}
}

func TestServiceSell_ProjectionInexecutableSaleIncludesDetailsAndDoesNotMutate(t *testing.T) {
	now := serviceTestTime()
	current := &dmarkets.UserPosition{Username: "testuser03", MarketID: 1, NoSharesOwned: 34, Value: 34}
	sellable := &dmarkets.UserPosition{Username: "testuser03", MarketID: 1, NoSharesOwned: 17, Value: 17}
	projected := &dmarkets.UserPosition{Username: "testuser03", MarketID: 1, NoSharesOwned: 34, Value: 34}
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(1),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(current),
		func(f *serviceFixture) {
			f.markets.positions.getUserSellablePositionInMarketFunc = func(context.Context, int64, string, string) (*dmarkets.UserPosition, error) {
				return sellable, nil
			}
		},
		withFixtureProjection(projected),
		withFixtureUser(&dusers.User{Username: "testuser03"}),
	)

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "testuser03", MarketID: 1, Amount: 17, Outcome: "NO"})
	if !errors.Is(err, bets.ErrInsufficientShares) {
		t.Fatalf("expected ErrInsufficientShares, got %v", err)
	}
	var projectionErr bets.SaleProjectionNotExecutableError
	if !errors.As(err, &projectionErr) {
		t.Fatalf("expected SaleProjectionNotExecutableError, got %T %v", err, err)
	}
	if projectionErr.Details.PositionValue != 34 || projectionErr.Details.NominalUnlockedValue != 17 || projectionErr.Details.ExecutableSaleValue != 0 {
		t.Fatalf("unexpected projection details: %+v", projectionErr.Details)
	}
	if fixture.repo.created != nil || len(fixture.users.calls) != 0 {
		t.Fatalf("projection-inexecutable sale must not mutate ledger: repo=%+v users=%+v", fixture.repo.created, fixture.users.calls)
	}
}
```

- [ ] **Step 2: Run the focused tests and verify red**

Run:

```bash
(cd backend && go test ./internal/domain/bets -run 'TestService(QuoteSell|Sell)_ProjectionInexecutableSale' -count=1)
```

Expected: FAIL because `bets.SaleProjectionNotExecutableError` does not exist.

- [ ] **Step 3: Add domain error/details types**

Add this code to `backend/internal/domain/bets/errors.go` after the existing sell message constants:

```go
const ProjectionInexecutableSaleMessage = "Position value exists, but this Sale Order is not executable right now. Market accounting requires a sale to reduce your projected position before credits can be paid out. Wait for more market activity, then try again."

type SaleProjectionDetails struct {
	Outcome                       string
	RequestedCredits              int64
	PositionValue                 int64
	PositionOutcomeShares         int64
	NominalUnlockedValue          int64
	NominalUnlockedOutcomeShares  int64
	ProjectedPositionValue        int64
	ProjectedOutcomeShares        int64
	ExecutableSaleValue           int64
}

type SaleProjectionNotExecutableError struct {
	Details SaleProjectionDetails
}

func (e SaleProjectionNotExecutableError) Error() string {
	return ProjectionInexecutableSaleMessage
}

func (e SaleProjectionNotExecutableError) Message() string {
	return ProjectionInexecutableSaleMessage
}

func (e SaleProjectionNotExecutableError) Unwrap() error {
	return ErrInsufficientShares
}
```

Run `gofmt` after adding the type. If `gofmt` aligns struct fields differently, keep the formatter output.

- [ ] **Step 4: Populate the error from projection validation**

In `backend/internal/domain/bets/bet_sell.go`, change both calls to `validateProjectedSale` so they pass `sellablePosition`:

```go
if err := validateProjectedSale(ctx, s.markets, req, outcome, currentPosition, sellablePosition, sale, *bet); err != nil {
	return nil, err
}
```

and:

```go
if err := validateProjectedSale(txCtx, markets, req, outcome, currentPosition, sellablePosition, sale, *bet); err != nil {
	return err
}
```

Update the function signatures and validation helper:

```go
func validateProjectedSale(
	ctx context.Context,
	markets PositionProjector,
	req SellRequest,
	outcome string,
	current *dmarkets.UserPosition,
	sellable *dmarkets.UserPosition,
	sale SaleQuote,
	saleBet boundary.Bet,
) error {
	if sale.SharesToSell <= 0 {
		return ErrInsufficientShares
	}
	projected, err := markets.ProjectUserPositionAfterBet(ctx, int64(req.MarketID), req.Username, saleBet)
	if err != nil {
		return err
	}
	return validateSaleProjection(req, current, sellable, projected, outcome, sale)
}

func validateSaleProjection(req SellRequest, current *dmarkets.UserPosition, sellable *dmarkets.UserPosition, projected *dmarkets.UserPosition, outcome string, sale SaleQuote) error {
	if current == nil {
		return ErrNoPosition
	}
	if projected == nil {
		projected = &dmarkets.UserPosition{}
	}

	currentSoldShares, err := sharesForOutcome(current, outcome)
	if err != nil {
		return err
	}
	projectedSoldShares, err := sharesForOutcome(projected, outcome)
	if err != nil {
		return err
	}
	if currentSoldShares-projectedSoldShares < sale.SharesToSell {
		return newSaleProjectionNotExecutableError(req, current, sellable, projected, outcome)
	}

	currentOppositeShares, err := sharesForOutcome(current, oppositeOutcome(outcome))
	if err != nil {
		return err
	}
	projectedOppositeShares, err := sharesForOutcome(projected, oppositeOutcome(outcome))
	if err != nil {
		return err
	}
	if projectedOppositeShares > currentOppositeShares {
		return newSaleProjectionNotExecutableError(req, current, sellable, projected, outcome)
	}

	maxProjectedValue := current.Value - sale.SaleValue
	if maxProjectedValue < 0 {
		maxProjectedValue = 0
	}
	if projected.Value > maxProjectedValue {
		return newSaleProjectionNotExecutableError(req, current, sellable, projected, outcome)
	}

	return nil
}
```

Add helpers near `sharesForOutcome`:

```go
func newSaleProjectionNotExecutableError(req SellRequest, current *dmarkets.UserPosition, sellable *dmarkets.UserPosition, projected *dmarkets.UserPosition, outcome string) SaleProjectionNotExecutableError {
	return SaleProjectionNotExecutableError{
		Details: SaleProjectionDetails{
			Outcome:                      outcome,
			RequestedCredits:             req.Amount,
			PositionValue:                positionValueOrZero(current),
			PositionOutcomeShares:        sharesForOutcomeOrZero(current, outcome),
			NominalUnlockedValue:         positionValueOrZero(sellable),
			NominalUnlockedOutcomeShares: sharesForOutcomeOrZero(sellable, outcome),
			ProjectedPositionValue:       positionValueOrZero(projected),
			ProjectedOutcomeShares:       sharesForOutcomeOrZero(projected, outcome),
			ExecutableSaleValue:          0,
		},
	}
}

func positionValueOrZero(pos *dmarkets.UserPosition) int64 {
	if pos == nil {
		return 0
	}
	return pos.Value
}

func sharesForOutcomeOrZero(pos *dmarkets.UserPosition, outcome string) int64 {
	shares, err := sharesForOutcome(pos, outcome)
	if err != nil {
		return 0
	}
	return shares
}
```

- [ ] **Step 5: Run red test again and verify green**

Run:

```bash
gofmt -w backend/internal/domain/bets/errors.go backend/internal/domain/bets/bet_sell.go backend/internal/domain/bets/service_test.go
(cd backend && go test ./internal/domain/bets -run 'TestService(QuoteSell|Sell)_ProjectionInexecutableSale' -count=1)
```

Expected: PASS.

- [ ] **Step 6: Run focused domain sell tests**

Run:

```bash
(cd backend && go test ./internal/domain/bets -run 'TestService(Sell|QuoteSell)' -count=1)
```

Expected: PASS.

- [ ] **Step 7: Commit Task 1**

```bash
git add backend/internal/domain/bets/errors.go backend/internal/domain/bets/bet_sell.go backend/internal/domain/bets/service_test.go
git commit -m "fix: expose projection-inexecutable sell details"
```

---

### Task 2: HTTP Handler Mapping For Projection Details

**Files:**
- Modify: `backend/handlers/bets/selling/sellpositionhandler.go`
- Test: `backend/handlers/bets/selling/sellpositionhandler_test.go`

**Interfaces:**
- Consumes: `bets.SaleProjectionNotExecutableError`, `bets.ProjectionInexecutableSaleMessage`.
- Produces: `422 INSUFFICIENT_SHARES` failure envelopes with backend message and requester-only `details`.

- [ ] **Step 1: Write failing handler tests**

Append these tests near `TestSellQuoteHandler_InsufficientSharesIncludesUserGuidance` in `backend/handlers/bets/selling/sellpositionhandler_test.go`:

```go
func TestSellQuoteHandler_ProjectionInexecutableIncludesRequesterOnlyDetails(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	svc := &fakeSellService{quoteErr: bets.SaleProjectionNotExecutableError{Details: bets.SaleProjectionDetails{
		Outcome:                      "NO",
		RequestedCredits:             17,
		PositionValue:                34,
		PositionOutcomeShares:        34,
		NominalUnlockedValue:         17,
		NominalUnlockedOutcomeShares: 17,
		ProjectedPositionValue:       34,
		ProjectedOutcomeShares:       34,
		ExecutableSaleValue:          0,
	}}}
	users := &fakeUsersService{user: &dusers.User{Username: "testuser03"}}

	body, _ := json.Marshal(dto.SellBetRequest{MarketID: 1, Amount: 17, Outcome: "NO"})
	req := httptest.NewRequest(http.MethodPost, "/v0/sell/quote", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("testuser03"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	SellQuoteHandler(svc, users).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, rr.Code)
	}
	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonInsufficientShares) {
		t.Fatalf("expected insufficient-shares reason, got %q", resp.Reason)
	}
	if resp.Message != bets.ProjectionInexecutableSaleMessage {
		t.Fatalf("expected projection message, got %q", resp.Message)
	}
	if resp.Details["outcome"] != "NO" ||
		resp.Details["requestedCredits"] != float64(17) ||
		resp.Details["positionValue"] != float64(34) ||
		resp.Details["nominalUnlockedValue"] != float64(17) ||
		resp.Details["projectedPositionValue"] != float64(34) ||
		resp.Details["executableSaleValue"] != float64(0) {
		t.Fatalf("unexpected projection details: %+v", resp.Details)
	}
}

func TestSellPositionHandler_ProjectionInexecutableIncludesRequesterOnlyDetails(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	svc := &fakeSellService{err: bets.SaleProjectionNotExecutableError{Details: bets.SaleProjectionDetails{
		Outcome:                      "NO",
		RequestedCredits:             17,
		PositionValue:                34,
		PositionOutcomeShares:        34,
		NominalUnlockedValue:         17,
		NominalUnlockedOutcomeShares: 17,
		ProjectedPositionValue:       34,
		ProjectedOutcomeShares:       34,
		ExecutableSaleValue:          0,
	}}}
	users := &fakeUsersService{user: &dusers.User{Username: "testuser03"}}

	body, _ := json.Marshal(dto.SellBetRequest{MarketID: 1, Amount: 17, Outcome: "NO"})
	req := httptest.NewRequest(http.MethodPost, "/v0/sell", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("testuser03"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	SellPositionHandler(svc, users).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, rr.Code)
	}
	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonInsufficientShares) {
		t.Fatalf("expected insufficient-shares reason, got %q", resp.Reason)
	}
	if resp.Message != bets.ProjectionInexecutableSaleMessage {
		t.Fatalf("expected projection message, got %q", resp.Message)
	}
	if resp.Details["positionOutcomeShares"] != float64(34) ||
		resp.Details["nominalUnlockedOutcomeShares"] != float64(17) ||
		resp.Details["projectedOutcomeShares"] != float64(34) {
		t.Fatalf("unexpected share details: %+v", resp.Details)
	}
}
```

- [ ] **Step 2: Run handler tests and verify red**

Run:

```bash
(cd backend && go test ./handlers/bets/selling -run 'TestSell(Quote|Position)Handler_ProjectionInexecutable' -count=1)
```

Expected: FAIL because the handler currently falls through to the generic insufficient-shares details.

- [ ] **Step 3: Implement projection handler mapping**

In `backend/handlers/bets/selling/sellpositionhandler.go`, add this block in `handleSellError` after the dust-cap block and before the `switch`:

```go
	var projectionErr bets.SaleProjectionNotExecutableError
	if errors.As(err, &projectionErr) {
		_ = handlers.WriteFailureWithDetails(
			w,
			http.StatusUnprocessableEntity,
			handlers.ReasonInsufficientShares,
			bets.ProjectionInexecutableSaleMessage,
			saleProjectionDetails(projectionErr.Details),
		)
		return
	}
```

Add this helper below `handleSellError`:

```go
func saleProjectionDetails(details bets.SaleProjectionDetails) map[string]any {
	return map[string]any{
		"outcome":                       details.Outcome,
		"requestedCredits":              details.RequestedCredits,
		"positionValue":                 details.PositionValue,
		"positionOutcomeShares":         details.PositionOutcomeShares,
		"nominalUnlockedValue":          details.NominalUnlockedValue,
		"nominalUnlockedOutcomeShares":  details.NominalUnlockedOutcomeShares,
		"projectedPositionValue":        details.ProjectedPositionValue,
		"projectedOutcomeShares":        details.ProjectedOutcomeShares,
		"executableSaleValue":           details.ExecutableSaleValue,
		"hint":                          "This Position has value, but the requested Sale Order does not currently reduce the backend-projected position enough to pay credits safely.",
	}
}
```

- [ ] **Step 4: Run handler tests and verify green**

Run:

```bash
gofmt -w backend/handlers/bets/selling/sellpositionhandler.go backend/handlers/bets/selling/sellpositionhandler_test.go
(cd backend && go test ./handlers/bets/selling -run 'TestSell(Quote|Position)Handler_ProjectionInexecutable|TestSellQuoteHandler_InsufficientSharesIncludesUserGuidance|TestSellPositionHandler_ErrorMapping' -count=1)
```

Expected: PASS.

- [ ] **Step 5: Commit Task 2**

```bash
git add backend/handlers/bets/selling/sellpositionhandler.go backend/handlers/bets/selling/sellpositionhandler_test.go
git commit -m "fix: return sell projection details in API errors"
```

---

### Task 3: HTTP Integration Coverage For Reported Sequence

**Files:**
- Modify: `integrationtest/scripts/sell-shares-overcashout.mjs`
- Modify: `integrationtest/cases/sell-shares-overcashout.md`
- Modify: `integrationtest/README.md`

**Interfaces:**
- Consumes: existing seeded users `admin`, `testuser01`, `testuser02`, `testuser03`.
- Produces: integration assertions that projection-inexecutable quote/sell failures include backend-computed detail fields and do not mutate state.

- [ ] **Step 1: Add failing integration assertions**

In `integrationtest/scripts/sell-shares-overcashout.mjs`, add a counterparty user near the existing user constants:

```js
const counterparty = args.counterparty || 'testuser03';
```

Add this sequence near the existing attachment constants:

```js
const projectionInexecutableSequence = [
  { seq: 1, user: 'bettor', type: 'buy', outcome: 'NO', amount: 50 },
  { seq: 2, user: 'counterparty', type: 'buy', outcome: 'NO', amount: 25 },
  { seq: 3, user: 'bettor', type: 'sell', outcome: 'NO', amount: 75 },
  { seq: 4, user: 'bettor', type: 'buy', outcome: 'NO', amount: 75 },
  { seq: 5, user: 'counterparty', type: 'buy', outcome: 'NO', amount: 10 },
];

const projectionInexecutableAttempt = { user: 'counterparty', outcome: 'NO', amount: 17 };
```

Add helpers:

```js
function message(raw) {
  return raw?.data?.message || raw?.result?.message || '';
}

function errorDetails(raw) {
  return raw?.data?.details || raw?.result?.details || {};
}

function assertProjectionDetails(prefix, raw) {
  const details = errorDetails(raw);
  check(`${prefix} includes projection message`, message(raw).includes('Position value exists'), message(raw));
  check(`${prefix} details include outcome`, details.outcome === 'NO', JSON.stringify(details));
  check(`${prefix} details include requested credits`, Number(details.requestedCredits) === projectionInexecutableAttempt.amount, JSON.stringify(details));
  check(`${prefix} details include position value`, Number(details.positionValue) > 0, JSON.stringify(details));
  check(`${prefix} details include nominal unlocked value`, Number(details.nominalUnlockedValue) > 0, JSON.stringify(details));
  sameInt(`${prefix} details executable sale value`, details.executableSaleValue, 0);
  check(`${prefix} details include projected position value`, Number.isFinite(Number(details.projectedPositionValue)), JSON.stringify(details));
  check(`${prefix} details include projected outcome shares`, Number.isFinite(Number(details.projectedOutcomeShares)), JSON.stringify(details));
}
```

Add replay and assertion helpers:

```js
async function replayProjectionInexecutableSequence(tokens, marketId) {
  for (const step of projectionInexecutableSequence) {
    const token = tokens[step.user];
    if (step.type === 'buy') {
      await place(token, marketId, step.outcome, step.amount);
      continue;
    }
    const quote = (await quoteRaw(token, marketId, step.outcome, step.amount)).result;
    check(`projection setup seq ${step.seq} quote allowed`, quote.allowed === true);
    const sell = (await sellRaw(token, marketId, step.outcome, step.amount)).result;
    check(`projection setup seq ${step.seq} sell succeeded`, sell.sharesSold > 0 && sell.netProceeds >= 0, `shares=${sell.sharesSold}, net=${sell.netProceeds}`);
  }
}

async function assertProjectionInexecutableRejected({ token, username, marketId }) {
  const beforeFinancial = await financial(username);
  const beforePosition = await position(token, marketId);
  const beforeDetails = await details(marketId);

  check('projection setup has aggregate value', positionValue(beforePosition) > 0, `value=${positionValue(beforePosition)}`);
  check('projection setup has NO shares', shares(beforePosition, 'NO') > 0, `shares=${shares(beforePosition, 'NO')}`);

  const quote = await quoteRaw(token, marketId, projectionInexecutableAttempt.outcome, projectionInexecutableAttempt.amount, [422]);
  check('projection-inexecutable quote rejected', quote.status === 422, `status=${quote.status}`);
  check('projection-inexecutable quote uses insufficient-shares contract', reason(quote) === 'INSUFFICIENT_SHARES', `reason=${reason(quote)}`);
  assertProjectionDetails('projection-inexecutable quote', quote);

  const sell = await sellRaw(token, marketId, projectionInexecutableAttempt.outcome, projectionInexecutableAttempt.amount, [422]);
  check('projection-inexecutable sell rejected', sell.status === 422, `status=${sell.status}`);
  check('projection-inexecutable sell uses insufficient-shares contract', reason(sell) === 'INSUFFICIENT_SHARES', `reason=${reason(sell)}`);
  assertProjectionDetails('projection-inexecutable sell', sell);

  const afterFinancial = await financial(username);
  const afterPosition = await position(token, marketId);
  const afterDetails = await details(marketId);
  assertUnchangedAfterRejection('projection-inexecutable', beforeFinancial, afterFinancial, beforePosition, afterPosition, beforeDetails, afterDetails);
}
```

Update `main()` to login the counterparty and run the new projection market:

```js
  const counterpartyToken = await login(counterparty);
  check('login seeded moderator, bettor, counterparty, admin', true);
```

and after the existing sad path:

```js
  const projectionMarketId = await createMarket('projection', modToken, adminToken);
  await replayProjectionInexecutableSequence({ bettor: bettorToken, counterparty: counterpartyToken }, projectionMarketId);
  await assertProjectionInexecutableRejected({
    token: counterpartyToken,
    username: counterparty,
    marketId: projectionMarketId,
  });
```

- [ ] **Step 2: Run syntax validation**

Run:

```bash
node --check integrationtest/scripts/sell-shares-overcashout.mjs
```

Expected: PASS for syntax. Against a backend without Task 2, the scenario would fail on missing `details`.

- [ ] **Step 3: Update integration docs**

In `integrationtest/cases/sell-shares-overcashout.md`, extend the sad path section with:

```md
Projection-inexecutable path:

- Replays the reported two-user NO sequence:
  - `testuser02 NO 50`
  - `testuser03 NO 25`
  - `testuser02 NO -75`
  - `testuser02 NO 75`
  - `testuser03 NO 10`
- Attempts the `testuser03` NO sale that has aggregate Position Value and nominal unlocked value but does not reduce the backend-projected DBPM position enough to pay credits.
- Asserts quote and sell both return `422 INSUFFICIENT_SHARES` with backend-owned message text and requester-only projection details.
- Asserts the rejected quote/sell do not change user balance, user position, market dust, or market volume.
```

In `integrationtest/README.md`, update the sell-shares over-cashout row to:

```md
| Sell shares over-cashout API scenario runner | `node integrationtest/scripts/sell-shares-overcashout.mjs --base-url http://localhost:8080 --api-prefix /v0` | Covers valid quote/sell, dust accounting, rejected over-cashout, and projection-inexecutable sell error details |
```

- [ ] **Step 4: Run integration scenario if local stack is available**

Run after `./SocialPredict dev-bootstrap-users` and a local backend is serving:

```bash
node integrationtest/scripts/sell-shares-overcashout.mjs --base-url http://localhost:8080 --api-prefix /v0
```

Expected: PASS summary. If no local stack is running, record that this verification was not run and keep `node --check` as completed validation.

- [ ] **Step 5: Commit Task 3**

```bash
git add integrationtest/scripts/sell-shares-overcashout.mjs integrationtest/cases/sell-shares-overcashout.md integrationtest/README.md
git commit -m "test: cover projection-inexecutable sell errors"
```

---

### Task 4: Frontend Trade Adapter And Sell Explanation UI

**Files:**
- Modify: `frontend/src/api/httpClient.js`
- Create: `frontend/src/api/tradeApi.js`
- Create: `frontend/src/api/tradeApi.test.jsx`
- Modify: `frontend/package.json`
- Modify: `frontend/package-lock.json`
- Modify: `frontend/src/components/layouts/trade/TradeUtils.jsx`
- Modify: `frontend/src/components/layouts/trade/BuySharesLayout.jsx`
- Modify: `frontend/src/components/layouts/trade/SellSharesLayout.jsx`

**Interfaces:**
- Consumes: backend `reason`, `message`, and `details` fields from API failure envelopes.
- Produces:
  - `placeBet({ token, marketId, outcome, amount })`
  - `fetchUserPosition({ token, marketId })`
  - `quoteSale({ token, marketId, outcome, amount })`
  - `sellShares({ token, marketId, outcome, amount })`
  - `fetchTradingFees({ token })`
  - `TradeUtils` callback wrappers backed by the adapter.

- [ ] **Step 1: Enable focused frontend tests when missing**

Check whether Vitest is runnable:

```bash
test -x frontend/node_modules/.bin/vitest && echo "vitest present" || echo "vitest missing"
```

If it prints `vitest missing`, run:

```bash
npm --prefix frontend install --save-dev vitest jsdom
```

Then add this script to `frontend/package.json`:

```json
"test": "vitest run"
```

Expected: `frontend/package-lock.json` updates if dependencies were missing.

- [ ] **Step 2: Write failing adapter test**

Create `frontend/src/api/tradeApi.test.jsx`:

```jsx
import { afterEach, describe, expect, it, vi } from 'vitest';
import { quoteSale } from './tradeApi';

afterEach(() => {
  vi.restoreAllMocks();
});

describe('tradeApi', () => {
  it('preserves backend projection details on rejected sale quotes', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({
      ok: false,
      reason: 'INSUFFICIENT_SHARES',
      message: 'Position value exists, but this Sale Order is not executable right now.',
      details: {
        outcome: 'NO',
        requestedCredits: 17,
        positionValue: 34,
        nominalUnlockedValue: 17,
        projectedPositionValue: 34,
        executableSaleValue: 0,
      },
    }), { status: 422, headers: { 'Content-Type': 'application/json' } })));

    await expect(quoteSale({
      token: 'token-1',
      marketId: 7,
      outcome: 'NO',
      amount: 17,
    })).rejects.toMatchObject({
      status: 422,
      reason: 'INSUFFICIENT_SHARES',
      message: 'Position value exists, but this Sale Order is not executable right now.',
      details: {
        outcome: 'NO',
        requestedCredits: 17,
        positionValue: 34,
        nominalUnlockedValue: 17,
        projectedPositionValue: 34,
        executableSaleValue: 0,
      },
    });
  });
});
```

- [ ] **Step 3: Run the adapter test and verify red**

Run:

```bash
npm --prefix frontend run test -- src/api/tradeApi.test.jsx
```

Expected: FAIL because `frontend/src/api/tradeApi.js` does not exist or because `httpClient` does not attach `details`.

- [ ] **Step 4: Preserve details in `httpClient`**

In `frontend/src/api/httpClient.js`, extend the non-OK branch:

```js
    if (!response.ok) {
      const error = new Error(getApiErrorMessage(response, data, fallbackMessage, reasonMessages));
      error.status = response.status;
      error.reason = data?.reason || '';
      error.details = data?.details || null;
      error.payload = data;
      throw error;
    }
```

- [ ] **Step 5: Create the trade API adapter**

Create `frontend/src/api/tradeApi.js`:

```js
import { apiRequest, authenticatedApiRequest } from './httpClient';

export const NO_SELLABLE_SHARES_MESSAGE = [
  'No sellable shares yet.',
  'Initial value cannot be sold until a follow-up order from another user has been placed.',
  'Wait for another order from another user, then try selling again.',
].join(' ');

const jsonHeaders = {
  'Content-Type': 'application/json',
};

const validateOrder = ({ marketId, amount, outcome }, actionLabel) => {
  if (!marketId || !amount || !outcome) {
    throw new Error(`Missing required ${actionLabel} data (marketId, amount, outcome)`);
  }
  if (Number(amount) < 1) {
    throw new Error(`${actionLabel} amount must be at least 1`);
  }
  if (outcome !== 'YES' && outcome !== 'NO') {
    throw new Error(`${actionLabel} outcome must be YES or NO`);
  }
};

export const placeBet = ({ token, marketId, outcome, amount }) => {
  if (!token) {
    throw new Error('Please log in to place a bet.');
  }
  validateOrder({ marketId, amount, outcome }, 'Bet');

  return authenticatedApiRequest('/v0/bet', {
    method: 'POST',
    headers: jsonHeaders,
    authToken: token,
    body: JSON.stringify({ marketId, outcome, amount }),
    fallbackMessage: 'Bet failed. Please try again.',
  });
};

export const fetchUserPosition = ({ token, marketId }) => {
  if (!token) {
    throw new Error('Please log in again to view your Position.');
  }
  if (!marketId) {
    throw new Error('Market ID is required.');
  }

  return authenticatedApiRequest(`/v0/userposition/${marketId}`, {
    authToken: token,
    reasonMessages: {
      INVALID_TOKEN: 'Please log in again to view your Position.',
      AUTHORIZATION_DENIED: 'Please log in again to view your Position.',
      PASSWORD_CHANGE_REQUIRED: 'Please update your password before viewing your Position.',
      NO_POSITION: NO_SELLABLE_SHARES_MESSAGE,
    },
    fallbackMessage: NO_SELLABLE_SHARES_MESSAGE,
  });
};

export const quoteSale = ({ token, marketId, outcome, amount }) => {
  if (!token) {
    throw new Error('Please log in to sell shares.');
  }
  validateOrder({ marketId, amount, outcome }, 'Sale');

  return authenticatedApiRequest('/v0/sell/quote', {
    method: 'POST',
    headers: jsonHeaders,
    authToken: token,
    body: JSON.stringify({ marketId, outcome, amount }),
    fallbackMessage: 'Sale quote failed. Please try again.',
  });
};

export const sellShares = ({ token, marketId, outcome, amount }) => {
  if (!token) {
    throw new Error('Please log in to sell shares.');
  }
  validateOrder({ marketId, amount, outcome }, 'Sale');

  return authenticatedApiRequest('/v0/sell', {
    method: 'POST',
    headers: jsonHeaders,
    authToken: token,
    body: JSON.stringify({ marketId, outcome, amount }),
    fallbackMessage: 'Sale failed. Please try again.',
  });
};

export const fetchTradingFees = async ({ token } = {}) => {
  const setup = await apiRequest('/v0/setup', {
    authenticated: Boolean(token),
    authToken: token,
    fallbackMessage: 'Failed to load trading fees.',
  });
  return setup?.betting?.betFees || null;
};
```

- [ ] **Step 6: Delegate `TradeUtils` to the adapter**

Replace transport logic in `frontend/src/components/layouts/trade/TradeUtils.jsx` with:

```jsx
import {
  fetchUserPosition,
  NO_SELLABLE_SHARES_MESSAGE,
  placeBet,
  quoteSale,
  sellShares,
} from '../../../api/tradeApi';

export { NO_SELLABLE_SHARES_MESSAGE };

export const submitBet = (betData, token, onSuccess, onError) => {
  placeBet({ token, ...betData })
    .then(onSuccess)
    .catch(onError);
};

export const fetchUserShares = async (marketId, token) => fetchUserPosition({ token, marketId });

export const fetchSaleQuote = async (saleData, token) => quoteSale({ token, ...saleData });

export const submitSale = (saleData, token, onSuccess, onError) => {
  sellShares({ token, ...saleData })
    .then(onSuccess)
    .catch(onError);
};
```

- [ ] **Step 7: Load fees through the adapter**

In `frontend/src/components/layouts/trade/BuySharesLayout.jsx`, remove the `API_URL` import and add:

```js
import { fetchTradingFees } from '../../../api/tradeApi';
```

Replace the `fetchFeeData` body with:

```js
            try {
                setFeeData(await fetchTradingFees({ token }));
            } catch {
                setFeeData(null);
            } finally {
                setIsLoading(false);
            }
```

In `frontend/src/components/layouts/trade/SellSharesLayout.jsx`, remove the `API_URL` import, add:

```js
import { fetchTradingFees } from '../../../api/tradeApi';
```

and make the same `fetchFeeData` replacement.

- [ ] **Step 8: Treat Position Value as informational in sell UI**

In `frontend/src/components/layouts/trade/SellSharesLayout.jsx`:

Rename:

```js
const maxSaleCredits = Math.max(0, Number(shares.value) || 0);
```

to:

```js
const positionValue = Math.max(0, Number(shares.value) || 0);
```

Change the quote error state from string to object:

```js
const [quoteError, setQuoteError] = useState(null);
```

Change all quote-error resets from `setQuoteError('')` to:

```js
setQuoteError(null);
```

Replace `handleSellAmountChange` with:

```js
    const handleSellAmountChange = (event) => {
        const newAmount = parseInt(event.target.value, 10) || 0;
        if (newAmount < 0) {
            return;
        }
        setSaleQuote(null);
        setQuoteError(null);
        setSellAmount(newAmount);
    };
```

Change both quote catch blocks to preserve the backend error:

```js
                setSaleQuote(null);
                setQuoteError(error);
                return null;
```

and:

```js
                alert(error.message);
                setQuoteError(error);
                setIsSubmitting(false);
```

Change:

```js
const sellUnavailableNotice = hasOwnedShares && maxSaleCredits <= 0
    ? NO_SELLABLE_SHARES_MESSAGE
    : sharesNotice;
```

to:

```js
const sellUnavailableNotice = sharesNotice;
```

Change the position display to use `positionValue`:

```jsx
<span className="text-green-300">{positionValue}</span>
```

Change the sale input from:

```jsx
max={maxSaleCredits || 1}
```

to:

```jsx
max={undefined}
```

Change `defaultSaleAmount` to:

```js
const defaultSaleAmount = () => 1;
```

- [ ] **Step 9: Render backend projection details without computing math**

In `frontend/src/components/layouts/trade/SellSharesLayout.jsx`, replace the quote-error branch in `SaleQuotePanel`:

```jsx
    if (quoteError) {
        const message = typeof quoteError === 'string' ? quoteError : quoteError.message;
        const details = typeof quoteError === 'object' ? quoteError.details : null;
        return (
            <div className="mb-4 rounded-lg border border-red-400 bg-red-950/40 p-3 text-sm text-red-100">
                <p>{message}</p>
                <ProjectionErrorDetails details={details} />
            </div>
        );
    }
```

Add this component above `QuoteMetric`:

```jsx
const ProjectionErrorDetails = ({ details }) => {
    if (!details) {
        return null;
    }

    const rows = [
        ['Position Value', details.positionValue],
        ['Nominal Unlocked Value', details.nominalUnlockedValue],
        ['Currently Executable Sale Value', details.executableSaleValue],
    ].filter(([, value]) => value !== undefined && value !== null);

    if (rows.length === 0) {
        return null;
    }

    return (
        <div className="mt-3 grid gap-2 sm:grid-cols-3">
            {rows.map(([label, value]) => (
                <QuoteMetric key={label} label={label} value={value} />
            ))}
        </div>
    );
};
```

- [ ] **Step 10: Run frontend tests and build**

Run:

```bash
npm --prefix frontend run test -- src/api/tradeApi.test.jsx
npm --prefix frontend run build:report
```

Expected: test PASS and build PASS. Existing build warnings may remain if unrelated.

- [ ] **Step 11: Commit Task 4**

If `package.json` and `package-lock.json` changed:

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/api/httpClient.js frontend/src/api/tradeApi.js frontend/src/api/tradeApi.test.jsx frontend/src/components/layouts/trade/TradeUtils.jsx frontend/src/components/layouts/trade/BuySharesLayout.jsx frontend/src/components/layouts/trade/SellSharesLayout.jsx
git commit -m "fix: show backend sell projection guidance in trade UI"
```

If no package files changed:

```bash
git add frontend/src/api/httpClient.js frontend/src/api/tradeApi.js frontend/src/api/tradeApi.test.jsx frontend/src/components/layouts/trade/TradeUtils.jsx frontend/src/components/layouts/trade/BuySharesLayout.jsx frontend/src/components/layouts/trade/SellSharesLayout.jsx
git commit -m "fix: show backend sell projection guidance in trade UI"
```

---

### Task 5: Final Verification And Branch Preparation

**Files:**
- Read: changed files from Tasks 1-4
- No required source edits unless verification exposes a defect.

**Interfaces:**
- Consumes: all task outputs.
- Produces: verified branch ready for PR.

- [ ] **Step 1: Run focused backend tests**

```bash
(cd backend && go test ./internal/domain/bets ./internal/domain/math/positions ./handlers/bets/selling)
```

Expected: PASS.

- [ ] **Step 2: Run broad backend tests**

```bash
(cd backend && JWT_SIGNING_KEY=test-secret-key-for-testing go test ./...)
```

Expected: PASS.

- [ ] **Step 3: Run frontend tests and build**

```bash
npm --prefix frontend run test -- src/api/tradeApi.test.jsx
npm --prefix frontend run build:report
```

Expected: PASS. Existing unrelated build warnings can be reported.

- [ ] **Step 4: Run integration syntax checks**

```bash
node --check integrationtest/scripts/sell-shares-overcashout.mjs
node --check integrationtest/scripts/sell-shares-two-share-cap.mjs
```

Expected: PASS.

- [ ] **Step 5: Run local HTTP integration when stack is available**

```bash
node integrationtest/scripts/sell-shares-overcashout.mjs --base-url http://localhost:8080 --api-prefix /v0
```

Expected: PASS. If no local backend is running, record that it was not run.

- [ ] **Step 6: Inspect final diff**

```bash
git status --short
git diff origin/main... --stat
git diff origin/main... -- backend/internal/domain/bets/errors.go backend/internal/domain/bets/bet_sell.go backend/handlers/bets/selling/sellpositionhandler.go frontend/src/api/httpClient.js frontend/src/api/tradeApi.js frontend/src/components/layouts/trade/SellSharesLayout.jsx
```

Expected: diff contains only projection-aware sell guidance, adapter seam work, tests, and integration docs.

- [ ] **Step 7: Push branch and create PR when requested**

```bash
git push -u origin HEAD
```

Use PR title:

```text
Fix projection-aware sell guidance and backend details
```

Use PR body:

```md
## Summary
- keeps backend quote/sell projection as the only authority for executable Sale Orders
- returns requester-only projection details on existing `422 INSUFFICIENT_SHARES` failures
- routes trade UI calls through an API adapter and displays backend-owned sell guidance
- extends sell over-cashout integration coverage for the reported two-user sequence

## Validation
- [ ] `(cd backend && go test ./internal/domain/bets ./internal/domain/math/positions ./handlers/bets/selling)`
- [ ] `(cd backend && JWT_SIGNING_KEY=test-secret-key-for-testing go test ./...)`
- [ ] `npm --prefix frontend run test -- src/api/tradeApi.test.jsx`
- [ ] `npm --prefix frontend run build:report`
- [ ] `node --check integrationtest/scripts/sell-shares-overcashout.mjs`
- [ ] `node integrationtest/scripts/sell-shares-overcashout.mjs --base-url http://localhost:8080 --api-prefix /v0`
```
