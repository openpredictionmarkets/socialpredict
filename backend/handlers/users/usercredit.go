package usershandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/setup"
	"socialpredict/util"

	"gorm.io/gorm"
)

// appConfig holds the loaded application configuration accessible within the package
var appConfig *setup.EconomicConfig

func init() {
	var err error
	appConfig, err = setup.LoadEconomicsConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
}

type UserCreditResponse struct {
	Credit int `json:"credit"`
}

// for usage on sidebar or continuously throughout application in order to continuously show available spend
func GetUserCreditResponse(w http.ResponseWriter, r *http.Request) {

	// accept get requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Use database connection
	db := util.GetDB()

	// Validate the token and get the user
	user, httperr := middleware.ValidateTokenAndGetUser(r, db)
	if httperr != nil {
		http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
		return
	}

	// The username is extracted from the token
	username := user.Username

	userCredit := calculateUserCredit(db, username)

	response := UserCreditResponse{
		Credit: userCredit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func calculateUserCredit(db *gorm.DB, username string) int {

	userPublicInfo := GetPublicUserInfo(db, username)

	// add the maximum debt from the setup file and he account balance, which may be negative
	userCredit := appConfig.Economics.User.MaximumDebtAllowed + userPublicInfo.AccountBalance

	return int(userCredit)
}
