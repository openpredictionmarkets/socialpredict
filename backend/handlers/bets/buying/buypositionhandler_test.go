package buybetshandlers

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
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

func TestPlaceBetCore_BalanceAdjustment(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	initialBalance := int64(1000)
	user := modelstesting.GenerateUser("testuser", initialBalance)
	market := modelstesting.GenerateMarket(1, "testuser")

	db.Create(&user)
	db.Create(&market)

	betRequest := models.Bet{
		MarketID: 1,
		Amount:   100,
		Outcome:  "YES",
	}

	// Call PlaceBetCore directly (no HTTP server)
	bet, err := PlaceBetCore(&user, betRequest, db, func() *setup.EconomicConfig {
		return modelstesting.GenerateEconomicConfig()
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Reload user from DB to verify updated balance
	var updatedUser models.User
	db.First(&updatedUser, "username = ?", "testuser")

	expectedBalance := initialBalance - betRequest.Amount - modelstesting.GenerateEconomicConfig().Economics.Betting.BetFees.InitialBetFee
	if updatedUser.AccountBalance != expectedBalance {
		t.Fatalf("Expected balance %d, got %d", expectedBalance, updatedUser.AccountBalance)
	}

	// Verify that the bet was created successfully
	if bet == nil {
		t.Fatalf("Expected bet to be created, got nil")
	}
	if bet.Username != "testuser" {
		t.Errorf("Expected bet username 'testuser', got %s", bet.Username)
	}
}
