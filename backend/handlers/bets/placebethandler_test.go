package betshandlers

import (
	"testing"

	"socialpredict/models"
	"socialpredict/setup"
)

func TestCheckUserBalance_CustomConfig(t *testing.T) {
	user := &models.User{
		PublicUser: models.PublicUser{
			Username:       "testuser",
			AccountBalance: 0,
		},
	}

	// Define a custom loadEconConfig function with MaximumDebtAllowed to use in the test
	loadEconConfig := func() *setup.EconomicConfig {
		return &setup.EconomicConfig{
			Economics: setup.Economics{
				User: setup.User{
					MaximumDebtAllowed: 100,
				},
			},
		}
	}

	tests := []struct {
		name         string
		betRequest   models.Bet
		sumOfBetFees int64
		expectsError bool
	}{
		// Buying Shares Cases
		{
			// Starting with AccountBalance 0, MaximumDebtAllowed 100, place a bet of 99, fee 1
			name: "Sufficient balance.",
			betRequest: models.Bet{
				Amount: 99,
			},
			sumOfBetFees: 1,
			expectsError: false,
		},
		{
			// Starting with AccountBalance 0, MaximumDebtAllowed 100, place a bet of 1, fee 99
			name: "Sufficient balance.",
			betRequest: models.Bet{
				Amount: 1,
			},
			sumOfBetFees: 99,
			expectsError: false,
		},
		{
			// Starting with AccountBalance 0, MaximumDebtAllowed 100, place a bet of 100, fee 1
			name: "Insufficient balance, fee prevents bet",
			betRequest: models.Bet{
				Amount: 100,
			},
			sumOfBetFees: 1,
			expectsError: true,
		},
		{
			// Starting with AccountBalance 0, MaximumDebtAllowed 100, place a bet of 1, fee 100
			name: "Insufficient balance, fee prevents bet",
			betRequest: models.Bet{
				Amount: 1,
			},
			sumOfBetFees: 100,
			expectsError: true,
		},
		// Selling Shares Cases
		{
			// Starting with AccountBalance 0, MaximumDebtAllowed 100, sell 1, fee 101
			name: "Sufficient balance.",
			betRequest: models.Bet{
				Amount: -1,
			},
			sumOfBetFees: 101,
			expectsError: false,
		},
		{
			// Starting with AccountBalance 0, MaximumDebtAllowed 100, sell 1, fee 102
			name: "Insufficient balance, fee prevents bet",
			betRequest: models.Bet{
				Amount: -1,
			},
			sumOfBetFees: 102,
			expectsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkUserBalance(user, tt.betRequest, tt.sumOfBetFees, loadEconConfig)
			if (err != nil) != tt.expectsError {
				t.Errorf("got error = %v, expected error = %v", err != nil, tt.expectsError)
			}
		})
	}
}
