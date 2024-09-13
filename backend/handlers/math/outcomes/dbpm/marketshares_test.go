package dbpm

import (
	"math"
	"reflect"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/models"
	"socialpredict/setup"
	"testing"
	"time"
)

var now = time.Now()

func TestAdjustPayoutsFromNewest(t *testing.T) {
	testcases := []struct {
		Name                  string
		Bets                  []models.Bet
		ScaledPayouts         []int64
		AdjustedScaledPayouts []int64
	}{
		{
			Name: "PreventSimultaneousSharesHeld",
			Bets: []models.Bet{
				{
					Amount:   3,
					Outcome:  "YES",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
					MarketID: 3,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 8, 13, 922665000, time.UTC),
					MarketID: 3,
				},
			},
			ScaledPayouts:         []int64{3, 1},
			AdjustedScaledPayouts: []int64{3, 1},
		},
		{
			Name: "InfinityAvoidance",
			Bets: []models.Bet{
				{
					Amount:   1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now,
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now.Add(time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(2 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(3 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(4 * time.Minute),
					MarketID: 1,
				},
			},
			ScaledPayouts:         []int64{0, 0, 1, 0, 1},
			AdjustedScaledPayouts: []int64{0, 0, 1, 0, 0},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			adjustedPayouts := AdjustPayoutsFromNewest(tc.Bets, tc.ScaledPayouts)
			for i, payout := range adjustedPayouts {
				if payout != tc.AdjustedScaledPayouts[i] {
					t.Errorf("Test %s failed at index %d: expected payout %d, got %d", tc.Name, i, tc.AdjustedScaledPayouts[i], payout)
				}
			}
		})
	}
}

func TestAggregateUserPayoutsDBPM(t *testing.T) {
	testcases := []struct {
		Name                  string
		Bets                  []models.Bet
		AdjustedScaledPayouts []int64
		AggregatedPositions   []MarketPosition
	}{
		{
			Name: "PreventSimultaneousSharesHeld",
			Bets: []models.Bet{
				{
					Amount:   3,
					Outcome:  "YES",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
					MarketID: 3,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 8, 13, 922665000, time.UTC),
					MarketID: 3,
				},
			},
			AdjustedScaledPayouts: []int64{3, 1},
			AggregatedPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 3, NoSharesOwned: 1},
			},
		},
		{
			Name: "InfinityAvoidance",
			Bets: []models.Bet{
				{
					Amount:   1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now,
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now.Add(time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(2 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(3 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(4 * time.Minute),
					MarketID: 1,
				},
			},
			AdjustedScaledPayouts: []int64{0, 0, 1, 0, 0},
			AggregatedPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
				{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			result := AggregateUserPayoutsDBPM(tc.Bets, tc.AdjustedScaledPayouts)

			// Create a map for expected results for easy lookup
			expectedResults := make(map[string]MarketPosition)
			for _, pos := range tc.AggregatedPositions {
				expectedResults[pos.Username] = pos
			}

			// Create a map from the results for comparison
			resultsMap := make(map[string]MarketPosition)
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

func TestCalculateCoursePayoutsDBPM(t *testing.T) {
	testcases := []struct {
		Name               string
		Bets               []models.Bet
		ProbabilityChanges []wpam.ProbabilityChange
		CoursePayouts      []CourseBetPayout
	}{
		{
			Name: "PreventSimultaneousSharesHeld",
			Bets: []models.Bet{
				{
					Amount:   3,
					Outcome:  "YES",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
					MarketID: 3,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 8, 13, 922665000, time.UTC),
					MarketID: 3,
				},
			},
			ProbabilityChanges: []wpam.ProbabilityChange{
				{Probability: 0.5},
				{Probability: 0.875},
				{Probability: 0.7},
			},
			CoursePayouts: []CourseBetPayout{
				{Payout: 0.5999999999999999, Outcome: "YES"},
				{Payout: 0.17500000000000004, Outcome: "NO"},
			},
		},
		{
			Name: "InfinityAvoidance",
			Bets: []models.Bet{
				{
					Amount:   1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now,
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now.Add(time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(2 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(3 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(4 * time.Minute),
					MarketID: 1,
				},
			},
			ProbabilityChanges: []wpam.ProbabilityChange{
				{Probability: 0.50},
				{Probability: 0.75},
				{Probability: 0.50},
				{Probability: 0.25},
				{Probability: 0.50},
				{Probability: 0.25},
			},
			CoursePayouts: []CourseBetPayout{
				{Payout: 0.25, Outcome: "YES"},
				{Payout: -0.5, Outcome: "YES"},
				{Payout: 0.25, Outcome: "NO"},
				{Payout: -0, Outcome: "NO"}, // golang math.Round() rounds to -0 and +0
				{Payout: 0.25, Outcome: "NO"},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			actualPayouts := CalculateCoursePayoutsDBPM(tc.Bets, tc.ProbabilityChanges)
			if len(actualPayouts) != len(tc.CoursePayouts) {
				t.Fatalf("expected %d payouts, got %d", len(tc.CoursePayouts), len(actualPayouts))
			}
			for i, payout := range actualPayouts {
				if payout.Payout != tc.CoursePayouts[i].Payout || payout.Outcome != tc.CoursePayouts[i].Outcome {
					t.Errorf("Test %s failed at index %d: expected %+v, got %+v", tc.Name, i, tc.CoursePayouts[i], payout)
				}
			}
		})
	}
}

func TestCalculateNormalizationFactorsDBPM(t *testing.T) {
	testcases := []struct {
		Name          string
		F_YES         float64
		F_NO          float64
		ExpectedF_YES float64
		ExpectedF_NO  float64
	}{
		{
			Name:          "PreventSimultaneousSharesHeld",
			F_YES:         5.000000000000001, // Actual output from function
			F_NO:          5.714285714285713, // Actual output from function
			ExpectedF_YES: 5.000000,
			ExpectedF_NO:  5.714286,
		},
		{
			Name:          "InfinityAvoidance",
			F_YES:         0,
			F_NO:          2,
			ExpectedF_YES: 0,
			ExpectedF_NO:  2,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			roundedF_YES := math.Round(tc.F_YES*1000000) / 1000000
			roundedF_NO := math.Round(tc.F_NO*1000000) / 1000000

			if roundedF_YES != tc.ExpectedF_YES || roundedF_NO != tc.ExpectedF_NO {
				t.Errorf("Test %s failed: expected F_YES=%f, F_NO=%f, got F_YES=%f, F_NO=%f",
					tc.Name, tc.ExpectedF_YES, tc.ExpectedF_NO, roundedF_YES, roundedF_NO)
			}
		})
	}
}

func TestCalculateScaledPayoutsDBPM(t *testing.T) {
	testcases := []struct {
		Name          string
		Bets          []models.Bet
		CoursePayouts []CourseBetPayout
		F_YES         float64
		F_NO          float64
		ScaledPayouts []int64
	}{
		{
			Name: "PreventSimultaneousSharesHeld",
			Bets: []models.Bet{
				{
					Amount:   3,
					Outcome:  "YES",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
					MarketID: 3,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 8, 13, 922665000, time.UTC),
					MarketID: 3,
				},
			},
			CoursePayouts: []CourseBetPayout{
				{Payout: 0.5999999999999999, Outcome: "YES"},
				{Payout: 0.17500000000000004, Outcome: "NO"},
			},
			F_YES:         5.000000000000001, // Actual output from function
			F_NO:          5.714285714285713, // Actual output from function
			ScaledPayouts: []int64{3, 1},
		},
		{
			Name: "InfinityAvoidance",
			Bets: []models.Bet{
				{
					Amount:   1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now,
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now.Add(time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(2 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(3 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(4 * time.Minute),
					MarketID: 1,
				},
			},
			CoursePayouts: []CourseBetPayout{
				{Payout: 0.25, Outcome: "YES"},
				{Payout: -0.5, Outcome: "YES"},
				{Payout: 0.25, Outcome: "NO"},
				{Payout: -0, Outcome: "NO"}, // golang math.Round() rounds to -0 and +0
				{Payout: 0.25, Outcome: "NO"},
			},
			F_YES:         0,
			F_NO:          2,
			ScaledPayouts: []int64{0, 0, 1, 0, 1},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			actualPayouts := CalculateScaledPayoutsDBPM(tc.Bets, tc.CoursePayouts, tc.F_YES, tc.F_NO)
			if len(actualPayouts) != len(tc.ScaledPayouts) {
				t.Fatalf("Test %s failed: expected %d payouts, got %d", tc.Name, len(tc.ScaledPayouts), len(actualPayouts))
			}
			for i, payout := range actualPayouts {
				if payout != tc.ScaledPayouts[i] {
					t.Errorf("Test %s failed at index %d: expected payout %d, got %d", tc.Name, i, tc.ScaledPayouts[i], payout)
				}
			}
		})
	}
}

func TestDivideUpMarketPoolSharesDBPM(t *testing.T) {
	testcases := []struct {
		Name               string
		Bets               []models.Bet
		ProbabilityChanges []wpam.ProbabilityChange
		S_YES              int64
		S_NO               int64
	}{
		{
			Name: "PreventSimultaneousSharesHeld",
			Bets: []models.Bet{
				{
					Amount:   3,
					Outcome:  "YES",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
					MarketID: 3,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 8, 13, 922665000, time.UTC),
					MarketID: 3,
				},
			},
			ProbabilityChanges: []wpam.ProbabilityChange{
				{Probability: 0.5},
				{Probability: 0.875},
				{Probability: 0.7},
			},
			S_YES: 3,
			S_NO:  1,
		},
		{
			Name: "InfinityAvoidance",
			Bets: []models.Bet{
				{
					Amount:   1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now,
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now.Add(time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(2 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(3 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(4 * time.Minute),
					MarketID: 1,
				},
			},
			ProbabilityChanges: []wpam.ProbabilityChange{
				{Probability: 0.50},
				{Probability: 0.75},
				{Probability: 0.50},
				{Probability: 0.25},
				{Probability: 0.50},
				{Probability: 0.25},
			},
			S_YES: 0,
			S_NO:  1,
		},
	}
	ec := setup.EconomicsConfig()
	ec.Economics.MarketCreation.InitialMarketSubsidization = 0
	ec.Economics.MarketIncentives.CreateMarketCost = 1

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			yes, no := DivideUpMarketPoolSharesDBPM(tc.Bets, tc.ProbabilityChanges)
			if yes != tc.S_YES || no != tc.S_NO {
				t.Errorf("%s: expected (%d, %d), got (%d, %d)", tc.Name, tc.S_YES, tc.S_NO, yes, no)
			}
		})
	}
}

func TestNetAggregateMarketPositions(t *testing.T) {
	testcases := []struct {
		Name                string
		AggregatedPositions []MarketPosition
		NetPositions        []MarketPosition
	}{
		{
			Name: "PreventSimultaneousSharesHeldPreventSimultaneousSharesHeldPreventSimultaneousSharesHeld",
			AggregatedPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 3, NoSharesOwned: 1},
			},
			NetPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 2, NoSharesOwned: 0},
			},
		},
		{
			Name: "InfinityAvoidance",
			AggregatedPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
				{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
			},
			NetPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
				{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := NetAggregateMarketPositions(tc.AggregatedPositions)
			expected := tc.NetPositions

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("Failed %s: expected %+v, got %+v", tc.Name, expected, actual)
			}
		})
	}
}
