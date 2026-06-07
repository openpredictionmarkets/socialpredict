package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// RefreshSystemMetricsSnapshot recomputes and stores the display-only system
// metrics snapshot from canonical data.
func (s *Service) RefreshSystemMetricsSnapshot(ctx context.Context) (*SystemMetricsReadModel, error) {
	metrics, err := s.ComputeSystemMetrics(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	snapshot := AnalyticsReadModelSnapshot{
		Key:         SystemMetricsSnapshotKey,
		Kind:        AnalyticsSnapshotKindSystemMetrics,
		PayloadJSON: payload,
		GeneratedAt: time.Now().UTC(),
		Source:      "read_model",
	}
	repo, err := s.analyticsReadModelSnapshotRepo()
	if err != nil {
		return nil, err
	}
	if err := repo.UpsertAnalyticsReadModelSnapshot(ctx, snapshot); err != nil {
		return nil, err
	}
	return &SystemMetricsReadModel{
		Metrics:   *metrics,
		Freshness: snapshot.Freshness(SystemMetricsSnapshotTargetFreshness),
	}, nil
}

// GetSystemMetricsReadModel returns the stored display-only system metrics
// snapshot. A missing snapshot is not an error.
func (s *Service) GetSystemMetricsReadModel(ctx context.Context) (*SystemMetricsReadModel, error) {
	repo, err := s.analyticsReadModelSnapshotRepo()
	if err != nil {
		return nil, err
	}
	snapshot, err := repo.GetAnalyticsReadModelSnapshot(ctx, SystemMetricsSnapshotKey)
	if err != nil || snapshot == nil {
		return nil, err
	}
	var metrics SystemMetrics
	if err := json.Unmarshal(snapshot.PayloadJSON, &metrics); err != nil {
		return nil, err
	}
	return &SystemMetricsReadModel{
		Metrics:   metrics,
		Freshness: snapshot.Freshness(SystemMetricsSnapshotTargetFreshness),
	}, nil
}

// RefreshGlobalLeaderboardSnapshot recomputes and stores the display-only
// global leaderboard snapshot from canonical data.
func (s *Service) RefreshGlobalLeaderboardSnapshot(ctx context.Context) (*GlobalLeaderboardReadModel, error) {
	leaderboard, err := s.ComputeGlobalLeaderboardSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(leaderboard.Result())
	if err != nil {
		return nil, err
	}
	snapshot := AnalyticsReadModelSnapshot{
		Key:         GlobalLeaderboardSnapshotKey,
		Kind:        AnalyticsSnapshotKindGlobalLeaderboard,
		PayloadJSON: payload,
		GeneratedAt: time.Now().UTC(),
		Source:      "read_model",
	}
	repo, err := s.analyticsReadModelSnapshotRepo()
	if err != nil {
		return nil, err
	}
	if err := repo.UpsertAnalyticsReadModelSnapshot(ctx, snapshot); err != nil {
		return nil, err
	}
	return &GlobalLeaderboardReadModel{
		Entries:   leaderboard.Result(),
		Freshness: snapshot.Freshness(GlobalLeaderboardSnapshotTargetFreshness),
	}, nil
}

// GetGlobalLeaderboardReadModel returns a page of the stored display-only
// global leaderboard snapshot. A missing snapshot is not an error.
func (s *Service) GetGlobalLeaderboardReadModel(ctx context.Context, limit, offset int) (*GlobalLeaderboardReadModel, error) {
	repo, err := s.analyticsReadModelSnapshotRepo()
	if err != nil {
		return nil, err
	}
	snapshot, err := repo.GetAnalyticsReadModelSnapshot(ctx, GlobalLeaderboardSnapshotKey)
	if err != nil || snapshot == nil {
		return nil, err
	}
	var entries []GlobalUserProfitability
	if err := json.Unmarshal(snapshot.PayloadJSON, &entries); err != nil {
		return nil, err
	}
	paged := (&GlobalLeaderboardSnapshot{Entries: entries}).ResultPage(limit, offset)
	return &GlobalLeaderboardReadModel{
		Entries:   paged,
		Freshness: snapshot.Freshness(GlobalLeaderboardSnapshotTargetFreshness),
	}, nil
}

// MarkAnalyticsReadModelsStale marks aggregate analytics snapshots stale after
// canonical mutations. It is safe to call even before snapshots exist.
func (s *Service) MarkAnalyticsReadModelsStale(ctx context.Context, reason string) error {
	repo, err := s.analyticsReadModelSnapshotRepo()
	if err != nil {
		return err
	}
	if err := repo.MarkAnalyticsReadModelSnapshotStale(ctx, SystemMetricsSnapshotKey, reason); err != nil {
		return err
	}
	return repo.MarkAnalyticsReadModelSnapshotStale(ctx, GlobalLeaderboardSnapshotKey, reason)
}

func (s *Service) analyticsReadModelSnapshotRepo() (AnalyticsReadModelSnapshotRepository, error) {
	repo, ok := s.repo.(AnalyticsReadModelSnapshotRepository)
	if !ok {
		return nil, errors.New("analytics read-model snapshot repository not provided")
	}
	return repo, nil
}
