package betutils

import (
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup/setuptesting"
	"testing"
	"time"
)

func TestGetUserInitialBetFee(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := db.AutoMigrate(&models.Bet{}, &models.User{}); err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	appConfig = setuptesting.MockEconomicConfig()
	user := &models.User{Username: "testuser", AccountBalance: 1000, ApiKey: "unique_api_key_1"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to save user to database: %v", err)
	}

	marketID := uint(1)

	// getUserInitialBetFee function to include both initial and buy share fees
	// For testing purpose, assuming getUserInitialBetFee function does this calculation correctly
	initialBetFee := getUserInitialBetFee(db, marketID, user) + appConfig.Economics.Betting.BetFees.EachBetFee
	wantFee := appConfig.Economics.Betting.BetFees.InitialBetFee + appConfig.Economics.Betting.BetFees.EachBetFee
	if initialBetFee != wantFee {
		t.Errorf("getUserInitialBetFee(db, %d, %s) = %d, want %d", marketID, user.Username, initialBetFee, wantFee)
	}

	// Place a bet for the user on Market 1
	bet := models.Bet{Username: "testuser", MarketID: marketID, Amount: 100, PlacedAt: time.Now()}
	if err := db.Create(&bet).Error; err != nil {
		t.Fatalf("Failed to save bet to database: %v", err)
	}

	// Scenario 2: User places another bet on Market 1 where they already have a bet
	initialBetFee = getUserInitialBetFee(db, marketID, user)
	wantFee = 0
	if initialBetFee != wantFee {
		t.Errorf("getUserInitialBetFee(db, %d, %s) = %d, want %d after placing a bet", marketID, user.Username, initialBetFee, wantFee)
	}

	// Update the market ID for a new scenario
	marketID = 2

	// Scenario 3: User places a bet on Market 2 where they have no prior bets
	initialBetFee = getUserInitialBetFee(db, marketID, user)
	if initialBetFee != appConfig.Economics.Betting.BetFees.InitialBetFee {
		t.Errorf("getUserInitialBetFee(db, %d, %s) = %d, want %d", marketID, user.Username, initialBetFee, appConfig.Economics.Betting.BetFees.InitialBetFee)
	}
}

func TestGetTransactionFee(t *testing.T) {
	// Mock the appConfig with test data
	appConfig = setuptesting.MockEconomicConfig()

	// Test buy scenario
	buyBet := models.Bet{Amount: 100}
	transactionFee := getTransactionFee(buyBet)
	if transactionFee != appConfig.Economics.Betting.BetFees.EachBetFee {
		t.Errorf("Expected buy transaction fee to be %d, got %d", appConfig.Economics.Betting.BetFees.EachBetFee, transactionFee)
	}

	// Test sell scenario
	sellBet := models.Bet{Amount: -100}
	transactionFee = getTransactionFee(sellBet)
	if transactionFee != appConfig.Economics.Betting.BetFees.SellSharesFee {
		t.Errorf("Expected sell transaction fee to be %d, got %d", appConfig.Economics.Betting.BetFees.SellSharesFee, transactionFee)
	}
}

func TestGetSumBetFees(t *testing.T) {
	// Set up in-memory SQLite database
	db := modelstesting.NewFakeDB(t)

	// Migrate the Bet model
	if err := db.AutoMigrate(&models.Bet{}); err != nil {
		t.Fatalf("Failed to auto migrate bets model %v", err)
	}

	// Mock the appConfig with test data
	appConfig = setuptesting.MockEconomicConfig()

	// Create a test user
	user := &models.User{Username: "testuser"}

	// Scenario 1: User has no bets, buys shares, gets initial fee
	buyBet := models.Bet{MarketID: 1, Amount: 100}
	sumOfBetFees := GetBetFees(db, user, buyBet)
	expectedSum := appConfig.Economics.Betting.BetFees.InitialBetFee +
		appConfig.Economics.Betting.BetFees.EachBetFee
	if sumOfBetFees != expectedSum {
		t.Errorf("Expected sum of bet fees to be %d, got %d", expectedSum, sumOfBetFees)
	}

	// Create a test bet
	bets := []models.Bet{
		{Username: "testuser", MarketID: 1, Amount: 100, PlacedAt: time.Now()},
	}
	db.Create(&bets)

	// Scenario 2: User has one bet, buys shares
	sumOfBetFees = GetBetFees(db, user, buyBet)
	expectedSum = appConfig.Economics.Betting.BetFees.EachBetFee
	if sumOfBetFees != expectedSum {
		t.Errorf("Expected sum of bet fees to be %d, got %d", expectedSum, sumOfBetFees)
	}

	// Scenario 3: User has one bet, sells shares
	sellBet := models.Bet{MarketID: 1, Amount: -1}
	sumOfBetFees = GetBetFees(db, user, sellBet)
	expectedSum = appConfig.Economics.Betting.BetFees.SellSharesFee
	if sumOfBetFees != expectedSum {
		t.Errorf("Expected sum of bet fees to be %d, got %d", expectedSum, sumOfBetFees)
	}

}
