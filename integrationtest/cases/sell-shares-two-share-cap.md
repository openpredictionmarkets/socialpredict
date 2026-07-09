# Sell Shares Two-Share Backend Cap

## Purpose

Verify through backend HTTP endpoints that a user's initial buy is not sellable
until a later buy by another user unlocks the prior DBPM-rounded value. Once
unlocked, the user can sell the value DBPM assigns to that prior buy, but cannot
sell beyond the unlocked position value.

This scenario covers both sell quote and settlement behavior:

- `/v0/sell/quote` must reject an immediate sell attempt with the existing
  `NO_POSITION` contract and a backend-provided follow-up-order message.
- After another user buys, `/v0/sell/quote` must cap `sharesSold` at the
  unlocked DBPM-rounded shares and cap `saleValue` at the unlocked position
  value.
- `/v0/sell` must apply the same cap, credit only `netProceeds`, retain dust in
  market accounting, and exhaust the position.
- The latest buyer's own value remains locked until another qualifying buy.
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

Defaults assume seeded users `admin`, `testuser01`, `testuser03`, and
`testuser04` all use password `password`.

## Scenario

- Login a seeded moderator, bettor, follower, and admin.
- Create a fresh binary market through `/v0/markets` and approve it through the
  admin route when market governance creates a proposal.
- Buy `1` credit of NO as the bettor.
- Assert immediate quote and sell attempts are rejected through `NO_POSITION`,
  include the backend follow-up-order message, and do not mutate balance,
  position value, or market dust.
- Buy `1` credit of NO as a different follower user.
- Assert the follower's latest value is still not sellable.
- Assert the original bettor now has two unlocked NO shares and position value
  `2`, matching the DBPM-rounded value assigned to the prior buy.
- Submit an oversized NO sale order for `3` credits through `/v0/sell/quote`.
- Assert the quote reports `sharesSold=2`, `saleValue=2`, `dust=1`,
  `netProceeds=1`, and `valuePerShare=1`.
- Submit the same oversized order through `/v0/sell`.
- Assert settlement reports the same bounded sale, credits only net proceeds,
  leaves the user with zero NO shares and zero position value, and retains dust
  in market detail accounting.
- Attempt another quote and sell after exhaustion and assert both are rejected
  through the existing `NO_POSITION` or `INSUFFICIENT_SHARES` path.
