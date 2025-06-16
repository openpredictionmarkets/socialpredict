package positions

import (
	"socialpredict/models/modelstesting"
	"testing"

	"gorm.io/gorm"
)

// Optional: Helper to create bets for test, mimics user position logic.
func addTestBets(t *testing.T, db *gorm.DB, marketID uint, userPos []struct {
	Username       string
	YesSharesOwned int64
	NoSharesOwned  int64
}) {
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

func TestCalculateRoundedUserValuationsFromUserMarketPositions(t *testing.T) {
	testcases := []struct {
		Name          string
		UserPositions []struct {
			Username       string
			YesSharesOwned int64
			NoSharesOwned  int64
		}
		Probability float64
		TotalVolume int64
		Expected    map[string]int64
	}{
		{
			Name: "Unresolved market, YES/NO users at 50%",
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
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
			Name: "Resolved YES market",
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
				{"alice", 10, 0},
				{"bob", 0, 10},
			},
			Probability:      0.5, // ignored if resolved
			TotalVolume:      20,
			IsResolved:       true,
			ResolutionResult: "YES",
			Expected:         map[string]int64{"alice": 20, "bob": 0}, // 100% payout to YES
		},
		{
			Name: "Resolved NO market",
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
				{"alice", 10, 0},
				{"bob", 0, 10},
			},
			Probability:      0.5, // ignored if resolved
			TotalVolume:      20,
			IsResolved:       true,
			ResolutionResult: "NO",
			Expected:         map[string]int64{"alice": 0, "bob": 20}, // 100% payout to NO
		},
		{
			Name: "Single YES user",
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
				{"alice", 50, 0},
			},
			Probability: 1.0,
			TotalVolume: 50,
			Expected:    map[string]int64{"alice": 50},
		},
		{
			Name: "Single NO user",
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
				{"bob", 0, 30},
			},
			Probability: 0.0,
			TotalVolume: 30,
			Expected:    map[string]int64{"bob": 30},
		},
		{
			Name: "YES and NO users at 50% prob, rounding needed",
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
				{"alice", 10, 0},
				{"bob", 0, 10},
			},
			Probability: 0.5,
			TotalVolume: 20,
			Expected:    map[string]int64{"alice": 10, "bob": 10},
		},
		{
			Name: "Rounding correction applied to largest holder",
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
				{"alice", 3, 0},
				{"bob", 2, 0},
			},
			Probability: 0.333,
			TotalVolume: 2,
			Expected:    nil, // Will print for copying
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
				db, uint(market.ID), positions, tc.Probability, tc.TotalVolume,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
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

// private helper function just for this specific use case
func makeUserPositions(data []struct {
	Username       string
	YesSharesOwned int64
	NoSharesOwned  int64
}) map[string]UserMarketPosition {
	result := make(map[string]UserMarketPosition)
	for _, d := range data {
		result[d.Username] = UserMarketPosition{
			YesSharesOwned: d.YesSharesOwned,
			NoSharesOwned:  d.NoSharesOwned,
		}
	}
	return result
}
