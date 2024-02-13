package marketshandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/errors"
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"

	"github.com/gorilla/mux"
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

func MarketDBPMPositionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]

	// open up database to utilize connection pooling
	db := util.GetDB()

	marketDBPMPositions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIdStr)
	if errors.HandleHTTPError(w, err, http.StatusBadRequest, "Invalid request or data processing error.") {
		return // Stop execution if there was an error.
	}

	// Respond with the bets display information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(marketDBPMPositions)
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
	publicResponseMarket, err := GetPublicResponseMarketByID(db, marketIdStr)
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

	// Adjust payouts to align with available betting pool
	finalPayouts := dbpm.AdjustPayoutsFromNewest(allBetsOnMarket, scaledPayouts)

	// Aggregate user payouts into market positions
	marketPositions := dbpm.AggregateUserPayoutsDBPM(allBetsOnMarket, finalPayouts)

	return marketPositions, nil
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

// CheckOppositeSharesOwned determines the number of opposite shares a user holds and needs to sell,
// as well as the type of those shares (YES or NO).
func CheckOppositeSharesOwned(db *gorm.DB, marketIDStr string, username string, betRequestOutcome string) (int64, string, error) {
	userMarketPositions, err := CalculateMarketPositionForUser_WPAM_DBPM(db, marketIDStr, username)
	if err != nil {
		return 0, "", err // Return the error if fetching the market positions fails.
	}

	switch betRequestOutcome {
	case "NO":
		// If the user wants to buy NO shares, check YES shares to sell.
		if userMarketPositions.YesSharesOwned > 0 {
			return userMarketPositions.YesSharesOwned, "YES", nil
		}
	case "YES":
		// If the user wants to buy YES shares, check NO shares to sell.
		if userMarketPositions.NoSharesOwned > 0 {
			return userMarketPositions.NoSharesOwned, "NO", nil
		}
	}

	// If the user has no opposite shares to sell, return 0 and an empty string for the outcome.
	return 0, "", nil
}
