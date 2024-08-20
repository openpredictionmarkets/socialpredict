package test

import (
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/logging"
	test "socialpredict/tests"
	"testing"
)

func TestCalculateMarketProbabilitiesWPAM(t *testing.T) {
	for _, tc := range test.TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Call the function under test
			probChanges, totalYes, totalNo := wpam.CalculateMarketProbabilitiesWPAM(tc.Bets[0].PlacedAt, tc.Bets)
			logging.LogAnyType(totalYes, "totalYes")
			logging.LogAnyType(totalNo, "totalNo")

			if len(probChanges) != len(tc.ProbabilityChanges) {
				t.Fatalf("expected %d probability changes, got %d", len(tc.ProbabilityChanges), len(probChanges))
			}

			for i, pc := range probChanges {
				expected := tc.ProbabilityChanges[i]
				if pc.Probability != expected.Probability {
					t.Errorf("at index %d, expected probability %f, got %f", i, expected.Probability, pc.Probability)
				}
			}
		})
	}
}
