package positionsmath

import (
	"fmt"
	"math"
	"time"
)

type UserValuationResult struct {
	Username     string
	RoundedValue int64
}

func CalculateRoundedUserValuationsFromUserMarketPositions(
	userPositions map[string]UserMarketPosition,
	currentProbability float64,
	totalVolume int64,
	isResolved bool,
	resolutionResult string,
	earliestBets map[string]time.Time,
) (map[string]UserValuationResult, error) {
	result := make(map[string]UserValuationResult)
	var finalProb float64

	finalProb = getFinalProbabilityFromMarketModel(currentProbability, isResolved, resolutionResult)

	for username, pos := range userPositions {
		var floatVal float64

		if isResolved {
			floatVal = 0
			if resolutionResult == "YES" {
				floatVal = float64(pos.YesSharesOwned)
			} else if resolutionResult == "NO" {
				floatVal = float64(pos.NoSharesOwned)
			}
		} else {
			if pos.YesSharesOwned > 0 {
				floatVal = float64(pos.YesSharesOwned) * finalProb
			} else if pos.NoSharesOwned > 0 {
				floatVal = float64(pos.NoSharesOwned) * (1 - finalProb)
			}
		}

		fmt.Printf("user=%s YES=%d NO=%d isResolved=%v result=%s val=%v\n",
			username, pos.YesSharesOwned, pos.NoSharesOwned, isResolved, resolutionResult, floatVal)

		roundedVal := int64(math.Round(floatVal))

		result[username] = UserValuationResult{
			Username:     username,
			RoundedValue: roundedVal,
		}
	}

	adjusted := AdjustUserValuationsToMarketVolume(result, earliestBets, totalVolume)
	return adjusted, nil
}

func getFinalProbabilityFromMarketModel(
	currentProbability float64,
	isResolved bool,
	resolutionResult string,
) float64 {
	if isResolved {
		switch resolutionResult {
		case "YES":
			return 1.0
		case "NO":
			return 0.0
		default:
			return currentProbability
		}
	}
	return currentProbability
}
