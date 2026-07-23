# Sell Shares Over-Cashout Regression

## Purpose

Verify that selling shares cannot repeatedly cash out more value than the user's
remaining market position supports.

This scenario covers both sell quote and settlement behavior:

- `/v0/sell/quote` must reject unsafe projected sales before returning an
  allowed quote.
- `/v0/sell` must reject the same unsafe sale through the existing
  insufficient-shares API contract.
- Rejected sells must not credit the user, append a sell row, increase position
  value, or increase market dust/volume.

## Setup

Run against a local dev stack after seeded users exist:

```bash
./SocialPredict dev-bootstrap-users
node integrationtest/scripts/sell-shares-overcashout.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0
```

Defaults assume seeded users `admin`, `testuser01`, and `testuser02` all use
password `password`.

## Scenario

The runner creates fresh binary markets through `/v0/markets` and approves them
through the admin route when market governance creates proposals.

Happy path:

- Replays the opening trades from the reported sequence.
- Quotes and sells the valid seq 4 YES sale.
- Replays additional buys and quotes/sells the valid seq 10 dust-generating YES
  sale.
- Asserts `sharesSold`, `saleValue`, `dust`, `netProceeds`, user balance,
  projected position value, and market detail dust/volume fields.

Sad path:

- Replays the reported sequence through seq 18.
- Attempts the seq 19 NO sale that would previously set up the alternating
  tiny-share large-cashout tail.
- Asserts both quote and sell return the existing `INSUFFICIENT_SHARES` failure
  reason and that user balance, position, market dust, and market volume remain
  unchanged.

Projection-inexecutable path:

- Replays the reported two-user NO sequence:
  - `testuser02 NO 50`
  - `testuser03 NO 25`
  - `testuser02 NO -75`
  - `testuser02 NO 75`
  - `testuser03 NO 10`
- Attempts the `testuser03` NO sale that has aggregate Position Value and
  nominal unlocked value but does not reduce the backend-projected DBPM position
  enough to pay credits.
- Asserts quote and sell both return `422 INSUFFICIENT_SHARES` with
  backend-owned message text and requester-only projection details.
- Asserts the rejected quote/sell do not change user balance, user position,
  market dust, or market volume.

The input is based on `.context/attachments/pPjgi8/sell_market.json`.
