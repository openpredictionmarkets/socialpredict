package betshandlers

import (
	"testing"
	"time"

	buybetshandlers "socialpredict/handlers/bets/buying"
	sellbetshandlers "socialpredict/handlers/bets/selling"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
	"socialpredict/util"

	"gorm.io/gorm"
)

// TestMarketStatusValidation tests that betting operations properly validate market status
func TestMarketStatusValidation(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	// Create test user
	testUser := modelstesting.GenerateUser("testuser", 1000)
	db.Create(&testUser)

	// Create economic configuration loader for tests
	loadEconConfig := func() *setup.EconomicConfig {
		return modelstesting.GenerateEconomicConfig()
	}

	t.Run("BuyingOperations", func(t *testing.T) {
		testBuyingOperations(t, db, &testUser, loadEconConfig)
	})

	t.Run("SellingOperations", func(t *testing.T) {
		testSellingOperations(t, db, &testUser, loadEconConfig)
	})
}

func testBuyingOperations(t *testing.T, db *gorm.DB, testUser *models.User, loadEconConfig setup.EconConfigLoader) {
	// Test buying on active market (should succeed)
	t.Run("BuyOnActiveMarket", func(t *testing.T) {
		activeMarket := createTestMarket(db, "Active Market", time.Now().Add(24*time.Hour), false, "")

		betRequest := models.Bet{
			MarketID: uint(activeMarket.ID),
			Amount:   10,
			Outcome:  "YES",
		}

		_, err := buybetshandlers.PlaceBetCore(testUser, betRequest, db, loadEconConfig)
		if err != nil {
			t.Errorf("Expected buying on active market to succeed, got error: %v", err)
		}
	})

	// Test buying on closed market (should fail)
	t.Run("BuyOnClosedMarket", func(t *testing.T) {
		closedMarket := createTestMarket(db, "Closed Market", time.Now().Add(-1*time.Hour), false, "")

		betRequest := models.Bet{
			MarketID: uint(closedMarket.ID),
			Amount:   10,
			Outcome:  "YES",
		}

		_, err := buybetshandlers.PlaceBetCore(testUser, betRequest, db, loadEconConfig)
		if err == nil {
			t.Error("Expected buying on closed market to fail, but it succeeded")
		}
		if err.Error() != "cannot place a bet on a closed market" {
			t.Errorf("Expected 'cannot place a bet on a closed market' error, got: %v", err)
		}
	})

	// Test buying on resolved market (should fail)
	t.Run("BuyOnResolvedMarket", func(t *testing.T) {
		resolvedMarket := createTestMarket(db, "Resolved Market", time.Now().Add(-1*time.Hour), true, "YES")

		betRequest := models.Bet{
			MarketID: uint(resolvedMarket.ID),
			Amount:   10,
			Outcome:  "YES",
		}

		_, err := buybetshandlers.PlaceBetCore(testUser, betRequest, db, loadEconConfig)
		if err == nil {
			t.Error("Expected buying on resolved market to fail, but it succeeded")
		}
		if err.Error() != "cannot place a bet on a resolved market" {
			t.Errorf("Expected 'cannot place a bet on a resolved market' error, got: %v", err)
		}
	})

	// Edge case: Test buying on market that closes exactly now
	t.Run("BuyOnMarketClosingNow", func(t *testing.T) {
		// Market closing within 1 second of now (should be closed by the time we check)
		almostClosedMarket := createTestMarket(db, "Almost Closed Market", time.Now().Add(1*time.Millisecond), false, "")

		// Wait a small amount to ensure market is closed
		time.Sleep(10 * time.Millisecond)

		betRequest := models.Bet{
			MarketID: uint(almostClosedMarket.ID),
			Amount:   10,
			Outcome:  "YES",
		}

		_, err := buybetshandlers.PlaceBetCore(testUser, betRequest, db, loadEconConfig)
		if err == nil {
			t.Error("Expected buying on market closing now to fail, but it succeeded")
		}
		if err.Error() != "cannot place a bet on a closed market" {
			t.Errorf("Expected 'cannot place a bet on a closed market' error, got: %v", err)
		}
	})
}

