package markets_test

import (
	"context"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestServiceRefreshMarketLeaderboardSnapshotMatchesRawRecomputation(t *testing.T) {
	service, db, _ := setupServiceWithDB(t)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("leaderboard_snapshot_creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}
	market := modelstesting.GenerateMarket(8082, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}
	bets := []models.Bet{
		modelstesting.GenerateBet(100, "YES", "alice", uint(market.ID), time.Minute),
		modelstesting.GenerateBet(70, "NO", "bob", uint(market.ID), 2*time.Minute),
		modelstesting.GenerateBet(40, "YES", "carol", uint(market.ID), 3*time.Minute),
	}
	for i := range bets {
		if err := db.Create(&bets[i]).Error; err != nil {
			t.Fatalf("create bet %d: %v", i, err)
		}
	}

	raw, err := service.GetMarketLeaderboard(ctx, market.ID, markets.Page{Limit: 1000})
	if err != nil {
		t.Fatalf("GetMarketLeaderboard returned error: %v", err)
	}
	snapshot, err := service.RefreshMarketLeaderboardSnapshot(ctx, market.ID)
	if err != nil {
		t.Fatalf("RefreshMarketLeaderboardSnapshot returned error: %v", err)
	}
	if len(snapshot.Rows) != len(raw) {
		t.Fatalf("snapshot row count = %d, want %d", len(snapshot.Rows), len(raw))
	}
	if snapshot.Freshness().TransactionSafeRead {
		t.Fatalf("market leaderboard snapshot must not be transaction safe")
	}

	page, err := service.GetMarketLeaderboardReadModel(ctx, market.ID, markets.Page{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("GetMarketLeaderboardReadModel returned error: %v", err)
	}
	if page == nil || len(page.Rows) != 1 {
		t.Fatalf("expected one paged row, got %+v", page)
	}
	if page.Rows[0].Username != raw[1].Username || page.Rows[0].Rank != raw[1].Rank {
		t.Fatalf("paged row mismatch:\ngot  %+v\nwant %+v", page.Rows[0], raw[1])
	}

	if err := service.MarkMarketLeaderboardSnapshotStale(ctx, market.ID, "bet_accepted"); err != nil {
		t.Fatalf("MarkMarketLeaderboardSnapshotStale returned error: %v", err)
	}
	stale, err := service.GetMarketLeaderboardReadModel(ctx, market.ID, markets.Page{Limit: 20})
	if err != nil {
		t.Fatalf("GetMarketLeaderboardReadModel stale returned error: %v", err)
	}
	if stale == nil || !stale.Freshness().IsStale {
		t.Fatalf("expected stale freshness, got %+v", stale)
	}
}
