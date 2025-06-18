package payout

import "math"

// MarketPosition is a simplified struct for payout calculation.
// (This can mirror your positions.MarketPosition if you want.)
type MarketPosition struct {
	Username       string
	YesSharesOwned int64
	NoSharesOwned  int64
}

// PayoutResult holds the result per user.
type PayoutResult struct {
	Username string
	Payout   int64
	Shares   int64
}

// CalculatePayouts distributes the totalVolume to the winning side according to resolution ("YES"/"NO").
// Returns a slice of payout results (Username, payout, shares) and the total paid out.
func CalculatePayouts(
	resolution string,
	positions []MarketPosition,
	totalVolume int64,
) ([]PayoutResult, int64) {
	var results []PayoutResult
	var totalWinningShares int64

	// 1. Filter for the winning side & count total shares
	for _, pos := range positions {
		var userShares int64
		if resolution == "YES" {
			userShares = pos.YesSharesOwned
		} else if resolution == "NO" {
			userShares = pos.NoSharesOwned
		}
		if userShares > 0 {
			results = append(results, PayoutResult{
				Username: pos.Username,
				Shares:   userShares,
				Payout:   0, // To be filled in next step
			})
			totalWinningShares += userShares
		}
	}

	if totalWinningShares == 0 {
		return nil, 0 // nothing to pay out
	}

	// 2. Allocate payouts, rounding, track sum
	var sumPayouts int64
	maxShares := int64(0)
	topUserIndex := 0

	for i := range results {
		r := &results[i]
		r.Payout = int64(math.Round(float64(r.Shares) / float64(totalWinningShares) * float64(totalVolume)))
		sumPayouts += r.Payout

		if r.Shares > maxShares {
			maxShares = r.Shares
			topUserIndex = i
		}
	}

	// 3. Assign any remainder to the largest holder (for rounding correction)
	delta := totalVolume - sumPayouts
	if delta > 0 && len(results) > 0 {
		results[topUserIndex].Payout += delta
	}

	return results, sumPayouts + delta
}
