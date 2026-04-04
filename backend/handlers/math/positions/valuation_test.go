package positionsmath

import (
	"socialpredict/models/modelstesting"
	"testing"

	"gorm.io/gorm"
)

type userPositionInput struct {
	Username       string
	YesSharesOwned int64
	NoSharesOwned  int64
}

func addTestBets(t *testing.T, db *gorm.DB, marketID uint, userPos []userPositionInput) {
	t.Helper()

	for _, pos := range userPos {
		if pos.YesSharesOwned > 0 {
			bet := modelstesting.GenerateBet(
				pos.YesSharesOwned, "YES", pos.Username, marketID, 0,
			)
			db.Create(&bet)
		}
		if pos.NoSharesOwned > 0 {
			bet := modelstesting.GenerateBet(
				pos.NoSharesOwned, "NO", pos.Username, marketID, 0,
			)
			db.Create(&bet)
		}
	}
}

func makeUserPositions(data []userPositionInput) map[string]UserMarketPosition {
	positions := make(map[string]UserMarketPosition, len(data))
	for _, d := range data {
		positions[d.Username] = newUserMarketPosition(d.YesSharesOwned, d.NoSharesOwned)
	}
	return positions
}

func newUserMarketPosition(yesSharesOwned, noSharesOwned int64) UserMarketPosition {
	return UserMarketPosition{
		YesSharesOwned: yesSharesOwned,
		NoSharesOwned:  noSharesOwned,
	}
}

func TestCalculateRoundedUserValuationsFromUserMarketPositions(t *testing.T) {
	testcases := []struct {
		Name             string
		UserPositions    []userPositionInput
		Probability      float64
		TotalVolume      int64
		IsResolved       bool
		ResolutionResult string
		Expected         map[string]int64
	}{
		{
			Name: "Unresolved market, YES/NO users at 50%",
			UserPositions: []userPositionInput{
				{"alice", 10, 0},
				{"bob", 0, 10},
			},
			Probability:      0.5,
			TotalVolume:      20,
			IsResolved:       false,
			ResolutionResult: "",
			Expected:         map[string]int64{"alice": 10, "bob": 10},
		},
		{
			Name: "Resolved market: YES wins",
			UserPositions: []userPositionInput{
				{"alice", 10, 0},
				{"bob", 0, 10},
			},
			Probability:      0.5, // ignored if resolved
			TotalVolume:      20,
			IsResolved:       true,
			ResolutionResult: "YES",
			Expected:         map[string]int64{"alice": 20, "bob": 0},
		},
		{
			Name: "Resolved market: NO wins",
			UserPositions: []userPositionInput{
				{"alice", 10, 0},
				{"bob", 0, 10},
			},
			Probability:      0.5, // ignored if resolved
			TotalVolume:      20,
			IsResolved:       true,
			ResolutionResult: "NO",
			Expected:         map[string]int64{"alice": 0, "bob": 20},
		},
		{
			Name: "Resolved market: All YES, NO wins (all get zero)",
			UserPositions: []userPositionInput{
				{"alice", 10, 0},
				{"bob", 5, 0},
			},
			Probability:      0.8,
			TotalVolume:      15,
			IsResolved:       true,
			ResolutionResult: "NO",
			Expected:         map[string]int64{"alice": 0, "bob": 0},
		},
		{
			Name: "Resolved market: All NO, YES wins (all get zero)",
			UserPositions: []userPositionInput{
				{"alice", 0, 10},
				{"bob", 0, 5},
			},
			Probability:      0.2,
			TotalVolume:      15,
			IsResolved:       true,
			ResolutionResult: "YES",
			Expected:         map[string]int64{"alice": 0, "bob": 0},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			db := modelstesting.NewFakeDB(t)
			market := modelstesting.GenerateMarket(1, "creator")
			db.Create(&market)
			addTestBets(t, db, uint(market.ID), tc.UserPositions)
			positions := makeUserPositions(tc.UserPositions)

			actual, err := CalculateRoundedUserValuationsFromUserMarketPositions(
				db, uint(market.ID), positions, tc.Probability, tc.TotalVolume, tc.IsResolved, tc.ResolutionResult,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Log for debug
			for user, val := range actual {
				t.Logf("user=%s: value=%d", user, val.RoundedValue)
			}

			if tc.Expected != nil {
				for user, want := range tc.Expected {
					got := actual[user].RoundedValue
					if got != want {
						t.Errorf("user %s: expected value %d, got %d", user, want, got)
					}
				}
			} else {
				for user, val := range actual {
					t.Logf("%s: %d", user, val.RoundedValue)
				}
			}
		})
	}
}
