package markets_test

import (
	"context"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestServiceRefreshMarketPositionsSnapshotMatchesRawRecomputation(t *testing.T) {
	service, db, _ := setupServiceWithDB(t)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("positions_snapshot_creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}
	market := modelstesting.GenerateMarket(8083, creator.Username)
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

	raw, err := service.GetMarketPositionsPage(ctx, market.ID, markets.Page{Limit: 1000})
	if err != nil {
		t.Fatalf("GetMarketPositionsPage returned error: %v", err)
	}
	snapshot, err := service.RefreshMarketPositionsSnapshot(ctx, market.ID)
	if err != nil {
		t.Fatalf("RefreshMarketPositionsSnapshot returned error: %v", err)
	}
	if len(snapshot.Positions) != len(raw) {
		t.Fatalf("snapshot position count = %d, want %d", len(snapshot.Positions), len(raw))
	}
	if snapshot.Freshness().TransactionSafeRead {
		t.Fatalf("market positions snapshot must not be transaction safe")
	}

	page, err := service.GetMarketPositionsReadModel(ctx, market.ID, markets.Page{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("GetMarketPositionsReadModel returned error: %v", err)
	}
	if page == nil || len(page.Positions) != 1 {
		t.Fatalf("expected one paged position, got %+v", page)
	}
	if page.Positions[0].Username != raw[1].Username || page.Positions[0].YesSharesOwned != raw[1].YesSharesOwned {
		t.Fatalf("paged position mismatch:\ngot  %+v\nwant %+v", page.Positions[0], raw[1])
	}

	if err := service.MarkMarketPositionsSnapshotStale(ctx, market.ID, "bet_accepted"); err != nil {
		t.Fatalf("MarkMarketPositionsSnapshotStale returned error: %v", err)
	}
	stale, err := service.GetMarketPositionsReadModel(ctx, market.ID, markets.Page{Limit: 20})
	if err != nil {
		t.Fatalf("GetMarketPositionsReadModel stale returned error: %v", err)
	}
	if stale == nil || !stale.Freshness().IsStale {
		t.Fatalf("expected stale freshness, got %+v", stale)
	}
}
