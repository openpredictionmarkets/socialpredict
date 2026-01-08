package positionsmath

import (
	marketmath "socialpredict/internal/domain/math/market"
	"socialpredict/internal/domain/math/outcomes/dbpm"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
	"sort"
	"time"
)

// holds the number of YES and NO shares owned by all users in a market
type MarketPosition struct {
	Username         string `json:"username"`
	MarketID         uint   `json:"marketId"`
	NoSharesOwned    int64  `json:"noSharesOwned"`
	YesSharesOwned   int64  `json:"yesSharesOwned"`
	Value            int64  `json:"value"`
	TotalSpent       int64  `json:"totalSpent"`       // Total amount user spent in this market
	TotalSpentInPlay int64  `json:"totalSpentInPlay"` // Amount spent in unresolved markets only
	IsResolved       bool   `json:"isResolved"`       // From market.IsResolved
	ResolutionResult string `json:"resolutionResult"` // From market.ResolutionResult
}

// UserMarketPosition holds the number of YES and NO shares owned by a user in a market.
type UserMarketPosition struct {
	NoSharesOwned    int64  `json:"noSharesOwned"`
	YesSharesOwned   int64  `json:"yesSharesOwned"`
	Value            int64  `json:"value"`
	TotalSpent       int64  `json:"totalSpent"`
	TotalSpentInPlay int64  `json:"totalSpentInPlay"`
	IsResolved       bool   `json:"isResolved"`
	ResolutionResult string `json:"resolutionResult"`
}

// MarketSnapshot captures the minimal market context needed for position calculations.
type MarketSnapshot struct {
	ID               int64
	CreatedAt        time.Time
	IsResolved       bool
	ResolutionResult string
}

