package betshandlers

import (
	"encoding/json"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	"socialpredict/handlers/positions"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"
	"strconv"
	"time"
)

func SellPositionHandler(loadEconConfig setup.EconConfigLoader) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		db := util.GetDB()
		user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
		if httperr != nil {
			http.Error(w, httperr.Error(), httperr.StatusCode)
			return
		}

		var redeemRequest models.Bet
		err := json.NewDecoder(r.Body).Decode(&redeemRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate the request (check if market exists, if not closed/resolved, etc.)
		betutils.CheckMarketStatus(db, redeemRequest.MarketID)

		// get the marketID in string format to be able to use CalculateMarketPositionForUser_WPAM_DBPM
		marketIDStr := strconv.FormatUint(uint64(redeemRequest.MarketID), 10)

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

}
