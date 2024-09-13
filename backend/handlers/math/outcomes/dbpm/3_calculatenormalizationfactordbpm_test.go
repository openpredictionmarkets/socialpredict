package dbpm

import (
	"math"
	"testing"
)

func TestCalculateNormalizationFactorsDBPM(t *testing.T) {

	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			roundedF_YES := math.Round(tc.F_YES*1000000) / 1000000
			roundedF_NO := math.Round(tc.F_NO*1000000) / 1000000

			if roundedF_YES != tc.ExpectedF_YES || roundedF_NO != tc.ExpectedF_NO {
				t.Errorf("Test %s failed: expected F_YES=%f, F_NO=%f, got F_YES=%f, F_NO=%f",
					tc.Name, tc.ExpectedF_YES, tc.ExpectedF_NO, roundedF_YES, roundedF_NO)
			}
		})
	}
}
