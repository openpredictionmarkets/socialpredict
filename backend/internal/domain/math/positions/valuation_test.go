package positionsmath

import (
	"testing"
	"time"
)

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

func TestCalculateRoundedUserValuationsFromUserMarketPositions(t *testing.T) {
	testcases := []struct {
		Name          string
		UserPositions []struct {
			Username       string
			YesSharesOwned int64
			NoSharesOwned  int64
		}
		Probability      float64
		TotalVolume      int64
		IsResolved       bool
		ResolutionResult string
		Expected         map[string]int64
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
			Name: "Resolved market: YES wins",
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
			Expected:         map[string]int64{"alice": 20, "bob": 0},
		},
		{
			Name: "Resolved market: NO wins",
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
			Expected:         map[string]int64{"alice": 0, "bob": 20},
		},
		{
			Name: "Resolved market: All YES, NO wins (all get zero)",
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
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
			UserPositions: []struct {
				Username       string
				YesSharesOwned int64
				NoSharesOwned  int64
			}{
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
			positions := makeUserPositions(tc.UserPositions)
			earliest := make(map[string]time.Time)
			base := time.Now()
			for i, pos := range tc.UserPositions {
				earliest[pos.Username] = base.Add(time.Duration(i) * time.Minute)
			}

			actual, err := CalculateRoundedUserValuationsFromUserMarketPositions(
				positions,
				tc.Probability,
				tc.TotalVolume,
				tc.IsResolved,
				tc.ResolutionResult,
				earliest,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for user, want := range tc.Expected {
				got := actual[user].RoundedValue
				if got != want {
					t.Errorf("user %s: expected value %d, got %d", user, want, got)
				}
			}
		})
	}
}
