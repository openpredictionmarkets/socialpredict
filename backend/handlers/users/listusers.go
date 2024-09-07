package usershandlers

import (
	"log"
	"socialpredict/models"

	"gorm.io/gorm"
)

// ListUserMarkets lists all markets that a specific user is betting in, ordered by the date of the last bet.
func ListUserMarkets(db *gorm.DB, userID int64) ([]models.Market, error) {
	var markets []models.Market

	// Query to find all markets where the user has bets, ordered by the date of the last bet
	query := db.Table("markets").
		Joins("join bets on bets.market_id = markets.id").
		Where("bets.user_id = ?", userID).
		Order("bets.created_at DESC").
		Distinct("markets.*").
		Find(&markets)

	if query.Error != nil {
		log.Printf("Error fetching user's markets: %v", query.Error)
		return nil, query.Error
	}

	return markets, nil
}
