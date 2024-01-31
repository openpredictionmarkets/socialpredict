package dbpm

import (
	"math"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/models"
)

// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func CalculateTotalSharesDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) uint {
	var totalShares uint = 0

	// Start from 1 since the first entry in probabilityChanges is just the initial condition
	for i, bet := range bets {
		// Since probabilityChanges has one extra entry at the beginning, use i+1
		probability := probabilityChanges[i+1].Probability

		if probability > 0 {
			shares := bet.Amount / probability
			totalShares += uint(shares)
		}
	}

	return totalShares
}

// DivideUpMarketPoolShares divides the market pool into YES and NO pools based on the resolution probability.
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func DivideUpMarketPoolSharesDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) (uint, uint) {
	if len(probabilityChanges) == 0 {
		return 0, 0
	}

	// Get the last probability change which is the resolution probability
	R := probabilityChanges[len(probabilityChanges)-1].Probability

	// Get the total share pool S as a float for precision
	S := float64(marketmath.GetMarketVolume(bets))

	// Calculate YES and NO pools using floating-point arithmetic
	// Note, fractional shares will be lost here
	S_YES := math.Round(S * R)
	S_NO := math.Round(S * (1 - R))

	// Convert results to uint
	return uint(S_YES), uint(S_NO)
}

// CalculateCoursePayoutsDBPM calculates the course payout for each bet in the market,
// separating the payouts for YES and NO outcomes.
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func CalculateCoursePayoutsDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) ([]float64, []float64) {
	if len(probabilityChanges) == 0 {
		return nil, nil
	}

	// Get the last probability change which is the resolution probability
	R := probabilityChanges[len(probabilityChanges)-1].Probability

	yesCoursePayouts := make([]float64, 0)
	noCoursePayouts := make([]float64, 0)

	for i, bet := range bets {
		// Assuming that the index of the bet corresponds to the index in probabilityChanges
		p_i := probabilityChanges[i].Probability

		// Calculate the reward factor (d_i)
		d_i := math.Abs(R - p_i)

		// Calculate the course payout (C_i)
		C_i := d_i * bet.Amount

		// Separate YES and NO course payouts
		if bet.Outcome == "YES" {
			yesCoursePayouts = append(yesCoursePayouts, C_i)
		} else if bet.Outcome == "NO" {
			noCoursePayouts = append(noCoursePayouts, C_i)
		}
	}

	return yesCoursePayouts, noCoursePayouts
}
