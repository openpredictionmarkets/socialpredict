package positionsmath

import (
	"socialpredict/errors"
	"socialpredict/handlers/marketpublicresponse"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"
	"strconv"

	"gorm.io/gorm"
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
	NoSharesOwned  int64 `json:"noSharesOwned"`
	YesSharesOwned int64 `json:"yesSharesOwned"`
	Value          int64 `json:"value"`
}

// FetchMarketPositions fetches and summarizes positions for a given market.
// It returns a slice of MarketPosition as defined in the dbpm package.
func CalculateMarketPositions_WPAM_DBPM(db *gorm.DB, marketIdStr string) ([]MarketPosition, error) {

	// marketIDUint for needed areas
	marketIDUint64, err := strconv.ParseUint(marketIdStr, 10, 64)
	if errors.ErrorLogger(err, "Can't convert string.") {
		return nil, err
	}

	marketIDUint := uint(marketIDUint64)

	// Assuming a function to fetch the market creation time
	publicResponseMarket, err := marketpublicresponse.GetPublicResponseMarketByID(db, marketIdStr)
	if errors.ErrorLogger(err, "Can't convert marketIdStr to publicResponseMarket.") {
		return nil, err
	}

	// Fetch bets for the market
	var allBetsOnMarket []models.Bet
	allBetsOnMarket = tradingdata.GetBetsForMarket(db, marketIDUint)

	// Get a timeline of probability changes for the market
	allProbabilityChangesOnMarket := wpam.CalculateMarketProbabilitiesWPAM(publicResponseMarket.CreatedAt, allBetsOnMarket)

	// Calculate the distribution of YES and NO shares based on DBPM
	S_YES, S_NO := dbpm.DivideUpMarketPoolSharesDBPM(allBetsOnMarket, allProbabilityChangesOnMarket)

	// Calculate course payout pools
	coursePayouts := dbpm.CalculateCoursePayoutsDBPM(allBetsOnMarket, allProbabilityChangesOnMarket)

	// Calculate normalization factors
	F_YES, F_NO := dbpm.CalculateNormalizationFactorsDBPM(S_YES, S_NO, coursePayouts)

	// Calculate scaled payouts
	scaledPayouts := dbpm.CalculateScaledPayoutsDBPM(allBetsOnMarket, coursePayouts, F_YES, F_NO)

	// Adjust payouts to align with the available betting pool using modularized functions
	finalPayouts := dbpm.AdjustPayouts(allBetsOnMarket, scaledPayouts)

	// Aggregate user payouts into market positions
	aggreatedPositions := dbpm.AggregateUserPayoutsDBPM(allBetsOnMarket, finalPayouts)

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
	totalVolume := marketmath.GetMarketVolume(allBetsOnMarket)

	// Step 4: Calculate valuations
	valuations, err := CalculateRoundedUserValuationsFromUserMarketPositions(
		db,
		marketIDUint,
		userPositionMap,
		currentProbability,
		totalVolume,
		publicResponseMarket.IsResolved,
		publicResponseMarket.ResolutionResult,
	)
	if err != nil {
		return nil, err
	}

	// Step 5: Calculate user bet totals for TotalSpent and TotalSpentInPlay
	userBetTotals := make(map[string]struct {
		TotalSpent       int64
		TotalSpentInPlay int64
	})

	for _, bet := range allBetsOnMarket {
		totals := userBetTotals[bet.Username]
		totals.TotalSpent += bet.Amount
		if !publicResponseMarket.IsResolved {
			totals.TotalSpentInPlay += bet.Amount
		}
		userBetTotals[bet.Username] = totals
	}

	// Step 6: Append valuation to each MarketPosition struct
	// Convert to []positions.MarketPosition for external use
	var displayPositions []MarketPosition
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
			IsResolved:       publicResponseMarket.IsResolved,
			ResolutionResult: publicResponseMarket.ResolutionResult,
		})
	}

	return displayPositions, nil

}

// CalculateMarketPositionForUser_WPAM_DBPM fetches and summarizes the position for a given user in a specific market.
func CalculateMarketPositionForUser_WPAM_DBPM(db *gorm.DB, marketIdStr string, username string) (UserMarketPosition, error) {
	marketPositions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIdStr)
	if err != nil {
		return UserMarketPosition{}, err
	}

	for _, position := range marketPositions {
		if position.Username == username {
			return UserMarketPosition{
				NoSharesOwned:  position.NoSharesOwned,
				YesSharesOwned: position.YesSharesOwned,
				Value:          position.Value,
			}, nil
		}
	}

	return UserMarketPosition{}, nil
}

// CalculateAllUserMarketPositions_WPAM_DBPM fetches and summarizes positions for a given user across all markets where they have bets.
// Optimized to only process markets where the user has positions (O(user_bets + unique_user_markets))
func CalculateAllUserMarketPositions_WPAM_DBPM(db *gorm.DB, username string) ([]MarketPosition, error) {
	// Step 1: Get all user bets (single query - O(user_bets))
	var userBets []models.Bet
	if err := db.Where("username = ?", username).Find(&userBets).Error; err != nil {
		return nil, err
	}

	// Step 2: Build stack of unique market IDs where user has positions
	marketIDSet := make(map[uint]bool)
	userBetsByMarket := make(map[uint][]models.Bet)

	for _, bet := range userBets {
		marketIDSet[bet.MarketID] = true
		userBetsByMarket[bet.MarketID] = append(userBetsByMarket[bet.MarketID], bet)
	}

	// Step 3: Get market resolution info for all relevant markets (single query)
	marketIDs := make([]uint, 0, len(marketIDSet))
	for id := range marketIDSet {
		marketIDs = append(marketIDs, id)
	}

	var markets []models.Market
	if err := db.Where("id IN ?", marketIDs).Find(&markets).Error; err != nil {
		return nil, err
	}

	marketResolutionMap := make(map[uint]models.Market)
	for _, market := range markets {
		marketResolutionMap[uint(market.ID)] = market
	}

	// Step 4: Calculate positions only for markets where user has bets
	var allPositions []MarketPosition
	for marketID := range marketIDSet {
		marketIDStr := strconv.Itoa(int(marketID))
		positions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIDStr)
		if err != nil {
			return nil, err
		}

		// Find user's position in this market
		for _, pos := range positions {
			if pos.Username == username {
				// Position already has all the enhanced fields from CalculateMarketPositions_WPAM_DBPM
				allPositions = append(allPositions, pos)
				break
			}
		}
	}

	return allPositions, nil
}