// CalculateMarketPositions_WPAM_DBPM summarizes positions for a given market using WPAM/DBPM math.
func CalculateMarketPositions_WPAM_DBPM(snapshot MarketSnapshot, bets []models.Bet) ([]MarketPosition, error) {
	marketIDUint := uint(snapshot.ID)

	// Ensure bets are processed in chronological order
	sortedBets := make([]models.Bet, len(bets))
	copy(sortedBets, bets)
	sort.Slice(sortedBets, func(i, j int) bool {
		return sortedBets[i].PlacedAt.Before(sortedBets[j].PlacedAt)
	})

	// Get a timeline of probability changes for the market
	allProbabilityChangesOnMarket := wpam.CalculateMarketProbabilitiesWPAM(snapshot.CreatedAt, sortedBets)

	// Calculate the distribution of YES and NO shares based on DBPM
	S_YES, S_NO := dbpm.DivideUpMarketPoolSharesDBPM(sortedBets, allProbabilityChangesOnMarket)

	// Calculate course payout pools
	coursePayouts := dbpm.CalculateCoursePayoutsDBPM(sortedBets, allProbabilityChangesOnMarket)

	// Calculate normalization factors
	F_YES, F_NO := dbpm.CalculateNormalizationFactorsDBPM(S_YES, S_NO, coursePayouts)

	// Calculate scaled payouts
	scaledPayouts := dbpm.CalculateScaledPayoutsDBPM(sortedBets, coursePayouts, F_YES, F_NO)

	// Adjust payouts to align with the available betting pool using modularized functions
	finalPayouts := dbpm.AdjustPayouts(sortedBets, scaledPayouts)

	// Aggregate user payouts into market positions
	aggreatedPositions := dbpm.AggregateUserPayoutsDBPM(sortedBets, finalPayouts)

	// enforce all users are betting on either one side or the other, or net zero
	netPositions := dbpm.NetAggregateMarketPositions(aggreatedPositions)

	// === Add valuation logic below ===

	// Step 1: Map to positions.UserMarketPosition
	userPositionMap := make(map[string]UserMarketPosition)
	for _, p := range netPositions {
		userPositionMap[p.Username] = UserMarketPosition{
			YesSharesOwned: p.YesSharesOwned,
			NoSharesOwned:  p.NoSharesOwned,
		}
	}

	// Step 2: Get current market probability
	currentProbability := wpam.GetCurrentProbability(allProbabilityChangesOnMarket)

	// Step 3: Get total volume
	totalVolume := marketmath.GetMarketVolume(sortedBets)

	// Step 4: Determine earliest bet per user
	earliestBets := computeEarliestBets(sortedBets)

	// Step 5: Calculate valuations
	valuations, err := CalculateRoundedUserValuationsFromUserMarketPositions(
		userPositionMap,
		currentProbability,
		totalVolume,
		snapshot.IsResolved,
		snapshot.ResolutionResult,
		earliestBets,
	)
	if err != nil {
		return nil, err
	}

	// Step 6: Calculate user bet totals
	userBetTotals := make(map[string]struct {
		TotalSpent       int64
		TotalSpentInPlay int64
	})

	for _, bet := range sortedBets {
		totals := userBetTotals[bet.Username]
		totals.TotalSpent += bet.Amount
		if !snapshot.IsResolved {
			totals.TotalSpentInPlay += bet.Amount
		}
		userBetTotals[bet.Username] = totals
	}

	// Step 7: Append valuation to each MarketPosition struct
	// Convert to []positions.MarketPosition for external use
	var (
		displayPositions []MarketPosition
		seenUsers        = make(map[string]bool)
	)
	for _, p := range netPositions {
		val := valuations[p.Username]
		betTotals := userBetTotals[p.Username]
		displayPositions = append(displayPositions, MarketPosition{
			Username:         p.Username,
			MarketID:         marketIDUint,
			YesSharesOwned:   p.YesSharesOwned,
			NoSharesOwned:    p.NoSharesOwned,
			Value:            val.RoundedValue,
			TotalSpent:       betTotals.TotalSpent,
			TotalSpentInPlay: betTotals.TotalSpentInPlay,
			IsResolved:       snapshot.IsResolved,
			ResolutionResult: snapshot.ResolutionResult,
		})
		seenUsers[p.Username] = true
	}

	// Ensure every bettor appears in the output even if their net position is zero.
	for username, totals := range userBetTotals {
		if seenUsers[username] {
			continue
		}

		displayPositions = append(displayPositions, MarketPosition{
			Username:         username,
			MarketID:         marketIDUint,
			YesSharesOwned:   0,
			NoSharesOwned:    0,
			Value:            valuations[username].RoundedValue,
			TotalSpent:       totals.TotalSpent,
			TotalSpentInPlay: totals.TotalSpentInPlay,
			IsResolved:       snapshot.IsResolved,
			ResolutionResult: snapshot.ResolutionResult,
		})
	}

	return displayPositions, nil

}

func computeEarliestBets(bets []models.Bet) map[string]time.Time {
	earliest := make(map[string]time.Time)
	for _, bet := range bets {
		if existing, ok := earliest[bet.Username]; !ok || bet.PlacedAt.Before(existing) {
			earliest[bet.Username] = bet.PlacedAt
		}
	}
	return earliest
}

// CalculateMarketPositionForUser_WPAM_DBPM fetches and summarizes the position for a given user in a specific market.
func CalculateMarketPositionForUser_WPAM_DBPM(snapshot MarketSnapshot, bets []models.Bet, username string) (UserMarketPosition, error) {
	marketPositions, err := CalculateMarketPositions_WPAM_DBPM(snapshot, bets)
	if err != nil {
		return UserMarketPosition{}, err
	}

	for _, position := range marketPositions {
		if position.Username == username {
			return UserMarketPosition{
				NoSharesOwned:    position.NoSharesOwned,
				YesSharesOwned:   position.YesSharesOwned,
				Value:            position.Value,
				TotalSpent:       position.TotalSpent,
				TotalSpentInPlay: position.TotalSpentInPlay,
				IsResolved:       position.IsResolved,
				ResolutionResult: position.ResolutionResult,
			}, nil
		}
	}

	return UserMarketPosition{}, nil
}

// CalculateAllUserMarketPositions_WPAM_DBPM is deprecated. Prefer computing positions via
// CalculateMarketPositions_WPAM_DBPM after fetching market snapshots and bet histories from a repository.
