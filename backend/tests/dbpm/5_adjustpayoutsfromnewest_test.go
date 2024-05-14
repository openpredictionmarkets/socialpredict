package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/models"
	"testing"
)

func TestAdjustPayoutsFromNewest(t *testing.T) {
	tests := []struct {
		name            string
		bets            []models.Bet
		scaledPayouts   []int64
		expectedPayouts []int64
	}{
		{
			name:            "normal adjustment",
			bets:            []models.Bet{{Amount: 100}, {Amount: 50}, {Amount: 150}},
			scaledPayouts:   []int64{100, 80, 120},
			expectedPayouts: []int64{100, 80, 70}, // 300 total, 100 + 50 + 150 = 300 available, adjust from newest
		},
		{
			name:            "no adjustment needed",
			bets:            []models.Bet{{Amount: 100}, {Amount: 50}, {Amount: 150}},
			scaledPayouts:   []int64{50, 80, 120},
			expectedPayouts: []int64{50, 80, 120}, // total 250, 300 available
		},
		{
			name:            "extreme adjustment",
			bets:            []models.Bet{{Amount: 50}, {Amount: 20}, {Amount: 10}},
			scaledPayouts:   []int64{100, 100, 100},
			expectedPayouts: []int64{30, 20, 10}, // 300 total, 50 + 20 + 10 = 80 available, heavy adjustment
		},
		{
			name:            "all payouts zero",
			bets:            []models.Bet{{Amount: 100}, {Amount: 50}, {Amount: 150}},
			scaledPayouts:   []int64{0, 0, 0},
			expectedPayouts: []int64{0, 0, 0}, // No need for adjustment
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adjustedPayouts := dbpm.AdjustPayoutsFromNewest(tc.bets, tc.scaledPayouts)
			for i, payout := range adjustedPayouts {
				if payout != tc.expectedPayouts[i] {
					t.Errorf("Test %s failed at index %d: expected payout %d, got %d", tc.name, i, tc.expectedPayouts[i], payout)
				}
			}
		})
	}
}
