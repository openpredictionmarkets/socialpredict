package dbpm

import (
	"testing"
)

func TestDivideUpMarketPoolSharesDBPM(t *testing.T) {
	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			yes, no := DivideUpMarketPoolSharesDBPM(tc.Bets, tc.ProbabilityChanges)
			if yes != tc.S_YES || no != tc.S_NO {
				t.Errorf("%s: expected (%d, %d), got (%d, %d)", tc.Name, tc.S_YES, tc.S_NO, yes, no)
			}
		})
	}
}
