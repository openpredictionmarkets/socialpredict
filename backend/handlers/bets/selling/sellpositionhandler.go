package sellbetshandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"
)

func SellPositionHandler(loadEconConfig setup.EconConfigLoader) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := util.GetDB()
		user, httpErr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode)
			return
		}

		var redeemRequest models.Bet
		err := json.NewDecoder(r.Body).Decode(&redeemRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Load economic configuration
		cfg := loadEconConfig()
		if cfg == nil {
			http.Error(w, "failed to load economic configuration", http.StatusInternalServerError)
			return
		}

		if err := ProcessSellRequest(db, &redeemRequest, user, cfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(redeemRequest)
	}
}
