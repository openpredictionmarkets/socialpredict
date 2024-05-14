package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"testing"
)

func TestCalculateCoursePayoutsDBPM(t *testing.T) {

	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			actualPayouts := dbpm.CalculateCoursePayoutsDBPM(tc.Bets, tc.ProbabilityChanges)
			if len(actualPayouts) != len(tc.ExpectedPayouts) {
				t.Fatalf("expected %d payouts, got %d", len(tc.ExpectedPayouts), len(actualPayouts))
			}
			for i, payout := range actualPayouts {
				if payout.Payout != tc.ExpectedPayouts[i].Payout || payout.Outcome != tc.ExpectedPayouts[i].Outcome {
					t.Errorf("Test %s failed at index %d: expected %+v, got %+v", tc.Name, i, tc.ExpectedPayouts[i], payout)
				}
			}
		})
	}
}
