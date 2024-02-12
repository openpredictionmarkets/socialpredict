package tradingdata

import (
	"socialpredict/logging"
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

	// Retrieve all bets for the market
	if err := db.Where("market_id = ?", marketID).Find(&bets).Error; err != nil {
		return nil
	}

	logging.LogAnyType(bets, "bets")

	return bets
}
