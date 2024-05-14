package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"testing"
)

func TestDivideUpMarketPoolSharesDBPM(t *testing.T) {
	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			yes, no := dbpm.DivideUpMarketPoolSharesDBPM(tc.Bets, tc.ProbabilityChanges)
			if yes != tc.ExpectedYES || no != tc.ExpectedNO {
				t.Errorf("%s: expected (%d, %d), got (%d, %d)", tc.Name, tc.ExpectedYES, tc.ExpectedNO, yes, no)
			}
		})
	}
}
