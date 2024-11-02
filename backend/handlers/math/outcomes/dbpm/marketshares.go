package dbpm

import (
	"fmt"
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

// DivideUpMarketPoolShares divides the market pool into YES and NO pools based on the resolution probability.
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func DivideUpMarketPoolSharesDBPM(bets []models.Bet, probabilityChanges []wpam.ProbabilityChange) (int64, int64) {
	if len(probabilityChanges) == 0 {
		return 0, 0
	}

	// Get the last probability change which is the resolution probability
	R := probabilityChanges[len(probabilityChanges)-1].Probability

	// Get the total share pool S as a float for precision
	// Include the initial market subsidization in the displayed volume
	S := float64(marketmath.GetMarketVolume(bets) + appConfig.Economics.MarketCreation.InitialMarketSubsidization)

	// initial condition, shares set to zero
	S_YES := int64(0)
	S_NO := int64(0)

	// Check case where there is single share, output
	// Calculate YES and NO pools using floating-point arithmetic
	// Note, fractional shares will be lost here
	if marketmath.GetMarketVolume(bets) == 1 {
		singleShareDirection := SingleShareYesNoAllocator(bets)
		if singleShareDirection == "YES" {
			S_YES = 1
		} else {
			S_NO = 1
		}
	} else {
		S_YES = int64(math.Round(S * R))
		S_NO = int64(math.Round(S * (1 - R)))
	}

	// Convert results to int64, rounding in predictable way
	return S_YES, S_NO
}

// Returns "YES", "NO", or "", indicating the outcome of the single share or no outcome if shares > 1.
func SingleShareYesNoAllocator(bets []models.Bet) string {
	total := int64(0)
	for _, bet := range bets {
		logging.LogMsg(fmt.Sprintf("Bet Outcome: %s", bet.Outcome))
		logging.LogMsg(fmt.Sprintf("Bet Amount: %d", bet.Amount))
		if bet.Outcome == "YES" {
			total += bet.Amount
		} else if bet.Outcome == "NO" {
			total -= bet.Amount
		}
	}

	if total > 0 {
		return "YES"
	} else if total < 0 {
		return "NO"
	} else {
		return "" // indeterminite
	}
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

	// Calculate normalization factor for YES
	if C_YES_SUM > 0 {
		F_YES = float64(S_YES) / C_YES_SUM
	} else {
		F_YES = 0
	}

	// Calculate normalization factor for NO
	if C_NO_SUM > 0 {
		F_NO = float64(S_NO) / C_NO_SUM
	} else {
		F_NO = 0
	}

	return math.Abs(F_YES), math.Abs(F_NO)
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
				scaledPayouts[i] -= 1 // deduct surplus from newest
				excess -= 1           // decrease excess until we get to zero
				if excess == 0 {
					break
				}
			}
		}
	}

	// Loop to add from oldest to newest until there's no excess
	for excess < 0 {
		for i := 0; i < len(scaledPayouts); i++ { // Iterate from the beginning to the end
			scaledPayouts[i] += 1 // Add surplus to oldest
			excess += 1           // Increment excess until we get to zero
			if excess == 0 {
				break
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
