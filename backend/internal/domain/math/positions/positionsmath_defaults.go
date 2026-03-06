package positionsmath

import (
	"time"

	"socialpredict/internal/domain/math/outcomes/dbpm"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
)

type defaultProbabilityProvider struct {
	calculator wpam.ProbabilityCalculator
}

// NewWPAMProbabilityProvider constructs a WPAM-backed probability provider with the supplied calculator.
func NewWPAMProbabilityProvider(calculator wpam.ProbabilityCalculator) ProbabilityProvider {
	return defaultProbabilityProvider{calculator: calculator}
}

func (p defaultProbabilityProvider) Calculate(createdAt time.Time, bets []models.Bet) []wpam.ProbabilityChange {
	calc := p.calculator
	if calc.Seeds().InitialSubsidization == 0 {
		calc = wpam.NewProbabilityCalculator(nil)
	}
	return calc.CalculateMarketProbabilitiesWPAM(createdAt, bets)
}

func (p defaultProbabilityProvider) Current(changes []wpam.ProbabilityChange) float64 {
	return wpam.GetCurrentProbability(changes)
}

type defaultPayoutModel struct{}

func (defaultPayoutModel) DivideShares(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) (int64, int64) {
	return dbpm.DivideUpMarketPoolSharesDBPM(bets, probabilityChanges)
}

func (defaultPayoutModel) CoursePayouts(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []dbpm.CourseBetPayout {
	return dbpm.CalculateCoursePayoutsDBPM(bets, probabilityChanges)
}

func (defaultPayoutModel) NormalizationFactors(yesShares, noShares int64, coursePayouts []dbpm.CourseBetPayout) (float64, float64) {
	return dbpm.CalculateNormalizationFactorsDBPM(yesShares, noShares, coursePayouts)
}

func (defaultPayoutModel) ScaledPayouts(bets []models.Bet, coursePayouts []dbpm.CourseBetPayout, yesFactor, noFactor float64) []int64 {
	return dbpm.CalculateScaledPayoutsDBPM(bets, coursePayouts, yesFactor, noFactor)
}

func (defaultPayoutModel) AdjustFinalPayouts(bets []models.Bet, scaledPayouts []int64) []int64 {
	return dbpm.AdjustPayouts(bets, scaledPayouts)
}

func (defaultPayoutModel) AggregateUserPayouts(bets []models.Bet, finalPayouts []int64) []dbpm.DBPMMarketPosition {
	return dbpm.AggregateUserPayoutsDBPM(bets, finalPayouts)
}

func (defaultPayoutModel) NetAggregateMarketPositions(positions []dbpm.DBPMMarketPosition) []dbpm.DBPMMarketPosition {
	return dbpm.NetAggregateMarketPositions(positions)
}
