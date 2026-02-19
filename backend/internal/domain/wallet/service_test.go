package wallet_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"socialpredict/internal/domain/wallet"
)

// --- Test doubles ---

type fakeRepo struct {
	balances map[string]int64
	entries  []*wallet.LedgerEntry

	getBalanceErr error
	updateErr     error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		balances: make(map[string]int64),
	}
}

func (r *fakeRepo) GetBalance(_ context.Context, username string) (int64, error) {
	if r.getBalanceErr != nil {
		return 0, r.getBalanceErr
	}
	bal, ok := r.balances[username]
	if !ok {
		return 0, wallet.ErrAccountNotFound
	}
	return bal, nil
}

func (r *fakeRepo) UpdateBalance(_ context.Context, username string, newBalance int64) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	if _, ok := r.balances[username]; !ok {
		return wallet.ErrAccountNotFound
	}
	r.balances[username] = newBalance
	return nil
}

func (r *fakeRepo) RecordTransaction(_ context.Context, entry *wallet.LedgerEntry) error {
	r.entries = append(r.entries, entry)
	return nil
}

func (r *fakeRepo) UpdateBalanceAndRecord(_ context.Context, username string, newBalance int64, entry *wallet.LedgerEntry) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	if _, ok := r.balances[username]; !ok {
		return wallet.ErrAccountNotFound
	}
	r.balances[username] = newBalance
	r.entries = append(r.entries, entry)
	return nil
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

// --- Credit tests ---

func TestCredit_HappyPath(t *testing.T) {
	tests := []struct {
		name       string
		txType     string
		balance    int64
		amount     int64
		wantBal    int64
		wantLedger int64
	}{
		{"WIN credit", wallet.TxWin, 500, 100, 600, 100},
		{"REFUND credit", wallet.TxRefund, 0, 250, 250, 250},
		{"SALE credit", wallet.TxSale, 1000, 1, 1001, 1},
		{"credit to negative balance", wallet.TxWin, -200, 50, -150, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)
			repo := newFakeRepo()
			repo.balances["alice"] = tt.balance
			svc := wallet.NewService(repo, fixedClock{now: now})

			err := svc.Credit(context.Background(), "alice", tt.amount, tt.txType)
			if err != nil {
				t.Fatalf("Credit returned error: %v", err)
			}

			if repo.balances["alice"] != tt.wantBal {
				t.Fatalf("expected balance %d, got %d", tt.wantBal, repo.balances["alice"])
			}

			if len(repo.entries) != 1 {
				t.Fatalf("expected 1 ledger entry, got %d", len(repo.entries))
			}
			entry := repo.entries[0]
			if entry.Username != "alice" {
				t.Fatalf("expected username alice, got %s", entry.Username)
			}
			if entry.Amount != tt.wantLedger {
				t.Fatalf("expected ledger amount %d, got %d", tt.wantLedger, entry.Amount)
			}
			if entry.Type != tt.txType {
				t.Fatalf("expected type %s, got %s", tt.txType, entry.Type)
			}
			if entry.Balance != tt.wantBal {
				t.Fatalf("expected ledger balance %d, got %d", tt.wantBal, entry.Balance)
			}
			if !entry.CreatedAt.Equal(now) {
				t.Fatalf("expected CreatedAt %v, got %v", now, entry.CreatedAt)
			}
		})
	}
}

func TestCredit_InvalidAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount int64
	}{
		{"zero", 0},
		{"negative", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.balances["alice"] = 500
			svc := wallet.NewService(repo, fixedClock{})

			err := svc.Credit(context.Background(), "alice", tt.amount, wallet.TxWin)
			if !errors.Is(err, wallet.ErrInvalidAmount) {
				t.Fatalf("expected ErrInvalidAmount, got %v", err)
			}
			if len(repo.entries) != 0 {
				t.Fatalf("expected no ledger entries, got %d", len(repo.entries))
			}
		})
	}
}

