package betutils

import (
	"socialpredict/models"
	"socialpredict/setup"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func mockEconomicConfig() *setup.EconomicConfig {
	return &setup.EconomicConfig{
		Economics: struct {
			MarketCreation struct {
				InitialMarketProbability   float64 `yaml:"initialMarketProbability"`
				InitialMarketSubsidization int64   `yaml:"initialMarketSubsidization"`
				InitialMarketYes           int64   `yaml:"initialMarketYes"`
				InitialMarketNo            int64   `yaml:"initialMarketNo"`
			} `yaml:"marketcreation"`
			MarketIncentives struct {
				CreateMarketCost int64 `yaml:"createMarketCost"`
				TraderBonus      int64 `yaml:"traderBonus"`
			} `yaml:"marketincentives"`
			User struct {
				InitialAccountBalance int64 `yaml:"initialAccountBalance"`
				MaximumDebtAllowed    int64 `yaml:"maximumDebtAllowed"`
			} `yaml:"user"`
			Betting struct {
				MinimumBet int64 `yaml:"minimumBet"`
				BetFees    struct {
					InitialBetFee int64 `yaml:"initialBetFee"`
					BuySharesFee  int64 `yaml:"buySharesFee"`
					SellSharesFee int64 `yaml:"sellSharesFee"`
				} `yaml:"betFees"`
			} `yaml:"betting"`
		}{
			MarketCreation: struct {
				InitialMarketProbability   float64 `yaml:"initialMarketProbability"`
				InitialMarketSubsidization int64   `yaml:"initialMarketSubsidization"`
				InitialMarketYes           int64   `yaml:"initialMarketYes"`
				InitialMarketNo            int64   `yaml:"initialMarketNo"`
			}{
				InitialMarketProbability:   0.5,
				InitialMarketSubsidization: 10,
				InitialMarketYes:           0,
				InitialMarketNo:            0,
			},
			MarketIncentives: struct {
				CreateMarketCost int64 `yaml:"createMarketCost"`
				TraderBonus      int64 `yaml:"traderBonus"`
			}{
				CreateMarketCost: 10,
				TraderBonus:      1,
			},
			User: struct {
				InitialAccountBalance int64 `yaml:"initialAccountBalance"`
				MaximumDebtAllowed    int64 `yaml:"maximumDebtAllowed"`
			}{
				InitialAccountBalance: 1000,
				MaximumDebtAllowed:    500,
			},
			Betting: struct {
				MinimumBet int64 `yaml:"minimumBet"`
				BetFees    struct {
					InitialBetFee int64 `yaml:"initialBetFee"`
					BuySharesFee  int64 `yaml:"buySharesFee"`
					SellSharesFee int64 `yaml:"sellSharesFee"`
				} `yaml:"betFees"`
			}{
				MinimumBet: 1,
				BetFees: struct {
					InitialBetFee int64 `yaml:"initialBetFee"`
					BuySharesFee  int64 `yaml:"buySharesFee"`
					SellSharesFee int64 `yaml:"sellSharesFee"`
				}{
					InitialBetFee: 1,
					BuySharesFee:  0,
					SellSharesFee: 0,
				},
			},
		},
	}
}

func TestGetUserInitialBetFee(t *testing.T) {
	// Set up in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	// Migrate the Bet and User models
	if err := db.AutoMigrate(&models.Bet{}, &models.User{}); err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	// Mock the appConfig with test data
	appConfig = mockEconomicConfig()

	// Create a test user with a unique api_key
	user := &models.User{
		Username:       "testuser",
		AccountBalance: 1000,
		ApiKey:         "unique_api_key_1", // Ensure this is unique
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to save user to database: %v", err)
	}

	// Initialize the market ID
	marketID := uint(1)

	// Scenario 1: User places a bet on Market 1 where they have no prior bets
	initialBetFee := getUserInitialBetFee(db, marketID, user)
	wantFee := appConfig.Economics.Betting.BetFees.InitialBetFee
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
	appConfig = mockEconomicConfig()

	// Test buy scenario
	buyBet := models.Bet{Amount: 100}
	transactionFee := getTransactionFee(buyBet)
	if transactionFee != appConfig.Economics.Betting.BetFees.BuySharesFee {
		t.Errorf("Expected buy transaction fee to be %d, got %d", appConfig.Economics.Betting.BetFees.BuySharesFee, transactionFee)
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
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	// Migrate the Bet model
	db.AutoMigrate(&models.Bet{})

	// Mock the appConfig with test data
	appConfig = mockEconomicConfig()

	// Create a test user
	user := &models.User{Username: "testuser"}

	// Scenario 1: User has no bets, buys shares, gets initial fee
	buyBet := models.Bet{MarketID: 1, Amount: 100}
	sumOfBetFees := GetBetFees(db, user, buyBet)
	expectedSum := appConfig.Economics.Betting.BetFees.InitialBetFee
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
	expectedSum = appConfig.Economics.Betting.BetFees.BuySharesFee
	if sumOfBetFees != expectedSum {
		t.Errorf("Expected sum of bet fees to be %d, got %d", expectedSum, sumOfBetFees)
	}

	// Scenario 3: User has one bet, sells shares
	sellBet := models.Bet{MarketID: 1, Amount: -100}
	sumOfBetFees = GetBetFees(db, user, sellBet)
	expectedSum = appConfig.Economics.Betting.BetFees.SellSharesFee
	if sumOfBetFees != expectedSum {
		t.Errorf("Expected sum of bet fees to be %d, got %d", expectedSum, sumOfBetFees)
	}

}
