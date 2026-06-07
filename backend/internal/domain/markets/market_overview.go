package markets

import (
	"context"

	"socialpredict/internal/domain/boundary"
)

// GetPublicMarket returns a public representation of a market.
func (s *Service) GetPublicMarket(ctx context.Context, marketID int64) (*PublicMarket, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.GetPublicMarket(ctx, marketID)
}

// ListMarkets returns a list of markets with filters.
func (s *Service) ListMarkets(ctx context.Context, filters ListFilters) ([]*Market, error) {
	return s.repo.List(ctx, filters)
}

// ListLifecycleMarkets returns non-public lifecycle queues for owner/admin views.
func (s *Service) ListLifecycleMarkets(ctx context.Context, filters ListFilters) ([]*Market, error) {
	status := NormalizeLifecycleStatus(filters.Status)
	switch status {
	case MarketStatusAll, MarketLifecycleProposed, MarketLifecyclePublished, MarketLifecycleRejected, MarketLifecycleClosed, MarketLifecycleResolved, MarketLifecycleCancelled:
	default:
		return nil, ErrInvalidInput
	}

	filters.Status = status
	repo, ok := s.repo.(LifecycleReadRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	markets, err := repo.ListByLifecycle(ctx, filters)
	if err != nil {
		return nil, err
	}
	if markets == nil {
		return []*Market{}, nil
	}
	return markets, nil
}

// GetMarketOverviews returns enriched market data with calculations.
func (s *Service) GetMarketOverviews(ctx context.Context, filters ListFilters) ([]*MarketOverview, error) {
	markets, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}
	if markets == nil {
		return []*MarketOverview{}, nil
	}

	var overviews []*MarketOverview
	for _, market := range markets {
		overview := &MarketOverview{
			Market:  market,
			Creator: s.buildCreatorSummary(ctx, market.CreatorUsername),
		}
		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// GetMarketDetails returns detailed market information with calculations.
func (s *Service) GetMarketDetails(ctx context.Context, marketID int64) (*MarketOverview, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
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

	boundaryBets := ToBoundaryBets(bets)
	accounting := NewMarketAccountingSnapshotCalculator(s.probabilityEngine, s.metricsCalculator, s.clock).
		Calculate(market, boundaryBets)

	return &MarketOverview{
		Market:                market,
		Creator:               s.buildCreatorSummary(ctx, market.CreatorUsername),
		ProbabilityChanges:    accounting.ProbabilityChanges,
		LastProbability:       accounting.LastProbability,
		NumUsers:              accounting.UserCount,
		TotalVolume:           accounting.VolumeWithDust,
		MarketDust:            accounting.MarketDust,
		DescriptionAmendments: s.approvedDescriptionAmendments(ctx, marketID),
	}, nil
}

func (s *Service) buildCreatorSummary(ctx context.Context, username string) *CreatorSummary {
	summary := &CreatorSummary{Username: username}
	if s.userService == nil {
		return summary
	}
	user, err := s.userService.GetPublicUser(ctx, username)
	if err != nil || user == nil {
		return summary
	}
	summary.DisplayName = user.DisplayName
	summary.PersonalEmoji = user.PersonalEmoji
	return summary
}

func convertToBoundaryBets(bets []*Bet) []boundary.Bet {
	return ToBoundaryBets(bets)
}

func countUniqueUsers(bets []boundary.Bet) int {
	if len(bets) == 0 {
		return 0
	}
	seen := make(map[string]struct{})
	for _, bet := range bets {
		if bet.Username == "" {
			continue
		}
		if _, ok := seen[bet.Username]; !ok {
			seen[bet.Username] = struct{}{}
		}
	}
	return len(seen)
}
