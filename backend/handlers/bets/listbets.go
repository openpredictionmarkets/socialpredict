package handlers

import (
	"encoding/json"
	"net/http"
	marketMathHandlers "socialpredict/handlers/math/market"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type BetDisplayInfo struct {
	Username    string    `json:"username"`
	Outcome     string    `json:"outcome"`
	Amount      float64   `json:"amount"`
	Probability float64   `json:"probability"`
	PlacedAt    time.Time `json:"placedAt"`
}

func MarketBetsDisplayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]
	// Convert marketId to uint
	marketIDUint, err := strconv.ParseUint(marketIdStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Database connection
	db := util.GetDB()

	// Fetch bets for the market
	bets := GetBetsForMarket(marketIDUint)

	var market models.Market

	// Process bets and calculate market probability at the time of each bet
	betsDisplayInfo := processBetsForDisplay(market, bets, db)

	// Respond with the bets display information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(betsDisplayInfo)
}

func processBetsForDisplay(market models.Market, bets []models.Bet, db *gorm.DB) []BetDisplayInfo {

	// Calculate probabilities using the fetched bets
	probabilityChanges := marketMathHandlers.CalculateMarketProbabilities(market, bets)

	var betsDisplayInfo []BetDisplayInfo

	// Iterate over each bet
	for _, bet := range bets {
		// Find the closest probability change that occurred before or at the time of the bet
		var matchedProbability float64 = probabilityChanges[0].Probability // Start with initial probability
		for _, probChange := range probabilityChanges {
			if probChange.Timestamp.After(bet.PlacedAt) {
				break
			}
			matchedProbability = probChange.Probability
		}

		// Append the bet and its matched probability to the slice
		betsDisplayInfo = append(betsDisplayInfo, BetDisplayInfo{
			Username:    bet.Username,
			Outcome:     bet.Outcome,
			Amount:      bet.Amount,
			Probability: matchedProbability,
			PlacedAt:    bet.PlacedAt,
		})
	}

	return betsDisplayInfo
}
