package wallet

import "errors"

// Account holds balance state for a user. Auth, identity, and market data live elsewhere.
type Account struct {
    ID       int64
    UserID   int64
    Balance  int64
    Currency string
}

var (
    ErrInvalidAccountID  = errors.New("account id must be positive")
    ErrInvalidUserID     = errors.New("user id must be positive")
    ErrInvalidAmount     = errors.New("amount must be positive")
    ErrInsufficientFunds = errors.New("insufficient funds")
)

// NewAccount constructs an account with a starting balance.
func NewAccount(id, userID, balance int64, currency string) (Account, error) {
    a := Account{
        ID:       id,
        UserID:   userID,
        Balance:  balance,
        Currency: currency,
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
        return ErrInvalidAmount
    }
    a.Balance += amount
    return nil
}

// Debit decreases the balance by amount if funds are available.
func (a *Account) Debit(amount int64) error {
    if amount <= 0 {
        return ErrInvalidAmount
    }
    if amount > a.Balance {
        return ErrInsufficientFunds
    }
    a.Balance -= amount
    return nil
}
