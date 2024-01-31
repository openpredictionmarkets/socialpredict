package marketshandlers

import (
	"encoding/json"
	"net/http"
	dbpm "socialpredict/handlers/math/outcomes/dbpm"
	usersHandlers "socialpredict/handlers/users"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func ResolveMarketHandler(w http.ResponseWriter, r *http.Request) {
	// Use database connection
	db := util.GetDB()

	// Retrieve marketId from URL parameters
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]
	marketId, err := strconv.ParseUint(marketIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Validate token and get user
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse request body for resolution outcome
	var resolutionData struct {
		Outcome string `json:"outcome"`
	}
	if err := json.NewDecoder(r.Body).Decode(&resolutionData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find the market by ID
	var market models.Market
	if result := db.First(&market, marketId); result.Error != nil {
		http.Error(w, "Market not found", http.StatusNotFound)
		return
	}

	// Check if the logged-in user is the creator of the market
	if market.CreatorUsername != user.Username {
		http.Error(w, "User is not the creator of the market", http.StatusUnauthorized)
		return
	}

	// Check if the market is already resolved
	if market.IsResolved {
		http.Error(w, "Market is already resolved", http.StatusBadRequest)
		return
	}

	// Validate the resolution outcome
	if resolutionData.Outcome != "YES" && resolutionData.Outcome != "NO" && resolutionData.Outcome != "N/A" {
		http.Error(w, "Invalid resolution outcome", http.StatusBadRequest)
		return
	}

	// Update the market with the resolution result
	market.IsResolved = true
	market.ResolutionResult = resolutionData.Outcome
	market.FinalResolutionDateTime = time.Now()

	// Handle payouts (if applicable)
	err = distributePayouts(&market, db)
	if err != nil {
		http.Error(w, "Error distributing payouts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save the market changes
	if err := db.Save(&market).Error; err != nil {
		http.Error(w, "Error saving market resolution: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a response back
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Market resolved successfully"})
}

// distributePayouts handles the logic for calculating and distributing payouts
func distributePayouts(market *models.Market, db *gorm.DB) error {
	// Handle the N/A outcome by refunding all bets
	if market.ResolutionResult == "N/A" {
		var bets []models.Bet
		if err := db.Where("market_id = ?", market.ID).Find(&bets).Error; err != nil {
			return err
		}

		for _, bet := range bets {
			if err := usersHandlers.UpdateUserBalance(bet.Username, bet.Amount, db, "refund"); err != nil {
				return err
			}
		}
		return nil
	}

	// Calculate and distribute payouts using the CPMM model
	return calculateDBPMPayouts(market, db)
}

// calculateCPMMPayouts calculates and updates user balances based on the CPMM model.
func calculateDBPMPayouts(market *models.Market, db *gorm.DB) error {
	// Retrieve all bets associated with the market
	var bets []models.Bet
	if err := db.Where("market_id = ?", market.ID).Find(&bets).Error; err != nil {
		return err
	}

	// Initialize variables to calculate total amounts for each outcome
	var totalYes, totalNo float64
	for _, bet := range bets {
		if bet.Outcome == "YES" {
			totalYes += bet.Amount
		} else if bet.Outcome == "NO" {
			totalNo += bet.Amount
		}
	}

	// Calculate payouts based on DBPM for YES and NO outcomes
	// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
	for _, bet := range bets {
		payout := dbpm.CalculatePayoutForOutcomeDBPM(bet, totalYes, totalNo, bet.Outcome, market.ResolutionResult)

		// Update user balance with the payout
		if payout > 0 {
			if err := usersHandlers.UpdateUserBalance(bet.Username, payout, db, "win"); err != nil {
				return err
			}
		}
	}

	return nil
}
