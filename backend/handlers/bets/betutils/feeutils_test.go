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
	db.AutoMigrate(&models.Bet{}, &models.User{})

	// Mock the appConfig with test data
	appConfig = mockEconomicConfig()

	// Create a test user with a unique api_key
	user := &models.User{
		Username:       "testuser",
		AccountBalance: 1000,
		ApiKey:         "unique_api_key_1", // Ensure this is unique
	}
	db.Create(&user) // Save the user to the database

	// Scenario 1: User places a bet on Market 1 where they have no prior bets
	initialBetFee := getUserInitialBetFee(db, 1, user)
	if initialBetFee != appConfig.Economics.Betting.BetFees.InitialBetFee {
		t.Errorf("Expected initial bet fee to be %d, got %d", appConfig.Economics.Betting.BetFees.InitialBetFee, initialBetFee)
	}

	// Place a bet for the user on Market 1
	bets := []models.Bet{
		{Username: "testuser", MarketID: 1, Amount: 100, PlacedAt: time.Now()},
	}
	db.Create(&bets) // Save the bet to the database

	// Scenario 2: User places another bet on Market 1 where they already have a bet
	initialBetFee = getUserInitialBetFee(db, 1, user)
	if initialBetFee != 0 {
		t.Errorf("Expected initial bet fee to be 0, got %d", initialBetFee)
	}

	// Scenario 3: User places a bet on Market 2 where they have no prior bets
	initialBetFee = getUserInitialBetFee(db, 2, user)
	if initialBetFee != appConfig.Economics.Betting.BetFees.InitialBetFee {
		t.Errorf("Expected initial bet fee to be %d, got %d", appConfig.Economics.Betting.BetFees.InitialBetFee, initialBetFee)
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
	sumOfBetFees := GetSumBetFees(db, user, buyBet)
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
	sumOfBetFees = GetSumBetFees(db, user, buyBet)
	expectedSum = appConfig.Economics.Betting.BetFees.BuySharesFee
	if sumOfBetFees != expectedSum {
		t.Errorf("Expected sum of bet fees to be %d, got %d", expectedSum, sumOfBetFees)
	}

	// Scenario 3: User has one bet, sells shares
	sellBet := models.Bet{MarketID: 1, Amount: -100}
	sumOfBetFees = GetSumBetFees(db, user, sellBet)
	expectedSum = appConfig.Economics.Betting.BetFees.SellSharesFee
	if sumOfBetFees != expectedSum {
		t.Errorf("Expected sum of bet fees to be %d, got %d", expectedSum, sumOfBetFees)
	}

}
