package markets_test

import (
	"context"
	"errors"
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

func TestServiceGetMarketDiscoverySummariesDeduplicatesAndPreservesHydratedMarkets(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	baseRepo := rmarkets.NewGormRepository(db)
	repo := &recordingDiscoveryEnrichmentRepo{
		GormRepository: baseRepo,
		enrichment: &markets.MarketDiscoveryEnrichment{
			AccountingByMarketID: map[int64]markets.MarketAccountingSnapshot{
				8101: {MarketID: 8101, VolumeWithDust: 101},
				8102: {MarketID: 8102, VolumeWithDust: 202},
			},
			CreatorsByUsername: map[string]markets.CreatorSummary{
				"alice": {Username: "alice", DisplayName: "Alice"},
				"bob":   {Username: "bob", DisplayName: "Bob"},
			},
		},
	}
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(marketsTestTime()), markets.Config{})
	marketA := &markets.Market{ID: 8101, CreatorUsername: "alice", QuestionTitle: "A"}
	marketB := &markets.Market{ID: 8102, CreatorUsername: "bob", QuestionTitle: "B"}

	got, err := service.GetMarketDiscoverySummaries(context.Background(), []*markets.Market{marketA, marketB, marketA, nil})
	if err != nil {
		t.Fatalf("GetMarketDiscoverySummaries returned error: %v", err)
	}
	if len(repo.marketIDs) != 2 || repo.marketIDs[0] != marketA.ID || repo.marketIDs[1] != marketB.ID {
		t.Fatalf("market IDs = %#v, want [%d %d]", repo.marketIDs, marketA.ID, marketB.ID)
	}
	if len(repo.creatorUsernames) != 2 || repo.creatorUsernames[0] != "alice" || repo.creatorUsernames[1] != "bob" {
		t.Fatalf("creator usernames = %#v", repo.creatorUsernames)
	}
	if got[marketA.ID].Market != marketA {
		t.Fatalf("service replaced hydrated market")
	}
	if got[marketA.ID].Creator.DisplayName != "Alice" {
		t.Fatalf("creator = %#v", got[marketA.ID].Creator)
	}
}

func TestServiceGetMarketDiscoverySummariesRefreshesMissingSnapshotOnce(t *testing.T) {
	service, db, _ := setupServiceWithDB(t)
	ctx := context.Background()
	creator := modelstesting.GenerateUser("batch_missing_creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}
	dbMarket := modelstesting.GenerateMarket(8103, creator.Username)
	if err := db.Create(&dbMarket).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}
	bet := modelstesting.GenerateBet(75, "YES", "alice", uint(dbMarket.ID), time.Minute)
	if err := db.Create(&bet).Error; err != nil {
		t.Fatalf("create bet: %v", err)
	}
	market := &markets.Market{ID: dbMarket.ID, CreatorUsername: dbMarket.CreatorUsername}

	first, err := service.GetMarketDiscoverySummaries(ctx, []*markets.Market{market})
	if err != nil {
		t.Fatalf("first discovery summaries: %v", err)
	}
	second, err := service.GetMarketDiscoverySummaries(ctx, []*markets.Market{market})
	if err != nil {
		t.Fatalf("second discovery summaries: %v", err)
	}
	if first[market.ID].Accounting.BetCount != 1 || first[market.ID].Accounting.VolumeWithDust != 75 {
		t.Fatalf("first accounting = %#v", first[market.ID].Accounting)
	}
	if !second[market.ID].Accounting.GeneratedAt.Equal(first[market.ID].Accounting.GeneratedAt) {
		t.Fatalf("snapshot refreshed twice: first=%s second=%s", first[market.ID].Accounting.GeneratedAt, second[market.ID].Accounting.GeneratedAt)
	}
}

func TestServiceGetMarketDiscoverySummariesPropagatesRepositoryError(t *testing.T) {
	errBatch := errors.New("batch failed")
	repo := &recordingDiscoveryEnrichmentRepo{
		GormRepository: rmarkets.NewGormRepository(modelstesting.NewFakeDB(t)),
		err:            errBatch,
	}
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(marketsTestTime()), markets.Config{})
	_, err := service.GetMarketDiscoverySummaries(context.Background(), []*markets.Market{{ID: 8104, CreatorUsername: "alice"}})
	if !errors.Is(err, errBatch) {
		t.Fatalf("error = %v, want %v", err, errBatch)
	}
}

type recordingDiscoveryEnrichmentRepo struct {
	*rmarkets.GormRepository
	marketIDs        []int64
	creatorUsernames []string
	enrichment       *markets.MarketDiscoveryEnrichment
	err              error
}

func (r *recordingDiscoveryEnrichmentRepo) GetMarketDiscoveryEnrichment(
	_ context.Context,
	marketIDs []int64,
	creatorUsernames []string,
) (*markets.MarketDiscoveryEnrichment, error) {
	r.marketIDs = append([]int64(nil), marketIDs...)
	r.creatorUsernames = append([]string(nil), creatorUsernames...)
	return r.enrichment, r.err
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
