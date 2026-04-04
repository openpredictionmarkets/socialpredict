package positionsmath

import (
	"socialpredict/models"
	"testing"
	"time"
)

var profitabilityTestTime = time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

func makeUserBets(entries []struct {
	Username string
	Amount   int64
	Offset   time.Duration
}) []models.Bet {
	bets := make([]models.Bet, 0, len(entries))
	for _, entry := range entries {
		bets = append(bets, models.Bet{
			Username: entry.Username,
			Amount:   entry.Amount,
			PlacedAt: profitabilityTestTime.Add(entry.Offset),
		})
	}
	return bets
}

func TestCalculateUserSpend(t *testing.T) {
	bets := makeUserBets([]struct {
		Username string
		Amount   int64
		Offset   time.Duration
	}{
		{Username: "alice", Amount: 100, Offset: 0},
		{Username: "alice", Amount: 50, Offset: time.Minute},
		{Username: "alice", Amount: -25, Offset: 2 * time.Minute},
		{Username: "bob", Amount: 200, Offset: 3 * time.Minute},
		{Username: "charlie", Amount: 75, Offset: 4 * time.Minute},
		{Username: "charlie", Amount: -75, Offset: 5 * time.Minute},
	})

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
				t.Fatalf("expected spend for %s to be %d, got %d", test.username, test.expectedSpend, spend)
			}
		})
	}
}

func TestGetEarliestBetTime(t *testing.T) {
	bets := makeUserBets([]struct {
		Username string
		Amount   int64
		Offset   time.Duration
	}{
		{Username: "alice", Offset: 2 * time.Hour},
		{Username: "alice", Offset: time.Hour},
		{Username: "alice", Offset: 3 * time.Hour},
		{Username: "bob", Offset: 30 * time.Minute},
		{Username: "bob", Offset: 4 * time.Hour},
	})

	tests := []struct {
		username     string
		expectedTime time.Time
	}{
		{"alice", profitabilityTestTime.Add(time.Hour)},
		{"bob", profitabilityTestTime.Add(30 * time.Minute)},
		{"charlie", time.Time{}},
	}

	for _, test := range tests {
		t.Run(test.username, func(t *testing.T) {
			earliestTime := GetEarliestBetTime(bets, test.username)
			if !earliestTime.Equal(test.expectedTime) {
				t.Fatalf("expected earliest time for %s to be %v, got %v",
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
				t.Fatalf("expected position type to be %s, got %s", test.expectedType, positionType)
			}
		})
	}
}

func TestCalculateMarketLeaderboard_EmptyBets(t *testing.T) {
	leaderboard, err := CalculateMarketLeaderboard(MarketSnapshot{ID: 1, CreatedAt: profitabilityTestTime}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(leaderboard) != 0 {
		t.Fatalf("expected empty leaderboard, got %d entries", len(leaderboard))
	}
}
