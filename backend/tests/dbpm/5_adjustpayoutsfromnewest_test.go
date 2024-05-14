package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"testing"
)

func TestAdjustPayoutsFromNewest(t *testing.T) {

	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			adjustedPayouts := dbpm.AdjustPayoutsFromNewest(tc.Bets, tc.ScaledPayouts)
			for i, payout := range adjustedPayouts {
				if payout != tc.AdjustedScaledPayouts[i] {
					t.Errorf("Test %s failed at index %d: expected payout %d, got %d", tc.Name, i, tc.AdjustedScaledPayouts[i], payout)
				}
			}
		})
	}
}
