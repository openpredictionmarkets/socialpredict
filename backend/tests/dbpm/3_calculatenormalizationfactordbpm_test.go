package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"testing"
)

func TestCalculateNormalizationFactorsDBPM(t *testing.T) {
	tests := []struct {
		name          string
		S_YES         int64
		S_NO          int64
		coursePayouts []dbpm.CourseBetPayout
		expectedF_YES float64
		expectedF_NO  float64
	}{
		{
			name:  "typical case",
			S_YES: 100,
			S_NO:  200,
			coursePayouts: []dbpm.CourseBetPayout{
				{Payout: 50, Outcome: "YES"},
				{Payout: 100, Outcome: "NO"},
			},
			expectedF_YES: 2.0,
			expectedF_NO:  2.0,
		},
		{
			name:  "division by zero in YES",
			S_YES: 100,
			S_NO:  200,
			coursePayouts: []dbpm.CourseBetPayout{
				{Payout: 0, Outcome: "YES"},
				{Payout: 100, Outcome: "NO"},
			},
			expectedF_YES: 1.0, // Default to 1 to avoid division by zero
			expectedF_NO:  2.0,
		},
		{
			name:  "division by zero in NO",
			S_YES: 100,
			S_NO:  200,
			coursePayouts: []dbpm.CourseBetPayout{
				{Payout: 50, Outcome: "YES"},
				{Payout: 0, Outcome: "NO"},
			},
			expectedF_YES: 2.0,
			expectedF_NO:  1.0, // Default to 1 to avoid division by zero
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			F_YES, F_NO := dbpm.CalculateNormalizationFactorsDBPM(tc.S_YES, tc.S_NO, tc.coursePayouts)
			if F_YES != tc.expectedF_YES || F_NO != tc.expectedF_NO {
				t.Errorf("Test %s failed: expected F_YES=%f, F_NO=%f, got F_YES=%f, F_NO=%f", tc.name, tc.expectedF_YES, tc.expectedF_NO, F_YES, F_NO)
			}
		})
	}
}
