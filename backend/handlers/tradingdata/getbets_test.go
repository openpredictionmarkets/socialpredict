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
	if err := db.AutoMigrate(&models.Bet{}); err != nil {
		t.Fatalf("Failed to auto-migrate Bet model: %v", err)
	}

	// Create some test data
	bets := []models.Bet{
		{Username: "user1", MarketID: 1, Amount: 100, PlacedAt: time.Now(), Outcome: "YES"},
		{Username: "user2", MarketID: 1, Amount: 200, PlacedAt: time.Now(), Outcome: "NO"},
		{Username: "user3", MarketID: 2, Amount: 150, PlacedAt: time.Now(), Outcome: "YES"},
	}
	if err := db.Create(&bets).Error; err != nil {
		t.Fatalf("Failed to create bets: %v", err)
	}

	// Test the function
	retrievedBets := GetBetsForMarket(db, 1)

	// Verify the number of bets retrieved
	if got, want := len(retrievedBets), 2; got != want {
		t.Errorf("GetBetsForMarket(db, 1) = %d bets, want %d bets", got, want)
	}

	// Check if the returned bets match the expected ones
	for _, bet := range retrievedBets {
		if got, want := int(bet.MarketID), 1; got != want {
			t.Errorf("GetBetsForMarket(db, 1) - retrieved bet with MarketID = %d, want %d", got, want)
		}
	}
}