func TestCredit_InvalidTransactionType(t *testing.T) {
	tests := []struct {
		name   string
		txType string
	}{
		{"debit type BUY", wallet.TxBuy},
		{"debit type FEE", wallet.TxFee},
		{"unknown type", "UNKNOWN"},
		{"empty type", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.balances["alice"] = 500
			svc := wallet.NewService(repo, fixedClock{})

			err := svc.Credit(context.Background(), "alice", 100, tt.txType)
			if !errors.Is(err, wallet.ErrInvalidTransaction) {
				t.Fatalf("expected ErrInvalidTransaction, got %v", err)
			}
			if len(repo.entries) != 0 {
				t.Fatalf("expected no ledger entries, got %d", len(repo.entries))
			}
		})
	}
}

func TestCredit_AccountNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := wallet.NewService(repo, fixedClock{})

	err := svc.Credit(context.Background(), "nonexistent", 100, wallet.TxWin)
	if !errors.Is(err, wallet.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestCredit_RepoFailure(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := newFakeRepo()
	repo.balances["alice"] = 500
	repo.updateErr = repoErr
	svc := wallet.NewService(repo, fixedClock{})

	err := svc.Credit(context.Background(), "alice", 100, wallet.TxWin)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

// --- Debit tests ---

func TestDebit_HappyPath(t *testing.T) {
	tests := []struct {
		name    string
		txType  string
		balance int64
		amount  int64
		maxDebt int64
		wantBal int64
	}{
		{"BUY with sufficient balance", wallet.TxBuy, 500, 100, 0, 400},
		{"FEE with sufficient balance", wallet.TxFee, 200, 50, 0, 150},
		{"BUY using debt allowance", wallet.TxBuy, 100, 200, 200, -100},
		{"BUY draining to zero", wallet.TxBuy, 100, 100, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)
			repo := newFakeRepo()
			repo.balances["alice"] = tt.balance
			svc := wallet.NewService(repo, fixedClock{now: now})

			err := svc.Debit(context.Background(), "alice", tt.amount, tt.maxDebt, tt.txType)
			if err != nil {
				t.Fatalf("Debit returned error: %v", err)
			}

			if repo.balances["alice"] != tt.wantBal {
				t.Fatalf("expected balance %d, got %d", tt.wantBal, repo.balances["alice"])
			}

			if len(repo.entries) != 1 {
				t.Fatalf("expected 1 ledger entry, got %d", len(repo.entries))
			}
			entry := repo.entries[0]
			if entry.Amount != -tt.amount {
				t.Fatalf("expected ledger amount %d, got %d", -tt.amount, entry.Amount)
			}
			if entry.Type != tt.txType {
				t.Fatalf("expected type %s, got %s", tt.txType, entry.Type)
			}
			if entry.Balance != tt.wantBal {
				t.Fatalf("expected ledger balance %d, got %d", tt.wantBal, entry.Balance)
			}
			if !entry.CreatedAt.Equal(now) {
				t.Fatalf("expected CreatedAt %v, got %v", now, entry.CreatedAt)
			}
		})
	}
}

func TestDebit_InsufficientBalance(t *testing.T) {
	tests := []struct {
		name    string
		balance int64
		amount  int64
		maxDebt int64
	}{
		{"exceeds balance no debt", 100, 200, 0},
		{"exceeds balance and debt", 100, 200, 50},
		{"zero balance no debt", 0, 1, 0},
		{"negative balance already at limit", -99, 2, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.balances["alice"] = tt.balance
			svc := wallet.NewService(repo, fixedClock{})

			err := svc.Debit(context.Background(), "alice", tt.amount, tt.maxDebt, wallet.TxBuy)
			if !errors.Is(err, wallet.ErrInsufficientBalance) {
				t.Fatalf("expected ErrInsufficientBalance, got %v", err)
			}
			if len(repo.entries) != 0 {
				t.Fatalf("expected no ledger entries on insufficient balance")
			}
		})
	}
}

