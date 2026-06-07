package markets

import "context"

// RefreshMarketAccountingSnapshot recomputes and stores a display/read-model
// accounting snapshot from canonical market and bet data.
func (s *Service) RefreshMarketAccountingSnapshot(ctx context.Context, marketID int64) (*MarketAccountingSnapshot, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}

	snapshotRepo, ok := s.repo.(MarketAccountingSnapshotRepository)
	if !ok {
		return nil, ErrInvalidState
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if market == nil {
		return nil, ErrMarketNotFound
	}

	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}

	snapshot := NewMarketAccountingSnapshotCalculator(s.probabilityEngine, s.metricsCalculator, s.clock).
		Calculate(market, ToBoundaryBets(bets))
	if err := snapshotRepo.UpsertMarketAccountingSnapshot(ctx, snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// MarkMarketAccountingSnapshotStale marks a market accounting read model stale
// after canonical market activity. It does not update market/bet truth.
func (s *Service) MarkMarketAccountingSnapshotStale(ctx context.Context, marketID int64, reason string) error {
	if marketID <= 0 {
		return ErrInvalidInput
	}
	snapshotRepo, ok := s.repo.(MarketAccountingSnapshotRepository)
	if !ok {
		return ErrInvalidState
	}
	return snapshotRepo.MarkMarketAccountingSnapshotStale(ctx, marketID, reason)
}
