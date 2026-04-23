package positionsmath

import (
	"time"

	"socialpredict/internal/domain/boundary"
	"socialpredict/internal/domain/math/outcomes/dbpm"
	"socialpredict/internal/domain/math/probabilities/wpam"
)

type defaultProbabilityProvider struct {
	calculator wpam.ProbabilityCalculator
}

type defaultNetPositionCalculator struct{}

// NewWPAMProbabilityProvider constructs a WPAM-backed probability provider with the supplied calculator.
func NewWPAMProbabilityProvider(calculator wpam.ProbabilityCalculator) ProbabilityProvider {
	return defaultProbabilityProvider{calculator: calculator}
}

func (p defaultProbabilityProvider) Calculate(createdAt time.Time, bets []boundary.Bet) []wpam.ProbabilityChange {
	calc := p.calculator
	if calc.Seeds().InitialSubsidization == 0 {
		calc = wpam.NewProbabilityCalculator(nil)
	}
	return calc.CalculateMarketProbabilitiesWPAM(createdAt, bets)
}

func (p defaultProbabilityProvider) Current(changes []wpam.ProbabilityChange) float64 {
	return wpam.GetCurrentProbability(changes)
}

func (defaultNetPositionCalculator) CalculateNetPositions(sortedBets []boundary.Bet, probabilityChanges []wpam.ProbabilityChange) []dbpm.DBPMMarketPosition {
	yesShares, noShares := dbpm.DivideUpMarketPoolSharesDBPM(sortedBets, probabilityChanges)
	coursePayouts := dbpm.CalculateCoursePayoutsDBPM(sortedBets, probabilityChanges)
	yesFactor, noFactor := dbpm.CalculateNormalizationFactorsDBPM(yesShares, noShares, coursePayouts)
	scaledPayouts := dbpm.CalculateScaledPayoutsDBPM(sortedBets, coursePayouts, yesFactor, noFactor)
	finalPayouts := dbpm.AdjustPayouts(sortedBets, scaledPayouts)
	aggregatedPositions := dbpm.AggregateUserPayoutsDBPM(sortedBets, finalPayouts)
	return dbpm.NetAggregateMarketPositions(aggregatedPositions)
}
