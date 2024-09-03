package betutils

import (
	"log"
	"socialpredict/handlers/tradingdata"
	"socialpredict/logging"
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
// If this is the first bet on this market, apply a fee.
func getUserInitialBetFee(db *gorm.DB, marketID uint, user *models.User) int64 {

	var initialBetFee int64

	// Fetch bets for the market
	allBetsOnMarket := tradingdata.GetBetsForMarket(db, marketID)

	var totalBetCount int64 = 0

	for _, bet := range allBetsOnMarket {
		if bet.Username == user.Username {
			totalBetCount += 1
		}
	}

	// if we have no bets on this market yet, then this is our first bet
	if totalBetCount == 0 {
		initialBetFee = appConfig.Economics.Betting.BetFees.InitialBetFee
	} else {
		initialBetFee = 0
	}

	logging.LogAnyType(initialBetFee, "initialBetFee")

	return initialBetFee
}

func getTransactionFee(betRequest models.Bet) int64 {

	var transactionFee int64

	// if amount > 0, buying share, else selling share
	if betRequest.Amount > 0 {
		transactionFee = appConfig.Economics.Betting.BetFees.BuySharesFee
	} else {
		transactionFee = appConfig.Economics.Betting.BetFees.SellSharesFee
	}

	logging.LogAnyType(transactionFee, "transactionFee")

	return transactionFee
}

func GetSumBetFees(db *gorm.DB, user *models.User, betRequest models.Bet) int64 {

	MarketID := betRequest.MarketID

	initialBetFee := getUserInitialBetFee(db, MarketID, user)
	transactionFee := getTransactionFee(betRequest)

	sumOfBetFees := initialBetFee + transactionFee

	logging.LogAnyType(sumOfBetFees, "sumOfBetFees in summing function")

	return sumOfBetFees
}
