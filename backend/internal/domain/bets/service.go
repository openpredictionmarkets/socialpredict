package bets

import (
	"context"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
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

// WalletService exposes balance checks and mutations needed by the bets domain.
// This is the single balance mutation dependency for bets.
type WalletService interface {
	ValidateBalance(ctx context.Context, username string, amount int64, maxDebt int64) error
	Debit(ctx context.Context, username string, amount int64, maxDebt int64, txType string) error
	Credit(ctx context.Context, username string, amount int64, txType string) error
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
	wallet  WalletService
	econ    *setup.EconomicConfig
	clock   Clock

	marketGate     marketGate
	fees           feeCalculator
	balances       balanceGuard
	ledger         betLedger
	saleCalculator saleCalculator
}

// NewServiceWithWallet constructs a bets service with explicit wallet dependency.
func NewServiceWithWallet(repo Repository, markets MarketService, wallet WalletService, econ *setup.EconomicConfig, clock Clock) *Service {
	if clock == nil {
		clock = serviceClock{}
	}
	return &Service{
		repo:    repo,
		markets: markets,
		wallet:  wallet,
		econ:    econ,
		clock:   clock,

		marketGate:     marketGate{markets: markets, clock: clock},
		fees:           feeCalculator{econ: econ},
		balances:       balanceGuard{maxDebtAllowed: int64(econ.Economics.User.MaximumDebtAllowed)},
		ledger:         betLedger{repo: repo, wallet: wallet, maxDebtAllowed: int64(econ.Economics.User.MaximumDebtAllowed)},
		saleCalculator: saleCalculator{maxDustPerSale: int64(econ.Economics.Betting.MaxDustPerSale)},
	}
}
