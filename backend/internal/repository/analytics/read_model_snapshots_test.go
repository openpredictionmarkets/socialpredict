package analytics

import (
	"context"
	"testing"
	"time"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestServiceRefreshSystemMetricsSnapshotMatchesRawRecomputation(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()
	service := newAnalyticsService(t, db, econ)
	ctx := context.Background()

	raw, err := service.ComputeSystemMetrics(ctx)
	if err != nil {
		t.Fatalf("ComputeSystemMetrics returned error: %v", err)
	}
	readModel, err := service.RefreshSystemMetricsSnapshot(ctx)
	if err != nil {
		t.Fatalf("RefreshSystemMetricsSnapshot returned error: %v", err)
	}
	if readModel.Metrics != *raw {
		t.Fatalf("snapshot metrics mismatch:\ngot  %+v\nwant %+v", readModel.Metrics, *raw)
	}
	if readModel.Freshness.TransactionSafeRead {
		t.Fatalf("system metrics read model must not be transaction safe")
	}

	stored, err := service.GetSystemMetricsReadModel(ctx)
	if err != nil {
		t.Fatalf("GetSystemMetricsReadModel returned error: %v", err)
	}
	if stored == nil || stored.Metrics != *raw {
		t.Fatalf("stored metrics mismatch: %+v want %+v", stored, raw)
	}
}

func TestServiceGlobalLeaderboardSnapshotPagesStoredReadModel(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)

	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}
	market := modelstesting.GenerateMarket(7070, creator.Username)
	market.CreatedAt = now.Add(-24 * time.Hour)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}
	for _, username := range []string{"alice", "bob", "carol"} {
		user := modelstesting.GenerateUser(username, 500)
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", username, err)
		}
	}
	bets := []models.Bet{
		modelstesting.GenerateBet(100, "YES", "alice", uint(market.ID), time.Minute),
		modelstesting.GenerateBet(75, "NO", "bob", uint(market.ID), 2*time.Minute),
		modelstesting.GenerateBet(50, "YES", "carol", uint(market.ID), 3*time.Minute),
	}
	for i := range bets {
		if err := db.Create(&bets[i]).Error; err != nil {
			t.Fatalf("create bet %d: %v", i, err)
		}
	}

	service := newAnalyticsService(t, db, econ)
	raw, err := service.ComputeGlobalLeaderboardSnapshot(context.Background())
	if err != nil {
		t.Fatalf("ComputeGlobalLeaderboardSnapshot returned error: %v", err)
	}
	if len(raw.Result()) < 2 {
		t.Fatalf("expected at least 2 raw leaderboard entries, got %d", len(raw.Result()))
	}
	if _, err := service.RefreshGlobalLeaderboardSnapshot(context.Background()); err != nil {
		t.Fatalf("RefreshGlobalLeaderboardSnapshot returned error: %v", err)
	}

	page, err := service.GetGlobalLeaderboardReadModel(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("GetGlobalLeaderboardReadModel returned error: %v", err)
	}
	if page == nil || len(page.Entries) != 1 {
		t.Fatalf("expected one paged entry, got %+v", page)
	}
	got := page.Entries[0]
	want := raw.Result()[1]
	if got.Username != want.Username ||
		got.TotalProfit != want.TotalProfit ||
		got.TotalCurrentValue != want.TotalCurrentValue ||
		got.TotalSpent != want.TotalSpent ||
		got.ActiveMarkets != want.ActiveMarkets ||
		got.ResolvedMarkets != want.ResolvedMarkets ||
		got.Rank != want.Rank ||
		!got.EarliestBet.Equal(want.EarliestBet) {
		t.Fatalf("paged entry mismatch:\ngot  %+v\nwant %+v", got, want)
	}
}