func TestDebit_ExactlyAtDebtLimit(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 0
	svc := wallet.NewService(repo, fixedClock{})

	err := svc.Debit(context.Background(), "alice", 500, 500, wallet.TxBuy)
	if err != nil {
		t.Fatalf("expected success at exact debt limit, got %v", err)
	}
	if repo.balances["alice"] != -500 {
		t.Fatalf("expected balance -500, got %d", repo.balances["alice"])
	}
}

func TestDebit_InvalidAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount int64
	}{
		{"zero", 0},
		{"negative", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.balances["alice"] = 500
			svc := wallet.NewService(repo, fixedClock{})

			err := svc.Debit(context.Background(), "alice", tt.amount, 0, wallet.TxBuy)
			if !errors.Is(err, wallet.ErrInvalidAmount) {
				t.Fatalf("expected ErrInvalidAmount, got %v", err)
			}
			if len(repo.entries) != 0 {
				t.Fatalf("expected no ledger entries, got %d", len(repo.entries))
			}
		})
	}
}

func TestDebit_InvalidTransactionType(t *testing.T) {
	tests := []struct {
		name   string
		txType string
	}{
		{"credit type WIN", wallet.TxWin},
		{"credit type REFUND", wallet.TxRefund},
		{"credit type SALE", wallet.TxSale},
		{"unknown type", "UNKNOWN"},
		{"empty type", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.balances["alice"] = 500
			svc := wallet.NewService(repo, fixedClock{})

			err := svc.Debit(context.Background(), "alice", 100, 0, tt.txType)
			if !errors.Is(err, wallet.ErrInvalidTransaction) {
				t.Fatalf("expected ErrInvalidTransaction, got %v", err)
			}
			if len(repo.entries) != 0 {
				t.Fatalf("expected no ledger entries, got %d", len(repo.entries))
			}
		})
	}
}

