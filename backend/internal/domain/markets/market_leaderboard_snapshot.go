package markets

import (
	"context"
	"time"

	"socialpredict/internal/domain/readmodels"
)

const MarketLeaderboardSnapshotTargetFreshness = 10 * time.Minute

// MarketLeaderboardSnapshot is a display/read-model snapshot for market
// leaderboard rows. It must not be used for payout, refund, or order logic.
type MarketLeaderboardSnapshot struct {
	MarketID            int64
	Rows                []*LeaderboardRow
	GeneratedAt         time.Time
	Source              string
	TransactionSafeRead bool
	IsStale             bool
	StaleReason         string
	MarkedStaleAt       *time.Time
}

// MarketLeaderboardSnapshotRepository persists display-only market leaderboard
// snapshots separately from transaction repository interfaces.
type MarketLeaderboardSnapshotRepository interface {
	GetMarketLeaderboardSnapshot(ctx context.Context, marketID int64) (*MarketLeaderboardSnapshot, error)
	UpsertMarketLeaderboardSnapshot(ctx context.Context, snapshot MarketLeaderboardSnapshot) error
	MarkMarketLeaderboardSnapshotStale(ctx context.Context, marketID int64, reason string) error
}

func (s MarketLeaderboardSnapshot) Freshness() readmodels.Freshness {
	if s.IsStale {
		return readmodels.NewStaleFreshness(
			s.GeneratedAt,
			s.Source,
			MarketLeaderboardSnapshotTargetFreshness,
			s.TransactionSafeRead,
			s.StaleReason,
			s.MarkedStaleAt,
		)
	}
	return readmodels.NewFreshness(
		s.GeneratedAt,
		s.Source,
		MarketLeaderboardSnapshotTargetFreshness,
		s.TransactionSafeRead,
	)
}

// RefreshMarketLeaderboardSnapshot recomputes and stores the display-only
// market leaderboard snapshot from canonical market and bet data.
func (s *Service) RefreshMarketLeaderboardSnapshot(ctx context.Context, marketID int64) (*MarketLeaderboardSnapshot, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}
	snapshotRepo, ok := s.repo.(MarketLeaderboardSnapshotRepository)
	if !ok {
		return nil, ErrInvalidState
	}

	rows, err := s.GetMarketLeaderboard(ctx, marketID, Page{Limit: 1000})
	if err != nil {
		return nil, err
	}
	snapshot := MarketLeaderboardSnapshot{
		MarketID:    marketID,
		Rows:        rows,
		GeneratedAt: time.Now().UTC(),
		Source:      "read_model",
	}
	if err := snapshotRepo.UpsertMarketLeaderboardSnapshot(ctx, snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// GetMarketLeaderboardReadModel returns a page from the stored display-only
// market leaderboard snapshot. A missing snapshot is not an error.
func (s *Service) GetMarketLeaderboardReadModel(ctx context.Context, marketID int64, p Page) (*MarketLeaderboardSnapshot, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}
	snapshotRepo, ok := s.repo.(MarketLeaderboardSnapshotRepository)
	if !ok {
		return nil, ErrInvalidState
	}
	snapshot, err := snapshotRepo.GetMarketLeaderboardSnapshot(ctx, marketID)
	if err != nil || snapshot == nil {
		return nil, err
	}
	p = s.statusPolicy.NormalizePage(p, 20, 100)
	return &MarketLeaderboardSnapshot{
		MarketID:            snapshot.MarketID,
		Rows:                paginateLeaderboardRows(snapshot.Rows, p),
		GeneratedAt:         snapshot.GeneratedAt,
		Source:              snapshot.Source,
		TransactionSafeRead: snapshot.TransactionSafeRead,
		IsStale:             snapshot.IsStale,
		StaleReason:         snapshot.StaleReason,
		MarkedStaleAt:       snapshot.MarkedStaleAt,
	}, nil
}

// MarkMarketLeaderboardSnapshotStale marks a display leaderboard snapshot stale
// after canonical market activity.
func (s *Service) MarkMarketLeaderboardSnapshotStale(ctx context.Context, marketID int64, reason string) error {
	if marketID <= 0 {
		return ErrInvalidInput
	}
	snapshotRepo, ok := s.repo.(MarketLeaderboardSnapshotRepository)
	if !ok {
		return ErrInvalidState
	}
	return snapshotRepo.MarkMarketLeaderboardSnapshotStale(ctx, marketID, reason)
}

func paginateLeaderboardRows(rows []*LeaderboardRow, p Page) []*LeaderboardRow {
	if len(rows) == 0 || p.Offset >= len(rows) {
		return []*LeaderboardRow{}
	}
	end := p.Offset + p.Limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[p.Offset:end]
}
