package betshandlers

import (
	"encoding/json"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	"socialpredict/handlers/positions"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
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
		if httpErr, ok := err.(*middleware.HTTPError); ok {
			http.Error(w, httpErr.Error(), httpErr.StatusCode)
			return
		}
		// Handle other types of errors generically
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var redeemRequest models.Bet
	err = json.NewDecoder(r.Body).Decode(&redeemRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get the marketID in string format
	marketIDStr := strconv.FormatUint(uint64(redeemRequest.MarketID), 10)

	// Validate the request similar to PlaceBetHandler
	betutils.CheckMarketStatus(db, redeemRequest.MarketID)

	// Calculate the net aggregate positions for the user
	userNetPosition, err := positions.CalculateMarketPositionForUser_WPAM_DBPM(db, marketIDStr, user.Username)
	if userNetPosition.NoSharesOwned == 0 && userNetPosition.YesSharesOwned == 0 {
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
	// For simplicity, we're just creating a negative bet to represent the sale
	redeemRequest.Amount = -redeemRequest.Amount // Negate the amount to indicate sale

	// Create a new Bet object
	bet := models.Bet{
		Username: user.Username,
		MarketID: redeemRequest.MarketID,
		Amount:   redeemRequest.Amount,
		PlacedAt: time.Now(), // Set the current time as the placement time
		Outcome:  redeemRequest.Outcome,
	}

	// Validate the final bet before putting into database
	if err := betutils.ValidateSale(db, &bet); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Deduct the bet and switching sides fee amount from the user's balance
	user.AccountBalance -= redeemRequest.Amount

	// Update the user's balance in the database
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Error updating user balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result := db.Create(&bet)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(redeemRequest)
}
