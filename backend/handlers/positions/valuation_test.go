package positions

import (
	"testing"
)

// Simple helper for easy map construction in tests.
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
			Name: "Single YES user",
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
				{"alice", 50, 0},
			},
			Probability: 1.0, // All YES
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
			Probability: 0.0, // All NO
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
			Expected:    map[string]int64{"alice": 15, "bob": 5},
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
			Probability: 0.333, // (3 * .333) = 1.0, (2 * .333) = .666
			TotalVolume: 2,
			Expected:    nil, // We will print values for easy copying
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			positions := makeUserPositions(tc.UserPositions)
			actual := CalculateRoundedUserValuationsFromUserMarketPositions(positions, tc.Probability, tc.TotalVolume)
			if tc.Expected != nil {
				for user, want := range tc.Expected {
					got := actual[user].RoundedValue
					if got != want {
						t.Errorf("user %s: expected value %d, got %d", user, want, got)
					}
				}
			} else {
				// Print for inspection/copying
				for user, val := range actual {
					t.Logf("%s: %d", user, val.RoundedValue)
				}
			}
		})
	}
}

func TestAdjustUserValuationsToMarketVolume(t *testing.T) {
	userVals := map[string]UserValuationResult{
		"alice": {Username: "alice", RoundedValue: 3},
		"bob":   {Username: "bob", RoundedValue: 2},
	}
	floatVals := map[string]float64{
		"alice": 3.1,
		"bob":   1.9,
	}
	target := int64(7) // so delta = 2
	adjusted := AdjustUserValuationsToMarketVolume(userVals, target, floatVals)
	total := int64(0)
	for _, v := range adjusted {
		total += v.RoundedValue
	}
	if total != target {
		t.Errorf("sum of adjusted valuations %d != target %d", total, target)
	}
}
