package wpam_test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/models"
	"testing"
	"time"
)

type TestCase struct {
	Name                  string
	Bets                  []models.Bet
	ProbabilityChanges    []wpam.ProbabilityChange
	S_YES                 int64
	S_NO                  int64
	CoursePayouts         []dbpm.CourseBetPayout
	F_YES                 float64
	F_NO                  float64
	ExpectedF_YES         float64
	ExpectedF_NO          float64
	ScaledPayouts         []int64
	AdjustedScaledPayouts []int64
	AggregatedPositions   []dbpm.DBPMMarketPosition
	NetPositions          []dbpm.DBPMMarketPosition
}

var now = time.Now() // Capture current time for consistent test data

var TestCases = []TestCase{
	{
		Name: "Prevent simultaneous shares held",
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
			{Probability: 0.61538461538461542},
			{Probability: 0.57142857142857140},
		},
		S_YES: 3,
		S_NO:  1,
		CoursePayouts: []dbpm.CourseBetPayout{
			{Payout: 0.5999999999999999, Outcome: "YES"},
			{Payout: 0.17500000000000004, Outcome: "NO"},
		},
		F_YES:                 5.000000000000001,
		F_NO:                  5.714285714285713,
		ExpectedF_YES:         5.000000,
		ExpectedF_NO:          5.714286,
		ScaledPayouts:         []int64{3, 1},
		AdjustedScaledPayouts: []int64{3, 1},
		AggregatedPositions: []dbpm.DBPMMarketPosition{
			{Username: "user1", YesSharesOwned: 3, NoSharesOwned: 1},
		},
		NetPositions: []dbpm.DBPMMarketPosition{
			{Username: "user1", YesSharesOwned: 2, NoSharesOwned: 0},
		},
	},
	{
		Name: "infinity avoidance",
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
			{Probability: 0.54545454545454541},
			{Probability: 0.50},
			{Probability: 0.45454545454545453},
			{Probability: 0.50},
			{Probability: 0.45454545454545453},
		},
		S_YES: 0,
		S_NO:  1,
		CoursePayouts: []dbpm.CourseBetPayout{
			{Payout: 0.25, Outcome: "YES"},
			{Payout: -0.5, Outcome: "YES"},
			{Payout: 0.25, Outcome: "NO"},
			{Payout: -0, Outcome: "NO"}, // golang math.Round() rounds to -0 and +0
			{Payout: 0.25, Outcome: "NO"},
		},
		F_YES:                 0,
		F_NO:                  2,
		ExpectedF_YES:         0,
		ExpectedF_NO:          2,
		ScaledPayouts:         []int64{0, 0, 1, 0, 1},
		AdjustedScaledPayouts: []int64{0, 0, 1, 0, 0},
		AggregatedPositions: []dbpm.DBPMMarketPosition{
			{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
			{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
		},
		NetPositions: []dbpm.DBPMMarketPosition{
			{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
			{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
		},
	},
}

func TestCalculateMarketProbabilitiesWPAM(t *testing.T) {
	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {

			// Call the function under test
			probChanges := wpam.CalculateMarketProbabilitiesWPAM(tc.Bets[0].PlacedAt, tc.Bets)

			if len(probChanges) != len(tc.ProbabilityChanges) {
				t.Fatalf("expected %d probability changes, got %d", len(tc.ProbabilityChanges), len(probChanges))
			}

			for i, pc := range probChanges {
				expected := tc.ProbabilityChanges[i]

				// Change fmt.Printf to t.Logf for debug logging
				t.Logf("at index %d, expected probability %.17f, got %.17f", i, expected.Probability, pc.Probability)

				if pc.Probability != expected.Probability {
					t.Errorf("at index %d, expected probability %.17f, got %.17f", i, expected.Probability, pc.Probability)
				}
			}

		})
	}
}
