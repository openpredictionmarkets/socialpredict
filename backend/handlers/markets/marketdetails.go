package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	betsHandlers "socialpredict/handlers/bets"
	marketMathHandlers "socialpredict/handlers/math/market"
	usersHandlers "socialpredict/handlers/users"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func MarketDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketId := vars["marketId"]

	// Use database connection
	db := util.GetDB()

	var market models.Market
	// Use Preload to fetch the Creator along with the Market
	result := db.Preload("Creator").Where("ID = ?", marketId).First(&market)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Market not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching market", http.StatusInternalServerError)
		}
		return
	}

	// Parsing a String to an Unsigned Integer, base10, 32bits
	marketIDUint, err := strconv.ParseUint(marketId, 10, 32)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Fetch all bets for the market once
	bets, err := betsHandlers.GetBetsForMarket(marketIDUint)
	if err != nil {
		http.Error(w, "Error retrieving bets.", http.StatusInternalServerError)
		return
	}

	// Calculate probabilities using the fetched bets
	probabilityChanges, err := marketMathHandlers.CalculateMarketProbabilities(market, bets)
	if err != nil {
		http.Error(w, "Error calculating market probabilities", http.StatusInternalServerError)
		return
	}

	numUsers := usersHandlers.GetMarketUsers(bets)
	if err != nil {
		http.Error(w, "Error retrieving number of users.", http.StatusInternalServerError)
		return
	}

	// Inside your handler
	marketVolume := marketMathHandlers.GetMarketVolume(bets)
	if err != nil {
		// Handle error
	}

	// get market creator

	// Update your response struct accordingly
	response := struct {
		Market             models.Market                          `json:"market"`
		CreatorUsername    string                                 `json:"creatorUsername"`
		ProbabilityChanges []marketMathHandlers.ProbabilityChange `json:"probabilityChanges"`
		NumUsers           int                                    `json:"numUsers"`
		TotalVolume        float64                                `json:"totalVolume"`
	}{
		Market:             market,
		CreatorUsername:    market.Creator.Username, // Include the creator's username
		ProbabilityChanges: probabilityChanges,
		NumUsers:           numUsers,
		TotalVolume:        marketVolume,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
