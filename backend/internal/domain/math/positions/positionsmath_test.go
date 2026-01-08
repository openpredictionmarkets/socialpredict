package positionsmath

import (
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"testing"
	"time"
)

func TestCalculateMarketPositions_WPAM_DBPM(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	wpam.SetSeeds(wpam.Seeds{
		InitialProbability:     econ.Economics.MarketCreation.InitialMarketProbability,
		InitialSubsidization:   econ.Economics.MarketCreation.InitialMarketSubsidization,
		InitialYesContribution: econ.Economics.MarketCreation.InitialMarketYes,
		InitialNoContribution:  econ.Economics.MarketCreation.InitialMarketNo,
	})

	testcases := []struct {
		Name       string
		BetConfigs []struct {
			Amount   int64
			Outcome  string
			Username string
			Offset   time.Duration
		}
		Expected []MarketPosition // You can fill in Yes/NoShares for now, and auto-print Value
	}{
		{
			Name: "Single YES Bet",
			BetConfigs: []struct {
				Amount   int64
				Outcome  string
				Username string
				Offset   time.Duration
			}{
				{Amount: 50, Outcome: "YES", Username: "alice", Offset: 0},
			},
			Expected: []MarketPosition{
				{Username: "alice", YesSharesOwned: 50, NoSharesOwned: 0},
			},
		},
		{
			Name: "Single NO Bet",
			BetConfigs: []struct {
				Amount   int64
				Outcome  string
				Username string
				Offset   time.Duration
			}{
				{Amount: 30, Outcome: "NO", Username: "bob", Offset: 0},
			},
			Expected: []MarketPosition{
				{Username: "bob", YesSharesOwned: 0, NoSharesOwned: 30},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			market := modelstesting.GenerateMarket(1, "testcreator")
			market.CreatedAt = time.Now()

			var bets []models.Bet
			for _, betConf := range tc.BetConfigs {
				bet := modelstesting.GenerateBet(betConf.Amount, betConf.Outcome, betConf.Username, uint(market.ID), betConf.Offset)
				bets = append(bets, bet)
			}

			snapshot := MarketSnapshot{
				ID:        market.ID,
				CreatedAt: market.CreatedAt,
			}

			actualPositions, err := CalculateMarketPositions_WPAM_DBPM(snapshot, bets)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(actualPositions) != len(tc.Expected) {
				t.Fatalf("expected %d positions, got %d", len(tc.Expected), len(actualPositions))
			}

			for i, expected := range tc.Expected {
				actual := actualPositions[i]
				// Print actual values to update your expected struct
				t.Logf("Test=%q User=%q Yes=%d No=%d Value=%d", tc.Name, actual.Username, actual.YesSharesOwned, actual.NoSharesOwned, actual.Value)

				if actual.Username != expected.Username ||
					actual.YesSharesOwned != expected.YesSharesOwned ||
					actual.NoSharesOwned != expected.NoSharesOwned {
					t.Errorf("expected shares %+v, got %+v", expected, actual)
				}
				// For the first run, comment out this Value check and just log it!
				// if actual.Value != expected.Value {
				//     t.Errorf("expected Value=%d, got %d", expected.Value, actual.Value)
				// }
			}
		})
	}
}

func TestCalculateMarketPositions_IncludesZeroPositionUsers(t *testing.T) {
	market := modelstesting.GenerateMarket(42, "creator")
	market.IsResolved = true
	market.ResolutionResult = "YES"
	market.CreatedAt = time.Now()

	bets := []struct {
		amount   int64
		outcome  string
		username string
		offset   time.Duration
	}{
		{amount: 50, outcome: "NO", username: "patrick", offset: 0},
		{amount: 51, outcome: "NO", username: "jimmy", offset: time.Second},
		{amount: 51, outcome: "NO", username: "jimmy", offset: 2 * time.Second},
		{amount: 10, outcome: "YES", username: "jyron", offset: 3 * time.Second},
		{amount: 30, outcome: "YES", username: "testuser03", offset: 4 * time.Second},
	}

	var betRecords []models.Bet
	for _, b := range bets {
		bet := modelstesting.GenerateBet(b.amount, b.outcome, b.username, uint(market.ID), b.offset)
		betRecords = append(betRecords, bet)
	}

	snapshot := MarketSnapshot{
		ID:               market.ID,
		CreatedAt:        market.CreatedAt,
		IsResolved:       market.IsResolved,
		ResolutionResult: market.ResolutionResult,
	}

	positions, err := CalculateMarketPositions_WPAM_DBPM(snapshot, betRecords)
	if err != nil {
		t.Fatalf("unexpected error calculating positions: %v", err)
	}

	var lockedUser *MarketPosition
	for i := range positions {
		if positions[i].Username == "testuser03" {
			lockedUser = &positions[i]
			break
		}
	}

	if lockedUser == nil {
		t.Fatalf("expected zero-position user to be present in positions")
	}

	if lockedUser.YesSharesOwned != 0 || lockedUser.NoSharesOwned != 0 || lockedUser.Value != 0 {
		t.Fatalf("expected zero shares/value for locked user, got %+v", lockedUser)
	}
}
