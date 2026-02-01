package wallet_test

import (
	"context"
	"testing"
	"time"

	dwallet "socialpredict/internal/domain/wallet"
	rwallet "socialpredict/internal/repository/wallet"
	"socialpredict/models/modelstesting"
)

func TestGormRepositoryGetBalance(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rwallet.NewGormRepository(db)

	user := modelstesting.GenerateUser("balance_user", 500)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	balance, err := repo.GetBalance(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}
	if balance != 500 {
		t.Fatalf("balance = %d, want 500", balance)
	}
}

func TestGormRepositoryGetBalanceMissingUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rwallet.NewGormRepository(db)

	ctx := context.Background()

	_, err := repo.GetBalance(ctx, "nonexistent")
	if err != dwallet.ErrAccountNotFound {
		t.Fatalf("GetBalance error = %v, want ErrAccountNotFound", err)
	}
}

func TestGormRepositoryUpdateBalance(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rwallet.NewGormRepository(db)

	user := modelstesting.GenerateUser("update_user", 100)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	if err := repo.UpdateBalance(ctx, user.Username, 250); err != nil {
		t.Fatalf("UpdateBalance returned error: %v", err)
	}

	balance, err := repo.GetBalance(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}
	if balance != 250 {
		t.Fatalf("balance = %d, want 250", balance)
	}
}

func TestGormRepositoryUpdateBalanceMissingUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rwallet.NewGormRepository(db)

	ctx := context.Background()

	err := repo.UpdateBalance(ctx, "nonexistent", 100)
	if err != dwallet.ErrAccountNotFound {
		t.Fatalf("UpdateBalance error = %v, want ErrAccountNotFound", err)
	}
}

func TestGormRepositoryRecordTransaction(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rwallet.NewGormRepository(db)

	user := modelstesting.GenerateUser("ledger_user", 100)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	entry := &dwallet.LedgerEntry{
		Username:  user.Username,
		Amount:    50,
		Type:      dwallet.TxWin,
		Balance:   150,
		CreatedAt: time.Now(),
	}

	if err := repo.RecordTransaction(ctx, entry); err != nil {
		t.Fatalf("RecordTransaction returned error: %v", err)
	}

	// Verify entry was created
	var count int64
	db.Table("ledger_entries").Where("username = ?", user.Username).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 ledger entry, got %d", count)
	}
}

func TestGormRepositoryUpdateBalanceAndRecord(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rwallet.NewGormRepository(db)

	user := modelstesting.GenerateUser("atomic_user", 100)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	entry := &dwallet.LedgerEntry{
		Username:  user.Username,
		Amount:    -30,
		Type:      dwallet.TxBuy,
		Balance:   70,
		CreatedAt: time.Now(),
	}

	if err := repo.UpdateBalanceAndRecord(ctx, user.Username, 70, entry); err != nil {
		t.Fatalf("UpdateBalanceAndRecord returned error: %v", err)
	}

	// Verify balance was updated
	balance, err := repo.GetBalance(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}
	if balance != 70 {
		t.Fatalf("balance = %d, want 70", balance)
	}

	// Verify ledger entry was created
	var count int64
	db.Table("ledger_entries").Where("username = ?", user.Username).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 ledger entry, got %d", count)
	}

	// Verify entry ID was set
	if entry.ID == 0 {
		t.Fatalf("expected entry ID to be set, got 0")
	}
}

func TestGormRepositoryUpdateBalanceAndRecordMissingUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rwallet.NewGormRepository(db)

	ctx := context.Background()

	entry := &dwallet.LedgerEntry{
		Username:  "nonexistent",
		Amount:    50,
		Type:      dwallet.TxWin,
		Balance:   50,
		CreatedAt: time.Now(),
	}

	err := repo.UpdateBalanceAndRecord(ctx, "nonexistent", 50, entry)
	if err != dwallet.ErrAccountNotFound {
		t.Fatalf("UpdateBalanceAndRecord error = %v, want ErrAccountNotFound", err)
	}

	// Verify no ledger entry was created (transaction rolled back)
	var count int64
	db.Table("ledger_entries").Where("username = ?", "nonexistent").Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 ledger entries after rollback, got %d", count)
	}
}

func TestGormRepositoryUpdateBalanceNegative(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rwallet.NewGormRepository(db)

	user := modelstesting.GenerateUser("negative_user", 100)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	// Update to negative balance (debt)
	if err := repo.UpdateBalance(ctx, user.Username, -50); err != nil {
		t.Fatalf("UpdateBalance returned error: %v", err)
	}

	balance, err := repo.GetBalance(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}
	if balance != -50 {
		t.Fatalf("balance = %d, want -50", balance)
	}
}
