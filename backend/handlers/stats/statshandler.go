package statshandlers

import (
	"net/http"
	"socialpredict/handlers"
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
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		if configService == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		economics := configService.Economics()

		financialStats, err := calculateFinancialStats(db, economics)
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		setupConfig := loadSetupConfiguration(economics)

		response := StatsResponse{
			FinancialStats:     financialStats,
			SetupConfiguration: setupConfig,
		}

		if err := handlers.WriteResult(w, http.StatusOK, response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func calculateFinancialStats(db *gorm.DB, economics configsvc.Economics) (FinancialStats, error) {
	var result FinancialStats

	totalMoney, err := calculateTotalMoney(db, economics.User)
	if err != nil {
		return result, err
	}
	result.TotalMoney = totalMoney

	result.TotalDebtExtended = 0
	result.TotalDebtUtilized = 0
	result.TotalFeesCollected = 0
	result.TotalBonusesPaid = 0
	result.OutstandingPayouts = 0
	result.TotalMoneyInCirculation = 0

	return result, nil
}

func calculateTotalMoney(db *gorm.DB, userConfig configsvc.User) (int64, error) {
	var userCount int64
	if err := db.Model(&models.User{}).Where("user_type = ?", "REGULAR").Count(&userCount).Error; err != nil {
		return 0, err
	}

	totalMoney := userConfig.InitialAccountBalance * userCount
	return totalMoney, nil
}

func loadSetupConfiguration(economics configsvc.Economics) SetupConfiguration {
	var result SetupConfiguration

	result.InitialMarketProbability = economics.MarketCreation.InitialMarketProbability
	result.InitialMarketSubsidization = economics.MarketCreation.InitialMarketSubsidization
	result.InitialMarketYes = economics.MarketCreation.InitialMarketYes
	result.InitialMarketNo = economics.MarketCreation.InitialMarketNo
	result.CreateMarketCost = economics.MarketIncentives.CreateMarketCost
	result.TraderBonus = economics.MarketIncentives.TraderBonus
	result.InitialAccountBalance = economics.User.InitialAccountBalance
	result.MaximumDebtAllowed = economics.User.MaximumDebtAllowed
	result.MinimumBet = economics.Betting.MinimumBet
	result.MaxDustPerSale = economics.Betting.MaxDustPerSale
	result.InitialBetFee = economics.Betting.BetFees.InitialBetFee
	result.BuySharesFee = economics.Betting.BetFees.BuySharesFee
	result.SellSharesFee = economics.Betting.BetFees.SellSharesFee

	return result
}
