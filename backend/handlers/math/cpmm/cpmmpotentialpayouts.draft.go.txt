package handlers

import (
	"socialpredict/models"
	"time"

	"gorm.io/gorm"
)

// PayoutInfo holds the calculated payout information for a user
type PayoutInfo struct {
	Username string
	Payout   float64
}

// calculatePotentialPayouts calculates potential payouts for a market at a given state based upon CPMM.
func calculatePotentialPayouts(market *models.Market, db *gorm.DB, atTime time.Time) ([]PayoutInfo, error) {
	var bets []models.Bet
	if err := db.Where("market_id = ? AND placed_at <= ?", market.ID, atTime).Find(&bets).Error; err != nil {
		return nil, err
	}

	var totalYes, totalNo float64
	for _, bet := range bets {
		if bet.Outcome == "YES" {
			totalYes += bet.Amount
		} else if bet.Outcome == "NO" {
			totalNo += bet.Amount
		}
	}

	var payoutInfos []PayoutInfo
	for _, bet := range bets {
		payout := CalculateCPMMPayoutForOutcome(bet, totalYes, totalNo, bet.Outcome, market.CurrentState)
		payoutInfos = append(payoutInfos, PayoutInfo{Username: bet.Username, Payout: payout})
	}

	return payoutInfos, nil
}
