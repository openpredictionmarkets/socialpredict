package statshandlers

import (
	"socialpredict/handlers/bets/betutils"
	"socialpredict/models"
	"socialpredict/repository"
	"socialpredict/setup"

	"gorm.io/gorm"
)

// calculateTotalRevenue sums up all revenue components from betting fees across all markets
func calculateTotalRevenue(db *gorm.DB) int64 {

	// fees asssessed from the initial market creation fee, for all markets
	marketCreationFees := sumAllMarketCreationFees(db)
	// initial transaction fees

	// buying fees across all bets

	// selling fees across all bets

	totalRevenue := marketCreationFees

	return totalRevenue
}

func sumAllMarketCreationFees(db *gorm.DB) int64 {

	economicConfig, err := setup.LoadEconomicsConfig()
	if err != nil {
		return 0
	}

	gormDatabase := &repository.GormDatabase{DB: db}

	marketRepo := repository.NewMarketRepository(gormDatabase)
	marketsCount, err := marketRepo.CountMarkets()
	if err != nil {
		return 0
	}

	return economicConfig.Economics.MarketIncentives.CreateMarketCost * marketsCount
}

// calculateMarketRevenue calculates the initial fees collected from a specific market
func calculateMarketInitialFees(db *gorm.DB, marketID uint) int64 {

	economicConfig, err := setup.LoadEconomicsConfig()
	if err != nil {
		return 0
	}

	var totalMarketRevenue int64 = 0

	gormDatabase := &repository.GormDatabase{DB: db}
	repo := repository.NewUserRepository(gormDatabase)

	users, err := repo.GetAllUsers()
	if err != nil {
		return 0
	}

	for _, user := range users {
		betRequest := models.Bet{
			MarketID: marketID,
			// Ensure you populate other necessary fields as required
		}
		// Assume GetBetFees uses PublicUserType and is adjusted to work with it
		fees := betutils.GetBetFees(db, &user, betRequest)
		totalMarketRevenue += fees
	}

	return totalMarketRevenue
}
