package betshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/security"
	"socialpredict/setup"
	"socialpredict/util"
)

func PlaceBetHandler(loadEconConfig setup.EconConfigLoader) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		// Initialize security service
		securityService := security.NewSecurityService()

		db := util.GetDB()
		user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
		if httperr != nil {
			http.Error(w, httperr.Error(), httperr.StatusCode)
			return
		}

		var betRequest models.Bet
		err := json.NewDecoder(r.Body).Decode(&betRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate and sanitize bet input using security service
		betInput := security.BetInput{
			MarketID: fmt.Sprintf("%d", betRequest.MarketID), // Convert uint to string
			Amount:   float64(betRequest.Amount),             // Convert int64 to float64
			Outcome:  betRequest.Outcome,
		}

		sanitizedBetInput, err := securityService.ValidateAndSanitizeBetInput(betInput)
		if err != nil {
			http.Error(w, "Invalid bet data: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Update the bet request with sanitized data - convert back to original types
		// Note: MarketID should not be changed after sanitization since it's a database reference
		betRequest.Amount = int64(sanitizedBetInput.Amount) // Convert float64 back to int64
		betRequest.Outcome = sanitizedBetInput.Outcome

		// Validate the request (check if market exists, if not closed/resolved, etc.)
		betutils.CheckMarketStatus(db, betRequest.MarketID)

		sumOfBetFees := betutils.GetBetFees(db, user, betRequest)

		// Check if the user's balance after the bet would be lower than the allowed maximum debt
		// deduct fees along with calculation to ensure fees can be paid.
		checkUserBalance(user, betRequest, sumOfBetFees, loadEconConfig)

		// Create a new Bet object, set at current time
		bet := models.CreateBet(user.Username, betRequest.MarketID, betRequest.Amount, betRequest.Outcome)

		// Validate the final bet before putting into database
		if err := betutils.ValidateBuy(db, &bet); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user.AccountBalance -= betRequest.Amount + sumOfBetFees
		if err := db.Save(&user).Error; err != nil {
			http.Error(w, "Failed to update user balance", http.StatusInternalServerError)
			return
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
}

func checkUserBalance(user *models.User, betRequest models.Bet, sumOfBetFees int64, loadEconConfig setup.EconConfigLoader) error {
	appConfig := loadEconConfig()
	maximumDebtAllowed := appConfig.Economics.User.MaximumDebtAllowed

	// Check if the user's balance after the bet would be lower than the allowed maximum debt
	if user.AccountBalance-betRequest.Amount-sumOfBetFees < -maximumDebtAllowed {
		return fmt.Errorf("Insufficient balance")
	}
	return nil
}
