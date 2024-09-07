package marketshandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/handlers/marketpublicresponse"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ProjectNewProbabilityHandler handles the projection of a new probability based on a new bet.
func ProjectNewProbabilityHandler(w http.ResponseWriter, r *http.Request) {

	// Parse market ID, amount, and outcome from the URL
	vars := mux.Vars(r)
	marketId := vars["marketId"]
	amountStr := vars["amount"]
	outcome := vars["outcome"]

	// Parse marketId string directly into a uint
	marketIDUint64, err := strconv.ParseUint(marketId, 10, strconv.IntSize)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Convert to uint (will be either uint32 or uint64 depending on platform)
	marketIDUint := uint(marketIDUint64)

	// Convert amount to int64
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid amount value", http.StatusBadRequest)
		return
	}

	// Create a new Bet object without a username
	newBet := models.Bet{
		Amount:   amount,
		Outcome:  outcome,
		PlacedAt: time.Now(), // Assuming the bet is placed now
		MarketID: marketIDUint,
	}

	// Open up database to utilize connection pooling
	db := util.GetDB()

	// Fetch all bets for the market
	currentBets := tradingdata.GetBetsForMarket(db, marketIDUint)

	// Fetch the market creation time using utility function
	publicResponseMarket, err := marketpublicresponse.GetPublicResponseMarketByID(db, marketId)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}
	marketCreatedAt := publicResponseMarket.CreatedAt

	// Project the new probability
	projectedProbability := wpam.ProjectNewProbabilityWPAM(marketCreatedAt, currentBets, newBet)

	// Set the content type to JSON and encode the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projectedProbability)
}
