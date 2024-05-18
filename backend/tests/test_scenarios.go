package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/models"
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
	ScaledPayouts         []int64
	AdjustedScaledPayouts []int64
	AggregatedPositions   []dbpm.MarketPosition
	NetPositions          []dbpm.MarketPosition
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
			{Probability: 0.875},
			{Probability: 0.7},
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
			{Probability: 0.75},
			{Probability: 0.50},
			{Probability: 0.25},
			{Probability: 0.50},
			{Probability: 0.25},
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
		ScaledPayouts:         []int64{0, 0, 1, 0, 1},
		AdjustedScaledPayouts: []int64{0, 0, 1, 0, 0},
		AggregatedPositions: []dbpm.MarketPosition{
			{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
			{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
		},
	},
}
