package modelstesting

import (
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/models"
	"time"
)

// GenerateBet is used for Generating fake bets for testing purposes
// helper function to generate bets succiently for testing
func GenerateBet(amount int64, outcome, username string, marketID uint, offset time.Duration) models.Bet {
	return models.Bet{
		Amount:   amount,
		Outcome:  outcome,
		Username: username,
		PlacedAt: time.Now().Add(offset),
		MarketID: marketID,
	}
}

// helper function to create wpam.ProbabilityChange points succiently
func GenerateProbability(probabilities ...float64) []wpam.ProbabilityChange {
	var changes []wpam.ProbabilityChange
	for _, p := range probabilities {
		changes = append(changes, wpam.ProbabilityChange{Probability: p})
	}
	return changes
}
