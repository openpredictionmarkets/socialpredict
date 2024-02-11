package tradingdata

import (
	"socialpredict/models"
	"time"

	"gorm.io/gorm"
)

type PublicBet struct {
	BetID    int64     `json:"betId"`
	MarketID int64     `json:"marketId"`
	Amount   float64   `json:"amount"`
	PlacedAt time.Time `json:"placedAt"`
	Outcome  string    `json:"outcome,omitempty"`
}

func GetBetsForMarket(db *gorm.DB, marketID uint) []models.Bet {
	var bets []models.Bet

	// Retrieve all bets for the market
	if err := db.Where("market_id = ?", marketID).Find(&bets).Error; err != nil {
		return nil
	}

	return bets
}
