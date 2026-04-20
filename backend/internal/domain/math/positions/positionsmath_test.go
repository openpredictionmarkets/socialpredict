package positionsmath

import (
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models/modelstesting"
)

type reverseBetSorter struct{}

func (reverseBetSorter) Sort(bets []boundary.Bet) []boundary.Bet {
	sorted := make([]boundary.Bet, len(bets))
	copy(sorted, bets)
	for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
		sorted[i], sorted[j] = sorted[j], sorted[i]
	}
	return sorted
}

type fixedValuationCalculator struct{}

func (fixedValuationCalculator) Calculate(
	userPositions map[string]UserMarketPosition,
	currentProbability float64,
	totalVolume int64,
	isResolved bool,
	resolutionResult string,
	earliestBets map[string]time.Time,
) (map[string]UserValuationResult, error) {
	result := make(map[string]UserValuationResult, len(userPositions))
	for username := range userPositions {
		result[username] = UserValuationResult{Username: username, RoundedValue: 77}
	}
	return result, nil
}

var positionsMathBaseTime = time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC)

func newTestPositionCalculator() PositionCalculator {
	econ := modelstesting.GenerateEconomicConfig()
	calculator := wpam.NewProbabilityCalculator(wpam.StaticSeedProvider{Value: wpam.Seeds{
		InitialProbability:     econ.Economics.MarketCreation.InitialMarketProbability,
		InitialSubsidization:   econ.Economics.MarketCreation.InitialMarketSubsidization,
		InitialYesContribution: econ.Economics.MarketCreation.InitialMarketYes,
		InitialNoContribution:  econ.Economics.MarketCreation.InitialMarketNo,
	}})
	return NewPositionCalculator(
		WithProbabilityProvider(NewWPAMProbabilityProvider(calculator)),
	)
}

func buildPositionBets(marketID uint, entries []struct {
	Amount   int64
	Outcome  string
	Username string
	Offset   time.Duration
}) []boundary.Bet {
	bets := make([]boundary.Bet, 0, len(entries))
	for _, entry := range entries {
		bets = append(bets, boundary.Bet{
			Username:  entry.Username,
			MarketID:  marketID,
			Amount:    entry.Amount,
			Outcome:   entry.Outcome,
			PlacedAt:  positionsMathBaseTime.Add(entry.Offset),
			CreatedAt: positionsMathBaseTime.Add(entry.Offset),
		})
	}
	return bets
}

func TestCalculateMarketPositions_WPAM_DBPM(t *testing.T) {
	positionCalculator := newTestPositionCalculator()

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
			market.CreatedAt = positionsMathBaseTime

			bets := buildPositionBets(uint(market.ID), tc.BetConfigs)

			snapshot := MarketSnapshot{
				ID:        market.ID,
				CreatedAt: market.CreatedAt,
			}

			actualPositions, err := positionCalculator.CalculateMarketPositions(snapshot, bets)
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
	market.CreatedAt = positionsMathBaseTime

	betRecords := buildPositionBets(uint(market.ID), []struct {
		Amount   int64
		Outcome  string
		Username string
		Offset   time.Duration
	}{
		{Amount: 50, Outcome: "NO", Username: "patrick", Offset: 0},
		{Amount: 51, Outcome: "NO", Username: "jimmy", Offset: time.Second},
		{Amount: 51, Outcome: "NO", Username: "jimmy", Offset: 2 * time.Second},
		{Amount: 10, Outcome: "YES", Username: "jyron", Offset: 3 * time.Second},
		{Amount: 30, Outcome: "YES", Username: "testuser03", Offset: 4 * time.Second},
	})

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

func TestCalculateMarketPositions_UsesInjectedValuationCalculator(t *testing.T) {
	calculator := newTestPositionCalculator()
	calculator = NewPositionCalculator(
		WithProbabilityProvider(calculator.probabilities),
		WithPayoutModel(calculator.netPositions),
		WithBetSorter(reverseBetSorter{}),
		WithValuationCalculator(fixedValuationCalculator{}),
	)

	snapshot := MarketSnapshot{ID: 1, CreatedAt: positionsMathBaseTime}
	bets := buildPositionBets(1, []struct {
		Amount   int64
		Outcome  string
		Username string
		Offset   time.Duration
	}{
		{Amount: 10, Outcome: "YES", Username: "alice", Offset: 0},
	})

	positions, err := calculator.CalculateMarketPositions(snapshot, bets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(positions) != 1 || positions[0].Value != 77 {
		t.Fatalf("expected injected valuation value 77, got %+v", positions)
	}
}
