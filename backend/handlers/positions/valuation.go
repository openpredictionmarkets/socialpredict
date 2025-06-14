package positions

import (
	"math"
)

type UserValuationResult struct {
	Username     string
	RoundedValue int64
}

func CalculateRoundedUserValuationsFromUserMarketPositions(
	userPositions map[string]UserMarketPosition,
	currentProbability float64,
	totalVolume int64,
) map[string]UserValuationResult {
	result := make(map[string]UserValuationResult)
	floatVals := make(map[string]float64)
	for username, pos := range userPositions {
		var floatVal float64
		if pos.YesSharesOwned > 0 {
			floatVal = float64(pos.YesSharesOwned) * currentProbability
		} else if pos.NoSharesOwned > 0 {
			floatVal = float64(pos.NoSharesOwned) * (1 - currentProbability)
		}
		floatVals[username] = floatVal
		roundedVal := int64(math.Round(floatVal))
		result[username] = UserValuationResult{
			Username:     username,
			RoundedValue: roundedVal,
		}
	}
	return AdjustUserValuationsToMarketVolume(result, totalVolume, floatVals)
}

// AdjustUserValuationsToMarketVolume ensures the sum of user valuations matches the market volume.
// It adds or subtracts the delta to the user with the largest float valuation (by absolute value).
func AdjustUserValuationsToMarketVolume(
	userVals map[string]UserValuationResult,
	targetTotal int64,
	floatVals map[string]float64,
) map[string]UserValuationResult {
	// Compute sum of rounded
	var totalRounded int64
	var maxUser string
	var maxFloatAbs float64
	for username, uv := range userVals {
		totalRounded += uv.RoundedValue
		if math.Abs(floatVals[username]) > maxFloatAbs {
			maxFloatAbs = math.Abs(floatVals[username])
			maxUser = username
		}
	}
	delta := targetTotal - totalRounded
	if delta != 0 {
		r := userVals[maxUser]
		r.RoundedValue += delta
		userVals[maxUser] = r
	}
	return userVals
}
