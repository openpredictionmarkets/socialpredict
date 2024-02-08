package betshandlers

import (
	"encoding/json"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
	"time"
)

func PlaceBetHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	// Validate JWT token and extract user information
	db := util.GetDB() // Get the database connection
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var betRequest models.Bet
	// Decode the request body into betRequest
	err = json.NewDecoder(r.Body).Decode(&betRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request (check if market exists, if not closed/resolved, etc.)
	betutils.CheckMarketStatus(db, betRequest.MarketID)

	// user-specific validation, sufficient balance,
	// Fetch the user's current balance
	if err := db.First(&user, user.ID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// sell opposite shares first
	// check users's current position on market, YES and NO
	// for betRequest of opposite Outcome held, first sell shares held at current price
	// Then go forward and adjust the betRequest to adjustedBetRequest and buy new shares with amount remaining

	// Check if the user has enough balance to place the bet
	// Use the appConfig for configuration values
	maximumDebtAllowed := betutils.Appconfig.MaximumDebtAllowed

	// Check if the user's balance after the bet would be lower than the allowed maximum debt
	if user.AccountBalance-betRequest.Amount < -maximumDebtAllowed {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// Deduct the bet amount from the user's balance
	user.AccountBalance -= betRequest.Amount

	// Update the user's balance in the database
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Error updating user balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new Bet object
	bet := models.Bet{
		Username: user.Username,
		MarketID: betRequest.MarketID,
		Amount:   betRequest.Amount,
		PlacedAt: time.Now(), // Set the current time as the placement time
		Outcome:  betRequest.Outcome,
	}

	// Save the Bet to the database
	result := db.Create(&bet)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bet)
}
