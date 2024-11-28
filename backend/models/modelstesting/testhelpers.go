package modelstesting

import (
	"socialpredict/models"
	"time"
)

// GenerateBet is used for Generating fake bets for testing purposes
// helper function to generate bets succinctly for testing
func GenerateBet(amount int64, outcome, username string, marketID uint, offset time.Duration) models.Bet {
	return models.Bet{
		Amount:   amount,
		Outcome:  outcome,
		Username: username,
		PlacedAt: time.Now().Add(offset),
		MarketID: marketID,
	}
}
