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

	if err := betutils.CheckMarketStatus(db, redeemRequest.MarketID); err != nil {
		return err
	}

	marketIDStr := strconv.FormatUint(uint64(redeemRequest.MarketID), 10)

	userNetPosition, err := GetUserNetPositionForMarket(db, marketIDStr, user.Username)
	if err != nil {
		return err
	}

	sharesOwned, err := GetSharesOwnedForOutcome(userNetPosition, redeemRequest.Outcome)
	if err != nil {
		return err
	}

	sharesToSell, actualSaleValue, err := CalculateSharesToSell(
		userNetPosition, sharesOwned, redeemRequest.Amount)
	if err != nil {
		return err
	}

	// dust := redeemRequest.Amount - actualSaleValue // remainder not paid out

	if sharesToSell == 0 {
		return errors.New("not enough value to sell at least one share")
	}

	bet := models.Bet{
		Username: user.Username,
		MarketID: redeemRequest.MarketID,
		Amount:   -sharesToSell, // negative share amount means sale
		PlacedAt: time.Now(),
		Outcome:  redeemRequest.Outcome,
	}

	if err := betutils.ValidateSale(db, &bet); err != nil {
		return err
	}

	if err := usershandlers.ApplyTransactionToUser(user.Username, actualSaleValue, db, usershandlers.TransactionSale); err != nil {
		return err
	}

	if err := db.Create(&bet).Error; err != nil {
		return err
	}

	return nil
}

func GetUserNetPositionForMarket(db *gorm.DB, marketIDStr string, username string) (positionsmath.UserMarketPosition, error) {
	userNetPosition, err := positionsmath.CalculateMarketPositionForUser_WPAM_DBPM(db, marketIDStr, username)
	if err != nil {
		return userNetPosition, err
	}
	if userNetPosition.NoSharesOwned == 0 && userNetPosition.YesSharesOwned == 0 {
		return userNetPosition, errors.New("no position found for the given market")
	}
	return userNetPosition, nil
}

func GetSharesOwnedForOutcome(userNetPosition positionsmath.UserMarketPosition, outcome string) (int64, error) {
	switch outcome {
	case "YES":
		if userNetPosition.YesSharesOwned == 0 {
			return 0, errors.New("no shares owned for selected outcome")
		}
		return userNetPosition.YesSharesOwned, nil
	case "NO":
		if userNetPosition.NoSharesOwned == 0 {
			return 0, errors.New("no shares owned for selected outcome")
		}
		return userNetPosition.NoSharesOwned, nil
	default:
		return 0, errors.New("invalid outcome")
	}
}

// CalculateSharesToSell determines how many shares a user can sell for a given credit amount.
func CalculateSharesToSell(userNetPosition positionsmath.UserMarketPosition, sharesOwned int64, creditsToSell int64) (int64, int64, error) {
	if userNetPosition.Value <= 0 {
		return 0, 0, errors.New("position value is non-positive")
	}
	valuePerShare := userNetPosition.Value / sharesOwned
	if creditsToSell < valuePerShare {
		return 0, 0, errors.New("requested credit amount is less than value of one share")
	}
	sharesToSell := creditsToSell / valuePerShare
	if sharesToSell > sharesOwned {
		sharesToSell = sharesOwned
	}
	actualSaleValue := sharesToSell * valuePerShare
	if sharesToSell == 0 {
		return 0, 0, errors.New("not enough value to sell at least one share")
	}
	return sharesToSell, actualSaleValue, nil
}
