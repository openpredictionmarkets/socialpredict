package analytics

import (
	"context"
	"testing"
	"time"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestGormRepositoryAnalyticsReadModelSnapshotUpsertReadAndStale(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	snapshot := AnalyticsReadModelSnapshot{
		Key:         SystemMetricsSnapshotKey,
		Kind:        AnalyticsSnapshotKindSystemMetrics,
		PayloadJSON: []byte(`{"ok":true}`),
		GeneratedAt: time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
		Source:      "read_model",
	}
	if err := repo.UpsertAnalyticsReadModelSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("upsert analytics snapshot: %v", err)
	}
	got, err := repo.GetAnalyticsReadModelSnapshot(ctx, snapshot.Key)
	if err != nil {
		t.Fatalf("get analytics snapshot: %v", err)
	}
	if got == nil || got.Key != snapshot.Key || string(got.PayloadJSON) != string(snapshot.PayloadJSON) {
		t.Fatalf("snapshot mismatch: got %+v want %+v", got, snapshot)
	}

	if err := repo.MarkAnalyticsReadModelSnapshotStale(ctx, snapshot.Key, "bet_accepted"); err != nil {
		t.Fatalf("mark stale: %v", err)
	}
	stale, err := repo.GetAnalyticsReadModelSnapshot(ctx, snapshot.Key)
	if err != nil {
		t.Fatalf("get stale snapshot: %v", err)
	}
	if !stale.IsStale || stale.StaleReason != "bet_accepted" || stale.MarkedStaleAt == nil {
		t.Fatalf("expected stale marker, got %+v", stale)
	}

	refreshed := snapshot
	refreshed.PayloadJSON = []byte(`{"ok":true,"version":2}`)
	if err := repo.UpsertAnalyticsReadModelSnapshot(ctx, refreshed); err != nil {
		t.Fatalf("refresh upsert: %v", err)
	}
	fresh, err := repo.GetAnalyticsReadModelSnapshot(ctx, snapshot.Key)
	if err != nil {
		t.Fatalf("get refreshed snapshot: %v", err)
	}
	if fresh.IsStale || fresh.StaleReason != "" || fresh.MarkedStaleAt != nil {
		t.Fatalf("expected refresh to clear stale marker, got %+v", fresh)
	}

	var rowCount int64
	if err := db.Model(&models.AnalyticsReadModelSnapshot{}).Where("snapshot_key = ?", snapshot.Key).Count(&rowCount).Error; err != nil {
		t.Fatalf("count snapshots: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("snapshot row count = %d, want 1", rowCount)
	}
}
