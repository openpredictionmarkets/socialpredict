package betshandlers

import (
	"socialpredict/models"
	"socialpredict/util"
	"time"
)

type PublicBet struct {
	BetID    uint      `json:"betId"`
	MarketID uint      `json:"marketId"`
	Amount   float64   `json:"amount"`
	PlacedAt time.Time `json:"placedAt"`
	Outcome  string    `json:"outcome,omitempty"`
}

func GetBetsForMarket(marketID uint64) []models.Bet {
	var bets []models.Bet

	// Retrieve all bets for the market
	db := util.GetDB()
	if err := db.Where("market_id = ?", marketID).Find(&bets).Error; err != nil {
		return nil
	}

	return bets
}
