package users

import "time"

// Account holds balance state for a user. Auth, identity, and market data live elsewhere.
type Account struct {
	ID      int64
	UserID  int64
	Balance int64
}

// LedgerEntry captures a single balance-affecting operation for auditing.
type LedgerEntry struct {
	AccountID int64
	Amount    int64
	Kind      string // e.g. credit, debit, win, refund, fee
	CreatedAt time.Time
}

// NewAccount constructs an account with a starting balance.
func NewAccount(id, userID, balance int64) (Account, error) {
	a := Account{
		ID:      id,
		UserID:  userID,
		Balance: balance,
	}
	if err := a.Validate(); err != nil {
		return Account{}, err
	}
	return a, nil
}

// Validate performs basic sanity checks.
func (a Account) Validate() error {
	if a.ID <= 0 {
		return ErrInvalidAccountID
	}
	if a.UserID <= 0 {
		return ErrInvalidUserID
	}
	return nil
}

// Credit increases the balance by amount.
func (a *Account) Credit(amount int64) error {
	if amount <= 0 {
		return ErrInsufficientBalance
	}
	a.Balance += amount
	return nil
}

// Debit decreases the balance by amount if funds are available.
func (a *Account) Debit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidTransactionType
	}
	if amount > a.Balance {
		return ErrInsufficientBalance
	}
	a.Balance -= amount
	return nil
}
