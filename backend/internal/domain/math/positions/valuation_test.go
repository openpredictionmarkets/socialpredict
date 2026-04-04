package positionsmath

import (
	"testing"
	"time"
)

type valuationPositionInput struct {
	Username       string
	YesSharesOwned int64
	NoSharesOwned  int64
}

var valuationTestBaseTime = time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC)

func makeUserPositions(data []valuationPositionInput) map[string]UserMarketPosition {
	result := make(map[string]UserMarketPosition)
	for _, d := range data {
		result[d.Username] = UserMarketPosition{
			YesSharesOwned: d.YesSharesOwned,
			NoSharesOwned:  d.NoSharesOwned,
		}
	}
	return result
}

func makeEarliestValuationBets(data []valuationPositionInput) map[string]time.Time {
	earliest := make(map[string]time.Time, len(data))
	for i, pos := range data {
		earliest[pos.Username] = valuationTestBaseTime.Add(time.Duration(i) * time.Minute)
	}
	return earliest
}

func assertValuations(t *testing.T, actual map[string]UserValuationResult, expected map[string]int64) {
	t.Helper()
	for user, want := range expected {
		if got := actual[user].RoundedValue; got != want {
			t.Fatalf("user %s: expected value %d, got %d", user, want, got)
		}
	}
}

func TestCalculateRoundedUserValuationsFromUserMarketPositions(t *testing.T) {
	testcases := []struct {
		Name             string
		UserPositions    []valuationPositionInput
		Probability      float64
		TotalVolume      int64
		IsResolved       bool
		ResolutionResult string
		Expected         map[string]int64
	}{
		{
			Name: "Unresolved market, YES/NO users at 50%",
			UserPositions: []valuationPositionInput{
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
			UserPositions: []valuationPositionInput{
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
			UserPositions: []valuationPositionInput{
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
			UserPositions: []valuationPositionInput{
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
			UserPositions: []valuationPositionInput{
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
			actual, err := CalculateRoundedUserValuationsFromUserMarketPositions(
				positions,
				tc.Probability,
				tc.TotalVolume,
				tc.IsResolved,
				tc.ResolutionResult,
				makeEarliestValuationBets(tc.UserPositions),
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertValuations(t, actual, tc.Expected)
		})
	}
}
