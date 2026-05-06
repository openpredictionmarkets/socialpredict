package wpam_test

import (
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
	"socialpredict/internal/domain/math/outcomes/dbpm"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models/modelstesting"
)

type fixedFormula struct{ probability float64 }

func (f fixedFormula) Calculate(wpam.Seeds, int64, int64) float64 { return f.probability }

type TestCase struct {
	Name                  string
	Bets                  []boundary.Bet
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

var wpamProbabilityBaseTime = time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)

var TestCases = []TestCase{
	{
		Name: "Prevent simultaneous shares held",
		Bets: []boundary.Bet{
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
		Bets: []boundary.Bet{
			{
				Amount:   1,
				Outcome:  "YES",
				Username: "user2",
				PlacedAt: wpamProbabilityBaseTime,
				MarketID: 1,
			},
			{
				Amount:   -1,
				Outcome:  "YES",
				Username: "user2",
				PlacedAt: wpamProbabilityBaseTime.Add(time.Minute),
				MarketID: 1,
			},
			{
				Amount:   1,
				Outcome:  "NO",
				Username: "user1",
				PlacedAt: wpamProbabilityBaseTime.Add(2 * time.Minute),
				MarketID: 1,
			},
			{
				Amount:   -1,
				Outcome:  "NO",
				Username: "user1",
				PlacedAt: wpamProbabilityBaseTime.Add(3 * time.Minute),
				MarketID: 1,
			},
			{
				Amount:   1,
				Outcome:  "NO",
				Username: "user1",
				PlacedAt: wpamProbabilityBaseTime.Add(4 * time.Minute),
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

func newWPAMTestCalculator() wpam.ProbabilityCalculator {
	econ := modelstesting.GenerateEconomicConfig()
	return wpam.NewProbabilityCalculator(wpam.StaticSeedProvider{Value: wpam.Seeds{
		InitialProbability:     econ.Economics.MarketCreation.InitialMarketProbability,
		InitialSubsidization:   econ.Economics.MarketCreation.InitialMarketSubsidization,
		InitialYesContribution: econ.Economics.MarketCreation.InitialMarketYes,
		InitialNoContribution:  econ.Economics.MarketCreation.InitialMarketNo,
	}})
}

func assertProbabilityChangesEqual(t *testing.T, actual, expected []wpam.ProbabilityChange) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("expected %d probability changes, got %d", len(expected), len(actual))
	}

	for i, change := range actual {
		if change.Probability != expected[i].Probability {
			t.Fatalf("at index %d, expected probability %.17f, got %.17f", i, expected[i].Probability, change.Probability)
		}
	}
}

func TestCalculateMarketProbabilitiesWPAM(t *testing.T) {
	calculator := newWPAMTestCalculator()

	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			probChanges := calculator.CalculateMarketProbabilitiesWPAM(tc.Bets[0].PlacedAt, tc.Bets)
			assertProbabilityChangesEqual(t, probChanges, tc.ProbabilityChanges)
		})
	}
}

func TestNewProbabilityCalculatorWithOptions(t *testing.T) {
	calculator := wpam.NewProbabilityCalculatorWithOptions(
		wpam.StaticSeedProvider{Value: wpam.Seeds{InitialProbability: 0.2}},
		wpam.WithProbabilityFormula(fixedFormula{probability: 0.8}),
	)

	bets := []boundary.Bet{{Amount: 5, Outcome: "YES", PlacedAt: wpamProbabilityBaseTime}}
	changes := calculator.CalculateMarketProbabilitiesWPAM(wpamProbabilityBaseTime, bets)

	if len(changes) != 2 {
		t.Fatalf("expected 2 probability changes, got %d", len(changes))
	}
	if changes[1].Probability != 0.8 {
		t.Fatalf("expected injected formula probability 0.8, got %f", changes[1].Probability)
	}
}
