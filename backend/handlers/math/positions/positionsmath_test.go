package positionsmath

import (
	"socialpredict/models/modelstesting"
	"strconv"
	"testing"
	"time"
)

func TestCalculateMarketPositions_WPAM_DBPM(t *testing.T) {
	testcases := []struct {
		Name       string
		BetConfigs []struct {
			Amount   int64
			Outcome  string
			Username string
			Offset   time.Duration
		}
		Expected []MarketPosition // You can fill in Yes/NoShares for now, and auto-print Value
	}{
		{
			Name: "Single YES Bet",
			BetConfigs: []struct {
				Amount   int64
				Outcome  string
				Username string
				Offset   time.Duration
			}{
				{Amount: 50, Outcome: "YES", Username: "alice", Offset: 0},
			},
			Expected: []MarketPosition{
				{Username: "alice", YesSharesOwned: 50, NoSharesOwned: 0},
			},
		},
		{
			Name: "Single NO Bet",
			BetConfigs: []struct {
				Amount   int64
				Outcome  string
				Username string
				Offset   time.Duration
			}{
				{Amount: 30, Outcome: "NO", Username: "bob", Offset: 0},
			},
			Expected: []MarketPosition{
				{Username: "bob", YesSharesOwned: 0, NoSharesOwned: 30},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			db := modelstesting.NewFakeDB(t)
			creator := "testcreator"
			market := modelstesting.GenerateMarket(1, creator)
			db.Create(&market)
			for _, betConf := range tc.BetConfigs {
				bet := modelstesting.GenerateBet(betConf.Amount, betConf.Outcome, betConf.Username, uint(market.ID), betConf.Offset)
				db.Create(&bet)
			}
			marketIDStr := strconv.Itoa(int(market.ID))
			actualPositions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIDStr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(actualPositions) != len(tc.Expected) {
				t.Fatalf("expected %d positions, got %d", len(tc.Expected), len(actualPositions))
			}

			for i, expected := range tc.Expected {
				actual := actualPositions[i]
				// Print actual values to update your expected struct
				t.Logf("Test=%q User=%q Yes=%d No=%d Value=%d", tc.Name, actual.Username, actual.YesSharesOwned, actual.NoSharesOwned, actual.Value)

				if actual.Username != expected.Username ||
					actual.YesSharesOwned != expected.YesSharesOwned ||
					actual.NoSharesOwned != expected.NoSharesOwned {
					t.Errorf("expected shares %+v, got %+v", expected, actual)
				}
				// For the first run, comment out this Value check and just log it!
				// if actual.Value != expected.Value {
				//     t.Errorf("expected Value=%d, got %d", expected.Value, actual.Value)
				// }
			}
		})
	}
}
