package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/models"
	"testing"
)

func TestCalculateScaledPayoutsDBPM(t *testing.T) {
	tests := []struct {
		name            string
		allBetsOnMarket []models.Bet
		coursePayouts   []dbpm.CourseBetPayout
		F_YES           float64
		F_NO            float64
		expectedPayouts []int64
	}{
		{
			name: "standard payouts",
			allBetsOnMarket: []models.Bet{
				{Amount: 100},
				{Amount: 50},
			},
			coursePayouts: []dbpm.CourseBetPayout{
				{Payout: 50, Outcome: "YES"},
				{Payout: 25, Outcome: "NO"},
			},
			F_YES: 2.0,
			F_NO:  1.5,
			expectedPayouts: []int64{
				100, // 50 * 2.0
				38,  // 25 * 1.5, rounded
			},
		},
		{
			name: "zero normalization factors",
			allBetsOnMarket: []models.Bet{
				{Amount: 100},
				{Amount: 50},
			},
			coursePayouts: []dbpm.CourseBetPayout{
				{Payout: 100, Outcome: "YES"},
				{Payout: 50, Outcome: "NO"},
			},
			F_YES: 0.0,
			F_NO:  0.0,
			expectedPayouts: []int64{
				0, // 100 * 0
				0, // 50 * 0
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualPayouts := dbpm.CalculateScaledPayoutsDBPM(tc.allBetsOnMarket, tc.coursePayouts, tc.F_YES, tc.F_NO)
			if len(actualPayouts) != len(tc.expectedPayouts) {
				t.Fatalf("Test %s failed: expected %d payouts, got %d", tc.name, len(tc.expectedPayouts), len(actualPayouts))
			}
			for i, payout := range actualPayouts {
				if payout != tc.expectedPayouts[i] {
					t.Errorf("Test %s failed at index %d: expected payout %d, got %d", tc.name, i, tc.expectedPayouts[i], payout)
				}
			}
		})
	}
}
