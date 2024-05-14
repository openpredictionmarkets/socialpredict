package test

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/models"
	"time"
)

type TestCase struct {
	Name               string
	Bets               []models.Bet
	ProbabilityChanges []wpam.ProbabilityChange
	ExpectedYES        int64
	ExpectedNO         int64
	ExpectedPayouts    []dbpm.CourseBetPayout
}

var now = time.Now() // Capture current time for consistent test data

var TestCases = []TestCase{
	{
		Name: "infinity avoidance",
		Bets: []models.Bet{
			{
				Amount:   1,
				Outcome:  "YES",
				PlacedAt: now,
				MarketID: 1,
			},
			{
				Amount:   -1,
				Outcome:  "YES",
				PlacedAt: now.Add(time.Minute),
				MarketID: 1,
			},
			{
				Amount:   1,
				Outcome:  "NO",
				PlacedAt: now.Add(2 * time.Minute),
				MarketID: 1,
			},
			{
				Amount:   -1,
				Outcome:  "NO",
				PlacedAt: now.Add(3 * time.Minute),
				MarketID: 1,
			},
			{
				Amount:   1,
				Outcome:  "NO",
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
		},
		ExpectedYES: 0,
		ExpectedNO:  1,
		ExpectedPayouts: []dbpm.CourseBetPayout{
			{Payout: 0, Outcome: "YES"},
			{Payout: -0.25, Outcome: "YES"},
			{Payout: 0, Outcome: "NO"},
			{Payout: -0.25, Outcome: "NO"},
			{Payout: 0, Outcome: "NO"},
		},
	},
}
