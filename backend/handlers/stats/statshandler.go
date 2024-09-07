package statshandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"

	"gorm.io/gorm"
)

type FinancialStats struct {
	TotalMoney              int64 `json:"totalMoney"`
	TotalDebtExtended       int64 `json:"totalDebtExtended"`
	TotalDebtUtilized       int64 `json:"totalDebtUtilized"`
	TotalFeesCollected      int64 `json:"totalFeesCollected"`
	TotalBonusesPaid        int64 `json:"totalBonusesPaid"`
	OutstandingPayouts      int64 `json:"outstandingPayouts"`
	TotalMoneyInCirculation int64 `json:"totalMoneyInCirculation"` // Money currently active in bets
}

// StatsHandler handles requests for financial stats
func StatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		db := util.GetDB()
		// Call the calculateFinancialStats function
		stats, err := calculateFinancialStats(db)
		if err != nil {
			http.Error(w, "Failed to calculate financial stats: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, "Failed to encode financial stats: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func calculateFinancialStats(db *gorm.DB) (FinancialStats, error) {
	var result FinancialStats

	// Calculate TotalMoney by calling the new function
	totalMoney, err := calculateTotalMoney(db)
	if err != nil {
		return result, err // Return the current result and the error
	}
	result.TotalMoney = totalMoney

	// Initialize other financial stats with dummy data (or implement real calculations)
	result.TotalDebtExtended = 0       // Real calculation should be implemented later
	result.TotalDebtUtilized = 0       // Real calculation should be implemented later
	result.TotalFeesCollected = 0      // Real calculation should be implemented later
	result.TotalBonusesPaid = 0        // Real calculation should be implemented later
	result.OutstandingPayouts = 0      // Real calculation should be implemented later
	result.TotalMoneyInCirculation = 0 // Real calculation should be implemented later

	return result, nil
}

// calculateTotalMoney calculates the total initial money in the system based on the number of regular users.
func calculateTotalMoney(db *gorm.DB) (int64, error) {
	// Load economic configuration
	economicConfig, err := setup.LoadEconomicsConfig()
	if err != nil {
		return 0, err // Return zero and the error if config can't be loaded
	}

	// Count the number of regular users
	var userCount int64
	if err := db.Model(&models.User{}).Where("user_type = ?", "REGULAR").Count(&userCount).Error; err != nil {
		return 0, err // Return zero and the error if the query fails
	}

	// Calculate total money based on the initial account balance and user count
	totalMoney := economicConfig.Economics.User.InitialAccountBalance * userCount
	return totalMoney, nil
}
