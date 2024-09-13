package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	test "socialpredict/tests"
	"testing"
)

func TestCalculateCoursePayoutsDBPM(t *testing.T) {

	for _, tc := range test.TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			actualPayouts := dbpm.CalculateCoursePayoutsDBPM(tc.Bets, tc.ProbabilityChanges)
			if len(actualPayouts) != len(tc.CoursePayouts) {
				t.Fatalf("expected %d payouts, got %d", len(tc.CoursePayouts), len(actualPayouts))
			}
			for i, payout := range actualPayouts {
				if payout.Payout != tc.CoursePayouts[i].Payout || payout.Outcome != tc.CoursePayouts[i].Outcome {
					t.Errorf("Test %s failed at index %d: expected %+v, got %+v", tc.Name, i, tc.CoursePayouts[i], payout)
				}
			}
		})
	}
}
