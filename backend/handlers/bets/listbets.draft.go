package handlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type BetDisplayInfo struct {
	Username    string    `json:"username"`
	Position    string    `json:"position"`
	Amount      float64   `json:"amount"`
	Probability float64   `json:"probability"`
	PlacedAt    time.Time `json:"placedAt"`
}

func marketBetsDisplayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketId := vars["marketId"]

	// Database connection
	db := util.GetDB()

	// Convert marketId to uint
	marketIDUint, err := strconv.ParseUint(marketId, 10, 32)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	var bets []models.Bet
	if err := db.Where("market_id = ?", marketIDUint).Find(&bets).Error; err != nil {
		http.Error(w, "Error fetching bets", http.StatusInternalServerError)
		return
	}

	// Process bets and calculate market probability at the time of each bet
	betsDisplayInfo := processBetsForDisplay(bets, db)

	// Respond with the bets display information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(betsDisplayInfo)
}

func processBetsForDisplay(bets []models.Bet, db *gorm.DB) []BetDisplayInfo {
	var betsDisplayInfo []BetDisplayInfo

	for _, bet := range bets {
		// Calculate the market probability at the time of the bet
		probabilityAtBetTime := calculateMarketProbabilityAtTime(bet.PlacedAt, bet.MarketID, db)

		betsDisplayInfo = append(betsDisplayInfo, BetDisplayInfo{
			Username:    bet.Username,
			Position:    bet.Outcome,
			Amount:      bet.Amount,
			Probability: probabilityAtBetTime,
			PlacedAt:    bet.PlacedAt,
		})
	}

	return betsDisplayInfo
}
