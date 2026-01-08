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

// UserService exposes the subset of user operations required by bets.
type UserService interface {
	GetUser(ctx context.Context, username string) (*dusers.User, error)
	ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error
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

	marketGate     marketGate
	fees           feeCalculator
	balances       balanceGuard
	ledger         betLedger
	saleCalculator saleCalculator
}

// NewService constructs a bets service.
func NewService(repo Repository, markets MarketService, users UserService, econ *setup.EconomicConfig, clock Clock) *Service {
	if clock == nil {
		clock = serviceClock{}
	}
	return &Service{
		repo:    repo,
		markets: markets,
		users:   users,
		econ:    econ,
		clock:   clock,

		marketGate:     marketGate{markets: markets, clock: clock},
		fees:           feeCalculator{econ: econ},
		balances:       balanceGuard{maxDebtAllowed: int64(econ.Economics.User.MaximumDebtAllowed)},
		ledger:         betLedger{repo: repo, users: users},
		saleCalculator: saleCalculator{maxDustPerSale: int64(econ.Economics.Betting.MaxDustPerSale)},
	}
}
