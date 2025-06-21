package sellbetshandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	positionsmath "socialpredict/handlers/math/positions"
	usershandlers "socialpredict/handlers/users"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"
	"strconv"
	"time"

	"gorm.io/gorm"
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

		if err := ProcessSellRequest(db, &redeemRequest, user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(redeemRequest)
	}
}

func ProcessSellRequest(db *gorm.DB, redeemRequest *models.Bet, user *models.User) error {
	// 1. Validate the market
	if err := betutils.CheckMarketStatus(db, redeemRequest.MarketID); err != nil {
		return err
	}

	marketIDStr := strconv.FormatUint(uint64(redeemRequest.MarketID), 10)

	// 2. Get user position and valuation
	userNetPosition, err := positionsmath.CalculateMarketPositionForUser_WPAM_DBPM(db, marketIDStr, user.Username)
	if err != nil {
		return err
	}
	if userNetPosition.NoSharesOwned == 0 && userNetPosition.YesSharesOwned == 0 {
		return errors.New("no position found for the given market")
	}

	// 3. Check oversell
	var sharesOwned int64
	if redeemRequest.Outcome == "YES" {
		sharesOwned = userNetPosition.YesSharesOwned
	} else if redeemRequest.Outcome == "NO" {
		sharesOwned = userNetPosition.NoSharesOwned
	} else {
		return errors.New("invalid outcome")
	}
	if redeemRequest.Amount > sharesOwned {
		return errors.New("redeem amount exceeds available position")
	}
	if sharesOwned == 0 {
		return errors.New("no shares owned for selected outcome")
	}

	// 4. Calculate value per share
	if userNetPosition.Value <= 0 {
		return errors.New("position value is non-positive")
	}
	valuePerShare := userNetPosition.Value / sharesOwned
	sharesToSell := redeemRequest.Amount
	saleValue := valuePerShare * sharesToSell

	// 5. Record the sale as a negative bet
	redeemRequest.Amount = -sharesToSell

	bet := models.Bet{
		Username: user.Username,
		MarketID: redeemRequest.MarketID,
		Amount:   redeemRequest.Amount,
		PlacedAt: time.Now(),
		Outcome:  redeemRequest.Outcome,
	}

	// 6. Final validation
	if err := betutils.ValidateSale(db, &bet); err != nil {
		return err
	}

	// 7. Credit the user's account with the **sale value**
	if err := usershandlers.ApplyTransactionToUser(user.Username, saleValue, db, usershandlers.TransactionSale); err != nil {
		return err
	}

	return nil
}
