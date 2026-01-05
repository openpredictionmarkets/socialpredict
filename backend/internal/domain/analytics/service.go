package analytics

import (
	"context"

	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/setup"
)

// DebtCalculator calculates debt-related metrics.
type DebtCalculator interface {
	Calculate(ctx context.Context, repo Repository, econ *setup.EconomicConfig) (*DebtStats, error)
}

// VolumeCalculator calculates market volume metrics.
type VolumeCalculator interface {
	Calculate(ctx context.Context, repo Repository, econ *setup.EconomicConfig) (*MarketVolumeStats, error)
}

// FeeCalculator calculates betting fee metrics.
type FeeCalculator interface {
	CalculateParticipationFees(ctx context.Context, repo Repository, econ *setup.EconomicConfig) (int64, error)
}

// MetricsAssembler combines calculator outputs into the final DTO.
type MetricsAssembler interface {
	Assemble(econ *setup.EconomicConfig, debt *DebtStats, volume *MarketVolumeStats, participationFees int64) *SystemMetrics
}

// Repository exposes the data access required by the analytics domain service.
type Repository interface {
	ListUsers(ctx context.Context) ([]models.User, error)
	ListMarkets(ctx context.Context) ([]models.Market, error)
	ListBetsForMarket(ctx context.Context, marketID uint) ([]models.Bet, error)
	ListBetsOrdered(ctx context.Context) ([]models.Bet, error)
	UserMarketPositions(ctx context.Context, username string) ([]positionsmath.MarketPosition, error)
}

// Service implements analytics calculations.
type Service struct {
	repo             Repository
	econLoader       setup.EconConfigLoader
	debtCalculator   DebtCalculator
	volumeCalculator VolumeCalculator
	feeCalculator    FeeCalculator
	metricsAssembler MetricsAssembler
}

// ServiceOption allows customizing analytics strategies.
type ServiceOption func(*Service)

// WithDebtCalculator overrides the default debt calculator.
func WithDebtCalculator(c DebtCalculator) ServiceOption {
	return func(s *Service) {
		if c != nil {
			s.debtCalculator = c
		}
	}
}

// WithVolumeCalculator overrides the default volume calculator.
func WithVolumeCalculator(c VolumeCalculator) ServiceOption {
	return func(s *Service) {
		if c != nil {
			s.volumeCalculator = c
		}
	}
}

// WithFeeCalculator overrides the default fee calculator.
func WithFeeCalculator(c FeeCalculator) ServiceOption {
	return func(s *Service) {
		if c != nil {
			s.feeCalculator = c
		}
	}
}

// WithMetricsAssembler overrides the default metrics assembler.
func WithMetricsAssembler(a MetricsAssembler) ServiceOption {
	return func(s *Service) {
		if a != nil {
			s.metricsAssembler = a
		}
	}
}

// NewService constructs an analytics service with optional strategy overrides.
func NewService(repo Repository, econLoader setup.EconConfigLoader, opts ...ServiceOption) *Service {
	service := &Service{
		repo:             repo,
		econLoader:       econLoader,
		debtCalculator:   DefaultDebtCalculator{},
		volumeCalculator: DefaultVolumeCalculator{},
		feeCalculator:    DefaultFeeCalculator{},
		metricsAssembler: DefaultMetricsAssembler{},
	}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func (s *Service) ensureStrategyDefaults() {
	if s.debtCalculator == nil {
		s.debtCalculator = DefaultDebtCalculator{}
	}
	if s.volumeCalculator == nil {
		s.volumeCalculator = DefaultVolumeCalculator{}
	}
	if s.feeCalculator == nil {
		s.feeCalculator = DefaultFeeCalculator{}
	}
	if s.metricsAssembler == nil {
		s.metricsAssembler = DefaultMetricsAssembler{}
	}
}
