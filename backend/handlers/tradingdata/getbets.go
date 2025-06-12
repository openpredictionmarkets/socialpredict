package tradingdata

import (
	"socialpredict/models"
	"time"

	"gorm.io/gorm"
)

type PublicBet struct {
	ID       uint      `json:"betId"`
	Username string    `json:"username"`
	MarketID uint      `json:"marketId"`
	Amount   int64     `json:"amount"`
	PlacedAt time.Time `json:"placedAt"`
	Outcome  string    `json:"outcome,omitempty"`
}

func GetBetsForMarket(db *gorm.DB, marketID uint) []models.Bet {
	var bets []models.Bet

	if err := db.
		Where("market_id = ?", marketID).
		Order("placed_at ASC").
		Find(&bets).Error; err != nil {
		return nil
	}

	return bets
}
