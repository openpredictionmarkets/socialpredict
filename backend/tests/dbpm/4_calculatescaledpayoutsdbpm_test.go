package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"testing"
)

func TestCalculateScaledPayoutsDBPM(t *testing.T) {
	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			actualPayouts := dbpm.CalculateScaledPayoutsDBPM(tc.Bets, tc.CoursePayouts, tc.F_YES, tc.F_NO)
			if len(actualPayouts) != len(tc.ScaledPayouts) {
				t.Fatalf("Test %s failed: expected %d payouts, got %d", tc.Name, len(tc.ScaledPayouts), len(actualPayouts))
			}
			for i, payout := range actualPayouts {
				if payout != tc.ScaledPayouts[i] {
					t.Errorf("Test %s failed at index %d: expected payout %d, got %d", tc.Name, i, tc.ScaledPayouts[i], payout)
				}
			}
		})
	}
}
