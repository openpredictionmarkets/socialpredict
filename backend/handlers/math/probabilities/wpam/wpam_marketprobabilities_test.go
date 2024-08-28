package wpam

import (
	"log"
	"socialpredict/models"
	"socialpredict/setup"
	"testing"
	"time"
)

type TestCase struct {
	Name               string
	Bets               []models.Bet
	ProbabilityChanges []ProbabilityChange
	S_YES              int64
	S_NO               int64
	//CoursePayouts         []dbpm.CourseBetPayout
	F_YES                 float64
	F_NO                  float64
	ExpectedF_YES         float64 // Ensure names match what you use in tests
	ExpectedF_NO          float64
	ScaledPayouts         []int64
	AdjustedScaledPayouts []int64
	//AggregatedPositions   []dbpm.MarketPosition
	//NetPositions          []dbpm.MarketPosition
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
		ProbabilityChanges: []ProbabilityChange{
			{Probability: 0.5},
			{Probability: 0.875},
			{Probability: 0.7},
		},
		S_YES: 3,
		S_NO:  1,
		/*
			CoursePayouts: []dbpm.CourseBetPayout{
				{Payout: 0.5999999999999999, Outcome: "YES"},
				{Payout: 0.17500000000000004, Outcome: "NO"},
			},*/
		F_YES:                 5.000000000000001, // Actual output from function
		F_NO:                  5.714285714285713, // Actual output from function
		ExpectedF_YES:         5.000000,
		ExpectedF_NO:          5.714286,
		ScaledPayouts:         []int64{3, 1},
		AdjustedScaledPayouts: []int64{3, 1},
		/*AggregatedPositions: []dbpm.MarketPosition{
			{Username: "user1", YesSharesOwned: 3, NoSharesOwned: 1},
		},
		NetPositions: []dbpm.MarketPosition{
			{Username: "user1", YesSharesOwned: 2, NoSharesOwned: 0},
		},*/
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
		S_YES: 0,
		S_NO:  1,
		/*CoursePayouts: []dbpm.CourseBetPayout{
			{Payout: 0.25, Outcome: "YES"},
			{Payout: -0.5, Outcome: "YES"},
			{Payout: 0.25, Outcome: "NO"},
			{Payout: -0, Outcome: "NO"}, // golang math.Round() rounds to -0 and +0
			{Payout: 0.25, Outcome: "NO"},
		},*/
		F_YES:                 0,
		F_NO:                  2,
		ExpectedF_YES:         0,
		ExpectedF_NO:          2,
		ScaledPayouts:         []int64{0, 0, 1, 0, 1},
		AdjustedScaledPayouts: []int64{0, 0, 1, 0, 0},
		/*AggregatedPositions: []dbpm.MarketPosition{
			{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
			{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
		},
		NetPositions: []dbpm.MarketPosition{
			{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
			{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
		},*/
	},
}

func TestCalculateMarketProbabilities(t *testing.T) {
	appConfig := setup.MustLoadEconomicsConfig()
	log.Printf("app config: %v", appConfig)
	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Call the function under test
			probChanges := CalculateMarketProbabilitiesWPAM(appConfig, tc.Bets[0].PlacedAt, tc.Bets)

			if len(probChanges) != len(tc.ProbabilityChanges) {
				t.Fatalf("expected %d probability changes, got %d", len(tc.ProbabilityChanges), len(probChanges))
			}

			for i, pc := range probChanges {
				expected := tc.ProbabilityChanges[i]
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
		appConfig *setup.EconomicConfig
		no        int64
		yes       int64
		want      float64
	}{
		{
			name:      "no bets",
			appConfig: buildInitialMarketAppConfig(t, .5, 10, 0, 0), //buildAppConfig(t, .5, 10, 0, 0, 10, 1, 0, 500, 1, 1, 0, 0),
			no:        0,
			yes:       0,
			want:      .5,
		},
		{
			name:      "3 yes",
			appConfig: buildInitialMarketAppConfig(t, .5, 1, 3, 0), //buildAppConfig(t, .5, 1, 0, 0, 10, 1, 0, 500, 1, 1, 0, 0),
			no:        0,
			yes:       3,
			want:      .875,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := calcProbability(test.appConfig.Economics.MarketCreation.InitialMarketProbability, test.appConfig.Economics.MarketCreation.InitialMarketSubsidization, test.yes, test.no)
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

// buildAppConfig builds an entire appConfig
func buildAppConfig(t *testing.T, initProbability float64, initSubsidization, initYes, initNo, createCost, bonus, initBalance, maxDebt, minBet, initBetFee, betFee, sellFee int64) *setup.EconomicConfig {
	t.Helper()
	return &setup.EconomicConfig{
		Economics: setup.Economics{
			MarketCreation: setup.MarketCreation{
				InitialMarketProbability:   initProbability,
				InitialMarketSubsidization: initSubsidization,
				InitialMarketYes:           initYes,
				InitialMarketNo:            initNo,
			},
			MarketIncentives: setup.MarketIncentives{
				CreateMarketCost: createCost,
				TraderBonus:      bonus,
			},
			User: setup.User{
				InitialAccountBalance: initBalance,
				MaximumDebtAllowed:    maxDebt,
			},
			Betting: setup.Betting{
				MinimumBet: minBet,
				BetFees: setup.BetFees{
					InitialBetFee: initBetFee,
					EachBetFee:    betFee,
					SellSharesFee: sellFee,
				},
			},
		},
	}
}
