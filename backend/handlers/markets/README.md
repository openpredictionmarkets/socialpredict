```
ResolveMarketHandler
│
├──> Validate + Load Market
│
├──> market.IsResolved ? (if yes, reject)
│
├──> Set IsResolved = true, set ResolutionResult
│
├──> payout.DistributePayoutsWithRefund(market, db)
│      │
│      ├──> If "N/A"   → refundAllBets (all users get original bet back)
│      │
│      └──> If "YES/NO"
│            └──> calculateAndAllocateProportionalPayouts
│                   ├──> Load bets, calc totalVolume
│                   ├──> positions.CalculateMarketPositions_WPAM_DBPM
│                   ├──> SelectWinningPositions
│                   └──> AllocateWinningSharePool
│                         ├──> For each winner: payout = round((shares/totalWinningShares)*totalVolume)
│                         ├──> Rounding correction to top holder
│                         └──> UpdateUserBalance(user, payout, ...)
│
└──> Save + Respond

```