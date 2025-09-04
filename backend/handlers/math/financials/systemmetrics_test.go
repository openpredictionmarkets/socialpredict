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

		// With no users, all metrics should be zero - use proper assertions
		if val, ok := metrics.MoneyCreated.UserDebtCapacity.Value.(int64); ok && val == 0 {
			t.Logf("✓ User debt capacity is 0 as expected")
		} else {
			t.Errorf("Expected user debt capacity 0, got %v", metrics.MoneyCreated.UserDebtCapacity.Value)
		}
		if val, ok := metrics.MoneyUtilized.TotalUtilized.Value.(int64); ok && val == 0 {
			t.Logf("✓ Total utilized is 0 as expected")
		} else {
			t.Errorf("Expected total utilized 0, got %v", metrics.MoneyUtilized.TotalUtilized.Value)
		}
		if val, ok := metrics.Verification.Balanced.Value.(bool); ok && val == true {
			t.Logf("✓ Metrics are balanced as expected")
		} else {
			t.Errorf("Expected balanced metrics for empty database, got %v", metrics.Verification.Balanced.Value)
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
		// Unused debt: (500-0) + (500-100) = 900
		// Market creation fees: 1 market × 50 = 50
		// Participation fees: 2 first-time bets × 5 = 10
		// Active bet volume: (50+30) = 80 (excludes subsidization)
		// Total utilized: 900 + 80 + 50 + 10 + 0 = 1040
		// Surplus: 1000 - 1040 = -40

		if val, ok := metrics.MoneyCreated.UserDebtCapacity.Value.(int64); ok && val == 1000 {
			t.Logf("✓ User debt capacity is 1000 as expected")
		} else {
			t.Errorf("Expected user debt capacity 1000, got %v", metrics.MoneyCreated.UserDebtCapacity.Value)
		}

		if val, ok := metrics.MoneyCreated.NumUsers.Value.(int64); ok && val == 2 {
			t.Logf("✓ Number of users is 2 as expected")
		} else {
			t.Errorf("Expected 2 users, got %v", metrics.MoneyCreated.NumUsers.Value)
		}

		if val, ok := metrics.MoneyUtilized.MarketCreationFees.Value.(int64); ok && val == 50 {
			t.Logf("✓ Market creation fees are 50 as expected")
		} else {
			t.Errorf("Expected market creation fees 50, got %v", metrics.MoneyUtilized.MarketCreationFees.Value)
		}

		if val, ok := metrics.MoneyUtilized.ParticipationFees.Value.(int64); ok && val == 10 {
			t.Logf("✓ Participation fees are 10 as expected")
		} else {
			t.Errorf("Expected participation fees 10, got %v", metrics.MoneyUtilized.ParticipationFees.Value)
		}

		if val, ok := metrics.MoneyUtilized.ActiveBetVolume.Value.(int64); ok && val == 80 {
			t.Logf("✓ Active bet volume is 80 as expected")
		} else {
			t.Errorf("Expected active bet volume 80, got %v", metrics.MoneyUtilized.ActiveBetVolume.Value)
		}

		if val, ok := metrics.MoneyUtilized.UnusedDebt.Value.(int64); ok && val == 900 {
			t.Logf("✓ Unused debt is 900 as expected")
		} else {
			t.Errorf("Expected unused debt 900, got %v", metrics.MoneyUtilized.UnusedDebt.Value)
		}

		if val, ok := metrics.MoneyUtilized.TotalUtilized.Value.(int64); ok && val == 1040 {
			t.Logf("✓ Total utilized is 1040 as expected")
		} else {
			t.Errorf("Expected total utilized 1040, got %v", metrics.MoneyUtilized.TotalUtilized.Value)
		}

		if val, ok := metrics.Verification.Surplus.Value.(int64); ok && val == -40 {
			t.Logf("✓ Surplus is -40 as expected")
		} else {
			t.Errorf("Expected surplus -40, got %v", metrics.Verification.Surplus.Value)
		}

		if val, ok := metrics.Verification.Balanced.Value.(bool); ok && val == false {
			t.Logf("✓ Metrics are unbalanced as expected")
		} else {
			t.Errorf("Expected unbalanced (false), got %v", metrics.Verification.Balanced.Value)
		}

		// Verify formulas and explanations exist
		if metrics.MoneyCreated.UserDebtCapacity.Formula == "" {
			t.Error("Expected formula for user debt capacity")
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
