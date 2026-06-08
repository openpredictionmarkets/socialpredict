# Sell Flow And Dust

A sell request asks for a credit amount, not a share count:

```text
POST /v0/sell
{ marketId, outcome, amount }
```

`amount` is the number of credits the user is requesting from the sale. The sell domain calculates the current integer value per share, rounds the sale down to whole shares, retains any dust fee, credits the user with net proceeds, and reports the retained remainder as transaction-time dust.

```text
valuePerShare = position.Value / sharesOwned
sharesSold    = requestedCredits / valuePerShare
saleValue     = sharesSold * valuePerShare
rawDust       = requestedCredits - saleValue
dust          = min(rawDust, maxDustPerSale)
netProceeds   = saleValue - dust
```

If the raw remainder exceeds `economics.betting.maxDustPerSale` from `setup.yaml`, the executable Sale Order is rounded down to the nearest allowed value. With the default `maxDustPerSale: 1`, an over-remainder sale is still allowed, but the charged dust fee is capped at `1`.

## Sale Quote Preview

Use the quote endpoint before submitting a sale:

```text
POST /v0/sell/quote
{ marketId, outcome, amount }
```

The quote endpoint is read-only. It uses the same backend sale calculator as
`POST /v0/sell`, then returns:

```json
{
  "requestedCredits": 31,
  "sharesSold": 3,
  "saleValue": 30,
  "dust": 1,
  "netProceeds": 29,
  "maxDust": 1,
  "valuePerShare": 10,
  "dustCapCoverage": 0.2,
  "allowed": true,
  "suggestedAmounts": [30, 31],
  "message": "This Sale Order can be submitted. It would include a 1 credit dust fee from whole-share rounding."
}
```

That keeps frontend previews aligned with `maxDustPerSale` and the current
`valuePerShare`: the frontend displays the quote and suggested amounts, but it
does not recalculate dust independently.

## Persistence

The sale creates a bet ledger row where:

```text
amount = -sharesSold
```

That keeps dust stateless:

- `amount` remains the share/position ledger entry used by the market math.
- User balance is credited by `netProceeds`, not the gross `saleValue` or the originally requested credits.
- `saleValue` remains the gross whole-share sale value before dust is retained.
- The transaction response reports `dust`, but dust is not persisted as a separate bet column.

## API Response

A successful sale returns:

```json
{
  "sharesSold": 2,
  "saleValue": 20,
  "dust": 1,
  "netProceeds": 19
}
```

The frontend should tell the user when `dust > 0` so the rounding fee is visible at the time of sale.

## Market Dust

Market dust is derived from the bet vector without additional database state. The current stateless convention counts one retained dust unit per historical sell row, while the sale endpoint enforces `maxDustPerSale` at transaction time.
