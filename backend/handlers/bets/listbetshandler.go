package betshandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"socialpredict/handlers/marketpublicresponse"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"
	"socialpredict/util"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type BetDisplayInfo struct {
	Username    string    `json:"username"`
	Outcome     string    `json:"outcome"`
	Amount      int64     `json:"amount"`
	Probability float64   `json:"probability"`
	PlacedAt    time.Time `json:"placedAt"`
}

func MarketBetsDisplayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]

	parsedUint64, err := strconv.ParseUint(marketIdStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}
	marketIDUint := uint(parsedUint64)

	db := util.GetDB()

	market, err := marketpublicresponse.GetPublicResponseMarketByID(db, marketIdStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Market not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching market", http.StatusInternalServerError)
		}
		return
	}

	bets := tradingdata.GetBetsForMarket(db, marketIDUint)

	betsDisplayInfo := processBetsForDisplay(market.CreatedAt, bets, db)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(betsDisplayInfo)
}

func processBetsForDisplay(marketCreatedAtTime time.Time, bets []models.Bet, db *gorm.DB) []BetDisplayInfo {

	// Calculate probabilities using the fetched bets
	probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(marketCreatedAtTime, bets)

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

	// Sort betsDisplayInfo by PlacedAt in ascending order (most recent on top)
	sort.Slice(betsDisplayInfo, func(i, j int) bool {
		return betsDisplayInfo[i].PlacedAt.Before(betsDisplayInfo[j].PlacedAt)
	})

	return betsDisplayInfo
}
