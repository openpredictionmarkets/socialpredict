package markets

import (
	"context"
	"time"

	"socialpredict/internal/domain/readmodels"
)

const MarketPositionsSnapshotTargetFreshness = 10 * time.Minute

// MarketPositionsSnapshot is a display/read-model snapshot for market position
// rows. It must not be used for order, payout, refund, or balance decisions.
type MarketPositionsSnapshot struct {
	MarketID            int64
	Positions           MarketPositions
	GeneratedAt         time.Time
	Source              string
	TransactionSafeRead bool
	IsStale             bool
	StaleReason         string
	MarkedStaleAt       *time.Time
}

// MarketPositionsSnapshotRepository persists display-only market position
// snapshots separately from transaction repository interfaces.
type MarketPositionsSnapshotRepository interface {
	GetMarketPositionsSnapshot(ctx context.Context, marketID int64) (*MarketPositionsSnapshot, error)
	UpsertMarketPositionsSnapshot(ctx context.Context, snapshot MarketPositionsSnapshot) error
	MarkMarketPositionsSnapshotStale(ctx context.Context, marketID int64, reason string) error
}

func (s MarketPositionsSnapshot) Freshness() readmodels.Freshness {
	if s.IsStale {
		return readmodels.NewStaleFreshness(
			s.GeneratedAt,
			s.Source,
			MarketPositionsSnapshotTargetFreshness,
			s.TransactionSafeRead,
			s.StaleReason,
			s.MarkedStaleAt,
		)
	}
	return readmodels.NewFreshness(
		s.GeneratedAt,
		s.Source,
		MarketPositionsSnapshotTargetFreshness,
		s.TransactionSafeRead,
	)
}

// RefreshMarketPositionsSnapshot recomputes and stores the display-only market
// positions snapshot from canonical market and bet data.
func (s *Service) RefreshMarketPositionsSnapshot(ctx context.Context, marketID int64) (*MarketPositionsSnapshot, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}
	snapshotRepo, ok := s.repo.(MarketPositionsSnapshotRepository)
	if !ok {
		return nil, ErrInvalidState
	}

	positions, err := s.GetMarketPositions(ctx, marketID)
	if err != nil {
		return nil, err
	}
	positions = activeMarketPositions(positions)
	sortMarketPositionsByTotalShares(positions)

	snapshot := MarketPositionsSnapshot{
		MarketID:    marketID,
		Positions:   positions,
		GeneratedAt: time.Now().UTC(),
		Source:      "read_model",
	}
	if err := snapshotRepo.UpsertMarketPositionsSnapshot(ctx, snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// GetMarketPositionsReadModel returns a page from the stored display-only
// market positions snapshot. A missing snapshot is not an error.
func (s *Service) GetMarketPositionsReadModel(ctx context.Context, marketID int64, p Page) (*MarketPositionsSnapshot, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}
	snapshotRepo, ok := s.repo.(MarketPositionsSnapshotRepository)
	if !ok {
		return nil, ErrInvalidState
	}
	snapshot, err := snapshotRepo.GetMarketPositionsSnapshot(ctx, marketID)
	if err != nil || snapshot == nil {
		return nil, err
	}
	p = s.statusPolicy.NormalizePage(p, 20, 100)
	return &MarketPositionsSnapshot{
		MarketID:            snapshot.MarketID,
		Positions:           paginateMarketPositions(snapshot.Positions, p),
		GeneratedAt:         snapshot.GeneratedAt,
		Source:              snapshot.Source,
		TransactionSafeRead: snapshot.TransactionSafeRead,
		IsStale:             snapshot.IsStale,
		StaleReason:         snapshot.StaleReason,
		MarkedStaleAt:       snapshot.MarkedStaleAt,
	}, nil
}

// MarkMarketPositionsSnapshotStale marks a display positions snapshot stale
// after canonical market activity. It does not update market/bet truth.
func (s *Service) MarkMarketPositionsSnapshotStale(ctx context.Context, marketID int64, reason string) error {
	if marketID <= 0 {
		return ErrInvalidInput
	}
	snapshotRepo, ok := s.repo.(MarketPositionsSnapshotRepository)
	if !ok {
		return ErrInvalidState
	}
	return snapshotRepo.MarkMarketPositionsSnapshotStale(ctx, marketID, reason)
}
