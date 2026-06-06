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
