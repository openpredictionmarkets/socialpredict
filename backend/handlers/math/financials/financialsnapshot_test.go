package financials

import (
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
	"testing"
	"time"
)

func TestComputeUserFinancials_NewUser_NoPositions(t *testing.T) {
	// Test case: Clean new user with no bets/positions
	db := modelstesting.NewFakeDB(t)

	// Create a user with initial balance
	user := modelstesting.GenerateUser("testuser", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Mock economic config
	econ := &setup.EconomicConfig{
		Economics: setup.Economics{
			User: setup.User{
				MaximumDebtAllowed: 500,
			},
		},
	}

	// Compute financial snapshot
	snapshot, err := ComputeUserFinancials(db, user.Username, user.AccountBalance, econ)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify core financial metrics for clean user
	expected := map[string]int64{
		"accountBalance":     1000,
		"maximumDebtAllowed": 500,
		"amountInPlay":       0,
		"amountBorrowed":     0,
		"retainedEarnings":   1000, // account balance - amount in play (0)
		"equity":             1000, // retained earnings + amount in play - amount borrowed
		"tradingProfits":     0,
		"workProfits":        0,
		"totalProfits":       0,
		"amountInPlayActive": 0,
		"totalSpent":         0,
		"totalSpentInPlay":   0,
		"realizedProfits":    0,
		"potentialProfits":   0,
		"realizedValue":      0,
		"potentialValue":     0,
	}

	for key, expectedVal := range expected {
		if snapshot[key] != expectedVal {
			t.Errorf("For %s: expected %d, got %d", key, expectedVal, snapshot[key])
		}
	}
}

func TestComputeUserFinancials_NegativeBalance_Borrowing(t *testing.T) {
	// Test case: User with negative balance (borrowing money)
	db := modelstesting.NewFakeDB(t)

	// Create a user with negative balance
	user := modelstesting.GenerateUser("borrower", -50)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Mock economic config
	econ := &setup.EconomicConfig{
		Economics: setup.Economics{
			User: setup.User{
				MaximumDebtAllowed: 500,
			},
		},
	}

	// Compute financial snapshot
	snapshot, err := ComputeUserFinancials(db, user.Username, user.AccountBalance, econ)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify borrowing calculations
	if snapshot["accountBalance"] != -50 {
		t.Errorf("Expected accountBalance -50, got %d", snapshot["accountBalance"])
	}
	if snapshot["amountBorrowed"] != 50 {
		t.Errorf("Expected amountBorrowed 50, got %d", snapshot["amountBorrowed"])
	}
	if snapshot["retainedEarnings"] != -50 { // account balance - amount in play (0)
		t.Errorf("Expected retainedEarnings -50, got %d", snapshot["retainedEarnings"])
	}
	// equity = retainedEarnings + amountInPlay - amountBorrowed
	// equity = -50 + 0 - 50 = -100
	expectedEquity := int64(-100)
	if snapshot["equity"] != expectedEquity {
		t.Errorf("Expected equity %d, got %d", expectedEquity, snapshot["equity"])
	}
}

func TestComputeUserFinancials_WithActivePositions(t *testing.T) {
	// Test case: User with positions in active (unresolved) markets
	db := modelstesting.NewFakeDB(t)

	// Create user and market
	user := modelstesting.GenerateUser("trader", 500)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	market := modelstesting.GenerateMarket(1, user.Username)
	market.IsResolved = false
	market.ResolutionResult = ""
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("Failed to create market: %v", err)
	}

	// Create bets for the user
	bet1 := modelstesting.GenerateBet(100, "YES", user.Username, uint(market.ID), 0)
	if err := db.Create(&bet1).Error; err != nil {
		t.Fatalf("Failed to create bet1: %v", err)
	}

	bet2 := modelstesting.GenerateBet(50, "NO", user.Username, uint(market.ID), time.Minute)
	if err := db.Create(&bet2).Error; err != nil {
		t.Fatalf("Failed to create bet2: %v", err)
	}

	// Mock economic config
	econ := &setup.EconomicConfig{
		Economics: setup.Economics{
			User: setup.User{
				MaximumDebtAllowed: 500,
			},
		},
	}

	// Note: Since we're testing the financial logic, we would need to mock
	// the position calculation results. For this test, let's assume the position
	// calculations work correctly and focus on the financial aggregation logic.

	// This test would need to be completed with actual market position data
	// For now, let's verify the function can be called without error
	_, err := ComputeUserFinancials(db, user.Username, user.AccountBalance, econ)
	if err != nil {
		t.Fatalf("Expected no error with active positions, got: %v", err)
	}
}

