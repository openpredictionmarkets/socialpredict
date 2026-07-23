# Projection-Aware Sell Quote Design

Status: design approved for review
Date: 2026-07-22
Branch: `pwdel/debug-sell-sequence-20260722`

## Context

SocialPredict treats the backend as the source of domain truth for market lifecycle, trading, resolution, accounting, canonical outcome semantics, and authentication. The frontend is an experience boundary: it owns workflow presentation, labels, recovery copy, accessibility behavior, and display mapping, but it must not compute trading or accounting rules.

The requested architecture constraints also require frontend changes to pass through adapter seams instead of adding direct `fetch`, `localStorage`, or raw API-envelope handling. Public UI language for domain failures is treated as part of the contract surface, so failure copy should originate from the backend when it describes trading/accounting state.

`lib/design/design-plan.json` was referenced by the user but is not present in this checkout. This design uses the architecture notes supplied in the conversation as authoritative.

## Problem

A Participant can have aggregate Position Value in a market and can also have nominal unlocked value from prior buy rows. In the reported sequence, however, appending a proposed negative Sale Order row does not reduce the seller's projected DBPM position enough to safely pay out credits. The existing backend projection guard correctly rejects the sale with `422 INSUFFICIENT_SHARES`, but the UI can still imply that the aggregate Position Value is directly sellable because it reads `/v0/userposition/:marketId` and treats that value as a sale limit.

That makes a correct backend rejection look like a product bug. The issue is not that the frontend needs to reimplement DBPM. The issue is that the backend quote/sell result needs to be the only authority for executable sales, and the frontend needs backend-owned language and requester-only values to explain why a visible Position Value may not currently be executable.

## Goals

- Keep backend quote and sell simulation as the only source of truth for Sale Order executability.
- Preserve the public request/response DTO contract for `/v0/sell/quote` and `/v0/sell`.
- Keep rejected unsafe sales on the existing `422 INSUFFICIENT_SHARES` path.
- Include requester-only diagnostic values in the existing error-envelope `details` object so the frontend can display Position Value versus currently executable Sale Order value without computing domain math.
- Move the trade UI path toward the existing HTTP/auth adapter seam instead of adding more direct transport code.
- Add regression coverage for the reported buy/sell sequence at backend and integration levels.

## Non-Goals

- Do not change DBPM settlement economics in this PR.
- Do not add durable backend state or database migrations.
- Do not introduce Redux, RTK Query, server-managed sessions, browser telemetry, CSP work, offline behavior, or broad frontend platform changes.
- Do not let frontend code calculate unlocked value, projected position reduction, dust, sale proceeds, or any other domain accounting rule.
- Do not perform historical production repair.

## Chosen Approach

Use a projection-aware quote/sell explanation while preserving the existing API shape.

`/v0/sell/quote` remains the preflight authority. `/v0/sell` remains the settlement authority. Both paths continue to run the same backend projection invariant: append the proposed Sale Order in memory, recalculate projected user position, and reject if the sale does not reduce the sold-outcome shares/value enough or causes unsafe projected state.

When that projection fails after nominal unlocked value exists, the service should return a more specific internal error that still matches `ErrInsufficientShares`. The HTTP handler should keep returning `422 INSUFFICIENT_SHARES`, but with clearer backend-owned message text and requester-only details.

## Backend Design

Add a domain error type for a projection-inexecutable sale. It should wrap or match `ErrInsufficientShares` so existing callers and handlers remain compatible.

The error should be raised from the projection validation path used by both quote and settlement. It should carry values the requester is already entitled to know about their own Position:

- `outcome`: canonical Outcome Code for the attempted sale.
- `requestedCredits`: the requested Sale Order amount.
- `positionValue`: the Participant's current aggregate Position Value for the market.
- `positionOutcomeShares`: the Participant's current shares for the attempted Outcome Code.
- `nominalUnlockedValue`: backend-computed unlocked value before projection validation.
- `nominalUnlockedOutcomeShares`: backend-computed unlocked shares before projection validation.
- `projectedPositionValue`: projected Position Value after appending the proposed Sale Order.
- `projectedOutcomeShares`: projected shares for the attempted Outcome Code after appending the proposed Sale Order.
- `executableSaleValue`: `0` for this rejected quote/sell attempt.

The handler should continue returning the existing error reason:

```json
{
  "ok": false,
  "reason": "INSUFFICIENT_SHARES",
  "message": "Position value exists, but this Sale Order is not executable right now. Market accounting requires a sale to reduce your projected position before credits can be paid out. Wait for more market activity, then try again.",
  "details": {
    "outcome": "NO",
    "requestedCredits": 17,
    "positionValue": 34,
    "positionOutcomeShares": 34,
    "nominalUnlockedValue": 17,
    "nominalUnlockedOutcomeShares": 17,
    "projectedPositionValue": 34,
    "projectedOutcomeShares": 34,
    "executableSaleValue": 0,
    "hint": "This Position has value, but the requested Sale Order does not currently reduce the backend-projected position enough to pay credits safely."
  }
}
```

The exact numeric values will vary by market state. The key contract is that these are backend-computed, requester-only explanation fields in the existing error envelope, not new request or success-response DTO fields.

