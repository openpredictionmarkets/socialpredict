package wpam

import (
	"fmt"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/setup/setuptesting"
	"testing"
	"time"
)

var now = time.Now() // Capture current time for consistent test data

func TestCalculateMarketProbabilitiesWPAM(t *testing.T) {
	testcases := []struct {
		Name                 string
		Bets                 models.Bets
		ProbabilityChanges   []ProbabilityChange
		MarketCreationConfig setup.MarketCreationLoader
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
			ProbabilityChanges: []ProbabilityChange{
				{Probability: 0.5},
				{Probability: 0.875},
				{Probability: 0.7},
			},
			MarketCreationConfig: setuptesting.BuildInitialMarketAppConfig(t, .5, 1, 0, 0),
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
			ProbabilityChanges: []ProbabilityChange{
				{Probability: 0.50},
				{Probability: 0.75},
				{Probability: 0.50},
				{Probability: 0.25},
				{Probability: 0.50},
				{Probability: 0.25},
			},
			MarketCreationConfig: setuptesting.BuildInitialMarketAppConfig(t, .5, 1, 0, 0),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {

			// Call the function under test
			probChanges := CalculateMarketProbabilitiesWPAM(tc.MarketCreationConfig, tc.Bets[0].PlacedAt, tc.Bets)

			if len(probChanges) != len(tc.ProbabilityChanges) {
				t.Fatalf("expected %d probability changes, got %d", len(tc.ProbabilityChanges), len(probChanges))
			}

			for i, pc := range probChanges {
				expected := tc.ProbabilityChanges[i]

				fmt.Printf("at index %d, expected probability %f, got %f\n", i, expected.Probability, pc.Probability)

				if pc.Probability != expected.Probability {
					t.Errorf("at index %d, expected probability %f, got %f", i, expected.Probability, pc.Probability)
				}
			}

		})
	}
}
