package dbpm

import (
	"math"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/logging"
	"socialpredict/models"
)

// holds betting payout information
type CourseBetPayout struct {
	Payout  float64
	Outcome string
}

type MarketPosition struct {
	Username       string
	NoSharesOwned  int64
	YesSharesOwned int64
}

// DivideUpMarketPoolShares divides the market pool into YES and NO pools based on the resolution probability.
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func DivideUpMarketPoolSharesDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) (int64, int64) {
	if len(probabilityChanges) == 0 {
		return 0, 0
	}

	// Get the last probability change which is the resolution probability
	R := probabilityChanges[len(probabilityChanges)-1].Probability
	logging.LogAnyType(R, "R")

	// Get the total share pool S as a float for precision
	S := float64(marketmath.GetMarketVolume(bets))
	logging.LogAnyType(S, "S")

	// Calculate YES and NO pools using floating-point arithmetic
	// Note, fractional shares will be lost here
	S_YES := math.Round(S * R)
	logging.LogAnyType(S_YES, "S_YES")
	S_NO := math.Round(S * (1 - R))
	logging.LogAnyType(S_NO, "S_NO")

	// Convert results to int64, rounding in predictable way
	return int64(S_YES), int64(S_NO)
}

// CalculateCoursePayoutsDBPM calculates the course payout for each bet in the market,
// separating the payouts for YES and NO outcomes.
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func CalculateCoursePayoutsDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []CourseBetPayout {
	if len(probabilityChanges) == 0 {
		return nil
	}

	var coursePayouts []CourseBetPayout

	// Get the last probability change which is the resolution probability
	R := probabilityChanges[len(probabilityChanges)-1].Probability

	for i, bet := range bets {
		// Distance to last (current) probability times bet amount
		C_i := math.Abs(R-probabilityChanges[i].Probability) * float64(bet.Amount)
		coursePayouts = append(coursePayouts, CourseBetPayout{Payout: C_i, Outcome: bet.Outcome})
	}

	logging.LogAnyType(coursePayouts, "coursePayouts")

	return coursePayouts
}

func CalculateNormalizationFactorsDBPM(S_YES int64, S_NO int64, coursePayouts []CourseBetPayout) (float64, float64) {
	var F_YES, F_NO float64
	var C_YES_SUM, C_NO_SUM float64

	// Iterate over coursePayouts to sum payouts based on outcome
	for _, payout := range coursePayouts {
		if payout.Outcome == "YES" {
			C_YES_SUM += payout.Payout
		} else if payout.Outcome == "NO" {
			C_NO_SUM += payout.Payout
		}
	}

	logging.LogAnyType(C_YES_SUM, "C_YES_SUM")
	logging.LogAnyType(C_NO_SUM, "C_NO_SUM")

	// Calculate normalization factor for YES
	if C_YES_SUM > 0 {
		F_YES = float64(S_YES) / C_YES_SUM
	} else {
		F_YES = 1 // Avoid division by zero
	}

	// Calculate normalization factor for NO
	if C_NO_SUM > 0 {
		F_NO = float64(S_NO) / C_NO_SUM
	} else {
		F_NO = 1 // Avoid division by zero
	}

	logging.LogAnyType(F_YES, "F_YES")
	logging.LogAnyType(F_NO, "F_NO")

	return F_YES, F_NO
}

// max returns the minimum of two float64 values.
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// CalculateFinalPayouts calculates the final payouts for each bet, adjusted by normalization factors.
func CalculateScaledPayoutsDBPM(allBetsOnMarket []models.Bet, coursePayouts []CourseBetPayout, F_YES, F_NO float64) []int64 {
	scaledPayouts := make([]int64, len(allBetsOnMarket))

	for i, payout := range coursePayouts {
		var scaledPayout float64
		if payout.Outcome == "YES" {
			scaledPayout = payout.Payout * F_YES
		} else if payout.Outcome == "NO" {
			scaledPayout = payout.Payout * F_NO
		}

		scaledPayouts[i] = int64(math.Round(scaledPayout))
	}

	logging.LogAnyType(scaledPayouts, "scaledPayouts")

	return scaledPayouts
}

// adjust payouts to account for case where calculated payouts > available
func AdjustPayoutsFromNewest(bets []models.Bet, scaledPayouts []int64) []int64 {
	// Calculate the sum of scaledPayouts
	var sumScaledPayouts int64
	for _, payout := range scaledPayouts {
		sumScaledPayouts += payout
	}

	availablePool := marketmath.GetMarketVolume(bets)

	// Determine the excess amount
	excess := sumScaledPayouts - availablePool

	// Loop to deduct from newest to oldest until there's no excess
	for excess > 0 {
		for i := len(scaledPayouts) - 1; i >= 0; i-- {
			if scaledPayouts[i] > 0 { // Ensure we don't deduct from a zero payout
				scaledPayouts[i] -= 1
				excess -= 1
				if excess == 0 {
					break
				}
			}
		}
	}

	return scaledPayouts
}

// AggregateUserPayouts aggregates YES and NO payouts for each user.
func AggregateUserPayoutsDBPM(bets []models.Bet, finalPayouts []int64) []MarketPosition {
	userPayouts := make(map[string]*MarketPosition)

	for i, bet := range bets {
		payout := finalPayouts[i]

		// Initialize the user's market position if it doesn't exist
		if _, exists := userPayouts[bet.Username]; !exists {
			userPayouts[bet.Username] = &MarketPosition{Username: bet.Username}
		}

		// Aggregate payouts based on the outcome
		if bet.Outcome == "YES" {
			userPayouts[bet.Username].YesSharesOwned += payout
		} else if bet.Outcome == "NO" {
			userPayouts[bet.Username].NoSharesOwned += payout
		}
	}

	// Convert map to slice for output
	var positions []MarketPosition
	for _, pos := range userPayouts {
		// Check and adjust negative shares to 0
		if pos.YesSharesOwned < 0 {
			pos.YesSharesOwned = 0
		}
		if pos.NoSharesOwned < 0 {
			pos.NoSharesOwned = 0
		}
		positions = append(positions, *pos)
	}

	return positions
}

// Function to normalize market positions such that for each user,
// only one of YesSharesOwned or NoSharesOwned is greater than 0,
// with the other being 0, and the value is the net difference.
func NetAggregateMarketPositions(positions []MarketPosition) []MarketPosition {
	var normalizedPositions []MarketPosition

	for _, position := range positions {
		var normalizedPosition MarketPosition
		normalizedPosition.Username = position.Username

		if position.YesSharesOwned > position.NoSharesOwned {
			normalizedPosition.YesSharesOwned = position.YesSharesOwned - position.NoSharesOwned
			normalizedPosition.NoSharesOwned = 0
		} else {
			normalizedPosition.NoSharesOwned = position.NoSharesOwned - position.YesSharesOwned
			normalizedPosition.YesSharesOwned = 0
		}

		normalizedPositions = append(normalizedPositions, normalizedPosition)
	}

	return normalizedPositions
}
