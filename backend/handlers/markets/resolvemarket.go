package marketshandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"socialpredict/handlers/math/payout"
	"socialpredict/logging"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func ResolveMarketHandler(w http.ResponseWriter, r *http.Request) {
	logging.LogMsg("Attempting to use ResolveMarketHandler.")

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
	user, httpErr := middleware.ValidateTokenAndGetUser(r, db)
	if httpErr != nil {
		http.Error(w, "Invalid token: "+httpErr.Error(), http.StatusUnauthorized)
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

	var market models.Market
	result := db.First(&market, marketId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Market not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error accessing database", http.StatusInternalServerError)
		return
	}

	if &market == nil {
		// handle nil market if necessary, this is just precautionary, as gorm.First should return found object or error
		http.Error(w, "No market found with provided ID", http.StatusNotFound)
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

	// Save the market changes first so payout calculation sees the resolved state
	if err := db.Save(&market).Error; err != nil {
		http.Error(w, "Error saving market resolution: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle payouts (if applicable) - after market is saved as resolved
	err = payout.DistributePayoutsWithRefund(&market, db)
	if err != nil {
		http.Error(w, "Error distributing payouts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a response back
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Market resolved successfully"})
}
