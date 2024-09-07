package betutils

import (
	"log"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"
	"socialpredict/setup"

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
// If this is the first bet on this market for the user, apply a fee.
func getUserInitialBetFee(db *gorm.DB, marketID uint, user *models.User) int64 {
	// Fetch bets for the market
	allBetsOnMarket := tradingdata.GetBetsForMarket(db, marketID)

	// Check if the user has placed any bets on this market
	for _, bet := range allBetsOnMarket {
		if bet.Username == user.Username {
			// User has placed a bet, so no initial fee is applicable
			return 0
		}
	}

	// This is the user's first bet on this market, apply the initial bet fee
	return appConfig.Economics.Betting.BetFees.InitialBetFee
}

func getTransactionFee(betRequest models.Bet) int64 {

	var transactionFee int64

	// if amount > 0, buying share, else selling share
	if betRequest.Amount > 0 {
		transactionFee = appConfig.Economics.Betting.BetFees.BuySharesFee
	} else {
		transactionFee = appConfig.Economics.Betting.BetFees.SellSharesFee
	}

	return transactionFee
}

func GetBetFees(db *gorm.DB, user *models.User, betRequest models.Bet) int64 {

	MarketID := betRequest.MarketID

	initialBetFee := getUserInitialBetFee(db, MarketID, user)
	transactionFee := getTransactionFee(betRequest)

	sumOfBetFees := initialBetFee + transactionFee

	return sumOfBetFees
}
