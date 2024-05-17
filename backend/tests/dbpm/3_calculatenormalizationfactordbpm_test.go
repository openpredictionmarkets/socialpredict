package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	test "socialpredict/tests"
	"testing"
)

func TestCalculateNormalizationFactorsDBPM(t *testing.T) {

	for _, tc := range test.TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			F_YES, F_NO := dbpm.CalculateNormalizationFactorsDBPM(tc.S_YES, tc.S_NO, tc.CoursePayouts)
			if F_YES != tc.F_YES || F_NO != tc.F_NO {
				t.Errorf("Test %s failed: expected F_YES=%f, F_NO=%f, got F_YES=%f, F_NO=%f", tc.Name, tc.F_YES, tc.F_NO, F_YES, F_NO)
			}
		})
	}
}