func TestDebit_AccountNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := wallet.NewService(repo, fixedClock{})

	err := svc.Debit(context.Background(), "nonexistent", 100, 0, wallet.TxBuy)
	if !errors.Is(err, wallet.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestDebit_RepoFailure(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := newFakeRepo()
	repo.balances["alice"] = 500
	repo.updateErr = repoErr
	svc := wallet.NewService(repo, fixedClock{})

	err := svc.Debit(context.Background(), "alice", 100, 0, wallet.TxBuy)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

// --- ValidateBalance tests ---

func TestValidateBalance_Sufficient(t *testing.T) {
	tests := []struct {
		name    string
		balance int64
		amount  int64
		maxDebt int64
	}{
		{"balance covers amount", 500, 100, 0},
		{"exact match", 100, 100, 0},
		{"debt allowance covers shortfall", 50, 200, 200},
		{"at exact debt limit", 0, 300, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.balances["alice"] = tt.balance
			svc := wallet.NewService(repo, fixedClock{})

			err := svc.ValidateBalance(context.Background(), "alice", tt.amount, tt.maxDebt)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidateBalance_Insufficient(t *testing.T) {
	tests := []struct {
		name    string
		balance int64
		amount  int64
		maxDebt int64
	}{
		{"exceeds balance no debt", 100, 200, 0},
		{"exceeds balance and debt", 0, 500, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.balances["alice"] = tt.balance
			svc := wallet.NewService(repo, fixedClock{})

			err := svc.ValidateBalance(context.Background(), "alice", tt.amount, tt.maxDebt)
			if !errors.Is(err, wallet.ErrInsufficientBalance) {
				t.Fatalf("expected ErrInsufficientBalance, got %v", err)
			}
		})
	}
}

func TestValidateBalance_AccountNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := wallet.NewService(repo, fixedClock{})

	err := svc.ValidateBalance(context.Background(), "nonexistent", 100, 0)
	if !errors.Is(err, wallet.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

// --- GetBalance tests ---

func TestGetBalance_Success(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 750
	svc := wallet.NewService(repo, fixedClock{})

	bal, err := svc.GetBalance(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}
	if bal != 750 {
		t.Fatalf("expected balance 750, got %d", bal)
	}
}

func TestGetBalance_AccountNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := wallet.NewService(repo, fixedClock{})

	_, err := svc.GetBalance(context.Background(), "nonexistent")
	if !errors.Is(err, wallet.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

// --- GetCredit tests ---

func TestGetCredit_ExistingUser(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 200
	svc := wallet.NewService(repo, fixedClock{})

	credit, err := svc.GetCredit(context.Background(), "alice", 500)
	if err != nil {
		t.Fatalf("GetCredit returned error: %v", err)
	}
	if credit != 700 {
		t.Fatalf("expected credit 700, got %d", credit)
	}
}

func TestGetCredit_NegativeBalance(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = -200
	svc := wallet.NewService(repo, fixedClock{})

	credit, err := svc.GetCredit(context.Background(), "alice", 500)
	if err != nil {
		t.Fatalf("GetCredit returned error: %v", err)
	}
	if credit != 300 {
		t.Fatalf("expected credit 300, got %d", credit)
	}
}

func TestGetCredit_AccountNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := wallet.NewService(repo, fixedClock{})

	credit, err := svc.GetCredit(context.Background(), "nonexistent", 500)
	if err != nil {
		t.Fatalf("expected no error for missing account, got %v", err)
	}
	if credit != 500 {
		t.Fatalf("expected credit to equal maxDebt (500), got %d", credit)
	}
}

func TestGetCredit_OtherError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := newFakeRepo()
	repo.getBalanceErr = repoErr
	svc := wallet.NewService(repo, fixedClock{})

	credit, err := svc.GetCredit(context.Background(), "alice", 500)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if credit != 0 {
		t.Fatalf("expected credit 0 on error, got %d", credit)
	}
}

// --- Ledger timestamp test ---

func TestLedgerEntryTimestamp(t *testing.T) {
	now := time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)
	repo := newFakeRepo()
	repo.balances["alice"] = 1000
	svc := wallet.NewService(repo, fixedClock{now: now})

	if err := svc.Credit(context.Background(), "alice", 50, wallet.TxWin); err != nil {
		t.Fatalf("Credit returned error: %v", err)
	}

	if len(repo.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(repo.entries))
	}
	if !repo.entries[0].CreatedAt.Equal(now) {
		t.Fatalf("expected CreatedAt %v, got %v", now, repo.entries[0].CreatedAt)
	}
}

// --- Generic GetBalance error propagation ---

func TestCredit_GetBalanceGenericError(t *testing.T) {
	dbErr := errors.New("connection refused")
	repo := newFakeRepo()
	repo.getBalanceErr = dbErr
	svc := wallet.NewService(repo, fixedClock{})

	err := svc.Credit(context.Background(), "alice", 100, wallet.TxWin)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected generic GetBalance error, got %v", err)
	}
}

func TestDebit_GetBalanceGenericError(t *testing.T) {
	dbErr := errors.New("connection refused")
	repo := newFakeRepo()
	repo.getBalanceErr = dbErr
	svc := wallet.NewService(repo, fixedClock{})

	err := svc.Debit(context.Background(), "alice", 100, 0, wallet.TxBuy)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected generic GetBalance error, got %v", err)
	}
}

func TestValidateBalance_GetBalanceGenericError(t *testing.T) {
	dbErr := errors.New("connection refused")
	repo := newFakeRepo()
	repo.getBalanceErr = dbErr
	svc := wallet.NewService(repo, fixedClock{})

	err := svc.ValidateBalance(context.Background(), "alice", 100, 0)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected generic GetBalance error, got %v", err)
	}
}

// --- Balance unchanged on failure ---

