package financials

import (
	"fmt"
	"socialpredict/handlers/positions"
	"socialpredict/models"

	"gorm.io/gorm"
)

// profit = selling price - cost price
// cost price = amount bet on particular market
// selling price = current position calculated

// profit(user(market)) -> individual user and market
// profit(market(users)) -> function that shows all user profits on a market
// profit(user(markets)) -> function that sums up for a user, all profits on markets they are in

// If we want to build a market leaderboard we need profit(user(market))
// For all users on a market, calculate profit(user(market))
// When done, rank and display them in order, greatest (+) to least (-)

func userMarketProfit(db *gorm.DB, username string, marketId string) (int64, error) {

	// get yes and no shares, maximum of the two will be the selling price
	// user can only own either yes or no shares
	sellingPrice, err := positions.GetMaxShares_WPAM_DBPM(db, marketId, username)
	if err != nil {
		return 0, err
	}

	// amount spent on market so far
	// we can do a custom gorm query on the bets table to extract the user in sql
	// rather than extracting the entire bets table and filtering in golang
	var totalSpent int64 = 0

	err = db.Model(&models.Bet{}).
		Select("COALESCE(SUM(amount), 0) AS total_amount").
		Where("market_id = ? AND username = ?", marketId, username).
		Scan(&totalSpent).Error

	if err != nil {
		return 0, fmt.Errorf("failed to calculate total bet amount: %w", err)
	}

	return (sellingPrice - totalSpent), err

}
