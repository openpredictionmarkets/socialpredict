package buybetshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	betutils "socialpredict/handlers/bets/betutils"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"

	"gorm.io/gorm"
)

func PlaceBetHandler(loadEconConfig setup.EconConfigLoader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := util.GetDB()
		user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUserFromDB(r, db)
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

		bet, err := PlaceBetCore(user, betRequest, db, loadEconConfig)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Return a success response
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(bet)
	}
}

// PlaceBetCore handles the core logic of placing a bet.
// It assumes user authentication and JSON decoding is already done.
func PlaceBetCore(user *models.User, betRequest models.Bet, db *gorm.DB, loadEconConfig setup.EconConfigLoader) (*models.Bet, error) {
	// Validate the request (check if market exists, if not closed/resolved, etc.)
	if err := betutils.CheckMarketStatus(db, betRequest.MarketID); err != nil {
		return nil, err
	}

	sumOfBetFees := betutils.GetBetFees(db, user, betRequest)

	// Check if the user's balance after the bet would be lower than the allowed maximum debt
	if err := checkUserBalance(user, betRequest, sumOfBetFees, loadEconConfig); err != nil {
		return nil, err
	}

	// Create a new Bet object
	bet := models.CreateBet(user.Username, betRequest.MarketID, betRequest.Amount, betRequest.Outcome)

	// Validate the final bet before putting into database
	if err := betutils.ValidateBuy(db, &bet); err != nil {
		return nil, err
	}

	// Deduct bet amount and fee from user balance
	totalCost := bet.Amount + sumOfBetFees
	user.AccountBalance -= totalCost

	// Save updated user balance
	if err := db.Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user balance: %w", err)
	}

	// Save the Bet
	if err := db.Create(&bet).Error; err != nil {
		return nil, fmt.Errorf("failed to create bet: %w", err)
	}

	return &bet, nil
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
