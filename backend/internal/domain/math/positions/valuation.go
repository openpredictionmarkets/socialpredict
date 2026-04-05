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

type ValuationModel interface {
	FinalProbability(currentProbability float64, isResolved bool, resolutionResult string) float64
	PositionValue(position UserMarketPosition, finalProbability float64, isResolved bool, resolutionResult string) float64
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

type valuationModel struct{}

type ValuationCalculator struct {
	model    ValuationModel
	adjuster UserValuationAdjuster
}

var defaultValuationCalculator = ValuationCalculator{
	model:    valuationModel{},
	adjuster: defaultUserValuationAdjuster,
}

func CalculateRoundedUserValuationsFromUserMarketPositions(
	userPositions map[string]UserMarketPosition,
	currentProbability float64,
	totalVolume int64,
	isResolved bool,
	resolutionResult string,
	earliestBets map[string]time.Time,
) (map[string]UserValuationResult, error) {
	return defaultValuationCalculator.Calculate(
		userPositions,
		currentProbability,
		totalVolume,
		isResolved,
		resolutionResult,
		earliestBets,
	)
}

func (c ValuationCalculator) Calculate(
	userPositions map[string]UserMarketPosition,
	currentProbability float64,
	totalVolume int64,
	isResolved bool,
	resolutionResult string,
	earliestBets map[string]time.Time,
) (map[string]UserValuationResult, error) {
	c = c.withDefaults()
	result := make(map[string]UserValuationResult)
	finalProb := c.model.FinalProbability(currentProbability, isResolved, resolutionResult)

	for username, pos := range userPositions {
		roundedVal := int64(math.Round(c.model.PositionValue(pos, finalProb, isResolved, resolutionResult)))

		result[username] = UserValuationResult{
			Username:     username,
			RoundedValue: roundedVal,
		}
	}

	adjusted := c.adjuster.Adjust(result, earliestBets, totalVolume)
	return adjusted, nil
}

func (c ValuationCalculator) withDefaults() ValuationCalculator {
	if c.model == nil {
		c.model = valuationModel{}
	}
	if c.adjuster == nil {
		c.adjuster = defaultUserValuationAdjuster
	}
	return c
}

func getFinalProbabilityFromMarketModel(
	currentProbability float64,
	isResolved bool,
	resolutionResult string,
) float64 {
	return valuationModel{}.FinalProbability(currentProbability, isResolved, resolutionResult)
}

func (valuationModel) FinalProbability(currentProbability float64, isResolved bool, resolutionResult string) float64 {
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
	return valuationModel{}.PositionValue(position, finalProbability, isResolved, resolutionResult)
}

func (valuationModel) PositionValue(
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
