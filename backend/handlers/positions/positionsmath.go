package positions

import (
	"socialpredict/errors"
	"socialpredict/handlers/marketpublicresponse"
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"
	"strconv"

	"gorm.io/gorm"
)

// holds the number of YES and NO shares owned by all users in a market
type MarketPosition struct {
	Username       string
	NoSharesOwned  int64
	YesSharesOwned int64
}

// UserMarketPosition holds the number of YES and NO shares owned by a user in a market.
type UserMarketPosition struct {
	NoSharesOwned  int64
	YesSharesOwned int64
}

// FetchMarketPositions fetches and summarizes positions for a given market.
// It returns a slice of MarketPosition as defined in the dbpm package.
func CalculateMarketPositions_WPAM_DBPM(db *gorm.DB, marketIdStr string) ([]dbpm.MarketPosition, error) {

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

	return netPositions, nil
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
			}, nil
		}
	}

	return UserMarketPosition{}, nil
}
