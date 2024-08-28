package wpam

import (
	"socialpredict/models"
	"socialpredict/setup"
	"testing"
	"time"
)

var now = time.Now() // Capture current time for consistent test data

func TestCalculateMarketProbabilities(t *testing.T) {
	tests := []struct {
		Name               string
		Bets               []models.Bet
		ProbabilityChanges []ProbabilityChange
		appConfig          func() *setup.EconomicConfig
	}{
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
			ProbabilityChanges: []ProbabilityChange{
				{Probability: 0.5},
				{Probability: 0.875},
				{Probability: 0.7},
			},
			appConfig: func() *setup.EconomicConfig { return buildInitialMarketAppConfig(t, .5, 1, 0, 0) },
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
			ProbabilityChanges: []ProbabilityChange{
				{Probability: 0.50},
				{Probability: 0.75},
				{Probability: 0.50},
				{Probability: 0.25},
				{Probability: 0.50},
				{Probability: 0.25},
			},
			appConfig: func() *setup.EconomicConfig { return buildInitialMarketAppConfig(t, .5, 1, 0, 0) },
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Call the function under test
			probChanges := CalculateMarketProbabilitiesWPAM(test.appConfig, test.Bets[0].PlacedAt, test.Bets)

			if len(probChanges) != len(test.ProbabilityChanges) {
				t.Fatalf("expected %d probability changes, got %d", len(test.ProbabilityChanges), len(probChanges))
			}

			for i, pc := range probChanges {
				expected := test.ProbabilityChanges[i]
				if pc.Probability != expected.Probability {
					t.Errorf("at index %d, expected probability %f, got %f", i, expected.Probability, pc.Probability)
				}
			}
		})
	}
}

func TestCalcProbability(t *testing.T) {
	tests := []struct {
		name      string
		appConfig func() *setup.EconomicConfig
		no        int64
		yes       int64
		want      float64
	}{
		{
			name:      "NoBets",
			appConfig: func() *setup.EconomicConfig { return buildInitialMarketAppConfig(t, .5, 10, 0, 0) }, //buildAppConfig(t, .5, 10, 0, 0, 10, 1, 0, 500, 1, 1, 0, 0),

			no:   0,
			yes:  0,
			want: .5,
		},
		{
			name:      "3YesBets",
			appConfig: func() *setup.EconomicConfig { return buildInitialMarketAppConfig(t, .5, 1, 3, 0) }, //buildAppConfig(t, .5, 1, 0, 0, 10, 1, 0, 500, 1, 1, 0, 0),
			no:        0,
			yes:       3,
			want:      .875,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			econConfig := test.appConfig()
			got := calcProbability(econConfig.Economics.MarketCreation.InitialMarketProbability, econConfig.Economics.MarketCreation.InitialMarketSubsidization, test.yes, test.no)
			if test.want != got {
				t.Errorf("Unexpected return value calculating probability, want %f, got %f", test.want, got)
			}
		})
	}
}

// buildInitialMarketAppConfig builds the MarketCreation portion of the app config
func buildInitialMarketAppConfig(t *testing.T, probability float64, subsidization, yes, no int64) *setup.EconomicConfig {
	t.Helper()
	return &setup.EconomicConfig{
		Economics: setup.Economics{
			MarketCreation: setup.MarketCreation{
				InitialMarketProbability:   probability,
				InitialMarketSubsidization: subsidization,
				InitialMarketYes:           yes,
				InitialMarketNo:            no,
			},
		},
	}
}
