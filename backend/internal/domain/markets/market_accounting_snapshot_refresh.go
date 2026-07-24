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

// GetMarketSummaryReadModel returns display-safe market summary data from the
// durable accounting snapshot. Existing stale snapshots are returned rather
// than synchronously refreshed so high-traffic display reads do not recompute
// market accounting on every write.
func (s *Service) GetMarketSummaryReadModel(ctx context.Context, marketID int64) (*MarketSummaryReadModel, error) {
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

	snapshot, err := snapshotRepo.GetMarketAccountingSnapshot(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		snapshot, err = s.RefreshMarketAccountingSnapshot(ctx, marketID)
		if err != nil {
			return nil, err
		}
	}

	return &MarketSummaryReadModel{
		Market:     market,
		Creator:    s.buildCreatorSummary(ctx, market.CreatorUsername),
		Accounting: *snapshot,
	}, nil
}

// GetMarketDiscoverySummaries assembles display summaries for hydrated
// discovery markets from batched repository reads.
func (s *Service) GetMarketDiscoverySummaries(
	ctx context.Context,
	input []*Market,
) (map[int64]*MarketSummaryReadModel, error) {
	repo, ok := s.repo.(MarketDiscoveryEnrichmentRepository)
	if !ok {
		return nil, ErrInvalidState
	}

	marketsByID := map[int64]*Market{}
	marketIDs := make([]int64, 0, len(input))
	creators := make([]string, 0, len(input))
	for _, market := range input {
		if market == nil || market.ID <= 0 {
			continue
		}
		if _, exists := marketsByID[market.ID]; exists {
			continue
		}
		marketsByID[market.ID] = market
		marketIDs = append(marketIDs, market.ID)
		creators = append(creators, market.CreatorUsername)
	}

	enrichment, err := repo.GetMarketDiscoveryEnrichment(ctx, marketIDs, creators)
	if err != nil {
		return nil, err
	}
	if enrichment == nil {
		return nil, ErrInvalidState
	}

	out := make(map[int64]*MarketSummaryReadModel, len(marketsByID))
	for _, marketID := range marketIDs {
		market := marketsByID[marketID]
		accounting, found := enrichment.AccountingByMarketID[marketID]
		if !found {
			refreshed, err := s.RefreshMarketAccountingSnapshot(ctx, marketID)
			if err != nil {
				return nil, err
			}
			accounting = *refreshed
		}

		creator := CreatorSummary{Username: market.CreatorUsername}
		if enriched, found := enrichment.CreatorsByUsername[market.CreatorUsername]; found {
			creator = enriched
		}
		out[marketID] = &MarketSummaryReadModel{
			Market:     market,
			Creator:    &creator,
			Accounting: accounting,
		}
	}
	return out, nil
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
