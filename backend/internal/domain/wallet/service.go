package wallet

import (
	"context"
	"time"
)

// Clock provides time functionality for testability.
type Clock interface {
	Now() time.Time
}

// Service defines the interface for wallet operations.
type Service interface {
	// Credit adds funds to a user's balance.
	// Returns ErrInvalidAmount if amount <= 0.
	// Returns ErrAccountNotFound if user doesn't exist.
	Credit(ctx context.Context, username string, amount int64, txType string) error

	// Debit removes funds from a user's balance.
	// Returns ErrInvalidAmount if amount <= 0.
	// Returns ErrInsufficientBalance if balance would go below -maxDebt.
	// Returns ErrAccountNotFound if user doesn't exist.
	Debit(ctx context.Context, username string, amount int64, maxDebt int64, txType string) error

	// ValidateBalance checks if a user has sufficient balance for an operation.
	// Returns ErrInsufficientBalance if balance - amount < -maxDebt.
	// Returns ErrAccountNotFound if user doesn't exist.
	ValidateBalance(ctx context.Context, username string, amount int64, maxDebt int64) error

	// GetBalance returns the current balance for a user.
	// Returns ErrAccountNotFound if user doesn't exist.
	GetBalance(ctx context.Context, username string) (int64, error)

	// GetCredit returns the available credit (balance + maxDebt) for a user.
	// Returns maxDebt if user doesn't exist.
	GetCredit(ctx context.Context, username string, maxDebt int64) (int64, error)
}

// WalletService implements the Service interface.
type WalletService struct {
	repo  Repository
	clock Clock
}

// NewService creates a new wallet service.
func NewService(repo Repository, clock Clock) *WalletService {
	return &WalletService{
		repo:  repo,
		clock: clock,
	}
}

// Ensure WalletService implements Service.
var _ Service = (*WalletService)(nil)

// Credit adds funds to a user's balance.
func (s *WalletService) Credit(ctx context.Context, username string, amount int64, txType string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if !IsCreditType(txType) {
		return ErrInvalidTransaction
	}

	balance, err := s.repo.GetBalance(ctx, username)
	if err != nil {
		return err
	}

	newBalance := balance + amount
	entry := &LedgerEntry{
		Username:  username,
		Amount:    amount,
		Type:      txType,
		Balance:   newBalance,
		CreatedAt: s.clock.Now(),
	}

	return s.repo.UpdateBalanceAndRecord(ctx, username, newBalance, entry)
}

// Debit removes funds from a user's balance.
func (s *WalletService) Debit(ctx context.Context, username string, amount int64, maxDebt int64, txType string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if !IsDebitType(txType) {
		return ErrInvalidTransaction
	}

	balance, err := s.repo.GetBalance(ctx, username)
	if err != nil {
		return err
	}

	newBalance := balance - amount
	if newBalance < -maxDebt {
		return ErrInsufficientBalance
	}

	entry := &LedgerEntry{
		Username:  username,
		Amount:    -amount, // negative for debit
		Type:      txType,
		Balance:   newBalance,
		CreatedAt: s.clock.Now(),
	}

	return s.repo.UpdateBalanceAndRecord(ctx, username, newBalance, entry)
}

// ValidateBalance checks if a user has sufficient balance for an operation.
func (s *WalletService) ValidateBalance(ctx context.Context, username string, amount int64, maxDebt int64) error {
	balance, err := s.repo.GetBalance(ctx, username)
	if err != nil {
		return err
	}

	if balance-amount < -maxDebt {
		return ErrInsufficientBalance
	}

	return nil
}

// GetBalance returns the current balance for a user.
func (s *WalletService) GetBalance(ctx context.Context, username string) (int64, error) {
	return s.repo.GetBalance(ctx, username)
}

// GetCredit returns the available credit (balance + maxDebt) for a user.
func (s *WalletService) GetCredit(ctx context.Context, username string, maxDebt int64) (int64, error) {
	balance, err := s.repo.GetBalance(ctx, username)
	if err != nil {
		if err == ErrAccountNotFound {
			return maxDebt, nil
		}
		return 0, err
	}

	return balance + maxDebt, nil
}
