package markets_test

import (
	"context"
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
	markets "socialpredict/internal/domain/markets"
	marketmath "socialpredict/internal/domain/math/market"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models/modelstesting"
)

type marketDetailsFixture struct {
	service    *markets.Service
	bets       []boundary.Bet
	market     *markets.Market
	calculator wpam.ProbabilityCalculator
}

func seedMarketDetailsFixture(t *testing.T) marketDetailsFixture {
	t.Helper()

	service, db, calculator := setupServiceWithDB(t)

	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(3001, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}
	if err := db.First(&market, market.ID).Error; err != nil {
		t.Fatalf("reload market: %v", err)
	}

	bets := []boundary.Bet{
		{Username: "alice", MarketID: uint(market.ID), Amount: 150, Outcome: "YES", PlacedAt: market.CreatedAt.Add(1 * time.Minute), CreatedAt: market.CreatedAt.Add(1 * time.Minute)},
		{Username: "bob", MarketID: uint(market.ID), Amount: 90, Outcome: "NO", PlacedAt: market.CreatedAt.Add(2 * time.Minute), CreatedAt: market.CreatedAt.Add(2 * time.Minute)},
		{Username: "alice", MarketID: uint(market.ID), Amount: -40, Outcome: "YES", PlacedAt: market.CreatedAt.Add(3 * time.Minute), CreatedAt: market.CreatedAt.Add(3 * time.Minute)},
	}

	for i, bet := range bets {
		dbBet := modelstesting.GenerateBet(bet.Amount, bet.Outcome, bet.Username, bet.MarketID, 0)
		dbBet.PlacedAt = bet.PlacedAt
		dbBet.CreatedAt = bet.CreatedAt
		if err := db.Create(&dbBet).Error; err != nil {
			t.Fatalf("create bet %d: %v", i, err)
		}
	}

	return marketDetailsFixture{
		service: service,
		bets:    bets,
		market: &markets.Market{
			ID:                 market.ID,
			QuestionTitle:      market.QuestionTitle,
			Description:        market.Description,
			OutcomeType:        market.OutcomeType,
			ResolutionDateTime: market.ResolutionDateTime,
			CreatorUsername:    market.CreatorUsername,
			CreatedAt:          market.CreatedAt,
			InitialProbability: market.InitialProbability,
		},
		calculator: calculator,
	}
}

func assertMarketDetailMetrics(t *testing.T, overview *markets.MarketOverview, market *markets.Market, bets []boundary.Bet, calculator wpam.ProbabilityCalculator) {
	t.Helper()

	expectedVolume := marketmath.GetMarketVolumeWithDust(bets)
	expectedDust := marketmath.GetMarketDust(bets)
	expectedProbabilities := calculator.CalculateMarketProbabilitiesWPAM(market.CreatedAt, bets)

	if overview.TotalVolume != expectedVolume {
		t.Fatalf("total volume = %d, want %d", overview.TotalVolume, expectedVolume)
	}
	if overview.MarketDust != expectedDust {
		t.Fatalf("market dust = %d, want %d", overview.MarketDust, expectedDust)
	}
	if overview.NumUsers != 2 {
		t.Fatalf("num users = %d, want 2", overview.NumUsers)
	}
	if overview.Creator == nil || overview.Creator.Username != market.CreatorUsername {
		t.Fatalf("creator username mismatch: got %+v want %s", overview.Creator, market.CreatorUsername)
	}
	if len(overview.ProbabilityChanges) != len(expectedProbabilities) {
		t.Fatalf("probability history length = %d, want %d", len(overview.ProbabilityChanges), len(expectedProbabilities))
	}
	if overview.LastProbability != expectedProbabilities[len(expectedProbabilities)-1].Probability {
		t.Fatalf("last probability = %f, want %f", overview.LastProbability, expectedProbabilities[len(expectedProbabilities)-1].Probability)
	}
}

func TestServiceGetMarketDetailsCalculatesMetrics(t *testing.T) {
	fixture := seedMarketDetailsFixture(t)

	overview, err := fixture.service.GetMarketDetails(context.Background(), fixture.market.ID)
	if err != nil {
		t.Fatalf("GetMarketDetails returned error: %v", err)
	}

	assertMarketDetailMetrics(t, overview, fixture.market, fixture.bets, fixture.calculator)

	fallback := markets.NewService(nil, newNoopUserService(), nil, markets.Config{}, markets.WithMetricsCalculator(nil), markets.WithClock(nil))
	if fallback == nil {
		t.Fatalf("expected fallback service")
	}
}

func TestServiceGetMarketDetails_InvalidAndMissing(t *testing.T) {
	service, _, _ := setupServiceWithDB(t)

	_, err := service.GetMarketDetails(context.Background(), 0)
	requireInvalidInput(t, err)

	_, err = service.GetMarketDetails(context.Background(), 999)
	requireMarketNotFound(t, err)

	if err := (noopUserService{}).DeductBalance(context.Background(), "alice", 1); err == nil {
		t.Fatalf("expected zero-value user service to fail predictably")
	}
}
