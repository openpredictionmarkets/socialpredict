package bets

import (
	"context"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
	"socialpredict/setup"
)

// Repository exposes the persistence layer needed by the bets domain service.
type Repository interface {
	Create(ctx context.Context, bet *models.Bet) error
	UserHasBet(ctx context.Context, marketID uint, username string) (bool, error)
}

// MarketService exposes the subset of market operations required by bets.
type MarketService interface {
	GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error)
	GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)
}

// MarketGate ensures market openness before betting operations.
type MarketGate interface {
	Open(ctx context.Context, marketID int64) (*dmarkets.Market, error)
}

// UserService exposes the subset of user operations required by bets.
type UserService interface {
	GetUser(ctx context.Context, username string) (*dusers.User, error)
	ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error
}

// PlaceValidator allows validation rules to be extended without changing the service.
type PlaceValidator interface {
	Validate(ctx context.Context, req PlaceRequest) (string, error)
}

// SellValidator allows sell validation rules to be extended without changing the service.
type SellValidator interface {
	Validate(ctx context.Context, req SellRequest) (string, error)
}

// SaleCalculator encapsulates sale pricing and dust rules.
type SaleCalculator interface {
	Calculate(pos *dmarkets.UserPosition, sharesOwned int64, creditsRequested int64) (SaleQuote, error)
}

// FeeCalculator encapsulates buy fee calculations.
type FeeCalculator interface {
	Calculate(hasBet bool, amount int64) betFees
}

// BalanceGuard validates user balances against debt limits.
type BalanceGuard interface {
	EnsureSufficient(balance, totalCost int64) error
}

// BetLedger encapsulates persistence and user accounting for bets.
type BetLedger interface {
	ChargeAndRecord(ctx context.Context, bet *models.Bet, totalCost int64) error
	CreditSale(ctx context.Context, bet *models.Bet, saleValue int64) error
}

// Clock allows time to be mocked in tests.
type Clock interface {
	Now() time.Time
}

type serviceClock struct{}

func (serviceClock) Now() time.Time { return time.Now() }

// ServiceInterface defines the behaviour offered by the bets domain.
type ServiceInterface interface {
	Place(ctx context.Context, req PlaceRequest) (*PlacedBet, error)
	Sell(ctx context.Context, req SellRequest) (*SellResult, error)
}

// Service implements the bets domain logic.
type Service struct {
	repo    Repository
	markets MarketService
	users   UserService
	econ    *setup.EconomicConfig
	clock   Clock

	placeValidator PlaceValidator
	sellValidator  SellValidator

	marketGate     MarketGate
	fees           FeeCalculator
	balances       BalanceGuard
	ledger         BetLedger
	saleCalculator SaleCalculator
}

var (
	_ ServiceInterface = (*Service)(nil)
	_ SaleCalculator   = saleCalculator{}
)

// ServiceOption configures bets Service collaborators.
type ServiceOption func(*Service)

// WithPlaceValidator overrides the place validator.
func WithPlaceValidator(v PlaceValidator) ServiceOption {
	return func(s *Service) {
		if v != nil {
			s.placeValidator = v
		}
	}
}

// WithSellValidator overrides the sell validator.
func WithSellValidator(v SellValidator) ServiceOption {
	return func(s *Service) {
		if v != nil {
			s.sellValidator = v
		}
	}
}

// WithMarketGate overrides the market gate.
func WithMarketGate(g MarketGate) ServiceOption {
	return func(s *Service) {
		if g != nil {
			s.marketGate = g
		}
	}
}

// WithFeeCalculator overrides the fee calculator.
func WithFeeCalculator(c FeeCalculator) ServiceOption {
	return func(s *Service) {
		if c != nil {
			s.fees = c
		}
	}
}

// WithBalanceGuard overrides the balance guard.
func WithBalanceGuard(g BalanceGuard) ServiceOption {
	return func(s *Service) {
		if g != nil {
			s.balances = g
		}
	}
}

// WithBetLedger overrides the bet ledger.
func WithBetLedger(l BetLedger) ServiceOption {
	return func(s *Service) {
		if l != nil {
			s.ledger = l
		}
	}
}

// WithSaleCalculator overrides the sale calculator.
func WithSaleCalculator(c SaleCalculator) ServiceOption {
	return func(s *Service) {
		if c != nil {
			s.saleCalculator = c
		}
	}
}

// WithClock overrides the service clock.
func WithClock(clock Clock) ServiceOption {
	return func(s *Service) {
		if clock != nil {
			s.clock = clock
		}
	}
}

// NewService constructs a bets service.
func NewService(repo Repository, markets MarketService, users UserService, econ *setup.EconomicConfig, clock Clock, opts ...ServiceOption) *Service {
	s := &Service{
		repo:    repo,
		markets: markets,
		users:   users,
		econ:    econ,
		clock:   clock,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.ensureDefaults()
	return s
}

func (s *Service) ensureDefaults() {
	if s.clock == nil {
		s.clock = serviceClock{}
	}
	if s.placeValidator == nil {
		s.placeValidator = defaultPlaceValidator{}
	}
	if s.sellValidator == nil {
		s.sellValidator = defaultSellValidator{}
	}
	if s.marketGate == nil {
		s.marketGate = marketGate{markets: s.markets, clock: s.clock}
	}
	if s.fees == nil {
		s.fees = feeCalculator{econ: s.econ}
	}
	if s.balances == nil {
		s.balances = balanceGuard{maxDebtAllowed: int64(s.econ.Economics.User.MaximumDebtAllowed)}
	}
	if s.ledger == nil {
		s.ledger = betLedger{repo: s.repo, users: s.users}
	}
	if s.saleCalculator == nil {
		s.saleCalculator = saleCalculator{maxDustPerSale: int64(s.econ.Economics.Betting.MaxDustPerSale)}
	}
}
