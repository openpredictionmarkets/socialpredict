package marketshandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"socialpredict/logging"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"
)

const maxQuestionTitleLength = 160

// appConfig holds the loaded application configuration accessible within the package
var appConfig *setup.EconomicConfig

func init() {
	appConfig = setup.MustLoadEconomicsConfig()
}

func checkQuestionTitleLength(title string) error {
	if len(title) > maxQuestionTitleLength || len(title) < 1 {
		return fmt.Errorf("question title exceeds %d characters or is blank", maxQuestionTitleLength)
	}
	return nil
}

func checkQuestionDescriptionLength(description string) error {
	if len(description) > 2000 {
		return errors.New("Question Description exceeds 2000 characters")
	}
	return nil
}

func CreateMarketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Use database connection, validate user based upon token
	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	var newMarket models.Market

	newMarket.CreatorUsername = user.Username

	err := json.NewDecoder(r.Body).Decode(&newMarket)
	if err != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		log.Printf("Error reading request body: %v, Body: %s", err, string(bodyBytes))
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	if err = checkQuestionTitleLength(newMarket.QuestionTitle); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = checkQuestionDescriptionLength(newMarket.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = util.CheckUserIsReal(db, newMarket.CreatorUsername); err != nil {
		if err.Error() == "creator user not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Subtract any Market Creation Fees from Creator, up to maximum debt
	marketCreateFee := appConfig.Economics.MarketIncentives.CreateMarketCost
	maximumDebtAllowed := appConfig.Economics.User.MaximumDebtAllowed

	// Maximum debt allowed check
	if user.AccountBalance-marketCreateFee < -maximumDebtAllowed {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// deduct fee
	logging.LogAnyType(user.AccountBalance, "user.AccountBalance before")
	// Deduct the bet and switching sides fee amount from the user's balance
	user.AccountBalance -= marketCreateFee
	logging.LogAnyType(user.AccountBalance, "user.AccountBalance after")

	// Update the user's balance in the database
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Error updating user balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the market in the database
	result := db.Create(&newMarket)
	if result.Error != nil {
		log.Printf("Error creating new market: %v", result.Error)
		http.Error(w, "Error creating new market", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Send a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newMarket)
}
