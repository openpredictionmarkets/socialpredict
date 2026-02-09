package bets

import (
	"context"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	dwallet "socialpredict/internal/domain/wallet"
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

// WalletService exposes balance checks and mutations needed by the bets domain.
// This will become the primary integration point as balance logic migrates out of users.
type WalletService interface {
	ValidateBalance(ctx context.Context, username string, amount int64, maxDebt int64) error
	Debit(ctx context.Context, username string, amount int64, maxDebt int64, txType string) error
	Credit(ctx context.Context, username string, amount int64, txType string) error
}

// userWalletAdapter temporarily adapts the legacy users service to the wallet port.
// It keeps constructor wiring stable while callers migrate to wallet.Service.
type userWalletAdapter struct {
	users UserService
}

var _ WalletService = userWalletAdapter{}

func (a userWalletAdapter) ValidateBalance(ctx context.Context, username string, amount int64, maxDebt int64) error {
	user, err := a.users.GetUser(ctx, username)
	if err != nil {
		return err
	}
	if user.AccountBalance-amount < -maxDebt {
		return dwallet.ErrInsufficientBalance
	}
	return nil
}

func (a userWalletAdapter) Debit(ctx context.Context, username string, amount int64, _ int64, txType string) error {
	return a.users.ApplyTransaction(ctx, username, amount, txType)
}

func (a userWalletAdapter) Credit(ctx context.Context, username string, amount int64, txType string) error {
	return a.users.ApplyTransaction(ctx, username, amount, txType)
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
	wallet  WalletService
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
		wallet:  userWalletAdapter{users: users},
		econ:    econ,
		clock:   clock,

		marketGate:     marketGate{markets: markets, clock: clock},
		fees:           feeCalculator{econ: econ},
		balances:       balanceGuard{maxDebtAllowed: int64(econ.Economics.User.MaximumDebtAllowed)},
		ledger:         betLedger{repo: repo, users: users},
		saleCalculator: saleCalculator{maxDustPerSale: int64(econ.Economics.Betting.MaxDustPerSale)},
	}
}
