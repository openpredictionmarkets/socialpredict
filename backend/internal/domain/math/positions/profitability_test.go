package positionsmath

import (
	"socialpredict/models"
	"testing"
	"time"
)

func TestCalculateUserSpend(t *testing.T) {
	// Create test bets data
	testTime := time.Now()
	bets := []models.Bet{
		{Username: "alice", Amount: 100, PlacedAt: testTime},   // Alice buys 100
		{Username: "alice", Amount: 50, PlacedAt: testTime},    // Alice buys 50 more
		{Username: "alice", Amount: -25, PlacedAt: testTime},   // Alice sells 25
		{Username: "bob", Amount: 200, PlacedAt: testTime},     // Bob buys 200
		{Username: "charlie", Amount: 75, PlacedAt: testTime},  // Charlie buys 75
		{Username: "charlie", Amount: -75, PlacedAt: testTime}, // Charlie sells all 75
	}

	tests := []struct {
		username      string
		expectedSpend int64
	}{
		{"alice", 125}, // 100 + 50 - 25 = 125
		{"bob", 200},   // 200
		{"charlie", 0}, // 75 - 75 = 0
		{"dave", 0},    // No bets = 0
	}

	for _, test := range tests {
		t.Run(test.username, func(t *testing.T) {
			spend := CalculateUserSpend(bets, test.username)
			if spend != test.expectedSpend {
				t.Errorf("Expected spend for %s to be %d, got %d", test.username, test.expectedSpend, spend)
			}
		})
	}
}

func TestGetEarliestBetTime(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	bets := []models.Bet{
		{Username: "alice", PlacedAt: baseTime.Add(2 * time.Hour)},  // 12:00
		{Username: "alice", PlacedAt: baseTime.Add(1 * time.Hour)},  // 11:00 (earliest for alice)
		{Username: "alice", PlacedAt: baseTime.Add(3 * time.Hour)},  // 13:00
		{Username: "bob", PlacedAt: baseTime.Add(30 * time.Minute)}, // 10:30 (earliest for bob)
		{Username: "bob", PlacedAt: baseTime.Add(4 * time.Hour)},    // 14:00
	}

	tests := []struct {
		username     string
		expectedTime time.Time
	}{
		{"alice", baseTime.Add(1 * time.Hour)},  // 11:00
		{"bob", baseTime.Add(30 * time.Minute)}, // 10:30
		{"charlie", time.Time{}},                // No bets, zero time
	}

	for _, test := range tests {
		t.Run(test.username, func(t *testing.T) {
			earliestTime := GetEarliestBetTime(bets, test.username)
			if !earliestTime.Equal(test.expectedTime) {
				t.Errorf("Expected earliest time for %s to be %v, got %v",
					test.username, test.expectedTime, earliestTime)
			}
		})
	}
}

func TestDeterminePositionType(t *testing.T) {
	tests := []struct {
		yesShares    int64
		noShares     int64
		expectedType string
	}{
		{100, 0, "YES"},     // Only YES shares
		{0, 150, "NO"},      // Only NO shares
		{50, 75, "NEUTRAL"}, // Both YES and NO shares
		{0, 0, "NONE"},      // No shares (shouldn't happen in practice)
	}

	for _, test := range tests {
		t.Run(test.expectedType, func(t *testing.T) {
			positionType := DeterminePositionType(test.yesShares, test.noShares)
			if positionType != test.expectedType {
				t.Errorf("Expected position type to be %s, got %s", test.expectedType, positionType)
			}
		})
	}
}

// Integration test would require database setup, so we'll keep it simple for now
// In a real implementation, you'd want to test CalculateMarketLeaderboard with test data
func TestCalculateMarketLeaderboard_EmptyBets(t *testing.T) {
	// This test would require more setup with database mocking
	// For now, we can test the core logic components above
	// In practice, you'd mock the database and test the full function
	t.Skip("Integration test requires database setup - core logic tested above")
}
