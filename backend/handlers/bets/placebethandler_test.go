package betshandlers_test

import (
	"testing"

	betshandlers "socialpredict/handlers/bets"
	"socialpredict/models"
	"socialpredict/setup"
)

func TestCheckUserBalance(t *testing.T) {
	// Define a fake user
	user := &models.User{
		Username:       "testuser",
		AccountBalance: 500,
	}

	// Define EconConfigLoader
	loadEconConfig := func() setup.EconConfig {
		return setup.EconConfig{
			Economics: setup.EconomicsConfig{
				User: setup.UserConfig{
					MaximumDebtAllowed: 100,
				},
			},
		}
	}

	// Test cases
	tests := []struct {
		name         string
		betRequest   models.Bet
		sumOfBetFees int64
		expectsError bool
	}{
		{
			name: "Sufficient balance",
			betRequest: models.Bet{
				Amount: 50,
			},
			sumOfBetFees: 10,
			expectsError: false,
		},
		{
			name: "Insufficient balance",
			betRequest: models.Bet{
				Amount: 450,
			},
			sumOfBetFees: 100,
			expectsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := betshandlers.CheckUserBalance(user, tt.betRequest, tt.sumOfBetFees, loadEconConfig)
			if (err != nil) != tt.expectsError {
				t.Errorf("got error = %v, expected error = %v", err != nil, tt.expectsError)
			}
		})
	}
}
