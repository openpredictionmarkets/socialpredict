# Sellable Position UI Design

## Problem

The July 31 container log contains only fast `422 Unprocessable Entity`
responses from `POST /v0/sell/quote`; it contains no timeout, panic, OOM, or
slow-request evidence. The responses occur when a user owns shares but has no
unlocked inventory, or requests more than the unlocked inventory permits.

Version 3.8 correctly keeps this rule on the server. The Sell view currently
loads only the aggregate position, however, so it presents the total position
value as though it could be sold. Users consequently discover the locked
portion by submitting a quote.

## Decision

Add a single sellability read to the existing authenticated user-position
response. It will return the current aggregate position plus server-derived
sellable values for YES and NO. The Sell view will display the sellable value,
select an outcome only when it has sellable shares, constrain the amount input
to the selected outcome's sellable value, and explain zero sellability without
calling the quote endpoint.

The quote and settlement endpoints remain the final authority. Their existing
validation and `422` responses remain necessary for stale state and concurrent
orders.

## Architecture

The markets service will add one position-read method that loads the market
and bet history once, calculates the current position and both outcome-specific
sellable positions from that shared input, then returns a small trade-position
read model. The user-position handler will serialize this read model through
its existing response envelope.

This avoids the tempting but inefficient alternative of calling
`GetUserSellablePositionInMarket` once per outcome in the HTTP handler, which
would reload and replay the same market history for each call. It also avoids
client-side sellability math, which would diverge from DBPM and the server's
accounting rules.

## API and UI Contract

`GET /v0/userposition/{marketId}` remains authenticated and backward
compatible. In addition to existing aggregate fields, its result gains:

- `yesSellableShares` and `noSellableShares`
- `yesSellableValue` and `noSellableValue`

All four fields are non-negative integers. A zero value means that outcome has
no sale that can currently be submitted. The frontend must treat these fields
as advisory and must preserve the existing quote-before-settlement flow.

The Sell view will show both total position value and selected-outcome sellable
value. For an owned but fully locked position it will show the established
unlock explanation, disable sale actions, and avoid a quote request. On a
mixed position it will make only outcomes with a positive sellable value
actionable. On each selection/change it clamps the requested amount to the
selected outcome's sellable value.

## Error Handling and Performance

Invalid/stale/concurrent requests continue to return the current backend
errors. The UI shows those errors as it does now. No retry loop or background
polling is introduced.

The new read performs one market/history load and reuses its payout sequence
for both outcome calculations. It replaces the current one-position read on
the Sell view, rather than adding quote traffic. Therefore the change removes
predictable invalid quote calls while keeping one bounded server calculation
per position refresh.

## Tests

- Go unit tests prove the trade-position read derives aggregate and both
  sellable outcome values from one history, including fully locked and mixed
  positions.
- Handler tests verify the new response fields for an authenticated user.
- React tests verify locked positions disable sale controls, sellable values
  render, and input changes cannot exceed the selected outcome's sellable
  value.
- Existing backend sell and quote tests continue to prove server-side
  enforcement when client data becomes stale.
