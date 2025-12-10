package analytics

import (
	"context"
	"testing"
	"time"

	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
)

func TestAggregateLeaderboardUserStats(t *testing.T) {
	markets := []leaderboardMarketData{
		{
			positions: []positionsmath.MarketPosition{
				{Username: "u1", Value: 200, TotalSpent: 100, IsResolved: true},
				{Username: "u2", Value: 50, TotalSpent: 80, IsResolved: false},
			},
		},
		{
			positions: []positionsmath.MarketPosition{
				{Username: "u1", Value: 120, TotalSpent: 60, IsResolved: false},
			},
		},
	}

	aggregates := aggregateLeaderboardUserStats(markets)

	if got := aggregates["u1"].totalProfit; got != 160 { // (200-100)+(120-60)
		t.Fatalf("u1 totalProfit = %d, want 160", got)
	}
	if got := aggregates["u1"].resolvedMarkets; got != 1 {
		t.Fatalf("u1 resolvedMarkets = %d, want 1", got)
	}
	if got := aggregates["u2"].activeMarkets; got != 1 {
		t.Fatalf("u2 activeMarkets = %d, want 1", got)
	}
}

func TestFindEarliestBetsPerUser(t *testing.T) {
	now := time.Now()
	markets := []leaderboardMarketData{
		{
			bets: []models.Bet{
				{Username: "u1", PlacedAt: now.Add(2 * time.Hour)},
				{Username: "u1", PlacedAt: now.Add(-time.Hour)},
				{Username: "u2", PlacedAt: now.Add(30 * time.Minute)},
			},
		},
	}

	aggregates := map[string]*leaderboardAggregate{
		"u1": {},
		"u2": {},
	}

	earliest := findEarliestBetsPerUser(markets, aggregates)

	if got := earliest["u1"]; !got.Equal(now.Add(-time.Hour)) {
		t.Fatalf("u1 earliest = %v, want %v", got, now.Add(-time.Hour))
	}
	if got := earliest["u2"]; !got.Equal(now.Add(30 * time.Minute)) {
		t.Fatalf("u2 earliest = %v, want %v", got, now.Add(30*time.Minute))
	}
}

func TestRankLeaderboardEntries_TieBreaksByEarliestBet(t *testing.T) {
	now := time.Now()
	entries := []GlobalUserProfitability{
		{Username: "late", TotalProfit: 100, EarliestBet: now.Add(time.Hour)},
		{Username: "early", TotalProfit: 100, EarliestBet: now},
	}

	ranked := rankLeaderboardEntries(entries)

	if ranked[0].Username != "early" || ranked[0].Rank != 1 {
		t.Fatalf("expected early user ranked first, got %+v", ranked[0])
	}
	if ranked[1].Rank != 2 {
		t.Fatalf("expected second rank to be 2, got %d", ranked[1].Rank)
	}
}

func TestComputeGlobalLeaderboard_OrdersByProfit(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()

	users := []models.User{
		modelstesting.GenerateUser("alice", 0),
		modelstesting.GenerateUser("bob", 0),
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	market := modelstesting.GenerateMarket(1, "alice")
	market.IsResolved = true
	market.ResolutionResult = "YES"
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(100, "YES", "alice", uint(market.ID), 0),
		modelstesting.GenerateBet(100, "NO", "bob", uint(market.ID), time.Minute),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	svc := NewService(NewGormRepository(db), func() *setup.EconomicConfig { return econ })

	results, err := svc.ComputeGlobalLeaderboard(context.Background())
	if err != nil {
		t.Fatalf("ComputeGlobalLeaderboard returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 leaderboard entries, got %d", len(results))
	}
	if results[0].Username != "alice" {
		t.Fatalf("expected alice to rank first, got %s", results[0].Username)
	}
	if results[0].Rank != 1 || results[1].Rank != 2 {
		t.Fatalf("expected ranks 1 and 2, got %d and %d", results[0].Rank, results[1].Rank)
	}
}
