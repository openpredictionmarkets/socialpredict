package dbpm

import (
	"testing"
)

func TestCalculateCoursePayoutsDBPM(t *testing.T) {

	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			actualPayouts := CalculateCoursePayoutsDBPM(tc.Bets, tc.ProbabilityChanges)
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
