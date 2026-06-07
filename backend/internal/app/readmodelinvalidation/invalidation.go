package readmodelinvalidation

import (
	"context"
	"errors"
)

// MarketInvalidator marks market-owned display read models stale.
type MarketInvalidator interface {
	MarkMarketAccountingSnapshotStale(ctx context.Context, marketID int64, reason string) error
	MarkMarketLeaderboardSnapshotStale(ctx context.Context, marketID int64, reason string) error
	MarkMarketPositionsSnapshotStale(ctx context.Context, marketID int64, reason string) error
}

// AnalyticsInvalidator marks analytics-owned display read models stale.
type AnalyticsInvalidator interface {
	MarkUserFinancialMetricSnapshotStale(ctx context.Context, username string, reason string) error
	MarkAnalyticsReadModelsStale(ctx context.Context, reason string) error
}

// DiscoveryInvalidator marks page/card discovery read models stale.
type DiscoveryInvalidator interface {
	MarkMarketDiscoverySnapshotsStale(ctx context.Context, reason string) error
}

// Service coordinates best-effort display read-model invalidation after
// canonical mutations. It must not participate in transaction decisions.
type Service struct {
	markets   MarketInvalidator
	analytics AnalyticsInvalidator
	discovery DiscoveryInvalidator
}

// New builds a read-model invalidator from optional collaborators.
func New(markets MarketInvalidator, analytics AnalyticsInvalidator, discovery DiscoveryInvalidator) *Service {
	return &Service{markets: markets, analytics: analytics, discovery: discovery}
}

// InvalidateAfterMarketTransaction marks affected display read models stale.
// The canonical transaction has already completed when this is called.
func (s *Service) InvalidateAfterMarketTransaction(ctx context.Context, username string, marketID int64, reason string) error {
	if s == nil {
		return nil
	}
	var errs []error
	if s.markets != nil {
		if err := s.markets.MarkMarketAccountingSnapshotStale(ctx, marketID, reason); err != nil {
			errs = append(errs, err)
		}
		if err := s.markets.MarkMarketLeaderboardSnapshotStale(ctx, marketID, reason); err != nil {
			errs = append(errs, err)
		}
		if err := s.markets.MarkMarketPositionsSnapshotStale(ctx, marketID, reason); err != nil {
			errs = append(errs, err)
		}
	}
	if s.analytics != nil {
		if username != "" {
			if err := s.analytics.MarkUserFinancialMetricSnapshotStale(ctx, username, reason); err != nil {
				errs = append(errs, err)
			}
		}
		if err := s.analytics.MarkAnalyticsReadModelsStale(ctx, reason); err != nil {
			errs = append(errs, err)
		}
	}
	if s.discovery != nil {
		if err := s.discovery.MarkMarketDiscoverySnapshotsStale(ctx, reason); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
