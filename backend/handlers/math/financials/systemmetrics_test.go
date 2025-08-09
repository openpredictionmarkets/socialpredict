package financials

import (
	"testing"

	"socialpredict/models/modelstesting"
	"socialpredict/setup"
)

func TestComputeSystemMetrics(t *testing.T) {
	// Mock economics config loader
	mockEconLoader := func() *setup.EconomicConfig {
		return &setup.EconomicConfig{
			Economics: setup.Economics{
				User: setup.User{
					InitialAccountBalance: 0,
					MaximumDebtAllowed:    500,
				},
				MarketIncentives: setup.MarketIncentives{
					CreateMarketCost: 50,
				},
				MarketCreation: setup.MarketCreation{
					InitialMarketSubsidization: 100,
				},
				Betting: setup.Betting{
					BetFees: setup.BetFees{
						InitialBetFee: 5,
					},
				},
			},
		}
	}

	// Test with empty database
	t.Run("EmptyDatabase", func(t *testing.T) {
		db := modelstesting.NewFakeDB(t)

		metrics, err := ComputeSystemMetrics(db, mockEconLoader)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// With no users, all metrics should be zero
		if metrics.MoneyCreated.UserDebtCapacity.Value != 0 {
			t.Errorf("Expected user debt capacity 0, got %d", metrics.MoneyCreated.UserDebtCapacity.Value)
		}
		if metrics.MoneyUtilized.TotalUtilized.Value != 0 {
			t.Errorf("Expected total utilized 0, got %d", metrics.MoneyUtilized.TotalUtilized.Value)
		}
		if metrics.Verification.Balanced.Value != 1 {
			t.Errorf("Expected balanced metrics for empty database")
		}
	})

	// Test with basic data
	t.Run("BasicData", func(t *testing.T) {
		db := modelstesting.NewFakeDB(t)

		// Create test users
		user1 := modelstesting.GenerateUser("user1", 950)
		user2 := modelstesting.GenerateUser("user2", -100)
		db.Create(&user1)
		db.Create(&user2)

		// Create test market
		market := modelstesting.GenerateMarket(1, "user1")
		market.IsResolved = false
		db.Create(&market)

		// Create test bets (first buy from each user)
		bet1 := modelstesting.GenerateBet(50, "YES", "user1", uint(market.ID), 0)
		bet2 := modelstesting.GenerateBet(30, "YES", "user2", uint(market.ID), 0)
		db.Create(&bet1)
		db.Create(&bet2)

		metrics, err := ComputeSystemMetrics(db, mockEconLoader)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Expected calculations:
		// User debt capacity: 2 users × 500 = 1000
		// Money in wallets: |950| + |-100| = 1050
		// Unused debt: (500-0) + (500-100) = 900
		// Market creation fees: 1 market × 50 = 50
		// Participation fees: 2 first-time bets × 5 = 10
		// Active bet volume: (50+30) + 100 subsidization = 180
		// Total utilized: 1050 + 900 + 180 + 50 + 10 + 0 = 2190
		// Surplus: 1000 - 2190 = -1190

		if metrics.MoneyCreated.UserDebtCapacity.Value != 1000 {
			t.Errorf("Expected user debt capacity 1000, got %d", metrics.MoneyCreated.UserDebtCapacity.Value)
		}

		if metrics.MoneyCreated.NumUsers.Value != 2 {
			t.Errorf("Expected 2 users, got %d", metrics.MoneyCreated.NumUsers.Value)
		}

		if metrics.MoneyUtilized.MoneyInWallets.Value != 1050 {
			t.Errorf("Expected money in wallets 1050, got %d", metrics.MoneyUtilized.MoneyInWallets.Value)
		}

		if metrics.MoneyUtilized.MarketCreationFees.Value != 50 {
			t.Errorf("Expected market creation fees 50, got %d", metrics.MoneyUtilized.MarketCreationFees.Value)
		}

		if metrics.MoneyUtilized.ParticipationFees.Value != 10 {
			t.Errorf("Expected participation fees 10, got %d", metrics.MoneyUtilized.ParticipationFees.Value)
		}

		if metrics.MoneyUtilized.ActiveBetVolume.Value != 180 {
			t.Errorf("Expected active bet volume 180, got %d", metrics.MoneyUtilized.ActiveBetVolume.Value)
		}

		if metrics.MoneyUtilized.UnusedDebt.Value != 900 {
			t.Errorf("Expected unused debt 900, got %d", metrics.MoneyUtilized.UnusedDebt.Value)
		}

		if metrics.MoneyUtilized.TotalUtilized.Value != 2190 {
			t.Errorf("Expected total utilized 2190, got %d", metrics.MoneyUtilized.TotalUtilized.Value)
		}

		if metrics.Verification.Surplus.Value != -1190 {
			t.Errorf("Expected surplus -1190, got %d", metrics.Verification.Surplus.Value)
		}

		if metrics.Verification.Balanced.Value != 0 {
			t.Errorf("Expected unbalanced (0), got %d", metrics.Verification.Balanced.Value)
		}

		// Verify formulas and explanations exist
		if metrics.MoneyCreated.UserDebtCapacity.Formula == "" {
			t.Error("Expected formula for user debt capacity")
		}
		if metrics.MoneyUtilized.MoneyInWallets.Explanation == "" {
			t.Error("Expected explanation for money in wallets")
		}
	})

	// Test error handling
	t.Run("NilDatabase", func(t *testing.T) {
		_, err := ComputeSystemMetrics(nil, mockEconLoader)
		if err == nil {
			t.Error("Expected error for nil database, got none")
		}
	})
}
