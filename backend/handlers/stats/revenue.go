package statshandlers

import (
	"log"
	"socialpredict/handlers/bets/betutils"
	marketshandlers "socialpredict/handlers/markets"
	"socialpredict/models"
	"socialpredict/repository"

	"gorm.io/gorm"
)

// calculateTotalRevenue sums up all revenue components from betting fees across all markets
func calculateTotalRevenue(db *gorm.DB) int64 {
	totalRevenue := int64(0)

	// Use the existing ListMarkets function to get all markets
	markets, err := marketshandlers.ListMarkets(db)
	if err != nil {
		log.Printf("Error fetching markets: %v", err)
		return 0
	}

	// Assuming you have a way to fetch all users or calculate fees without needing each user
	// If you need each user, you should fetch them or iterate over users differently
	for _, market := range markets {
		marketID := uint(market.ID)
		// This assumes you have a function or a way to calculate the total fees for a market
		marketRevenue := calculateMarketRevenue(db, marketID)
		totalRevenue += marketRevenue
	}

	return totalRevenue
}

// calculateMarketRevenue calculates the total fees collected from a specific market
func calculateMarketRevenue(db *gorm.DB, marketID uint) int64 {
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
