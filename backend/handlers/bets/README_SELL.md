```
User submits Sell Request (marketID, outcome, creditAmount)
│
├── Stream all bets in chronological order
│   │
│   ├── For each bet:
│   │   ├── If bet is a BUY (Amount > 0):
│   │   │   ├── Update user's position (shares, value)
│   │   │   └── Update total market volume
│   │   │
│   │   └── If bet is a SALE (Amount < 0):
│   │       ├── Look up current user position
│   │       ├── Calculate value/share = position.Value / shares
│   │       ├── Compute sharesSold = -bet.Amount
│   │       ├── Compute actualCredits = sharesSold × valuePerShare
│   │       ├── Compute dust = creditRequested - actualCredits
│   │       └── Accumulate dust in totalDust variable
│   │
│   └── Continue until all bets processed
│
└── Output:
    ├── Final market volume = sum(abs(bet.Amount)) + totalDust
    ├── Updated user balance = oldBalance + actualCredits
    ├── New negative bet recorded
    └── TotalDust available for audit / liquidity
```

Ramifications of this:

* GetMarketVolume is used in positions.
* GetMarketVolume, in order to be stateless, needs to compute market dust.

```
User submits Sell Request (marketID, outcome, creditAmount)
│
├── Stream all bets in chronological order
│   │
│   ├── For each bet:
│   │   ├── If bet is a BUY (Amount > 0):
│   │   │   ├── Update user's position (shares, value)
│   │   │   └── Update total market volume
│   │   │
│   │   └── If bet is a SALE (Amount < 0):
│   │       ├── Snapshot user state (position, value/share)
│   │       ├── ⬇️ Call CalculateMarketPositions_WPAM_DBPM()
│   │       │     ├── ⬇️ Calls GetMarketVolume(allBets)
│   │       │     │     ├── sum(abs(bet.Amount)) from allBets
│   │       │     │     └── ⬇️ + GetMarketDust(allBets)
│   │       │     │           └── Iterates over allBets, and for each SALE:
│   │       │     │               ├── Snapshot position at sale time
│   │       │     │               ├── Compute value/share
│   │       │     │               ├── Compute actual credits = shares × value/share
│   │       │     │               ├── Infer dust = requested - actualCredits
│   │       │     │               └── Accumulate to dustTotal
│   │       │     └── Uses totalVolume (with dust) to calculate valuations
│   │       ├── Compute value/share from user's position
│   │       ├── Compute sharesSold = -bet.Amount
│   │       ├── Compute actualCredits = sharesSold × value/share
│   │       ├── Compute dust = creditRequested - actualCredits
│   │       └── Accumulate dust in totalDust variable
│   │
│   └── Continue until all bets processed
│
└── Output:
    ├── Final market volume = sum(abs(bet.Amount)) + inferred dust
    ├── Updated user balance = oldBalance + actualCredits
    ├── New negative bet recorded
    └── TotalDust available for audit / liquidity

```

* Note that MarketDust is calculated along with volume.

```
func GetMarketDust(bets []models.Bet) int64 {
    var totalDust int64 = 0
    userPositions := make(map[string]positionsmath.UserMarketPosition)
    var totalVolume int64 = 0

    for _, bet := range bets {
        absAmount := bet.Amount
        if absAmount < 0 {
            absAmount = -absAmount
        }

        // Update volume
        totalVolume += absAmount

        // BUY bet → update user position
        if bet.Amount > 0 {
            pos := userPositions[bet.Username]
            if bet.Outcome == "YES" {
                pos.YesSharesOwned += bet.Amount
            } else {
                pos.NoSharesOwned += bet.Amount
            }
            userPositions[bet.Username] = pos
            continue
        }

        // SELL bet → infer dust
        pos := userPositions[bet.Username]
        var sharesOwned int64
        if bet.Outcome == "YES" {
            sharesOwned = pos.YesSharesOwned
        } else {
            sharesOwned = pos.NoSharesOwned
        }

        if sharesOwned == 0 {
            continue // skip invalid sale
        }

        // ⚠️ Placeholder logic for valuation (replace with real valuation logic)
        value := pos.YesSharesOwned + pos.NoSharesOwned
        valuePerShare := int64(0)
        if sharesOwned != 0 {
            valuePerShare = value / sharesOwned
        }

        sharesSold := -bet.Amount
        actualCredits := valuePerShare * sharesSold
        requestedCredits := actualCredits // or override if known
        dust := requestedCredits - actualCredits
        totalDust += dust

        // Burn sold shares
        if bet.Outcome == "YES" {
            pos.YesSharesOwned -= sharesSold
        } else {
            pos.NoSharesOwned -= sharesSold
        }
        userPositions[bet.Username] = pos
    }

    return totalDust
}
```