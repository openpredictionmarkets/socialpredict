package positionsmath

import (
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"testing"
	"time"
)

func fixedTestBets() []models.Bet {
	testTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	return []models.Bet{
		{Username: "alice", Amount: 100, PlacedAt: testTime},
		{Username: "alice", Amount: 50, PlacedAt: testTime.Add(1 * time.Minute)},
		{Username: "alice", Amount: -25, PlacedAt: testTime.Add(2 * time.Minute)},
		{Username: "bob", Amount: 200, PlacedAt: testTime.Add(3 * time.Minute)},
		{Username: "charlie", Amount: 75, PlacedAt: testTime.Add(4 * time.Minute)},
		{Username: "charlie", Amount: -75, PlacedAt: testTime.Add(5 * time.Minute)},
	}
}

func TestCalculateUserSpend(t *testing.T) {
	bets := fixedTestBets()

	tests := []struct {
		username      string
		expectedSpend int64
	}{
		{"alice", 125},
		{"bob", 200},
		{"charlie", 0},
		{"dave", 0},
	}

	for _, test := range tests {
		t.Run(test.username, func(t *testing.T) {
			assertSpend(t, bets, test.username, test.expectedSpend)
		})
	}
}

func assertSpend(t *testing.T, bets []models.Bet, username string, expected int64) {
	t.Helper()

	spend := CalculateUserSpend(bets, username)
	if spend != expected {
		t.Errorf("expected spend for %s to be %d, got %d", username, expected, spend)
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
		{"alice", baseTime.Add(1 * time.Hour)},
		{"bob", baseTime.Add(30 * time.Minute)},
		{"charlie", time.Time{}},
	}

	for _, test := range tests {
		t.Run(test.username, func(t *testing.T) {
			assertEarliestBet(t, bets, test.username, test.expectedTime)
		})
	}
}

func assertEarliestBet(t *testing.T, bets []models.Bet, username string, expected time.Time) {
	t.Helper()

	earliestTime := GetEarliestBetTime(bets, username)
	if !earliestTime.Equal(expected) {
		t.Errorf("expected earliest time for %s to be %v, got %v", username, expected, earliestTime)
	}
}

func TestDeterminePositionType(t *testing.T) {
	tests := []struct {
		yesShares    int64
		noShares     int64
		expectedType string
	}{
		{100, 0, "YES"},
		{0, 150, "NO"},
		{50, 75, "NEUTRAL"},
		{0, 0, "NONE"},
	}

	for _, test := range tests {
		t.Run(test.expectedType, func(t *testing.T) {
			positionType := DeterminePositionType(test.yesShares, test.noShares)
			if positionType != test.expectedType {
				t.Errorf("expected position type to be %s, got %s", test.expectedType, positionType)
			}
		})
	}
}

func TestCalculateMarketLeaderboard_EmptyBets(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	market := modelstesting.GenerateMarket(1, "creator")
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	leaderboard, err := CalculateMarketLeaderboard(db, "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(leaderboard) != 0 {
		t.Fatalf("expected empty leaderboard, got %+v", leaderboard)
	}
}
