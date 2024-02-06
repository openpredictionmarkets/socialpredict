package marketshandlers

import (
	"encoding/json"
	"net/http"
	betshandlers "socialpredict/handlers/bets"
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"

	"github.com/gorilla/mux"
)

type MarketPosition struct {
	Username       string
	NoSharesOwned  int64
	YesSharesOwned int64
}

func MarketDBPMPositionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]
	// Convert marketId to uint
	marketIDUint, err := strconv.ParseUint(marketIdStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// open up database to utilize connection pooling
	db := util.GetDB()

	// return the PublicResponse type with information about the market
	publicResponseMarket, err := GetPublicResponseMarketByID(db, marketIdStr)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Fetch bets for the market
	var allBetsOnMarket []models.Bet
	allBetsOnMarket = betshandlers.GetBetsForMarket(marketIDUint)

	// get a timeline of probability changes for the market
	// input the market the safe way
	allProbabilityChangesOnMarket := wpam.CalculateMarketProbabilitiesWPAM(publicResponseMarket.CreatedAt, allBetsOnMarket)

	// calculate number of shares that exist in the entire market, based upon dbpm, int64s
	S_YES, S_NO := dbpm.DivideUpMarketPoolSharesDBPM(allBetsOnMarket, allProbabilityChangesOnMarket)

	// calculate course payout pools, floats
	C_YES, C_NO := dbpm.CalculateCoursePayoutsDBPM(allBetsOnMarket, allProbabilityChangesOnMarket)

	// calculate scaling factor
	F_YES, F_NO := dbpm.CalculateNormalizationFactorsDBPM(S_YES, C_YES, S_NO, C_NO)

	// calculate normalized payout pools
	finalPayouts := dbpm.CalculateFinalPayoutsDBPM(allBetsOnMarket, F_YES, F_NO, C_YES, C_NO)

	// aggregate user payouts into list of positions including username, yes and no positions
	marketDBPMPositions := dbpm.AggregateUserPayoutsDBPM(allBetsOnMarket, finalPayouts)

	// Respond with the bets display information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(marketDBPMPositions)
}
