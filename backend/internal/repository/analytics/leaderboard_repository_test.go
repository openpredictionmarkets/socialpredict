package analytics

import (
	"context"
	"testing"
	"time"

	danalytics "socialpredict/internal/domain/analytics"
	"socialpredict/internal/domain/boundary"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

type globalLeaderboardComputer interface {
	ComputeGlobalLeaderboard(context.Context) ([]danalytics.GlobalUserProfitability, error)
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

func requireGlobalLeaderboard(t *testing.T, svc globalLeaderboardComputer) []danalytics.GlobalUserProfitability {
	t.Helper()

	results, err := svc.ComputeGlobalLeaderboard(context.Background())
	if err != nil {
		t.Fatalf("ComputeGlobalLeaderboard returned error: %v", err)
	}

	return results
}

func requireLeaderboardOrder(t *testing.T, entries []danalytics.GlobalUserProfitability, usernames ...string) {
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
