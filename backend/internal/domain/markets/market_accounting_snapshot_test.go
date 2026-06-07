package markets_test

import (
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
	markets "socialpredict/internal/domain/markets"
	marketmath "socialpredict/internal/domain/math/market"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models/modelstesting"
)

func TestMarketAccountingSnapshotCalculatorMatchesRawRecomputation(t *testing.T) {
	now := marketsTestTime()
	market := &markets.Market{
		ID:        77,
		CreatedAt: now.Add(-1 * time.Hour),
	}
	bets := []boundary.Bet{
		{ID: 11, Username: "alice", MarketID: uint(market.ID), Amount: 150, Outcome: "YES", PlacedAt: market.CreatedAt.Add(1 * time.Minute), CreatedAt: market.CreatedAt.Add(1 * time.Minute)},
		{ID: 12, Username: "bob", MarketID: uint(market.ID), Amount: 90, Outcome: "NO", PlacedAt: market.CreatedAt.Add(2 * time.Minute), CreatedAt: market.CreatedAt.Add(2 * time.Minute)},
		{ID: 13, Username: "alice", MarketID: uint(market.ID), Amount: -40, Outcome: "YES", PlacedAt: market.CreatedAt.Add(3 * time.Minute), CreatedAt: market.CreatedAt.Add(3 * time.Minute)},
	}
	calculator := testProbabilityCalculator()
	expectedProbabilities := calculator.CalculateMarketProbabilitiesWPAM(market.CreatedAt, bets)

	snapshot := markets.NewMarketAccountingSnapshotCalculator(
		markets.DefaultProbabilityEngine(calculator),
		nil,
		newFixedClock(now),
	).Calculate(market, bets)

	if snapshot.MarketID != market.ID {
		t.Fatalf("market id = %d, want %d", snapshot.MarketID, market.ID)
	}
	if !snapshot.GeneratedAt.Equal(now) {
		t.Fatalf("generated at = %s, want %s", snapshot.GeneratedAt, now)
	}
	if snapshot.NetBetVolume != marketmath.GetMarketVolume(bets) {
		t.Fatalf("net bet volume = %d, want %d", snapshot.NetBetVolume, marketmath.GetMarketVolume(bets))
	}
	if snapshot.MarketDust != marketmath.GetMarketDust(bets) {
		t.Fatalf("market dust = %d, want %d", snapshot.MarketDust, marketmath.GetMarketDust(bets))
	}
	if snapshot.VolumeWithDust != marketmath.GetMarketVolumeWithDust(bets) {
		t.Fatalf("volume with dust = %d, want %d", snapshot.VolumeWithDust, marketmath.GetMarketVolumeWithDust(bets))
	}
	if snapshot.UserCount != 2 {
		t.Fatalf("user count = %d, want 2", snapshot.UserCount)
	}
	if snapshot.BetCount != len(bets) {
		t.Fatalf("bet count = %d, want %d", snapshot.BetCount, len(bets))
	}
	if snapshot.LastProcessedBetID != 13 {
		t.Fatalf("last processed bet id = %d, want 13", snapshot.LastProcessedBetID)
	}
	if !snapshot.LastProcessedBetAt.Equal(bets[2].PlacedAt) {
		t.Fatalf("last processed bet at = %s, want %s", snapshot.LastProcessedBetAt, bets[2].PlacedAt)
	}
	if snapshot.Source != "read_model" {
		t.Fatalf("source = %q, want read_model", snapshot.Source)
	}
	if snapshot.TransactionSafeRead {
		t.Fatalf("snapshot must not be marked transaction safe")
	}
	if len(snapshot.ProbabilityChanges) != len(expectedProbabilities) {
		t.Fatalf("probability history length = %d, want %d", len(snapshot.ProbabilityChanges), len(expectedProbabilities))
	}
	if snapshot.LastProbability != expectedProbabilities[len(expectedProbabilities)-1].Probability {
		t.Fatalf("last probability = %f, want %f", snapshot.LastProbability, expectedProbabilities[len(expectedProbabilities)-1].Probability)
	}
}

func TestMarketAccountingSnapshotCalculatorUsesSafeDefaults(t *testing.T) {
	now := marketsTestTime()
	market := &markets.Market{ID: 88, CreatedAt: now}

	snapshot := markets.NewMarketAccountingSnapshotCalculator(nil, nil, newFixedClock(now)).
		Calculate(market, nil)

	if snapshot.MarketID != market.ID {
		t.Fatalf("market id = %d, want %d", snapshot.MarketID, market.ID)
	}
	if !snapshot.GeneratedAt.Equal(now) {
		t.Fatalf("generated at = %s, want %s", snapshot.GeneratedAt, now)
	}
	if snapshot.NetBetVolume != 0 || snapshot.MarketDust != 0 || snapshot.VolumeWithDust != 0 {
		t.Fatalf("empty snapshot volumes = net %d dust %d with dust %d, want all zero", snapshot.NetBetVolume, snapshot.MarketDust, snapshot.VolumeWithDust)
	}
	if snapshot.UserCount != 0 || snapshot.BetCount != 0 {
		t.Fatalf("empty snapshot counts = users %d bets %d, want zero", snapshot.UserCount, snapshot.BetCount)
	}
}

func testProbabilityCalculator() wpam.ProbabilityCalculator {
	econ := modelstesting.GenerateEconomicConfig()
	return wpam.NewProbabilityCalculator(wpam.StaticSeedProvider{Value: wpam.Seeds{
		InitialProbability:     econ.Economics.MarketCreation.InitialMarketProbability,
		InitialSubsidization:   econ.Economics.MarketCreation.InitialMarketSubsidization,
		InitialYesContribution: econ.Economics.MarketCreation.InitialMarketYes,
		InitialNoContribution:  econ.Economics.MarketCreation.InitialMarketNo,
	}})
}
