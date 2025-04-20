package betshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"
	"time"
)

type BetResponse struct {
    ID        int    `json:"id"`
    Username  string `json:"username"`
    MarketID  int    `json:"marketId"`
    Outcome   string `json:"outcome"`
    Amount    int64  `json:"amount"`
    PlacedAt  string `json:"placedAt"`
}

func PlaceBetHandler(loadEconConfig setup.EconConfigLoader) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		db := util.GetDB()
		user, httpErr := middleware.GetAuthenticatedUser(r,db)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode)
			return
		}

		var betRequest models.Bet
		err := json.NewDecoder(r.Body).Decode(&betRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

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

		// Save the Bet to the database
		result := db.Create(&bet)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		response := BetResponse{
			ID:       int(bet.ID),
			Username: bet.Username,
			MarketID: int(bet.MarketID),
			Outcome:  bet.Outcome,
			Amount:   bet.Amount,
			PlacedAt: bet.CreatedAt.Format(time.RFC3339),
		}
		
		// Return a success response
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
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
