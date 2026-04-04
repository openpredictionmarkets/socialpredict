package analytics

import (
	"context"

	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
	"socialpredict/setup"
)

// DebtRepository exposes only the user data needed for debt calculations.
type DebtRepository interface {
	ListUsers(ctx context.Context) ([]models.User, error)
}

// VolumeRepository exposes the market and bet data needed for volume calculations.
type VolumeRepository interface {
	ListMarkets(ctx context.Context) ([]models.Market, error)
	ListBetsForMarket(ctx context.Context, marketID uint) ([]models.Bet, error)
}

// FeeRepository exposes the ordered bet data needed for participation fee calculations.
type FeeRepository interface {
	ListBetsOrdered(ctx context.Context) ([]models.Bet, error)
}

// LeaderboardRepository provides the data required to compute leaderboards.
type LeaderboardRepository interface {
	DebtRepository
	VolumeRepository
}

// FinancialsRepository provides the data required for per-user financial snapshots.
type FinancialsRepository interface {
	UserMarketPositions(ctx context.Context, username string) ([]positionsmath.MarketPosition, error)
}

// DebtCalculator calculates debt-related metrics.
type DebtCalculator interface {
	Calculate(ctx context.Context, repo DebtRepository, econ *setup.EconomicConfig) (*DebtStats, error)
}

// VolumeCalculator calculates market volume metrics.
type VolumeCalculator interface {
	Calculate(ctx context.Context, repo VolumeRepository, econ *setup.EconomicConfig) (*MarketVolumeStats, error)
}

// FeeCalculator calculates betting fee metrics.
type FeeCalculator interface {
	CalculateParticipationFees(ctx context.Context, repo FeeRepository, econ *setup.EconomicConfig) (int64, error)
}

// MetricsAssembler combines calculator outputs into the final DTO.
type MetricsAssembler interface {
	Assemble(econ *setup.EconomicConfig, debt *DebtStats, volume *MarketVolumeStats, participationFees int64) *SystemMetrics
}

// MarketPositionCalculator calculates market positions for analytics consumers.
type MarketPositionCalculator interface {
	Calculate(snapshot positionsmath.MarketSnapshot, bets []models.Bet) ([]positionsmath.MarketPosition, error)
}

// Repository exposes the data access required by the analytics domain service.
type Repository interface {
	LeaderboardRepository
	FeeRepository
	FinancialsRepository
}

// Service implements analytics calculations.
type Service struct {
	repo             Repository
	econLoader       setup.EconConfigLoader
	debtCalculator   DebtCalculator
	volumeCalculator VolumeCalculator
	feeCalculator    FeeCalculator
	metricsAssembler MetricsAssembler
	positions        MarketPositionCalculator
}

// ServiceOption allows customizing analytics strategies.
type ServiceOption func(*Service)

func defaultDebtCalculator() DebtCalculator {
	return DefaultDebtCalculator{}
}

func defaultVolumeCalculator() VolumeCalculator {
	return DefaultVolumeCalculator{}
}

func defaultFeeCalculator() FeeCalculator {
	return DefaultFeeCalculator{}
}

func defaultMetricsAssembler() MetricsAssembler {
	return DefaultMetricsAssembler{}
}

func defaultPositionCalculator() MarketPositionCalculator {
	return defaultMarketPositionCalculator{}
}

func debtCalculatorOrDefault(calculator DebtCalculator) DebtCalculator {
	if calculator == nil {
		return defaultDebtCalculator()
	}
	return calculator
}

func volumeCalculatorOrDefault(calculator VolumeCalculator) VolumeCalculator {
	if calculator == nil {
		return defaultVolumeCalculator()
	}
	return calculator
}

func feeCalculatorOrDefault(calculator FeeCalculator) FeeCalculator {
	if calculator == nil {
		return defaultFeeCalculator()
	}
	return calculator
}

func metricsAssemblerOrDefault(assembler MetricsAssembler) MetricsAssembler {
	if assembler == nil {
		return defaultMetricsAssembler()
	}
	return assembler
}

func positionCalculatorOrDefaultStrategy(calculator MarketPositionCalculator) MarketPositionCalculator {
	if calculator == nil {
		return defaultPositionCalculator()
	}
	return calculator
}

// WithDebtCalculator overrides the default debt calculator.
func WithDebtCalculator(c DebtCalculator) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.debtCalculator = debtCalculatorOrDefault(c)
		}
	}
}

// WithVolumeCalculator overrides the default volume calculator.
func WithVolumeCalculator(c VolumeCalculator) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.volumeCalculator = volumeCalculatorOrDefault(c)
		}
	}
}

// WithFeeCalculator overrides the default fee calculator.
func WithFeeCalculator(c FeeCalculator) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.feeCalculator = feeCalculatorOrDefault(c)
		}
	}
}

// WithMetricsAssembler overrides the default metrics assembler.
func WithMetricsAssembler(a MetricsAssembler) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.metricsAssembler = metricsAssemblerOrDefault(a)
		}
	}
}

// WithPositionCalculator overrides the default position calculator.
func WithPositionCalculator(c MarketPositionCalculator) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.positions = positionCalculatorOrDefaultStrategy(c)
		}
	}
}

// NewService constructs an analytics service with optional strategy overrides.
func NewService(repo Repository, econLoader setup.EconConfigLoader, opts ...ServiceOption) *Service {
	service := &Service{
		repo:       repo,
		econLoader: econLoader,
	}

	for _, opt := range opts {
		opt(service)
	}

	service.ensureStrategyDefaults()

	return service
}

func (s *Service) ensureStrategyDefaults() {
	if s == nil {
		return
	}
	s.debtCalculator = debtCalculatorOrDefault(s.debtCalculator)
	s.volumeCalculator = volumeCalculatorOrDefault(s.volumeCalculator)
	s.feeCalculator = feeCalculatorOrDefault(s.feeCalculator)
	s.metricsAssembler = metricsAssemblerOrDefault(s.metricsAssembler)
	s.positions = positionCalculatorOrDefaultStrategy(s.positions)
}

var (
	_ DebtCalculator           = defaultDebtCalculator()
	_ VolumeCalculator         = defaultVolumeCalculator()
	_ FeeCalculator            = defaultFeeCalculator()
	_ MetricsAssembler         = defaultMetricsAssembler()
	_ MarketPositionCalculator = defaultPositionCalculator()
)
