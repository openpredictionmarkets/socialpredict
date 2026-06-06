# Sell Flow And Dust

A sell request asks for a credit amount, not a share count:

```text
POST /v0/sell
{ marketId, outcome, amount }
```

`amount` is the number of credits the user is requesting from the sale. The sell domain calculates the current integer value per share, rounds the sale down to whole shares, credits the user with the actual sale value, and records any remainder as dust.

```text
valuePerShare = position.Value / sharesOwned
sharesSold    = requestedCredits / valuePerShare
saleValue     = sharesSold * valuePerShare
dust          = requestedCredits - saleValue
```

If `dust` exceeds `economics.betting.maxDustPerSale` from `setup.yaml`, the sale is rejected with `DUST_CAP_EXCEEDED`.

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
  "requestedCredits": 33,
  "sharesSold": 3,
  "saleValue": 30,
  "dust": 3,
  "maxDust": 2,
  "valuePerShare": 10,
  "dustCapCoverage": 0.3,
  "allowed": false,
  "suggestedAmounts": [30, 31, 32],
  "message": "This sale would create 3 dust, above the configured maximum of 2. Try a different requested credit amount."
}
```

That keeps frontend previews aligned with `maxDustPerSale` and the current
`valuePerShare`: the frontend displays the quote and suggested amounts, but it
does not recalculate dust independently.

## Persistence

The sale creates a bet ledger row where:

```text
amount = -sharesSold
dust   = exact assessed dust
```

That keeps share accounting and dust accounting separate:

- `amount` remains the share/position ledger entry used by the market math.
- `dust` is explicit transaction metadata used to calculate market dust.
- User balance is credited by `saleValue`, not the originally requested credits.

## API Response

A successful sale returns:

```json
{
  "sharesSold": 2,
  "saleValue": 20,
  "dust": 1
}
```

The frontend should tell the user when `dust > 0` so the rounding fee is visible at the time of sale.

## Market Dust

Market dust is no longer inferred from negative sale rows. It is summed from the persisted `bets.dust` values for sale rows. Legacy sale rows without recorded dust use the historical fallback dust value so old markets continue to display a conservative dust estimate.
