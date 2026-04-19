package statshandlers

import (
	"encoding/json"
	"net/http"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/models"

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

// StatsHandler handles requests for both financial stats and setup configuration.
func StatsHandler(db *gorm.DB, configService configsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "Failed to calculate financial stats: database connection is not initialized", http.StatusInternalServerError)
			return
		}
		if configService == nil {
			http.Error(w, "Failed to load setup configuration: configuration service unavailable", http.StatusInternalServerError)
			return
		}

		// Calculate financial stats
		financialStats, err := calculateFinancialStats(db, configService)
		if err != nil {
			http.Error(w, "Failed to calculate financial stats: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Load setup configuration
		setupConfig := loadSetupConfiguration(configService.Current())

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

func calculateFinancialStats(db *gorm.DB, configService configsvc.Service) (FinancialStats, error) {
	var result FinancialStats

	// Calculate TotalMoney by calling the new function
	totalMoney, err := calculateTotalMoney(db, configService.Current())
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
func calculateTotalMoney(db *gorm.DB, economicConfig *configsvc.AppConfig) (int64, error) {
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
func loadSetupConfiguration(economicConfig *configsvc.AppConfig) SetupConfiguration {
	var result SetupConfiguration

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

	return result
}
