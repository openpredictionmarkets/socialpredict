package statshandlers

import (
	"context"
	"net/http"
	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
	configsvc "socialpredict/internal/service/config"
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

type FinancialStatsService interface {
	ComputeFinancialStats(ctx context.Context, config analytics.StatsConfig) (analytics.FinancialStats, error)
}

// StatsHandler handles requests for both financial stats and setup configuration.
func StatsHandler(statsService FinancialStatsService, configService configsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if statsService == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		if configService == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		economics := configService.Economics()

		financialStats, err := statsService.ComputeFinancialStats(r.Context(), analytics.StatsConfig{
			InitialAccountBalance: economics.User.InitialAccountBalance,
		})
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		setupConfig := loadSetupConfiguration(economics)

		response := StatsResponse{
			FinancialStats:     mapFinancialStats(financialStats),
			SetupConfiguration: setupConfig,
		}

		if err := handlers.WriteResult(w, http.StatusOK, response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func mapFinancialStats(stats analytics.FinancialStats) FinancialStats {
	return FinancialStats{
		TotalMoney:              stats.TotalMoney,
		TotalDebtExtended:       stats.TotalDebtExtended,
		TotalDebtUtilized:       stats.TotalDebtUtilized,
		TotalFeesCollected:      stats.TotalFeesCollected,
		TotalBonusesPaid:        stats.TotalBonusesPaid,
		OutstandingPayouts:      stats.OutstandingPayouts,
		TotalMoneyInCirculation: stats.TotalMoneyInCirculation,
	}
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
