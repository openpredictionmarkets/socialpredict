package handlers

import (
	"encoding/json"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
	"time"
)

func SellShareHandler(w http.ResponseWriter, r *http.Request) {
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
	// ... similar steps as in PlaceBetHandler to authenticate and validate request ...

	var sellRequest models.Bet
	err = json.NewDecoder(r.Body).Decode(&sellRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request (check if market exists, if not closed/resolved, etc.)
	betutils.CheckMarketStatus(db, sellRequest.MarketID)

	// user-specific validation, sufficient balance,
	// Fetch the user's current balance
	if err := db.First(&user, user.ID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// we do not check whether the user has sufficient balance because this is a sale

	// calculate shares owned in market
	sharesOwned, err := marketmath.CalculateSharesOwned(db, user.ID, sellRequest.MarketID)
	if err != nil {
		// handle error
	}

	// Get the current market price
	marketPrice, err := marketmath.GetCurrentMarketPrice(db, sellRequest.MarketID)
	if err != nil {
		// handle error
	}

	// Calculate the total value of shares owned
	totalValue := marketmath.CalculateTotalValueOfShares(sharesOwned, marketPrice)

	// Convert the requested sell amount to the equivalent number of shares
	sharesToSell := marketmath.ConvertAmountToShares(sellRequest.Amount, marketPrice)

	// Check if the user has enough shares to sell
	if sharesToSell > sharesOwned {
		http.Error(w, "Not enough shares to sell", http.StatusBadRequest)
		return
	}

	// first, create new Bet object for selling
	// Create a new Bet object
	// The sale bet will be equivalent to the a buy bet in the opposite direction.
	// For example, selling a NO share at 1 point would be equivalent to buying a YES share at 1 point.
	bet := models.Bet{
		Username: user.Username,
		MarketID: sellRequest.MarketID,
		Amount:   sellRequest.Amount,
		PlacedAt: time.Now(),          // Set the current time as the placement time
		Outcome:  sellRequest.Outcome, // needs to be the opposite
	}

	// Save the Bet to the database
	result := db.Create(&bet)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Add to the user's balance the amount calculated in the sell
	user.AccountBalance += sellRequest.Amount

	// Update the user's balance in the database
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Error updating user balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bet)
}
