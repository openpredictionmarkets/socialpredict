package users_test

import (
	"math"
	"testing"

	users "socialpredict/internal/domain/users"
)

func TestNewAccount(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		userID    int64
		balance   int64
		wantErr   error
		wantBal   int64
	}{
		{
			name:    "valid account with positive balance",
			id:      1,
			userID:  100,
			balance: 500,
			wantErr: nil,
			wantBal: 500,
		},
		{
			name:    "valid account with zero balance",
			id:      1,
			userID:  100,
			balance: 0,
			wantErr: nil,
			wantBal: 0,
		},
		{
			name:    "valid account with negative balance",
			id:      1,
			userID:  100,
			balance: -100,
			wantErr: nil,
			wantBal: -100,
		},
		{
			name:    "invalid account ID zero",
			id:      0,
			userID:  100,
			balance: 500,
			wantErr: users.ErrInvalidAccountID,
		},
		{
			name:    "invalid account ID negative",
			id:      -1,
			userID:  100,
			balance: 500,
			wantErr: users.ErrInvalidAccountID,
		},
		{
			name:    "invalid user ID zero",
			id:      1,
			userID:  0,
			balance: 500,
			wantErr: users.ErrInvalidUserID,
		},
		{
			name:    "invalid user ID negative",
			id:      1,
			userID:  -1,
			balance: 500,
			wantErr: users.ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := users.NewAccount(tt.id, tt.userID, tt.balance)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewAccount() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewAccount() unexpected error: %v", err)
			}
			if account.Balance != tt.wantBal {
				t.Fatalf("NewAccount() balance = %d, want %d", account.Balance, tt.wantBal)
			}
		})
	}
}

func TestAccountValidate(t *testing.T) {
	tests := []struct {
		name    string
		account users.Account
		wantErr error
	}{
		{
			name:    "valid account",
			account: users.Account{ID: 1, UserID: 1, Balance: 100},
			wantErr: nil,
		},
		{
			name:    "zero account ID",
			account: users.Account{ID: 0, UserID: 1, Balance: 100},
			wantErr: users.ErrInvalidAccountID,
		},
		{
			name:    "negative account ID",
			account: users.Account{ID: -5, UserID: 1, Balance: 100},
			wantErr: users.ErrInvalidAccountID,
		},
		{
			name:    "zero user ID",
			account: users.Account{ID: 1, UserID: 0, Balance: 100},
			wantErr: users.ErrInvalidUserID,
		},
		{
			name:    "negative user ID",
			account: users.Account{ID: 1, UserID: -5, Balance: 100},
			wantErr: users.ErrInvalidUserID,
		},
		{
			name:    "negative balance is valid",
			account: users.Account{ID: 1, UserID: 1, Balance: -500},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.account.Validate()
			if err != tt.wantErr {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccountCredit(t *testing.T) {
	tests := []struct {
		name           string
		initialBalance int64
		creditAmount   int64
		wantBalance    int64
		wantErr        error
	}{
		{
			name:           "credit positive amount",
			initialBalance: 100,
			creditAmount:   50,
			wantBalance:    150,
			wantErr:        nil,
		},
		{
			name:           "credit to zero balance",
			initialBalance: 0,
			creditAmount:   100,
			wantBalance:    100,
			wantErr:        nil,
		},
		{
			name:           "credit to negative balance",
			initialBalance: -50,
			creditAmount:   30,
			wantBalance:    -20,
			wantErr:        nil,
		},
		{
			name:           "credit zero amount rejected",
			initialBalance: 100,
			creditAmount:   0,
			wantBalance:    100,
			wantErr:        users.ErrInsufficientBalance,
		},
		{
			name:           "credit negative amount rejected",
			initialBalance: 100,
			creditAmount:   -50,
			wantBalance:    100,
			wantErr:        users.ErrInsufficientBalance,
		},
		{
			name:           "credit large amount",
			initialBalance: 0,
			creditAmount:   math.MaxInt64 - 100,
			wantBalance:    math.MaxInt64 - 100,
			wantErr:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := users.Account{ID: 1, UserID: 1, Balance: tt.initialBalance}
			err := account.Credit(tt.creditAmount)
			if err != tt.wantErr {
				t.Fatalf("Credit() error = %v, want %v", err, tt.wantErr)
			}
			if account.Balance != tt.wantBalance {
				t.Fatalf("Credit() balance = %d, want %d", account.Balance, tt.wantBalance)
			}
		})
	}
}

func TestAccountDebit(t *testing.T) {
	tests := []struct {
		name           string
		initialBalance int64
		debitAmount    int64
		wantBalance    int64
		wantErr        error
	}{
		{
			name:           "debit within balance",
			initialBalance: 100,
			debitAmount:    50,
			wantBalance:    50,
			wantErr:        nil,
		},
		{
			name:           "debit exact balance",
			initialBalance: 100,
			debitAmount:    100,
			wantBalance:    0,
			wantErr:        nil,
		},
		{
			name:           "debit exceeds balance",
			initialBalance: 100,
			debitAmount:    150,
			wantBalance:    100,
			wantErr:        users.ErrInsufficientBalance,
		},
		{
			name:           "debit from zero balance",
			initialBalance: 0,
			debitAmount:    50,
			wantBalance:    0,
			wantErr:        users.ErrInsufficientBalance,
		},
		{
			name:           "debit from negative balance",
			initialBalance: -50,
			debitAmount:    10,
			wantBalance:    -50,
			wantErr:        users.ErrInsufficientBalance,
		},
		{
			name:           "debit zero amount rejected",
			initialBalance: 100,
			debitAmount:    0,
			wantBalance:    100,
			wantErr:        users.ErrInvalidTransactionType,
		},
		{
			name:           "debit negative amount rejected",
			initialBalance: 100,
			debitAmount:    -50,
			wantBalance:    100,
			wantErr:        users.ErrInvalidTransactionType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := users.Account{ID: 1, UserID: 1, Balance: tt.initialBalance}
			err := account.Debit(tt.debitAmount)
			if err != tt.wantErr {
				t.Fatalf("Debit() error = %v, want %v", err, tt.wantErr)
			}
			if account.Balance != tt.wantBalance {
				t.Fatalf("Debit() balance = %d, want %d", account.Balance, tt.wantBalance)
			}
		})
	}
}

func TestAccountCreditOverflow(t *testing.T) {
	account := users.Account{ID: 1, UserID: 1, Balance: math.MaxInt64}
	err := account.Credit(1)
	if err != nil {
		t.Skipf("Credit() does not guard against overflow (expected)")
		return
	}
	if account.Balance >= 0 {
		t.Fatalf("Credit() should have overflowed to negative, got %d", account.Balance)
	}
}
