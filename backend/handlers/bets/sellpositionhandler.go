package betshandlers

import (
	"encoding/json"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	marketshandlers "socialpredict/handlers/markets"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
	"time"
)

func SellPositionHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	db := util.GetDB() // Get the database connection
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var redeemRequest models.Bet
	err = json.NewDecoder(r.Body).Decode(&redeemRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request similar to PlaceBetHandler
	betutils.CheckMarketStatus(db, redeemRequest.MarketID)

	// Calculate the net aggregate positions for the user

	userNetPosition, err := marketshandlers.CalculateMarketPositionForUser_WPAM_DBPM(db, redeemRequest.MarketID, user.Username)
	if userNetPosition == nil {
		http.Error(w, "No position found for the given market", http.StatusBadRequest)
		return
	}

	// Check if the user is trying to redeem more than they own
	if (redeemRequest.Outcome == "YES" && redeemRequest.Amount > userNetPosition.YesSharesOwned) ||
		(redeemRequest.Outcome == "NO" && redeemRequest.Amount > userNetPosition.NoSharesOwned) {
		http.Error(w, "Redeem amount exceeds available position", http.StatusBadRequest)
		return
	}

	// Proceed with redemption logic
	// Here, you would typically update the user's balance and record the transaction
	// as a negative bet or a separate redemption record, depending on your data model.

	// For simplicity, we're just creating a negative bet to represent the sale
	// Note: Ensure your system correctly handles negative bets in all relevant calculations
	redeemRequest.Amount = -redeemRequest.Amount // Negate the amount to indicate sale
	redeemRequest.PlacedAt = time.Now()          // Set the current time as the redemption time

	result := db.Create(&redeemRequest)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(redeemRequest)
}