func TestCredit_BalanceUnchangedOnInvalidAmount(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 500
	svc := wallet.NewService(repo, fixedClock{})

	_ = svc.Credit(context.Background(), "alice", 0, wallet.TxWin)
	if repo.balances["alice"] != 500 {
		t.Fatalf("expected balance unchanged at 500, got %d", repo.balances["alice"])
	}
}

func TestCredit_BalanceUnchangedOnInvalidTxType(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 500
	svc := wallet.NewService(repo, fixedClock{})

	_ = svc.Credit(context.Background(), "alice", 100, wallet.TxBuy)
	if repo.balances["alice"] != 500 {
		t.Fatalf("expected balance unchanged at 500, got %d", repo.balances["alice"])
	}
}

func TestCredit_BalanceUnchangedOnRepoUpdateError(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 500
	repo.updateErr = errors.New("write failed")
	svc := wallet.NewService(repo, fixedClock{})

	_ = svc.Credit(context.Background(), "alice", 100, wallet.TxWin)
	// fakeRepo doesn't update balance when updateErr is set
	if repo.balances["alice"] != 500 {
		t.Fatalf("expected balance unchanged at 500, got %d", repo.balances["alice"])
	}
}

func TestDebit_BalanceUnchangedOnInsufficientBalance(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 100
	svc := wallet.NewService(repo, fixedClock{})

	_ = svc.Debit(context.Background(), "alice", 200, 0, wallet.TxBuy)
	if repo.balances["alice"] != 100 {
		t.Fatalf("expected balance unchanged at 100, got %d", repo.balances["alice"])
	}
}

func TestDebit_BalanceUnchangedOnInvalidAmount(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 500
	svc := wallet.NewService(repo, fixedClock{})

	_ = svc.Debit(context.Background(), "alice", -1, 0, wallet.TxBuy)
	if repo.balances["alice"] != 500 {
		t.Fatalf("expected balance unchanged at 500, got %d", repo.balances["alice"])
	}
}

func TestDebit_BalanceUnchangedOnRepoUpdateError(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 500
	repo.updateErr = errors.New("write failed")
	svc := wallet.NewService(repo, fixedClock{})

	_ = svc.Debit(context.Background(), "alice", 100, 0, wallet.TxBuy)
	if repo.balances["alice"] != 500 {
		t.Fatalf("expected balance unchanged at 500, got %d", repo.balances["alice"])
	}
}

// --- GetCredit wrapped ErrAccountNotFound ---

func TestGetCredit_WrappedAccountNotFoundDoesNotFallback(t *testing.T) {
	// Service uses direct equality (==), so a wrapped ErrAccountNotFound
	// will NOT trigger the maxDebt fallback — it returns (0, error) instead.
	wrappedErr := fmt.Errorf("repo layer: %w", wallet.ErrAccountNotFound)
	repo := newFakeRepo()
	repo.getBalanceErr = wrappedErr
	svc := wallet.NewService(repo, fixedClock{})

	credit, err := svc.GetCredit(context.Background(), "alice", 500)
	if err == nil {
		t.Fatalf("expected error for wrapped ErrAccountNotFound, got nil")
	}
	if credit == 500 {
		t.Fatalf("expected wrapped error NOT to trigger maxDebt fallback, but got %d", credit)
	}
	if credit != 0 {
		t.Fatalf("expected credit 0, got %d", credit)
	}
}

// --- Negative maxDebt behavior ---

