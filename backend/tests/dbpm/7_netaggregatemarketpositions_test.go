package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"testing"
)

func TestNetAggregateMarketPositions(t *testing.T) {
	tests := []struct {
		name              string
		inputPositions    []dbpm.MarketPosition
		expectedPositions []dbpm.MarketPosition
	}{
		{
			name: "simple net positive",
			inputPositions: []dbpm.MarketPosition{
				{Username: "user1", YesSharesOwned: 100, NoSharesOwned: 50},
			},
			expectedPositions: []dbpm.MarketPosition{
				{Username: "user1", YesSharesOwned: 50, NoSharesOwned: 0},
			},
		},
		{
			name: "simple net negative",
			inputPositions: []dbpm.MarketPosition{
				{Username: "user2", YesSharesOwned: 30, NoSharesOwned: 80},
			},
			expectedPositions: []dbpm.MarketPosition{
				{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 50},
			},
		},
		{
			name: "equal positions zero out",
			inputPositions: []dbpm.MarketPosition{
				{Username: "user3", YesSharesOwned: 50, NoSharesOwned: 50},
			},
			expectedPositions: []dbpm.MarketPosition{
				{Username: "user3", YesSharesOwned: 0, NoSharesOwned: 0},
			},
		},
		{
			name: "multiple users",
			inputPositions: []dbpm.MarketPosition{
				{Username: "user1", YesSharesOwned: 100, NoSharesOwned: 30},
				{Username: "user2", YesSharesOwned: 20, NoSharesOwned: 100},
			},
			expectedPositions: []dbpm.MarketPosition{
				{Username: "user1", YesSharesOwned: 70, NoSharesOwned: 0},
				{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 80},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := dbpm.NetAggregateMarketPositions(tc.inputPositions)
			if len(result) != len(tc.expectedPositions) {
				t.Fatalf("Test %s failed: expected %d results, got %d", tc.name, len(tc.expectedPositions), len(result))
			}
			for i, pos := range result {
				if pos.Username != tc.expectedPositions[i].Username || pos.YesSharesOwned != tc.expectedPositions[i].YesSharesOwned || pos.NoSharesOwned != tc.expectedPositions[i].NoSharesOwned {
					t.Errorf("Test %s failed at index %d: expected %+v, got %+v", tc.name, i, tc.expectedPositions[i], pos)
				}
			}
		})
	}
}