func TestComputeUserFinancials_WithResolvedPositions(t *testing.T) {
	// Test case: User with positions in resolved markets
	db := modelstesting.NewFakeDB(t)

	// Create user and resolved market
	user := modelstesting.GenerateUser("winner", 200)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	market := modelstesting.GenerateMarket(2, user.Username)
	market.IsResolved = true
	market.ResolutionResult = "YES"
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("Failed to create market: %v", err)
	}

	// Create bets for the user
	bet1 := modelstesting.GenerateBet(75, "YES", user.Username, uint(market.ID), 0)
	if err := db.Create(&bet1).Error; err != nil {
		t.Fatalf("Failed to create bet: %v", err)
	}

	// Mock economic config
	econ := &setup.EconomicConfig{
		Economics: setup.Economics{
			User: setup.User{
				MaximumDebtAllowed: 500,
			},
		},
	}

	// This test would need to be completed with actual resolved market position data
	_, err := ComputeUserFinancials(db, user.Username, user.AccountBalance, econ)
	if err != nil {
		t.Fatalf("Expected no error with resolved positions, got: %v", err)
	}
}

func TestComputeUserFinancials_MixedPositions(t *testing.T) {
	// Test case: User with both active and resolved positions
	db := modelstesting.NewFakeDB(t)

	// Create user
	user := modelstesting.GenerateUser("mixedtrader", 300)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create active market
	activeMarket := modelstesting.GenerateMarket(3, user.Username)
	activeMarket.IsResolved = false
	activeMarket.ResolutionResult = ""
	if err := db.Create(&activeMarket).Error; err != nil {
		t.Fatalf("Failed to create active market: %v", err)
	}

	activeBet := modelstesting.GenerateBet(100, "YES", user.Username, uint(activeMarket.ID), 0)
	if err := db.Create(&activeBet).Error; err != nil {
		t.Fatalf("Failed to create active bet: %v", err)
	}

	// Create resolved market
	resolvedMarket := modelstesting.GenerateMarket(4, user.Username)
	resolvedMarket.IsResolved = true
	resolvedMarket.ResolutionResult = "NO"
	if err := db.Create(&resolvedMarket).Error; err != nil {
		t.Fatalf("Failed to create resolved market: %v", err)
	}

	resolvedBet := modelstesting.GenerateBet(50, "NO", user.Username, uint(resolvedMarket.ID), time.Minute)
	if err := db.Create(&resolvedBet).Error; err != nil {
		t.Fatalf("Failed to create resolved bet: %v", err)
	}

	// Mock economic config
	econ := &setup.EconomicConfig{
		Economics: setup.Economics{
			User: setup.User{
				MaximumDebtAllowed: 500,
			},
		},
	}

	// Test mixed positions scenario
	snapshot, err := ComputeUserFinancials(db, user.Username, user.AccountBalance, econ)
	if err != nil {
		t.Fatalf("Expected no error with mixed positions, got: %v", err)
	}

	// Verify that we get a proper response structure
	requiredFields := []string{
		"accountBalance", "maximumDebtAllowed", "amountInPlay", "amountBorrowed",
		"retainedEarnings", "equity", "tradingProfits", "workProfits", "totalProfits",
		"amountInPlayActive", "totalSpent", "totalSpentInPlay", "realizedProfits",
		"potentialProfits", "realizedValue", "potentialValue",
	}

	for _, field := range requiredFields {
		if _, exists := snapshot[field]; !exists {
			t.Errorf("Missing required field: %s", field)
		}
	}
}

func TestSumWorkProfitsFromTransactions(t *testing.T) {
	// Test the work profits function (should return 0 since no transaction system exists)
	db := modelstesting.NewFakeDB(t)
	user := modelstesting.GenerateUser("worker", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	workProfits, err := sumWorkProfitsFromTransactions(db, user.Username)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if workProfits != 0 {
		t.Errorf("Expected work profits to be 0 (no transaction system), got: %d", workProfits)
	}
}
