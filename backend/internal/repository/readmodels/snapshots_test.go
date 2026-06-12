package readmodels

import (
	"context"
	"testing"
	"time"

	"socialpredict/models/modelstesting"
)

func TestMarkMarketDiscoverySnapshotsStaleSoftStaleForTransactions(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	generatedAt := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	key := "market_discovery:markets:status=active:tag=none:limit=21:offset=0"

	if err := repo.Upsert(ctx, Snapshot{
		Key:         key,
		Kind:        "market_discovery",
		PayloadJSON: `{"markets":[]}`,
		GeneratedAt: generatedAt,
		Source:      "read_model",
	}); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}

	if err := repo.MarkMarketDiscoverySnapshotsStale(ctx, "bet_accepted"); err != nil {
		t.Fatalf("mark stale: %v", err)
	}

	got, err := repo.Get(ctx, key)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if got == nil || !got.IsStale || got.StaleReason != "bet_accepted" || got.MarkedStaleAt == nil {
		t.Fatalf("expected soft-stale transaction marker, got %+v", got)
	}
	if !got.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("transaction invalidation should keep generated_at, got %s want %s", got.GeneratedAt, generatedAt)
	}
}

func TestMarkMarketDiscoverySnapshotsStaleExpiresStructuralChanges(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	generatedAt := time.Now().UTC()
	key := "market_discovery:markets:status=active:tag=none:limit=21:offset=0"

	if err := repo.Upsert(ctx, Snapshot{
		Key:         key,
		Kind:        "market_discovery",
		PayloadJSON: `{"markets":[]}`,
		GeneratedAt: generatedAt,
		Source:      "read_model",
	}); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}

	if err := repo.MarkMarketDiscoverySnapshotsStale(ctx, "market_group_created"); err != nil {
		t.Fatalf("mark stale: %v", err)
	}

	got, err := repo.Get(ctx, key)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if got == nil || !got.IsStale || got.StaleReason != "market_group_created" || got.MarkedStaleAt == nil {
		t.Fatalf("expected structural stale marker, got %+v", got)
	}
	if !got.GeneratedAt.Before(generatedAt.Add(-23 * time.Hour)) {
		t.Fatalf("structural invalidation should age generated_at, got %s original %s", got.GeneratedAt, generatedAt)
	}
}
