package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/setup"
	test "socialpredict/tests"
	"testing"
)

func TestDivideUpMarketPoolSharesDBPM(t *testing.T) {
	for _, tc := range test.TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			yes, no := dbpm.DivideUpMarketPoolSharesDBPM(func() *setup.EconomicConfig { return setup.BuildInitialMarketAppConfig(t, .5, 1, 0, 0) }, tc.Bets, tc.ProbabilityChanges)
			if yes != tc.S_YES || no != tc.S_NO {
				t.Errorf("%s: expected (%d, %d), got (%d, %d)", tc.Name, tc.S_YES, tc.S_NO, yes, no)
			}
		})
	}
}