func TestDebit_NegativeMaxDebt(t *testing.T) {
	// With negative maxDebt, -maxDebt becomes positive, raising the minimum balance
	// threshold. newBalance must be >= -maxDebt (which is positive).
	repo := newFakeRepo()
	repo.balances["alice"] = 100
	svc := wallet.NewService(repo, fixedClock{})

	// balance=100, amount=50, newBalance=50, maxDebt=-1, check: 50 < 1 → false → succeeds
	// Large positive balance still passes even with negative maxDebt
	err := svc.Debit(context.Background(), "alice", 50, -1, wallet.TxBuy)
	if err != nil {
		t.Fatalf("expected success when newBalance exceeds threshold, got %v", err)
	}

	// But draining to zero fails: newBalance=0, check: 0 < 1 → true → insufficient
	repo2 := newFakeRepo()
	repo2.balances["bob"] = 10
	svc2 := wallet.NewService(repo2, fixedClock{})

	err = svc2.Debit(context.Background(), "bob", 10, -1, wallet.TxBuy)
	if !errors.Is(err, wallet.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance when draining to zero with negative maxDebt, got %v", err)
	}
	if repo2.balances["bob"] != 10 {
		t.Fatalf("expected balance unchanged at 10, got %d", repo2.balances["bob"])
	}
}

func TestValidateBalance_NegativeMaxDebt(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 10
	svc := wallet.NewService(repo, fixedClock{})

	// balance=10, amount=10, check: 0 < 1 → true → insufficient
	err := svc.ValidateBalance(context.Background(), "alice", 10, -1)
	if !errors.Is(err, wallet.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}

	// balance=10, amount=9, check: 1 < 1 → false → passes
	err = svc.ValidateBalance(context.Background(), "alice", 9, -1)
	if err != nil {
		t.Fatalf("expected success when remaining balance exceeds threshold, got %v", err)
	}
}

func TestGetCredit_NegativeMaxDebt(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 100
	svc := wallet.NewService(repo, fixedClock{})

	// credit = balance + maxDebt = 100 + (-50) = 50
	credit, err := svc.GetCredit(context.Background(), "alice", -50)
	if err != nil {
		t.Fatalf("GetCredit returned error: %v", err)
	}
	if credit != 50 {
		t.Fatalf("expected credit 50 (100 + -50), got %d", credit)
	}
}

// --- ValidateBalance with zero/negative amount ---

func TestValidateBalance_ZeroAmount(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 100
	svc := wallet.NewService(repo, fixedClock{})

	// No amount validation in ValidateBalance — balance(100) - 0 = 100, 100 < 0 → false → passes
	err := svc.ValidateBalance(context.Background(), "alice", 0, 0)
	if err != nil {
		t.Fatalf("expected no error for zero amount, got %v", err)
	}
}

func TestValidateBalance_NegativeAmount(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = 0
	svc := wallet.NewService(repo, fixedClock{})

	// No amount validation — balance(0) - (-100) = 100, 100 < 0 → false → passes
	// This effectively makes the balance appear larger.
	err := svc.ValidateBalance(context.Background(), "alice", -100, 0)
	if err != nil {
		t.Fatalf("expected no error for negative amount (no validation), got %v", err)
	}
}

// --- int64 overflow boundary tests ---

func TestCredit_Overflow(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = math.MaxInt64
	svc := wallet.NewService(repo, fixedClock{})

	// MaxInt64 + 1 overflows to negative — service doesn't guard against this
	err := svc.Credit(context.Background(), "alice", 1, wallet.TxWin)
	if err != nil {
		t.Fatalf("Credit returned error: %v", err)
	}
	// Verify overflow actually happened (documents the behavior)
	if repo.balances["alice"] != math.MinInt64 {
		t.Fatalf("expected overflow to MinInt64, got %d", repo.balances["alice"])
	}
}

func TestDebit_Overflow(t *testing.T) {
	repo := newFakeRepo()
	repo.balances["alice"] = math.MinInt64
	svc := wallet.NewService(repo, fixedClock{})

	// MinInt64 - 1 overflows to MaxInt64 — would pass the < -maxDebt check
	// because MaxInt64 is not < 0
	err := svc.Debit(context.Background(), "alice", 1, 0, wallet.TxBuy)
	if err != nil {
		t.Fatalf("Debit returned error: %v", err)
	}
	// Documents overflow behavior
	if repo.balances["alice"] != math.MaxInt64 {
		t.Fatalf("expected overflow to MaxInt64, got %d", repo.balances["alice"])
	}
}
