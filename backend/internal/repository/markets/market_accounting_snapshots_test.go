package markets

import (
	"context"
	"errors"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestGormRepositoryMarketAccountingSnapshotUpsertAndRead(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	generatedAt := time.Date(2026, 6, 7, 9, 30, 0, 0, time.UTC)
	lastBetAt := generatedAt.Add(-1 * time.Minute)

	snapshot := dmarkets.MarketAccountingSnapshot{
		MarketID:            42,
		GeneratedAt:         generatedAt,
		LastProbability:     0.62,
		NetBetVolume:        150,
		MarketDust:          1,
		VolumeWithDust:      151,
		UserCount:           3,
		BetCount:            8,
		LastProcessedBetID:  99,
		LastProcessedBetAt:  lastBetAt,
		Source:              "read_model",
		TransactionSafeRead: false,
	}

	if err := repo.UpsertMarketAccountingSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("UpsertMarketAccountingSnapshot returned error: %v", err)
	}

	got, err := repo.GetMarketAccountingSnapshot(ctx, snapshot.MarketID)
	if err != nil {
		t.Fatalf("GetMarketAccountingSnapshot returned error: %v", err)
	}
	assertAccountingSnapshotEqual(t, got, snapshot)

	replacement := snapshot
	replacement.GeneratedAt = generatedAt.Add(5 * time.Minute)
	replacement.LastProbability = 0.71
	replacement.NetBetVolume = 260
	replacement.MarketDust = 2
	replacement.VolumeWithDust = 262
	replacement.UserCount = 4
	replacement.BetCount = 10
	replacement.LastProcessedBetID = 101
	replacement.LastProcessedBetAt = lastBetAt.Add(5 * time.Minute)

	if err := repo.UpsertMarketAccountingSnapshot(ctx, replacement); err != nil {
		t.Fatalf("replacement upsert returned error: %v", err)
	}

	got, err = repo.GetMarketAccountingSnapshot(ctx, snapshot.MarketID)
	if err != nil {
		t.Fatalf("GetMarketAccountingSnapshot replacement returned error: %v", err)
	}
	assertAccountingSnapshotEqual(t, got, replacement)

	var rowCount int64
	if err := db.Model(&models.MarketAccountingSnapshot{}).Where("market_id = ?", snapshot.MarketID).Count(&rowCount).Error; err != nil {
		t.Fatalf("count snapshots: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("snapshot row count = %d, want 1", rowCount)
	}
}

func TestGormRepositoryMarketAccountingSnapshotMissingAndInvalid(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	got, err := repo.GetMarketAccountingSnapshot(ctx, 999)
	if err != nil {
		t.Fatalf("GetMarketAccountingSnapshot missing returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("missing snapshot = %+v, want nil", got)
	}

	if _, err := repo.GetMarketAccountingSnapshot(ctx, 0); !errors.Is(err, dmarkets.ErrInvalidInput) {
		t.Fatalf("invalid get error = %v, want ErrInvalidInput", err)
	}
	if err := repo.UpsertMarketAccountingSnapshot(ctx, dmarkets.MarketAccountingSnapshot{}); !errors.Is(err, dmarkets.ErrInvalidInput) {
		t.Fatalf("invalid upsert error = %v, want ErrInvalidInput", err)
	}
}

func assertAccountingSnapshotEqual(t *testing.T, got *dmarkets.MarketAccountingSnapshot, want dmarkets.MarketAccountingSnapshot) {
	t.Helper()
	if got == nil {
		t.Fatalf("snapshot is nil")
	}
	if got.MarketID != want.MarketID ||
		got.LastProbability != want.LastProbability ||
		got.NetBetVolume != want.NetBetVolume ||
		got.MarketDust != want.MarketDust ||
		got.VolumeWithDust != want.VolumeWithDust ||
		got.UserCount != want.UserCount ||
		got.BetCount != want.BetCount ||
		got.LastProcessedBetID != want.LastProcessedBetID ||
		got.Source != want.Source ||
		got.TransactionSafeRead != false {
		t.Fatalf("snapshot mismatch:\ngot  %+v\nwant %+v", got, want)
	}
	if !got.GeneratedAt.Equal(want.GeneratedAt) {
		t.Fatalf("generated at = %s, want %s", got.GeneratedAt, want.GeneratedAt)
	}
	if !got.LastProcessedBetAt.Equal(want.LastProcessedBetAt) {
		t.Fatalf("last processed bet at = %s, want %s", got.LastProcessedBetAt, want.LastProcessedBetAt)
	}
}
