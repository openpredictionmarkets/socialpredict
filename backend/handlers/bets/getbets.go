package handlers

import (
	"socialpredict/models"
	"socialpredict/util"
)

func GetBetsForMarket(marketID uint64) ([]models.Bet, error) {
	var bets []models.Bet

	// Retrieve all bets for the market
	db := util.GetDB()
	if err := db.Where("market_id = ?", marketID).Find(&bets).Error; err != nil {
		return nil, err
	}

	return bets, nil
}
