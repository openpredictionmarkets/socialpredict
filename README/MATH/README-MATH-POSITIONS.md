### Different Types of Positions

* There are two different types of positions in our codebase, which might be confusing.
* The positions are meant to seperate the core math from what get served to the outside world and combined with other math.
* The idea is to keep the functions in the core math as pure and simple as possible, so that they can be combined together in a pipeline of functions that results in a final output.

####  DBPMMarketPosition

Purpose: Internal, math-only.

Usage: Output of your core pipeline (including NetAggregateMarketPositions), and used for payout resolution logic.

Fields: Only Username, YesSharesOwned, NoSharesOwned.

#### MarketPosition

Purpose: API-facing, informational.

Usage: Takes the result of the internal pipeline (DBPMMarketPosition slice), augments it (for example, by calculating Value using current probability and market volume), and presents it to the frontend/consumer.

Fields: Adds Value (mark-to-market estimate) for display, but does not affect core math.

#### Flowchart of How Each Position Type Used

```
[Raw Bets/Market] 
    |
    V
[ Core math functions: Aggregate, Normalize, etc. ]
    |
    V
[]dbpm.DBPMMarketPosition   <-- (internal math output, result of last in a pipeline of pure functions)
    |
    V
[]positions.MarketPosition  <-- (Run all of core functions and enrich []dbpm.DBPMMarketPosition, adding more information)
    |
    V
[ API response layer: append Value, format for HTTP ]

```