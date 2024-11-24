package dbpm

import (
	"log"
	"math"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/logging"
	"socialpredict/models"
	"socialpredict/setup"
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

// appConfig holds the loaded application configuration accessible within the package
var appConfig *setup.EconomicConfig

func init() {
	var err error
	appConfig, err = setup.LoadEconomicsConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
}

// DivideUpMarketPoolSharesDBPM divides the market pool into YES and NO pools based on the resolution probability.
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func DivideUpMarketPoolSharesDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) (int64, int64) {
	if len(probabilityChanges) == 0 {
		return 0, 0
	}

	// Get the last probability change, which is the resolution probability
	currentProbability := probabilityChanges[len(probabilityChanges)-1].Probability

	// Get the total share pool as a float for precision
	// Do not include the initial market subsidization in volume until market hits final resolution
	totalSharePool := float64(marketmath.GetMarketVolume(bets))

	// Initial condition, shares set to zero
	yesShares := int64(0)
	noShares := int64(0)

	// Check case where there is a single share, implying one bet
	if marketmath.GetMarketVolume(bets) == 1 {
		singleShareDirection := singleShareYesNoAllocator(bets, logging.DefaultLogger{})

		if singleShareDirection == "YES" {
			yesShares = 1
		} else {
			noShares = 1
		}
	} else {
		// Calculate YES and NO pools using floating-point arithmetic
		yesShares = int64(math.Round(totalSharePool * currentProbability))
		noShares = int64(math.Round(totalSharePool * (1 - currentProbability)))
	}

	// Return calculated shares
	return yesShares, noShares
}

// CalculateCoursePayoutsDBPM calculates the course payout for each bet in the market,
// separating the payouts for YES and NO outcomes.
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func CalculateCoursePayoutsDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []CourseBetPayout {
	if len(probabilityChanges) == 0 {
		return nil
	}

	var coursePayouts []CourseBetPayout

	// Get the current (final) probability for the market
	currentProbability := probabilityChanges[len(probabilityChanges)-1].Probability

	// Iterate over each bet to calculate its course payout
	for i, bet := range bets {
		// Probability at which the bet was placed is the bet index+1
		// The probability index is always the length of the bet index+1 because of the initial probability
		betProbabilityAtTimePlaced := probabilityChanges[i+1].Probability

		coursePaymentForBet := math.Abs(currentProbability-betProbabilityAtTimePlaced) * float64(bet.Amount)

		// Append the calculated payout to the result
		coursePayouts = append(coursePayouts, CourseBetPayout{Payout: coursePaymentForBet, Outcome: bet.Outcome})
	}

	return coursePayouts
}

// F_YES calculates the normalization factor for "YES" by dividing the total stake by the cumulative payout for "YES".
// F_NO calculates the normalization factor for "NO" by dividing the total stake by the cumulative payout for "NO".
// Return absolute values of normalization factors to ensure non-negative values for further calculations.
func CalculateNormalizationFactorsDBPM(yesShares int64, noShares int64, coursePayouts []CourseBetPayout) (float64, float64) {
	var yesNormalizationFactor, noNormalizationFactor float64
	var yesCoursePayoutsSum, noCoursePayoutsSum float64

	// Iterate over coursePayouts to sum payouts based on outcome
	for _, payout := range coursePayouts {
		if payout.Outcome == "YES" {
			yesCoursePayoutsSum += payout.Payout
		} else if payout.Outcome == "NO" {
			noCoursePayoutsSum += payout.Payout
		}
	}

	// Calculate normalization factor for YES
	if yesCoursePayoutsSum > 0 {
		yesNormalizationFactor = float64(yesShares) / yesCoursePayoutsSum
	} else {
		yesNormalizationFactor = 0
	}

	// Calculate normalization factor for NO
	if noCoursePayoutsSum > 0 {
		noNormalizationFactor = float64(noShares) / noCoursePayoutsSum
	} else {
		noNormalizationFactor = 0
	}

	return math.Abs(yesNormalizationFactor), math.Abs(noNormalizationFactor)
}

// CalculateFinalPayouts calculates the final payouts for each bet, adjusted by normalization factors.
func CalculateScaledPayoutsDBPM(allBetsOnMarket []models.Bet, coursePayouts []CourseBetPayout, yesNormalizationFactor, noNormalizationFactor float64) []int64 {
	scaledPayouts := make([]int64, len(allBetsOnMarket))

	for i, payout := range coursePayouts {
		var scaledPayout float64
		if payout.Outcome == "YES" {
			scaledPayout = payout.Payout * yesNormalizationFactor
		} else if payout.Outcome == "NO" {
			scaledPayout = payout.Payout * noNormalizationFactor
		}

		scaledPayouts[i] = int64(math.Round(scaledPayout))
	}

	return scaledPayouts
}

// calculateExcess determines the amount of credits unaccounted for by comparing calculated scaledPayouts to availablePool
func calculateExcess(bets []models.Bet, scaledPayouts []int64) int64 {
	var sumScaledPayouts int64
	for _, payout := range scaledPayouts {
		sumScaledPayouts += payout
	}
	availablePool := marketmath.GetMarketVolume(bets)
	return sumScaledPayouts - availablePool
}

// Adjust scaled payouts if excess is greater than 0
// This  should not be possible given how the preceeding pipeline works, but we adjust for it anyway.
func adjustForPositiveExcess(scaledPayouts []int64, excess int64) []int64 {
	for excess > 0 {
		for i := len(scaledPayouts) - 1; i >= 0; i-- {
			if scaledPayouts[i] > 0 { // Ensure we don't deduct from a zero payout
				scaledPayouts[i] -= 1 // Deduct surplus from newest
				excess -= 1           // Decrease excess
				if excess == 0 {
					break
				}
			}
		}
	}
	return scaledPayouts
}

func adjustForNegativeExcess(scaledPayouts []int64, excess int64) []int64 {
	// No adjustment needed if no payouts or excess is non-negative
	if excess >= 0 || len(scaledPayouts) == 0 {
		return scaledPayouts
	}

	numBets := int64(len(scaledPayouts)) // Total number of bets
	absoluteExcess := -excess            // Convert excess to positive for allocation

	// Calculate the base addition for each bet and the leftover remainder
	// int64 will apply floor division
	baseAddition := int64(absoluteExcess / numBets)
	totalAddition := baseAddition * numBets
	remainderAddition := absoluteExcess - totalAddition

	// Apply the base addition to all payouts
	for betIndex := range scaledPayouts {
		scaledPayouts[betIndex] += baseAddition
	}

	// Apply the remainder addition to the earliest bets
	for betIndex := int64(0); betIndex < remainderAddition; betIndex++ {
		scaledPayouts[betIndex] += 1
	}

	return scaledPayouts
}

// AdjustPayouts reconciles the additional or lacking funds from the betting pool by adjusting the payouts to past bets
func AdjustPayouts(bets []models.Bet, scaledPayouts []int64) []int64 {
	excess := calculateExcess(bets, scaledPayouts)

	if excess > 0 {
		scaledPayouts = adjustForPositiveExcess(scaledPayouts, excess)
	} else if excess < 0 {
		scaledPayouts = adjustForNegativeExcess(scaledPayouts, excess)
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

// singleShareYesNoAllocator determines the outcome of a single bet.
// Logs a fatal error and exits if the input condition (len(bets) == 1) is not met.
func singleShareYesNoAllocator(bets []models.Bet, logger logging.Logger) string {
	if len(bets) != 1 {
		logger.Fatalf("singleShareYesNoAllocator: expected len(bets) = 1, got %d", len(bets))
	}

	return bets[0].Outcome
}
