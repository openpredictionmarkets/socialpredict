package betshandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	betutils "socialpredict/handlers/bets/betutils"
	"socialpredict/handlers/positions"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"

	"gorm.io/gorm"
)

// SellPositionHandler is the HTTP handler for selling a user's position in a market.
func SellPositionHandler(loadEconConfig setup.EconConfigLoader) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		db := util.GetDB()
		user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
		if httperr != nil {
			http.Error(w, httperr.Error(), httperr.StatusCode)
			return
		}

		redeemRequest, err := parseRedeemRequest(w, r)
		if err != nil {
			return
		}

		// Validate the request (check if market exists, if not closed/resolved, etc.)
		if err := betutils.CheckMarketStatus(db, redeemRequest.MarketID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get the marketID in string format to be able to use CalculateMarketPositionForUser_WPAM_DBPM
		marketIDStr := strconv.FormatUint(uint64(redeemRequest.MarketID), 10)

		// Calculate the net aggregate positions for the user
		userNetPosition, err := positions.CalculateMarketPositionForUser_WPAM_DBPM(
			db,
			marketIDStr,
			user.Username,
		)
		if err != nil {
			http.Error(w, "Error calculating user net position: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if userNetPosition.NoSharesOwned == 0 && userNetPosition.YesSharesOwned == 0 {
			http.Error(w, "No position found for the given market", http.StatusBadRequest)
			return
		}

		if err := validateRedeemAmount(redeemRequest, userNetPosition); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// the redeemRequest.Amount should be turned into negative to create the negativeBet
		negativeBet := createNegativeBet(redeemRequest, user.Username)

		err = betutils.ValidateSale(db, negativeBet)

		err = reduceUseAccountBalance(db, user, negativeBet)

		err = createBetInDatabase(db, negativeBet)

		respondSuccess(w, redeemRequest)
	}
}

func parseRedeemRequest(w http.ResponseWriter, r *http.Request) (*models.Bet, error) {
	var redeemRequest models.Bet
	if err := json.NewDecoder(r.Body).Decode(&redeemRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	return &redeemRequest, nil
}

func validateRedeemAmount(redeemRequest *models.Bet, userNetPosition positions.UserMarketPosition) error {
	if (redeemRequest.Outcome == "YES" && redeemRequest.Amount > userNetPosition.YesSharesOwned) ||
		(redeemRequest.Outcome == "NO" && redeemRequest.Amount > userNetPosition.NoSharesOwned) {
		return errors.New("redeem amount exceeds available position")
	}
	return nil
}

func createNegativeBet(redeemRequest *models.Bet, username string) *models.Bet {

	// Ensure that the bet being created is a negative amount of the redeemed amount requested
	return &models.Bet{
		Username: username,
		MarketID: redeemRequest.MarketID,
		Amount:   -redeemRequest.Amount,
		PlacedAt: time.Now(),
		Outcome:  redeemRequest.Outcome,
	}
}

func reduceUseAccountBalance(db *gorm.DB, user *models.User, bet *models.Bet) error {

	// we now increase the user account by subtracting the negative bet amount
	user.AccountBalance -= bet.Amount
	if err := db.Save(user).Error; err != nil {
		return fmt.Errorf("error updating user balance: %w", err)
	}
	return nil
}

func createBetInDatabase(db *gorm.DB, bet *models.Bet) error {
	if result := db.Create(bet); result.Error != nil {
		return fmt.Errorf("error saving bet: %w", result.Error)
	}
	return nil
}

func respondSuccess(w http.ResponseWriter, redeemRequest *models.Bet) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(redeemRequest)
}
