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
	isResolved bool,
	resolutionResult string,
) (map[string]UserValuationResult, error) {
	result := make(map[string]UserValuationResult)
	var finalProb float64

	if isResolved {
		switch resolutionResult {
		case "YES":
			finalProb = 1.0
		case "NO":
			finalProb = 0.0
		default:
			finalProb = currentProbability // fallback
		}
	} else {
		finalProb = currentProbability
	}

	for username, pos := range userPositions {
		var floatVal float64
		if pos.YesSharesOwned > 0 {
			floatVal = float64(pos.YesSharesOwned) * finalProb
		} else if pos.NoSharesOwned > 0 {
			floatVal = float64(pos.NoSharesOwned) * (1 - finalProb)
		}
		roundedVal := int64(math.Round(floatVal))
		result[username] = UserValuationResult{
			Username:     username,
			RoundedValue: roundedVal,
		}
	}

	adjusted, err := AdjustUserValuationsToMarketVolume(db, marketID, result, totalVolume)
	if err != nil {
		return nil, err
	}
	return adjusted, nil
}
