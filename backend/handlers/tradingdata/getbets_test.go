package tradingdata

import (
	"socialpredict/models"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetBetsForMarket(t *testing.T) {
	// Set up in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	// Auto-migrate the Bet model
	db.AutoMigrate(&models.Bet{})

	// Create some test data
	bets := []models.Bet{
		{Username: "user1", MarketID: 1, Amount: 100, PlacedAt: time.Now(), Outcome: "YES"},
		{Username: "user2", MarketID: 1, Amount: 200, PlacedAt: time.Now(), Outcome: "NO"},
		{Username: "user3", MarketID: 2, Amount: 150, PlacedAt: time.Now(), Outcome: "YES"},
	}
	db.Create(&bets)

	// Test the function
	retrievedBets := GetBetsForMarket(db, 1)

	// Verify the result
	if len(retrievedBets) != 2 {
		t.Errorf("Expected 2 bets, got %d", len(retrievedBets))
	}

	// Check if the returned bets match the expected ones
	for _, bet := range retrievedBets {
		if bet.MarketID != 1 {
			t.Errorf("Expected MarketID to be 1, got %d", bet.MarketID)
		}
	}
}
