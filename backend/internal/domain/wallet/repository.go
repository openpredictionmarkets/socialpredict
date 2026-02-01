package wallet

import "context"

// Repository defines the interface for wallet data access.
type Repository interface {
	// GetBalance retrieves the current balance for a user.
	GetBalance(ctx context.Context, username string) (int64, error)

	// UpdateBalance sets a new balance for a user.
	// Returns ErrAccountNotFound if the user doesn't exist.
	UpdateBalance(ctx context.Context, username string, newBalance int64) error

	// RecordTransaction persists a ledger entry for audit purposes.
	RecordTransaction(ctx context.Context, entry *LedgerEntry) error

	// UpdateBalanceAndRecord atomically updates the balance and records the transaction.
	// This should be done in a single database transaction.
	UpdateBalanceAndRecord(ctx context.Context, username string, newBalance int64, entry *LedgerEntry) error
}
