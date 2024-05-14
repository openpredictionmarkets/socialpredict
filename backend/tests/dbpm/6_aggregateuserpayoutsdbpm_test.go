package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"testing"
)

func TestAggregateUserPayoutsDBPM(t *testing.T) {
	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			result := dbpm.AggregateUserPayoutsDBPM(tc.Bets, tc.AdjustedScaledPayouts)

			// Create a map for expected results for easy lookup
			expectedResults := make(map[string]dbpm.MarketPosition)
			for _, pos := range tc.AggregatedPositions {
				expectedResults[pos.Username] = pos
			}

			// Create a map from the results for comparison
			resultsMap := make(map[string]dbpm.MarketPosition)
			for _, pos := range result {
				resultsMap[pos.Username] = pos
			}

			// Check if the results match expected results for each user
			for username, expectedPos := range expectedResults {
				resultPos, ok := resultsMap[username]
				if !ok {
					t.Errorf("Test %s failed: missing position for username %s", tc.Name, username)
					continue
				}
				if resultPos.YesSharesOwned != expectedPos.YesSharesOwned || resultPos.NoSharesOwned != expectedPos.NoSharesOwned {
					t.Errorf("Test %s failed for %s: expected %+v, got %+v", tc.Name, username, expectedPos, resultPos)
				}
			}

			// Check for any unexpected extra users in the results
			for username := range resultsMap {
				if _, ok := expectedResults[username]; !ok {
					t.Errorf("Test %s failed: unexpected position for username %s", tc.Name, username)
				}
			}
		})
	}
}
