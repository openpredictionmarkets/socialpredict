package statshandlers

import (
	"fmt"
	"socialpredict/repository"
	"socialpredict/setup"

	"gorm.io/gorm"
)

// calculateTotalRevenue sums up all revenue components from betting fees across all markets
func calculateTotalRevenue(db *gorm.DB) int64 {

	// fees asssessed from the initial market creation fee, for all markets
	marketCreationFees := sumAllMarketCreationFees(db)
	// initial transaction fees from frist time bets
	totalInitialFees := calculateMarketInitialFees(db)

	// buying fees across all bets

	// selling fees across all bets

	totalRevenue := marketCreationFees + totalInitialFees

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

func calculateMarketInitialFees(db *gorm.DB) int64 {
	// Load the economic configuration
	economicConfig, err := setup.LoadEconomicsConfig()
	if err != nil {
		fmt.Println("Error loading economics config:", err) // Debugging config load error
		return 0
	}

	// Create a new GormDatabase object wrapping the db
	gormDatabase := &repository.GormDatabase{DB: db}

	// Create a new BetsRepository
	betsRepo := repository.NewBetsRepository(gormDatabase)

	// Call FirstTimeBets to get the total number of initial bets across all markets
	totalInitialBets, err := betsRepo.FirstTimeBets()
	if err != nil {
		fmt.Println("Error in FirstTimeBets:", err) // Debugging error output
		return 0
	}

	fmt.Println("Total initial bets:", totalInitialBets) // Debugging initial bets output

	// Multiply the total count of initial bets by the initial bet fee
	initialBetFee := economicConfig.Economics.Betting.BetFees.InitialBetFee
	fmt.Println("Initial bet fee:", initialBetFee) // Debugging initial bet fee output

	totalInitialFees := totalInitialBets * initialBetFee
	fmt.Println("Total initial fees:", totalInitialFees) // Debugging total fees output

	return totalInitialFees
}
