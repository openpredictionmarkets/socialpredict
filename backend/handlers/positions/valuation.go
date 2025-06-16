package positions

import (
	"math"

	"gorm.io/gorm"
)

type UserValuationResult struct {
	Username     string
	RoundedValue int64
}

func CalculateRoundedUserValuationsFromUserMarketPositions(
	db *gorm.DB,
	marketID uint,
	userPositions map[string]UserMarketPosition,
	currentProbability float64,
	totalVolume int64,
) (map[string]UserValuationResult, error) {
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
	// Now call the improved adjustment function
	adjusted, err := AdjustUserValuationsToMarketVolume(db, marketID, result, totalVolume)
	if err != nil {
		return nil, err
	}
	return adjusted, nil
}
