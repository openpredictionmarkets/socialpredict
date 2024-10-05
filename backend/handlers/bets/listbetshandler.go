package betshandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"
	"socialpredict/setup"
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

func MarketBetsDisplayHandler(mcl setup.MarketCreationLoader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]

		// Convert marketId to uint
		parsedUint64, err := strconv.ParseUint(marketIdStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid market ID", http.StatusBadRequest)
			return
		}

		// Convert uint64 to uint safely.
		marketIDUint := uint(parsedUint64)

		// Database connection
		db := util.GetDB()

		// Fetch bets for the market
		bets := tradingdata.GetBetsForMarket(db, marketIDUint)

		// feed in the time created
		// note we are not using GetPublicResponseMarketByID because of circular import
		var market models.Market
		result := db.Where("ID = ?", marketIdStr).First(&market)
		if result.Error != nil {
			// Handle error, for example:
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Market not found
			} else {
				// Other error fetching market
			}
			return // Make sure to return or appropriately handle the error
		}

		// Process bets and calculate market probability at the time of each bet
		betsDisplayInfo := processBetsForDisplay(mcl, market.CreatedAt, bets, db)

		// Respond with the bets display information
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(betsDisplayInfo)
	}
}

func processBetsForDisplay(mcl setup.MarketCreationLoader, marketCreatedAtTime time.Time, bets []models.Bet, db *gorm.DB) []BetDisplayInfo {

	// Calculate probabilities using the fetched bets
	probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(mcl, marketCreatedAtTime, bets)

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
