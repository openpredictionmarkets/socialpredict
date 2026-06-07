package analytics

import (
	"context"
	"testing"
	"time"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestGormRepositoryUserFinancialMetricSnapshotUpsertAndRead(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	generatedAt := time.Date(2026, 6, 7, 11, 0, 0, 0, time.UTC)

	snapshot := UserFinancialMetricSnapshot{
		Username:      "alice",
		GeneratedAt:   generatedAt,
		PositionCount: 3,
		Financial: FinancialSnapshot{
			AccountBalance:     500,
			MaximumDebtAllowed: 250,
			AmountInPlay:       125,
			RetainedEarnings:   375,
			Equity:             500,
			TradingProfits:     25,
			TotalProfits:       25,
			TotalSpent:         100,
			PotentialValue:     125,
		},
		Source:              "read_model",
		TransactionSafeRead: false,
	}

	if err := repo.UpsertUserFinancialMetricSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("UpsertUserFinancialMetricSnapshot returned error: %v", err)
	}
	got, err := repo.GetUserFinancialMetricSnapshot(ctx, snapshot.Username)
	if err != nil {
		t.Fatalf("GetUserFinancialMetricSnapshot returned error: %v", err)
	}
	assertUserFinancialMetricSnapshotEqual(t, got, snapshot)

	replacement := snapshot
	replacement.GeneratedAt = generatedAt.Add(5 * time.Minute)
	replacement.PositionCount = 4
	replacement.Financial.AccountBalance = 450
	replacement.Financial.AmountInPlay = 175
	replacement.Financial.RetainedEarnings = 275
	replacement.Financial.Equity = 450
	replacement.Financial.TradingProfits = 35
	replacement.Financial.TotalProfits = 35

	if err := repo.UpsertUserFinancialMetricSnapshot(ctx, replacement); err != nil {
		t.Fatalf("replacement upsert returned error: %v", err)
	}
	got, err = repo.GetUserFinancialMetricSnapshot(ctx, snapshot.Username)
	if err != nil {
		t.Fatalf("replacement get returned error: %v", err)
	}
	assertUserFinancialMetricSnapshotEqual(t, got, replacement)

	var rowCount int64
	if err := db.Model(&models.UserFinancialMetricSnapshot{}).Where("username = ?", snapshot.Username).Count(&rowCount).Error; err != nil {
		t.Fatalf("count snapshots: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("snapshot row count = %d, want 1", rowCount)
	}
}

func TestGormRepositoryUserFinancialMetricSnapshotMissingAndInvalid(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	got, err := repo.GetUserFinancialMetricSnapshot(ctx, "missing")
	if err != nil {
		t.Fatalf("missing get returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("missing snapshot = %+v, want nil", got)
	}

	if _, err := repo.GetUserFinancialMetricSnapshot(ctx, ""); err == nil {
		t.Fatalf("expected invalid get error")
	}
	if err := repo.UpsertUserFinancialMetricSnapshot(ctx, UserFinancialMetricSnapshot{}); err == nil {
		t.Fatalf("expected invalid upsert error")
	}
}

func TestGormRepositoryUserFinancialMetricSnapshotCanBeMarkedStaleAndRefreshed(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	snapshot := UserFinancialMetricSnapshot{
		Username:    "alice",
		GeneratedAt: time.Date(2026, 6, 7, 11, 0, 0, 0, time.UTC),
		Financial:   FinancialSnapshot{AccountBalance: 500},
		Source:      "read_model",
	}
	if err := repo.UpsertUserFinancialMetricSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}
	if err := repo.MarkUserFinancialMetricSnapshotStale(ctx, snapshot.Username, "bet_accepted"); err != nil {
		t.Fatalf("mark stale: %v", err)
	}

	stale, err := repo.GetUserFinancialMetricSnapshot(ctx, snapshot.Username)
	if err != nil {
		t.Fatalf("get stale snapshot: %v", err)
	}
	if !stale.IsStale || stale.StaleReason != "bet_accepted" || stale.MarkedStaleAt == nil {
		t.Fatalf("expected stale marker, got %+v", stale)
	}
	if !stale.Freshness().IsStale {
		t.Fatalf("freshness should report stale marker")
	}

	refreshed := snapshot
	refreshed.GeneratedAt = snapshot.GeneratedAt.Add(time.Minute)
	if err := repo.UpsertUserFinancialMetricSnapshot(ctx, refreshed); err != nil {
		t.Fatalf("refresh upsert: %v", err)
	}
	fresh, err := repo.GetUserFinancialMetricSnapshot(ctx, snapshot.Username)
	if err != nil {
		t.Fatalf("get refreshed snapshot: %v", err)
	}
	if fresh.IsStale || fresh.StaleReason != "" || fresh.MarkedStaleAt != nil {
		t.Fatalf("expected refresh to clear stale marker, got %+v", fresh)
	}
}

func assertUserFinancialMetricSnapshotEqual(t *testing.T, got *UserFinancialMetricSnapshot, want UserFinancialMetricSnapshot) {
	t.Helper()
	if got == nil {
		t.Fatalf("snapshot is nil")
	}
	if got.Username != want.Username ||
		got.PositionCount != want.PositionCount ||
		got.Source != want.Source ||
		got.TransactionSafeRead {
		t.Fatalf("snapshot metadata mismatch:\ngot  %+v\nwant %+v", got, want)
	}
	if !got.GeneratedAt.Equal(want.GeneratedAt) {
		t.Fatalf("generated at = %s, want %s", got.GeneratedAt, want.GeneratedAt)
	}
	if got.Financial != want.Financial {
		t.Fatalf("financial snapshot mismatch:\ngot  %+v\nwant %+v", got.Financial, want.Financial)
	}
}
