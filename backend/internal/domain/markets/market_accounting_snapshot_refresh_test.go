package markets_test

import (
	"context"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	rmarkets "socialpredict/internal/repository/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestServiceRefreshMarketAccountingSnapshotPersistsRawRecomputedSnapshot(t *testing.T) {
	service, db, _ := setupServiceWithDB(t)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("snapshot_creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(9091, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(150, "YES", "alice", uint(market.ID), 1*time.Minute),
		modelstesting.GenerateBet(90, "NO", "bob", uint(market.ID), 2*time.Minute),
		modelstesting.GenerateBet(-40, "YES", "alice", uint(market.ID), 3*time.Minute),
	}
	for i := range bets {
		if err := db.Create(&bets[i]).Error; err != nil {
			t.Fatalf("create bet %d: %v", i, err)
		}
	}

	snapshot, err := service.RefreshMarketAccountingSnapshot(ctx, market.ID)
	if err != nil {
		t.Fatalf("RefreshMarketAccountingSnapshot returned error: %v", err)
	}
	if snapshot.MarketID != market.ID {
		t.Fatalf("market id = %d, want %d", snapshot.MarketID, market.ID)
	}
	if snapshot.NetBetVolume != 200 || snapshot.MarketDust != 1 || snapshot.VolumeWithDust != 201 {
		t.Fatalf("unexpected snapshot volume/dust: %+v", snapshot)
	}
	if snapshot.UserCount != 2 || snapshot.BetCount != 3 {
		t.Fatalf("unexpected snapshot counts: %+v", snapshot)
	}
	if snapshot.TransactionSafeRead {
		t.Fatalf("snapshot must not be transaction safe")
	}

	repo := rmarkets.NewGormRepository(db)
	stored, err := repo.GetMarketAccountingSnapshot(ctx, market.ID)
	if err != nil {
		t.Fatalf("get stored snapshot: %v", err)
	}
	if stored == nil {
		t.Fatalf("stored snapshot is nil")
	}
	if stored.NetBetVolume != snapshot.NetBetVolume ||
		stored.MarketDust != snapshot.MarketDust ||
		stored.VolumeWithDust != snapshot.VolumeWithDust ||
		stored.BetCount != snapshot.BetCount {
		t.Fatalf("stored snapshot mismatch:\ngot  %+v\nwant %+v", stored, snapshot)
	}
}

func TestServiceGetMarketSummaryReadModelReturnsStaleSnapshotWithoutRecomputing(t *testing.T) {
	service, db, _ := setupServiceWithDB(t)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("summary_snapshot_creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(9092, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	firstBet := modelstesting.GenerateBet(100, "YES", "alice", uint(market.ID), time.Minute)
	if err := db.Create(&firstBet).Error; err != nil {
		t.Fatalf("create first bet: %v", err)
	}

	if _, err := service.RefreshMarketAccountingSnapshot(ctx, market.ID); err != nil {
		t.Fatalf("refresh snapshot: %v", err)
	}

	repo := rmarkets.NewGormRepository(db)
	if err := repo.MarkMarketAccountingSnapshotStale(ctx, market.ID, "bet_accepted"); err != nil {
		t.Fatalf("mark snapshot stale: %v", err)
	}

	secondBet := modelstesting.GenerateBet(50, "NO", "bob", uint(market.ID), 2*time.Minute)
	if err := db.Create(&secondBet).Error; err != nil {
		t.Fatalf("create second bet: %v", err)
	}

	summary, err := service.GetMarketSummaryReadModel(ctx, market.ID)
	if err != nil {
		t.Fatalf("GetMarketSummaryReadModel returned error: %v", err)
	}
	if summary.Accounting.BetCount != 1 || summary.Accounting.UserCount != 1 || summary.Accounting.VolumeWithDust != 100 {
		t.Fatalf("summary recomputed instead of returning stale snapshot: %+v", summary.Accounting)
	}
	freshness := summary.Accounting.Freshness()
	if !freshness.IsStale || freshness.StaleReason != "bet_accepted" {
		t.Fatalf("expected stale freshness from stored snapshot, got %+v", freshness)
	}
}

func TestServiceGetMarketSummaryReadModelComputesMissingSnapshotOnce(t *testing.T) {
	service, db, _ := setupServiceWithDB(t)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("summary_missing_creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(9093, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}
	bet := modelstesting.GenerateBet(75, "YES", "alice", uint(market.ID), time.Minute)
	if err := db.Create(&bet).Error; err != nil {
		t.Fatalf("create bet: %v", err)
	}

	summary, err := service.GetMarketSummaryReadModel(ctx, market.ID)
	if err != nil {
		t.Fatalf("GetMarketSummaryReadModel returned error: %v", err)
	}
	if summary.Accounting.BetCount != 1 || summary.Accounting.VolumeWithDust != 75 {
		t.Fatalf("unexpected computed summary: %+v", summary.Accounting)
	}

	repo := rmarkets.NewGormRepository(db)
	stored, err := repo.GetMarketAccountingSnapshot(ctx, market.ID)
	if err != nil {
		t.Fatalf("get stored snapshot: %v", err)
	}
	if stored == nil {
		t.Fatalf("expected missing snapshot to be stored after first summary read")
	}
}

func TestServiceRefreshMarketAccountingSnapshotRequiresSnapshotRepository(t *testing.T) {
	service := markets.NewService(&snapshotlessMarketRepo{}, newNoopUserService(), newFixedClock(marketsTestTime()), markets.Config{})

	if _, err := service.RefreshMarketAccountingSnapshot(context.Background(), 1); !markets.IsInvalidState(err) {
		t.Fatalf("RefreshMarketAccountingSnapshot error = %v, want ErrInvalidState", err)
	}
}

type snapshotlessMarketRepo struct{}

func (snapshotlessMarketRepo) GetByID(context.Context, int64) (*markets.Market, error) {
	return &markets.Market{}, nil
}
func (snapshotlessMarketRepo) List(context.Context, markets.ListFilters) ([]*markets.Market, error) {
	return nil, nil
}
func (snapshotlessMarketRepo) ListByStatus(context.Context, string, markets.Page) ([]*markets.Market, error) {
	return nil, nil
}
func (snapshotlessMarketRepo) Search(context.Context, string, markets.SearchFilters) ([]*markets.Market, error) {
	return nil, nil
}
func (snapshotlessMarketRepo) GetPublicMarket(context.Context, int64) (*markets.PublicMarket, error) {
	return nil, nil
}
func (snapshotlessMarketRepo) Create(context.Context, *markets.Market) error { return nil }
func (snapshotlessMarketRepo) UpdateLabels(context.Context, int64, string, string) error {
	return nil
}
func (snapshotlessMarketRepo) Delete(context.Context, int64) error { return nil }
func (snapshotlessMarketRepo) ResolveMarket(context.Context, int64, string) error {
	return nil
}
func (snapshotlessMarketRepo) GetUserPosition(context.Context, int64, string) (*markets.UserPosition, error) {
	return nil, nil
}
func (snapshotlessMarketRepo) ListMarketPositions(context.Context, int64) (markets.MarketPositions, error) {
	return nil, nil
}
func (snapshotlessMarketRepo) CalculatePayoutPositions(context.Context, int64) ([]*markets.PayoutPosition, error) {
	return nil, nil
}
func (snapshotlessMarketRepo) ListBetsForMarket(context.Context, int64) ([]*markets.Bet, error) {
	return nil, nil
}
