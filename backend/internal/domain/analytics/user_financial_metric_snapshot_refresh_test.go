package analytics

import (
	"context"
	"testing"
	"time"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestServiceRefreshUserFinancialMetricSnapshotPersistsRawRecomputedSnapshot(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()
	user := modelstesting.GenerateUser("financial_refresh_user", 500)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	market := modelstesting.GenerateMarket(8081, user.Username)
	market.IsResolved = false
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(100, "YES", user.Username, uint(market.ID), 1*time.Minute),
		modelstesting.GenerateBet(50, "NO", user.Username, uint(market.ID), 2*time.Minute),
	}
	for i := range bets {
		if err := db.Create(&bets[i]).Error; err != nil {
			t.Fatalf("create bet %d: %v", i, err)
		}
	}

	service := newAnalyticsService(t, db, econ)
	generatedAt := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	snapshot, err := service.RefreshUserFinancialMetricSnapshot(context.Background(), FinancialSnapshotRequest{
		Username:       user.Username,
		AccountBalance: user.AccountBalance,
	}, generatedAt)
	if err != nil {
		t.Fatalf("RefreshUserFinancialMetricSnapshot returned error: %v", err)
	}
	if snapshot.Username != user.Username {
		t.Fatalf("username = %s, want %s", snapshot.Username, user.Username)
	}
	if snapshot.PositionCount != 1 {
		t.Fatalf("position count = %d, want 1", snapshot.PositionCount)
	}
	if !snapshot.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("generated at = %s, want %s", snapshot.GeneratedAt, generatedAt)
	}
	if snapshot.Financial.AmountInPlay == 0 || snapshot.Financial.TotalSpent == 0 {
		t.Fatalf("expected computed financial values, got %+v", snapshot.Financial)
	}
	if snapshot.TransactionSafeRead {
		t.Fatalf("snapshot must not be transaction safe")
	}

	stored, err := NewGormRepository(db).GetUserFinancialMetricSnapshot(context.Background(), user.Username)
	if err != nil {
		t.Fatalf("get stored snapshot: %v", err)
	}
	if stored == nil {
		t.Fatalf("stored snapshot is nil")
	}
	if stored.Username != snapshot.Username ||
		stored.PositionCount != snapshot.PositionCount ||
		stored.Financial != snapshot.Financial ||
		!stored.GeneratedAt.Equal(snapshot.GeneratedAt) {
		t.Fatalf("stored snapshot mismatch:\ngot  %+v\nwant %+v", stored, snapshot)
	}
}

func TestServiceGetUserFinancialMetricReadModelReturnsStoredFreshness(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()
	user := modelstesting.GenerateUser("financial_read_model_user", 500)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := newAnalyticsService(t, db, econ)
	generatedAt := time.Date(2026, 6, 7, 13, 0, 0, 0, time.UTC)
	if _, err := service.RefreshUserFinancialMetricSnapshot(context.Background(), FinancialSnapshotRequest{
		Username:       user.Username,
		AccountBalance: user.AccountBalance,
	}, generatedAt); err != nil {
		t.Fatalf("refresh snapshot: %v", err)
	}

	readModel, err := service.GetUserFinancialMetricReadModel(context.Background(), user.Username)
	if err != nil {
		t.Fatalf("GetUserFinancialMetricReadModel returned error: %v", err)
	}
	if readModel == nil {
		t.Fatalf("read model is nil")
	}
	if readModel.Snapshot.Username != user.Username {
		t.Fatalf("snapshot username = %s, want %s", readModel.Snapshot.Username, user.Username)
	}
	if !readModel.Freshness.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("freshness generated at = %s, want %s", readModel.Freshness.GeneratedAt, generatedAt)
	}
	if readModel.Freshness.Source != "read_model" {
		t.Fatalf("freshness source = %q, want read_model", readModel.Freshness.Source)
	}
	if readModel.Freshness.TargetFreshnessSeconds != 600 {
		t.Fatalf("freshness target = %d, want 600", readModel.Freshness.TargetFreshnessSeconds)
	}
	if readModel.Freshness.TransactionSafeRead {
		t.Fatalf("read-model freshness must not be transaction safe")
	}

	missing, err := service.GetUserFinancialMetricReadModel(context.Background(), "missing_user")
	if err != nil {
		t.Fatalf("missing read model returned error: %v", err)
	}
	if missing != nil {
		t.Fatalf("missing read model = %+v, want nil", missing)
	}
}
