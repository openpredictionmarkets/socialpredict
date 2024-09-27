package usercredit

import (
	"encoding/json"
	"log"
	"net/http"
	usershandlers "socialpredict/handlers/users"
	"socialpredict/setup"
	"socialpredict/util"

	"github.com/gorilla/mux"
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

type UserCredit struct {
	Credit int64 `json:"credit"`
}

// gets the user's available credits for display
func GetUserCreditHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	db := util.GetDB()

	userCredit := CalculateUserCredit(db, username)

	response := UserCredit{
		Credit: userCredit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func CalculateUserCredit(db *gorm.DB, username string) int64 {

	userPublicInfo := usershandlers.GetPublicUserInfo(db, username)

	// add the maximum debt from the setup file and he account balance, which may be negative
	userCredit := appConfig.Economics.User.MaximumDebtAllowed + userPublicInfo.AccountBalance

	return int64(userCredit)
}
