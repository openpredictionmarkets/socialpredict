# Sell Shares Over-Cashout Regression

## Purpose

Verify that selling shares cannot repeatedly cash out more value than the user's
remaining market position supports.

This scenario covers both sell quote and settlement behavior:

- `/v0/sell/quote` must never return an allowed quote whose executed
  `sharesSold` or `saleValue` exceeds backend-derived unlocked inventory.
- `/v0/sell` must settle the same capped execution that quote returned.
- Oversized sale requests may round down to remaining executable value when the
  dust rules allow it.
- Sales must credit only `netProceeds`; dust is retained by the market.
- Previously unlocked value must remain executable after a later same-user buy.

## Setup

Run against a local dev stack after seeded users exist:

```bash
./SocialPredict dev-bootstrap-users
node integrationtest/scripts/sell-shares-overcashout.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0 \
  --delay 1100
```

Defaults assume seeded users `admin`, `testuser01`, and `testuser02` through
`testuser07` all use password `password`.

## Scenario

The runner creates fresh binary markets through `/v0/markets` and approves them
through the admin route when market governance creates proposals.

Happy path:

- Replays the opening trades from the reported sequence.
- Adds a later counterparty buy before each sale so the bettor's prior YES value
  is sellable under sequence-based unlocking.
- Quotes and sells the valid seq 4 YES sale after that unlock.
- Replays additional buys and quotes/sells the valid seq 10 dust-generating YES
  sale after a second counterparty unlock.
- Asserts `sharesSold`, `saleValue`, `dust`, `netProceeds`, user balance,
  position display values, and market detail dust/volume fields.

Capped over-request path:

- Replays the final two-user unlocked-lot NO sequence.
- Attempts the reported oversized `507` credit NO sale request against the
  remaining unlocked position.
- Asserts quote and sell cap the oversized request to remaining executable
  unlocked inventory.
- Asserts the capped sale does not produce large proceeds from tiny remaining
  shares.
- Asserts user balance increases only by `netProceeds` and market dust/volume
  reflect the capped sell row.

Sequence-based unlocked-lot path:

- Replays the reported two-user NO sequence:
  - `testuser02 NO 50`
  - `testuser03 NO 25`
  - `testuser02 NO -75`
  - `testuser02 NO 75`
  - `testuser03 NO 10`
  - `testuser02 NO 20`
- Uses separate fresh markets to assert both `testuser02` and `testuser03` can
  quote and sell a small amount from previously unlocked NO value.
- Asserts quote and sell match on `sharesSold`, `saleValue`, `dust`, and
  `netProceeds`.
- Asserts user balance increases only by `netProceeds` and market dust is
  retained by the market.

The input is based on `.context/attachments/pPjgi8/sell_market.json`.
