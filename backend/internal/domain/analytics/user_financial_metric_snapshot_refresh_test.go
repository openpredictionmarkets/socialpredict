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
