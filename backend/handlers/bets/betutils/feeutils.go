package betutils

import (
	"errors"
	"socialpredict/models"
	"time"

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

// Get initial bet fee, if applicable, for user on market.
// If this is the first bet on this market, apply a fee.
func GetUserInitialBetFee(db *gorm.DB, marketID uint, user *models.User) initialBetFee int64 {

	var initialBetFee int64

	// Fetch bets for the market
	var allBetsOnMarket []models.Bet
	allBetsOnMarket = tradingdata.GetBetsForMarket(db, marketID)

	var market models.Market
	if result := db.First(allBetsOnMarket.username, user); result.Error != nil {
		initialBetFee = appConfig.Economics.Betting.BetFees.initialBetFee
	} else {
		initialBetFee = 0
	}

	return initialBetFee
}