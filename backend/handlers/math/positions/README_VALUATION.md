| Feature                              | `valuation.go`                              | `resolvemarketpayouts.go`                           |
| ------------------------------------ | ------------------------------------------- | --------------------------------------------------- |
| Uses user share positions            | ✅ (`UserMarketPosition`)                    | ✅ (`DBPMMarketPosition → MarketPosition`)           |
| Resolves to YES/NO/probability       | ✅ (`isResolved` + `resolutionResult`)       | ❌ (currently only YES/NO supported)                 |
| Computes floating-point value first  | ✅ (`floatVal = shares * prob`)              | ✅ (`payout = round(shares / total * volume)`)       |
| Rounds final result to `int64`       | ✅ (`math.Round`)                            | ✅ (`math.Round`)                                    |
| Applies rounding adjustment/tiebreak | ✅ (`Adjust...()` with timestamp-based sort) | ✅ (simpler: gives delta to largest position holder) |
| Returns per-user valuation/payout    | ✅ (`UserValuationResult`)                   | ✅ (`UpdateUserBalance(...)`)                        |
| Uses total volume                    | ✅ (normalizes valuation to `marketVolume`)  | ✅ (used to fund payout pool directly)               |

---

Step-by-step strategy:
Start from resolvemarketpayouts.go

Extract reusable computation steps into helper functions without changing behavior.

E.g., extract share-weighted value computation, rounding logic, and adjustment logic into isolated funcs.

Unit test payout behavior

Ensure AllocateWinningSharePool() works identically after refactor.

Snapshot-based test with fixed inputs and expected balance deltas.

Create a core computation module

Place reusable functions inside handlers/math/valuation/ or handlers/math/payoutcore/.

Allow both valuation.go and resolvemarket*.go to import these.

Refactor valuation.go to use shared logic

Replace duplicated logic in CalculateRoundedUserValuationsFromUserMarketPositions() with call to shared computation module.

Keep timestamp-based adjustment for rounding delta, unless reused directly.

Verify correctness

Confirm valuation results match payout results when resolution is YES or NO.

Also test behavior when resolving to a probability between 0.0 and 1.0.

💡 Future Cleanups
Move valuation.go out of positions/ and into math/valuation/.

Treat valuation as a view-layer for prediction markets:

Used in UI/API to preview potential or current value.

Not responsible for state mutation or transfers.

Final shared structure:

bash
Copy
Edit
handlers/
  math/
    payout/                  # Finalize payouts to users
    valuation/               # Compute user value at any point
    payoutcore.go           # Shared math used by both
