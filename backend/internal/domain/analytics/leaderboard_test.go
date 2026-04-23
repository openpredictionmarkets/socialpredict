package analytics

import (
	"context"
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

type globalLeaderboardComputer interface {
	ComputeGlobalLeaderboard(context.Context) ([]GlobalUserProfitability, error)
}

func leaderboardMarketDataFixture(positions []positionsmath.MarketPosition, bets []boundary.Bet) leaderboardMarketData {
	return leaderboardMarketData{
		positions: positions,
		bets:      bets,
	}
}

func boundaryBetFromModel(bet models.Bet) boundary.Bet {
	return boundary.Bet{
		ID:        uint(bet.ID),
		Username:  bet.Username,
		MarketID:  bet.MarketID,
		Amount:    bet.Amount,
		PlacedAt:  bet.PlacedAt,
		Outcome:   bet.Outcome,
		CreatedAt: bet.CreatedAt,
	}
}

func requireLeaderboardOrder(t *testing.T, entries []GlobalUserProfitability, usernames ...string) {
	t.Helper()

	if len(entries) < len(usernames) {
		t.Fatalf("expected at least %d entries, got %d", len(usernames), len(entries))
	}
	for i, username := range usernames {
		if entries[i].Username != username {
			t.Fatalf("entry %d username = %s, want %s", i, entries[i].Username, username)
		}
	}
}

func requireGlobalLeaderboard(t *testing.T, svc globalLeaderboardComputer) []GlobalUserProfitability {
	t.Helper()

	results, err := svc.ComputeGlobalLeaderboard(context.Background())
	if err != nil {
		t.Fatalf("ComputeGlobalLeaderboard returned error: %v", err)
	}

	return results
}

func TestAggregateLeaderboardUserStats(t *testing.T) {
	markets := []leaderboardMarketData{
		leaderboardMarketDataFixture(
			[]positionsmath.MarketPosition{
				{Username: "u1", Value: 200, TotalSpent: 100, IsResolved: true},
				{Username: "u2", Value: 50, TotalSpent: 80, IsResolved: false},
			},
			nil,
		),
		leaderboardMarketDataFixture(
			[]positionsmath.MarketPosition{
				{Username: "u1", Value: 120, TotalSpent: 60, IsResolved: false},
			},
			nil,
		),
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
		leaderboardMarketDataFixture(
			nil,
			[]boundary.Bet{
				{Username: "u1", PlacedAt: now.Add(2 * time.Hour)},
				{Username: "u1", PlacedAt: now.Add(-time.Hour)},
				{Username: "u2", PlacedAt: now.Add(30 * time.Minute)},
			},
		),
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

	requireLeaderboardOrder(t, ranked, "early", "late")
	if ranked[0].Rank != 1 {
		t.Fatalf("expected early user rank 1, got %d", ranked[0].Rank)
	}
	if ranked[1].Rank != 2 {
		t.Fatalf("expected second rank to be 2, got %d", ranked[1].Rank)
	}
}

func TestComputeGlobalLeaderboard_OrdersByProfit(t *testing.T) {
	_ = modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

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

	bets := []boundary.Bet{
		boundaryBetFromModel(modelstesting.GenerateBet(100, "YES", "alice", uint(market.ID), 0)),
		boundaryBetFromModel(modelstesting.GenerateBet(100, "NO", "bob", uint(market.ID), time.Minute)),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	svc := newAnalyticsService(t, db, econ)
	results := requireGlobalLeaderboard(t, svc)
	if len(results) != 2 {
		t.Fatalf("expected 2 leaderboard entries, got %d", len(results))
	}
	requireLeaderboardOrder(t, results, "alice", "bob")
	if results[0].Rank != 1 || results[1].Rank != 2 {
		t.Fatalf("expected ranks 1 and 2, got %d and %d", results[0].Rank, results[1].Rank)
	}
}
