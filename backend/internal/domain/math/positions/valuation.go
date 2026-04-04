package positionsmath

import (
	"math"
	"time"
)

type valuationRule struct {
	matches func(UserMarketPosition) bool
	value   func(UserMarketPosition, float64) float64
}

type resolutionRule struct {
	finalProbability float64
	resolvedValue    func(UserMarketPosition) float64
}

var unresolvedValuationRules = []valuationRule{
	{
		matches: func(position UserMarketPosition) bool { return position.YesSharesOwned > 0 },
		value: func(position UserMarketPosition, probability float64) float64 {
			return float64(position.YesSharesOwned) * probability
		},
	},
	{
		matches: func(position UserMarketPosition) bool { return position.NoSharesOwned > 0 },
		value: func(position UserMarketPosition, probability float64) float64 {
			return float64(position.NoSharesOwned) * (1 - probability)
		},
	},
}

var resolvedValuationRules = map[string]resolutionRule{
	positionTypeYes: {
		finalProbability: 1.0,
		resolvedValue: func(position UserMarketPosition) float64 {
			return float64(position.YesSharesOwned)
		},
	},
	positionTypeNo: {
		finalProbability: 0.0,
		resolvedValue: func(position UserMarketPosition) float64 {
			return float64(position.NoSharesOwned)
		},
	},
}

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
		roundedVal := int64(math.Round(calculatePositionValue(pos, finalProb, isResolved, resolutionResult)))

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
	if !isResolved {
		return currentProbability
	}
	rule, ok := resolvedValuationRules[resolutionResult]
	if !ok {
		return currentProbability
	}
	return rule.finalProbability
}

func calculatePositionValue(
	position UserMarketPosition,
	finalProbability float64,
	isResolved bool,
	resolutionResult string,
) float64 {
	if isResolved {
		if rule, ok := resolvedValuationRules[resolutionResult]; ok {
			return rule.resolvedValue(position)
		}
		return 0
	}

	for _, rule := range unresolvedValuationRules {
		if rule.matches(position) {
			return rule.value(position, finalProbability)
		}
	}
	return 0
}
