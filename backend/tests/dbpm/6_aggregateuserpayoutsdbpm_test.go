package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/models"
	"testing"
)

func TestAggregateUserPayoutsDBPM(t *testing.T) {
	tests := []struct {
		name           string
		bets           []models.Bet
		finalPayouts   []int64
		expectedResult []dbpm.MarketPosition
	}{
		{
			name: "single user multiple bets",
			bets: []models.Bet{
				{Username: "user1", Outcome: "YES", Amount: 100},
				{Username: "user1", Outcome: "NO", Amount: 50},
			},
			finalPayouts: []int64{100, 50},
			expectedResult: []dbpm.MarketPosition{
				{Username: "user1", YesSharesOwned: 100, NoSharesOwned: 50},
			},
		},
		{
			name: "multiple users",
			bets: []models.Bet{
				{Username: "user1", Outcome: "YES", Amount: 100},
				{Username: "user2", Outcome: "NO", Amount: 50},
			},
			finalPayouts: []int64{100, 50},
			expectedResult: []dbpm.MarketPosition{
				{Username: "user1", YesSharesOwned: 100, NoSharesOwned: 0},
				{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 50},
			},
		},
		{
			name: "negative payouts adjusted",
			bets: []models.Bet{
				{Username: "user1", Outcome: "YES", Amount: 100},
				{Username: "user1", Outcome: "NO", Amount: 50},
			},
			finalPayouts: []int64{100, -10},
			expectedResult: []dbpm.MarketPosition{
				{Username: "user1", YesSharesOwned: 100, NoSharesOwned: 0}, // Negative payout adjusted to 0
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := dbpm.AggregateUserPayoutsDBPM(tc.bets, tc.finalPayouts)
			if len(result) != len(tc.expectedResult) {
				t.Fatalf("Test %s failed: expected %d results, got %d", tc.name, len(tc.expectedResult), len(result))
			}
			for i, pos := range result {
				expected := tc.expectedResult[i]
				if pos.Username != expected.Username || pos.YesSharesOwned != expected.YesSharesOwned || pos.NoSharesOwned != expected.NoSharesOwned {
					t.Errorf("Test %s failed at index %d: expected %+v, got %+v", tc.name, i, expected, pos)
				}
			}
		})
	}
}
