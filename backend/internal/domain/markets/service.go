package markets

import (
	"context"
	"errors"

	dusers "socialpredict/internal/domain/users"
	dwallet "socialpredict/internal/domain/wallet"
)

// Constants for validation limits.
const (
	MaxQuestionTitleLength = 160
	MaxDescriptionLength   = 2000
	MaxLabelLength         = 20
	MinLabelLength         = 1
)

// Config holds configuration for the markets service.
type Config struct {
	MinimumFutureHours float64
	CreateMarketCost   int64
	MaximumDebtAllowed int64
}

// Service implements the core market business logic.
type Service struct {
	repo                  Repository
	creatorProfileService CreatorProfileService
	walletService         WalletService
	clock                 Clock
	config                Config
}

// userWalletAdapter temporarily adapts the legacy users service to the wallet port.
// This keeps constructor wiring stable while markets migrates to wallet.Service.
type userWalletAdapter struct {
	users UserService
}

var _ WalletService = userWalletAdapter{}

func (a userWalletAdapter) ValidateBalance(ctx context.Context, username string, amount int64, maxDebt int64) error {
	if err := a.users.ValidateUserBalance(ctx, username, amount, maxDebt); err != nil {
		if errors.Is(err, dusers.ErrInsufficientBalance) {
			return dwallet.ErrInsufficientBalance
		}
		return err
	}
	return nil
}

func (a userWalletAdapter) Debit(ctx context.Context, username string, amount int64, maxDebt int64, _ string) error {
	if err := a.users.ValidateUserBalance(ctx, username, amount, maxDebt); err != nil {
		if errors.Is(err, dusers.ErrInsufficientBalance) {
			return dwallet.ErrInsufficientBalance
		}
		return err
	}
	// Legacy users service only supports raw balance deduction for debits.
	return a.users.DeductBalance(ctx, username, amount)
}

func (a userWalletAdapter) Credit(ctx context.Context, username string, amount int64, txType string) error {
	return a.users.ApplyTransaction(ctx, username, amount, txType)
}

// NewService creates a new markets service.
func NewService(repo Repository, userService UserService, clock Clock, config Config) *Service {
	var walletService WalletService
	if userService != nil {
		walletService = userWalletAdapter{users: userService}
	}
	return &Service{
		repo:                  repo,
		creatorProfileService: userService,
		walletService:         walletService,
		clock:                 clock,
		config:                config,
	}
}

// Compile-time interface compliance checks.
var (
	_ ServiceInterface   = (*Service)(nil)
	_ CoreService        = (*Service)(nil)
	_ SearchService      = (*Service)(nil)
	_ LeaderboardService = (*Service)(nil)
	_ PositionsService   = (*Service)(nil)
	_ ProjectionService  = (*Service)(nil)
)