Settlement must remain transaction-scoped. A rejected projection-inexecutable sale must not create a sell ledger row, credit user balance, increase market dust, or update market volume.

## Frontend Design

The frontend should stop treating aggregate `/v0/userposition/:marketId` value as an executable sale limit. It may display aggregate Position Value as informational, but sellability comes from `/v0/sell/quote` and `/v0/sell` only.

Create or extend a trade API adapter under `frontend/src/api/` that uses the existing HTTP/auth adapter primitives. The trade UI should call adapter functions for:

- placing a Buy Order,
- fetching the Participant's Position display data,
- requesting a Sale Order quote,
- submitting a Sale Order.

The adapter should unwrap backend envelopes and preserve structured error fields such as `reason`, `message`, and `details`. The UI should render backend-owned domain copy when present. It may format requester-only numeric details for a Terms/Conditions panel, but it must not derive whether a Sale Order is allowed from those values.

Suggested UI language:

- Use `Position Value` for aggregate value.
- Use `Sale Order` for the user's requested sale.
- Avoid labeling aggregate value as `sellable`.
- On projection-inexecutable failure, display the backend message and optionally show:
  - `Position Value`
  - `Nominal Unlocked Value`
  - `Currently Executable Sale Value`

For a rejected quote, `Currently Executable Sale Value` should be shown from backend `details.executableSaleValue`, not calculated in the browser.

## Data Flow

1. Participant opens Sell Shares.
2. Frontend adapter loads position display data for the market.
3. UI displays Position Value as informational.
4. Participant enters a Sale Order amount.
5. Frontend adapter calls `/v0/sell/quote`.
6. Backend validates market state, nominal unlocked position, dust, and projection safety.
7. If projection passes, backend returns the existing quote response and UI enables confirmation.
8. If projection fails, backend returns `422 INSUFFICIENT_SHARES` with message/details. UI displays that explanation and does not enable a misleading confirmation path.
9. On confirmation, frontend submits `/v0/sell`; backend repeats the transaction-scoped projection check before mutation.

## Error Handling

`NO_POSITION` and `NO_SELLABLE_SHARES` should continue to explain that no sellable shares exist yet.

`INSUFFICIENT_SHARES` should distinguish between:

- requested amount exceeds currently executable value,
- nominal unlocked value exists but projection safety prevents this Sale Order from paying credits now.

Both cases can keep the same public reason. The backend message and details provide the more precise user-facing explanation.

Frontend fallback copy is allowed only for network, auth, or unknown errors. Domain trading messages should come from the backend response.

## Tests

Backend tests:

- Add a domain regression for the reported sequence:
  - `testuser02 NO 50`
  - `testuser03 NO 25`
  - `testuser02 NO -75`
  - `testuser02 NO 75`
  - `testuser03 NO 10`
- Assert aggregate Position Value exists for the seller.
- Assert nominal unlocked value may exist.
- Assert quote and sell reject the attempted Sale Order when projection does not reduce the Position enough.
- Assert the error matches `ErrInsufficientShares` and carries projection details.
- Assert rejected settlement does not mutate ledger, user balance, dust, or volume.

Handler tests:

- Assert projection-inexecutable quote/sell returns `422 INSUFFICIENT_SHARES`.
- Assert response message is backend-owned and avoids raw HTTP-machine language.
- Assert `details` includes requester-only Position and projection fields.

Integration test:

- Add or update an executable scenario under `integrationtest/` that replays the reported sequence.
- Cover both `/v0/sell/quote` and `/v0/sell`.
- Assert rejected attempts do not increase user balance, market dust, market volume, or sale ledger rows.
- Assert returned error details can support the Terms/Conditions explanation without frontend domain math.

Frontend validation:

- Build the frontend after moving trade calls through the adapter seam.
- Where the test harness supports it, add focused tests around adapter error preservation and UI rendering of backend message/details.

## Validation Commands

Run focused backend tests:

```bash
go test ./internal/domain/bets ./internal/domain/math/positions ./handlers/bets/selling
```

Run broad backend tests when practical:

```bash
JWT_SIGNING_KEY=test-secret-key-for-testing go test ./...
```

Run frontend validation after adapter changes:

```bash
npm --prefix frontend run build:report
```

Run the integration scenario against a local dev stack after seeded users are bootstrapped:

```bash
node integrationtest/scripts/sell-shares-overcashout.mjs --base-url http://localhost:8080 --api-prefix /v0
```

If the existing overcashout script is not the right home for the projection-inexecutable case, add a separate script and case file with a precise name.

## Acceptance Criteria

- The frontend does not compute domain sellability, unlocked value, projected position, dust, or proceeds.
- Backend quote/sell are the sole authority for Sale Order executability.
- Existing request DTOs and success response DTOs remain compatible.
- Unsafe projection failures still use `422 INSUFFICIENT_SHARES`.
- Error details include backend-computed requester-only values for Terms/Conditions display.
- Reported sequence is covered by backend and integration regression tests.
- Rejected quote/sell attempts do not mutate balance, ledger, dust, or volume.
