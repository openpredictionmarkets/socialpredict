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

type SetupConfiguration struct {
	InitialMarketProbability   float64 `json:"initialMarketProbability"`
	InitialMarketSubsidization int64   `json:"initialMarketSubsidization"`
	InitialMarketYes           int64   `json:"initialMarketYes"`
	InitialMarketNo            int64   `json:"initialMarketNo"`
	CreateMarketCost           int64   `json:"createMarketCost"`
	TraderBonus                int64   `json:"traderBonus"`
	InitialAccountBalance      int64   `json:"initialAccountBalance"`
	MaximumDebtAllowed         int64   `json:"maximumDebtAllowed"`
	MinimumBet                 int64   `json:"minimumBet"`
	MaxDustPerSale             int64   `json:"maxDustPerSale"`
	InitialBetFee              int64   `json:"initialBetFee"`
	BuySharesFee               int64   `json:"buySharesFee"`
	SellSharesFee              int64   `json:"sellSharesFee"`
}

type StatsResponse struct {
	FinancialStats     FinancialStats     `json:"financialStats"`
	SetupConfiguration SetupConfiguration `json:"setupConfiguration"`
}

// StatsHandler handles requests for both financial stats and setup configuration
func StatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		db := util.GetDB()

		// Calculate financial stats
		financialStats, err := calculateFinancialStats(db)
		if err != nil {
			http.Error(w, "Failed to calculate financial stats: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Load setup configuration
		setupConfig, err := loadSetupConfiguration()
		if err != nil {
			http.Error(w, "Failed to load setup configuration: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := StatsResponse{
			FinancialStats:     financialStats,
			SetupConfiguration: setupConfig,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode stats response: "+err.Error(), http.StatusInternalServerError)
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

// loadSetupConfiguration loads the setup configuration from setup.yaml
func loadSetupConfiguration() (SetupConfiguration, error) {
	var result SetupConfiguration

	// Load economic configuration
	economicConfig, err := setup.LoadEconomicsConfig()
	if err != nil {
		return result, err
	}

	// Map configuration values to our response struct
	result.InitialMarketProbability = economicConfig.Economics.MarketCreation.InitialMarketProbability
	result.InitialMarketSubsidization = economicConfig.Economics.MarketCreation.InitialMarketSubsidization
	result.InitialMarketYes = economicConfig.Economics.MarketCreation.InitialMarketYes
	result.InitialMarketNo = economicConfig.Economics.MarketCreation.InitialMarketNo
	result.CreateMarketCost = economicConfig.Economics.MarketIncentives.CreateMarketCost
	result.TraderBonus = economicConfig.Economics.MarketIncentives.TraderBonus
	result.InitialAccountBalance = economicConfig.Economics.User.InitialAccountBalance
	result.MaximumDebtAllowed = economicConfig.Economics.User.MaximumDebtAllowed
	result.MinimumBet = economicConfig.Economics.Betting.MinimumBet
	result.MaxDustPerSale = economicConfig.Economics.Betting.MaxDustPerSale
	result.InitialBetFee = economicConfig.Economics.Betting.BetFees.InitialBetFee
	result.BuySharesFee = economicConfig.Economics.Betting.BetFees.BuySharesFee
	result.SellSharesFee = economicConfig.Economics.Betting.BetFees.SellSharesFee

	return result, nil
}
