package modelstesting

import (
	"fmt"
	"socialpredict/handlers/math/probabilities/wpam"
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

func GenerateMarket(id int64, creatorUsername string) models.Market {
	return models.Market{
		ID:                 id,
		QuestionTitle:      "Test Market",
		Description:        "Test Description",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(24 * time.Hour),
		InitialProbability: 0.5,
		CreatorUsername:    creatorUsername,
	}
}

func GenerateUser(username string, startingBalance int64) models.User {
	now := time.Now().UnixNano()
	return models.User{
		PublicUser: models.PublicUser{
			Username:              username,
			DisplayName:           fmt.Sprintf("%s_display_%d", username, now),
			UserType:              "regular",
			InitialAccountBalance: startingBalance,
			AccountBalance:        startingBalance,
		},
		PrivateUser: models.PrivateUser{
			Email:    fmt.Sprintf("%s_%d@example.com", username, now),
			APIKey:   fmt.Sprintf("api-key-%d", now),
			Password: "password",
		},
	}
}

// helper function to create wpam.ProbabilityChange points succinctly
func GenerateProbability(probabilities ...float64) []wpam.ProbabilityChange {
	var changes []wpam.ProbabilityChange
	for _, p := range probabilities {
		changes = append(changes, wpam.ProbabilityChange{Probability: p})
	}
	return changes
}
