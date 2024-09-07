package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/setup"
	"socialpredict/util"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type UserCredit struct {
	Credit int `json:"credit"`
}

// gets the user's available credits for display
func GetUserCreditHandler(loadEconomicsConfig setup.EconConfigLoader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// accept get requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Extract username from the URL path
		vars := mux.Vars(r)
		username := vars["username"]

		// Use database connection
		db := util.GetDB()

		userCredit := calculateUserCredit(loadEconomicsConfig, db, username)
		response := UserCredit{
			Credit: userCredit,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func calculateUserCredit(loadEconomicsConfig setup.EconConfigLoader, db *gorm.DB, username string) int {
	appConfig := loadEconomicsConfig()
	userPublicInfo := GetPublicUserInfo(db, username)

	// add the maximum debt from the setup file and he account balance, which may be negative
	userCredit := appConfig.Economics.User.MaximumDebtAllowed + userPublicInfo.AccountBalance

	return int(userCredit)
}