func testSellingOperations(t *testing.T, db *gorm.DB, testUser *models.User, loadEconConfig setup.EconConfigLoader) {
	// Note: For selling tests, we need to first create positions for the user
	// This is more complex, so we'll test the market status validation at the core level

	// Test selling on closed market (should fail)
	t.Run("SellOnClosedMarket", func(t *testing.T) {
		closedMarket := createTestMarket(db, "Closed Market for Selling", time.Now().Add(-1*time.Hour), false, "")

		// Create a mock economic config for selling tests
		cfg := loadEconConfig()

		sellRequest := models.Bet{
			MarketID: uint(closedMarket.ID),
			Amount:   10,
			Outcome:  "YES",
		}

		err := sellbetshandlers.ProcessSellRequest(db, &sellRequest, testUser, cfg)
		if err == nil {
			t.Error("Expected selling on closed market to fail, but it succeeded")
		}
		if err.Error() != "cannot place a bet on a closed market" {
			t.Errorf("Expected 'cannot place a bet on a closed market' error, got: %v", err)
		}
	})

	// Test selling on resolved market (should fail)
	t.Run("SellOnResolvedMarket", func(t *testing.T) {
		resolvedMarket := createTestMarket(db, "Resolved Market for Selling", time.Now().Add(-1*time.Hour), true, "YES")

		cfg := loadEconConfig()

		sellRequest := models.Bet{
			MarketID: uint(resolvedMarket.ID),
			Amount:   10,
			Outcome:  "YES",
		}

		err := sellbetshandlers.ProcessSellRequest(db, &sellRequest, testUser, cfg)
		if err == nil {
			t.Error("Expected selling on resolved market to fail, but it succeeded")
		}
		if err.Error() != "cannot place a bet on a resolved market" {
			t.Errorf("Expected 'cannot place a bet on a resolved market' error, got: %v", err)
		}
	})
}

// TestMarketNotFound tests behavior when trying to bet on non-existent markets
func TestMarketNotFound(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db // Set global DB for util.GetDB()

	testUser := modelstesting.GenerateUser("testuser", 1000)
	db.Create(&testUser)

	loadEconConfig := func() *setup.EconomicConfig {
		return modelstesting.GenerateEconomicConfig()
	}

	t.Run("BuyOnNonExistentMarket", func(t *testing.T) {
		betRequest := models.Bet{
			MarketID: 99999, // Non-existent market ID
			Amount:   10,
			Outcome:  "YES",
		}

		_, err := buybetshandlers.PlaceBetCore(&testUser, betRequest, db, loadEconConfig)
		if err == nil {
			t.Error("Expected buying on non-existent market to fail, but it succeeded")
		}
		if err.Error() != "market not found" {
			t.Errorf("Expected 'market not found' error, got: %v", err)
		}
	})

	t.Run("SellOnNonExistentMarket", func(t *testing.T) {
		cfg := loadEconConfig()

		sellRequest := models.Bet{
			MarketID: 99999, // Non-existent market ID
			Amount:   10,
			Outcome:  "YES",
		}

		err := sellbetshandlers.ProcessSellRequest(db, &sellRequest, &testUser, cfg)
		if err == nil {
			t.Error("Expected selling on non-existent market to fail, but it succeeded")
		}
		if err.Error() != "market not found" {
			t.Errorf("Expected 'market not found' error, got: %v", err)
		}
	})
}

// Helper function to create test markets with different statuses
func createTestMarket(db *gorm.DB, title string, resolutionDateTime time.Time, isResolved bool, resolutionResult string) *models.Market {
	market := &models.Market{
		QuestionTitle:      title,
		Description:        "Test market for validation testing",
		OutcomeType:        "BINARY",
		ResolutionDateTime: resolutionDateTime,
		IsResolved:         isResolved,
		ResolutionResult:   resolutionResult,
		InitialProbability: 0.5,
		CreatorUsername:    "testuser",
	}

	if err := db.Create(market).Error; err != nil {
		panic("Failed to create test market: " + err.Error())
	}

	return market
}
