package analytics

import (
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
	positionsmath "socialpredict/internal/domain/math/positions"
)

func leaderboardMarketDataFixture(positions []positionsmath.MarketPosition, bets []boundary.Bet) leaderboardMarketData {
	return leaderboardMarketData{
		positions: positions,
		bets:      bets,
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
