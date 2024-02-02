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

func CalculateNormalizationFactorsDBPM(S_YES uint, C_YES float64, S_NO uint, C_NO float64) (float64, float64) {
    var F_YES, F_NO float64

    // Calculate normalization factor for YES
    if C_YES > 0 {
		// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
        // minimum used to prevent balooning payouts edge case
		F_YES = min(1, float64(S_YES)/C_YES)
    } else {
        F_YES = 1 // Default to 1 if C_YES is 0 to avoid division by zero
    }

    // Calculate normalization factor for NO
    if C_NO > 0 {
		// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
		// minimum used to prevent balooning payouts edge case
        F_NO = min(1, float64(S_NO)/C_NO)
    } else {
        F_NO = 1 // Default to 1 if C_NO is 0 to avoid division by zero
    }

    return F_YES, F_NO
}

// min returns the minimum of two float64 values.
func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

// CalculateFinalPayouts calculates the final payouts for each bet, adjusted by normalization factors.
func CalculateFinalPayoutsDBPM(bets []models.Bet, F_YES, F_NO float64, C_YES, C_NO []float64) ([]uint) {
    finalPayouts := make([]uint, len(bets))

    for i, bet := range bets {
        var payout float64

        // Check the outcome of the bet and calculate the payout accordingly
        if bet.Outcome == "YES" {
            payout = C_YES[i] * F_YES
        } else if bet.Outcome == "NO" {
            payout = C_NO[i] * F_NO
        }

        // Convert the payout to uint, rounding as necessary
        finalPayouts[i] = uint(math.Round(payout))
    }

    return finalPayouts
}


// AggregateUserPayouts aggregates YES and NO payouts for each user.
func AggregateUserPayoutsDBPM(bets []models.Bet, finalPayouts []uint) []models.MarketPosition {
    userPayouts := make(map[string]*models.MarketPosition)

    for i, bet := range bets {
        payout := finalPayouts[i]

        // Initialize the user's market position if it doesn't exist
        if _, exists := userPayouts[bet.Username]; !exists {
            userPayouts[bet.Username] = &models.MarketPosition{Username: bet.Username}
        }

        // Aggregate payouts based on the outcome
        if bet.Outcome == "YES" {
            userPayouts[bet.Username].YesSharesOwned += payout
        } else if bet.Outcome == "NO" {
            userPayouts[bet.Username].NoSharesOwned += payout
        }
    }

    // Convert map to slice for output
    var positions []models.MarketPosition
    for _, pos := range userPayouts {
        positions = append(positions, *pos)
    }

    return positions
}