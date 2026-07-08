# Sell Shares Two-Share Backend Cap

## Purpose

Verify through backend HTTP endpoints that a user who buys `1` NO twice owns
two NO shares and cannot sell more than those two shares, even when the sell
order asks for more value than the position supports.

This scenario covers both sell quote and settlement behavior:

- `/v0/sell/quote` must cap `sharesSold` at the owned two shares and cap
  `saleValue` at the current position value.
- `/v0/sell` must apply the same cap, credit only `netProceeds`, retain dust in
  market accounting, and exhaust the position.
- A follow-up quote and sell after the position is exhausted must return the
  existing insufficient-position contract.

## Setup

Run against a local dev stack after seeded users exist:

```bash
./SocialPredict dev-bootstrap-users
node integrationtest/scripts/sell-shares-two-share-cap.mjs \
  --base-url http://localhost:8080 \
  --api-prefix /v0
```

Defaults assume seeded users `admin`, `testuser01`, and `testuser03` all use
password `password`.

## Scenario

- Login a seeded moderator, bettor, and admin.
- Create a fresh binary market through `/v0/markets` and approve it through the
  admin route when market governance creates a proposal.
- Buy `1` credit of NO twice as the bettor.
- Assert the bettor owns two NO shares and the position value is `2`.
- Submit an oversized NO sale order for `3` credits through `/v0/sell/quote`.
- Assert the quote reports `sharesSold=2`, `saleValue=2`, `dust=1`,
  `netProceeds=1`, and `valuePerShare=1`.
- Submit the same oversized order through `/v0/sell`.
- Assert settlement reports the same bounded sale, credits only net proceeds,
  leaves the user with zero NO shares and zero position value, and retains dust
  in market detail accounting.
- Attempt another quote and sell after exhaustion and assert both are rejected
  through the existing `NO_POSITION` or `INSUFFICIENT_SHARES` path.
