package bets

import (
	"context"
	"time"

	"socialpredict/internal/domain/boundary"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/setup"
)

// Repository exposes the persistence layer needed by the bets domain service.
type BetWriter interface {
	Create(ctx context.Context, bet *boundary.Bet) error
}

// BetHistoryReader exposes prior-participation lookups used for buy fee rules.
type BetHistoryReader interface {
	UserHasBet(ctx context.Context, marketID uint, username string) (bool, error)
}

// Repository exposes the persistence layer needed by the bets domain service.
type Repository interface {
	BetWriter
	BetHistoryReader
}

// MarketService exposes the subset of market operations required by bets.
type MarketReader interface {
	GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error)
}

// PositionReader exposes the position reads needed by share-sale flows.
type PositionReader interface {
	GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)
}

// MarketService exposes the subset of market operations required by bets.
type MarketService interface {
	MarketReader
	PositionReader
}

// MarketGate ensures market openness before betting operations.
type MarketGate interface {
	Open(ctx context.Context, marketID int64) (*dmarkets.Market, error)
}

// UserService exposes the subset of user operations required by bets.
type UserReader interface {
	GetUser(ctx context.Context, username string) (*dusers.User, error)
}

// TransactionRecorder exposes account mutations used by the betting ledger.
type TransactionRecorder interface {
	ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error
}

// UserService exposes the subset of user operations required by bets.
type UserService interface {
	UserReader
	TransactionRecorder
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
	ChargeAndRecord(ctx context.Context, bet *boundary.Bet, totalCost int64) error
	CreditSale(ctx context.Context, bet *boundary.Bet, saleValue int64) error
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

func defaultClock() Clock {
	return serviceClock{}
}

func defaultPlaceValidatorStrategy() PlaceValidator {
	return defaultPlaceValidator{}
}

func defaultSellValidatorStrategy() SellValidator {
	return defaultSellValidator{}
}

func defaultMarketGateStrategy(markets MarketReader, clock Clock) MarketGate {
	return marketGate{markets: markets, clock: clock}
}

func defaultFeeCalculatorStrategy(econ *setup.EconomicConfig) FeeCalculator {
	return feeCalculator{econ: econOrDefault(econ)}
}

func defaultBalanceGuardStrategy(econ *setup.EconomicConfig) BalanceGuard {
	econ = econOrDefault(econ)
	return balanceGuard{maxDebtAllowed: int64(econ.Economics.User.MaximumDebtAllowed)}
}

func defaultBetLedgerStrategy(repo BetWriter, users TransactionRecorder) BetLedger {
	return betLedger{repo: repo, users: users}
}

func defaultSaleCalculatorStrategy(econ *setup.EconomicConfig) SaleCalculator {
	econ = econOrDefault(econ)
	return saleCalculator{maxDustPerSale: int64(econ.Economics.Betting.MaxDustPerSale)}
}

func clockOrDefault(clock Clock) Clock {
	if clock == nil {
		return defaultClock()
	}
	return clock
}

func econOrDefault(econ *setup.EconomicConfig) *setup.EconomicConfig {
	if econ == nil {
		return &setup.EconomicConfig{}
	}
	return econ
}

func placeValidatorOrDefault(v PlaceValidator) PlaceValidator {
	if v == nil {
		return defaultPlaceValidatorStrategy()
	}
	return v
}

func sellValidatorOrDefault(v SellValidator) SellValidator {
	if v == nil {
		return defaultSellValidatorStrategy()
	}
	return v
}

func marketGateOrDefault(g MarketGate, markets MarketReader, clock Clock) MarketGate {
	if g == nil {
		return defaultMarketGateStrategy(markets, clock)
	}
	return g
}

func feeCalculatorOrDefault(c FeeCalculator, econ *setup.EconomicConfig) FeeCalculator {
	if c == nil {
		return defaultFeeCalculatorStrategy(econ)
	}
	return c
}

func balanceGuardOrDefault(g BalanceGuard, econ *setup.EconomicConfig) BalanceGuard {
	if g == nil {
		return defaultBalanceGuardStrategy(econ)
	}
	return g
}

func betLedgerOrDefault(l BetLedger, repo BetWriter, users TransactionRecorder) BetLedger {
	if l == nil {
		return defaultBetLedgerStrategy(repo, users)
	}
	return l
}

func saleCalculatorOrDefault(c SaleCalculator, econ *setup.EconomicConfig) SaleCalculator {
	if c == nil {
		return defaultSaleCalculatorStrategy(econ)
	}
	return c
}

// WithPlaceValidator overrides the place validator.
func WithPlaceValidator(v PlaceValidator) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.placeValidator = placeValidatorOrDefault(v)
		}
	}
}

// WithSellValidator overrides the sell validator.
func WithSellValidator(v SellValidator) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.sellValidator = sellValidatorOrDefault(v)
		}
	}
}

// WithMarketGate overrides the market gate.
func WithMarketGate(g MarketGate) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.marketGate = marketGateOrDefault(g, s.markets, clockOrDefault(s.clock))
		}
	}
}

// WithFeeCalculator overrides the fee calculator.
func WithFeeCalculator(c FeeCalculator) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.fees = feeCalculatorOrDefault(c, s.econ)
		}
	}
}

// WithBalanceGuard overrides the balance guard.
func WithBalanceGuard(g BalanceGuard) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.balances = balanceGuardOrDefault(g, s.econ)
		}
	}
}

// WithBetLedger overrides the bet ledger.
func WithBetLedger(l BetLedger) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.ledger = betLedgerOrDefault(l, s.repo, s.users)
		}
	}
}

// WithSaleCalculator overrides the sale calculator.
func WithSaleCalculator(c SaleCalculator) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.saleCalculator = saleCalculatorOrDefault(c, s.econ)
		}
	}
}

// WithClock overrides the service clock.
func WithClock(clock Clock) ServiceOption {
	return func(s *Service) {
		if s != nil {
			s.clock = clockOrDefault(clock)
		}
	}
}

// NewService constructs a bets service.
func NewService(repo Repository, markets MarketService, users UserService, econ *setup.EconomicConfig, clock Clock, opts ...ServiceOption) *Service {
	s := &Service{
		repo:    repo,
		markets: markets,
		users:   users,
		econ:    econOrDefault(econ),
		clock:   clockOrDefault(clock),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.ensureDefaults()
	return s
}

func (s *Service) ensureDefaults() {
	if s == nil {
		return
	}

	s.clock = clockOrDefault(s.clock)
	s.placeValidator = placeValidatorOrDefault(s.placeValidator)
	s.sellValidator = sellValidatorOrDefault(s.sellValidator)
	s.marketGate = marketGateOrDefault(s.marketGate, s.markets, s.clock)
	s.fees = feeCalculatorOrDefault(s.fees, s.econ)
	s.balances = balanceGuardOrDefault(s.balances, s.econ)
	s.ledger = betLedgerOrDefault(s.ledger, s.repo, s.users)
	s.saleCalculator = saleCalculatorOrDefault(s.saleCalculator, s.econ)
}
