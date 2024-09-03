package betutils

import (
	"socialpredict/models"
	"socialpredict/setup"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Mock the setup.EconomicConfig structure
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

	// Migrate the Bet model
	db.AutoMigrate(&models.Bet{})

	// Mock the appConfig with test data
	appConfig = mockEconomicConfig()

	// Create a test user
	user := &models.User{Username: "testuser"}

	// Scenario 1: User has no bets on the market
	initialBetFee := getUserInitialBetFee(db, 1, user)
	if initialBetFee != 0 {
		t.Errorf("Expected initial bet fee to be 0, got %d", initialBetFee)
	}

	// Create a test bet
	bets := []models.Bet{
		{Username: "testuser", MarketID: 1, Amount: 100, PlacedAt: time.Now()},
	}
	db.Create(&bets)

	// Scenario 2: User has one bet on the market
	initialBetFee = getUserInitialBetFee(db, 1, user)
	if initialBetFee != appConfig.Economics.Betting.BetFees.InitialBetFee {
		t.Errorf("Expected initial bet fee to be %d, got %d", appConfig.Economics.Betting.BetFees.InitialBetFee, initialBetFee)
	}

	// Create another bet for the same user
	db.Create(&models.Bet{Username: "testuser", MarketID: 1, Amount: 200, PlacedAt: time.Now()})

	// Scenario 3: User has multiple bets on the market
	initialBetFee = getUserInitialBetFee(db, 1, user)
	if initialBetFee != 0 {
		t.Errorf("Expected initial bet fee to be 0, got %d", initialBetFee)
	}
}
