package positions

import (
	"math"
)

// UserValuationResult represents a valuation for a single user.
type UserValuationResult struct {
	Username     string
	RoundedValue int64
}

// CalculateRoundedUserValuationsFromUserMarketPositions returns integer valuations per user.
func CalculateRoundedUserValuationsFromUserMarketPositions(
	userPositions map[string]UserMarketPosition,
	currentProb float64,
	totalVolume int64,
) map[string]UserValuationResult {
	result := make(map[string]UserValuationResult)
	var totalRounded int64
	var maxUser string
	var maxFloatAbs float64

	for username, pos := range userPositions {
		var floatVal float64
		if pos.YesSharesOwned > 0 {
			floatVal = float64(pos.YesSharesOwned) * currentProb
		} else if pos.NoSharesOwned > 0 {
			floatVal = float64(pos.NoSharesOwned) * (1 - currentProb)
		}

		roundedVal := int64(math.Round(floatVal))
		result[username] = UserValuationResult{
			Username:     username,
			RoundedValue: roundedVal,
		}

		totalRounded += roundedVal
		if math.Abs(floatVal) > maxFloatAbs {
			maxFloatAbs = math.Abs(floatVal)
			maxUser = username
		}
	}

	delta := totalVolume - totalRounded
	if delta != 0 {
		r := result[maxUser]
		r.RoundedValue += delta
		result[maxUser] = r
	}

	return result
}
