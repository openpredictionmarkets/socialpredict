package users_test

import (
	"context"
	"testing"

	users "socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	"socialpredict/models/modelstesting"
)

func TestServiceApplyTransaction(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo)

	user := modelstesting.GenerateUser("tx_user", 0)
	user.AccountBalance = 100
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	tests := []struct {
		name        string
		txType      string
		amount      int64
		wantBalance int64
		wantErr     bool
	}{
		{"win adds funds", users.TransactionWin, 50, 150, false},
		{"refund adds funds", users.TransactionRefund, 30, 180, false},
		{"sale adds funds", users.TransactionSale, 20, 200, false},
		{"buy subtracts funds", users.TransactionBuy, 40, 160, false},
		{"fee subtracts funds", users.TransactionFee, 10, 150, false},
		{"invalid type", "UNKNOWN", 5, 150, true},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ApplyTransaction(ctx, user.Username, tt.amount, tt.txType)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ApplyTransaction returned error: %v", err)
			}

			var updatedBalance int64
			if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&updatedBalance).Error; err != nil {
				t.Fatalf("scan balance: %v", err)
			}
			if updatedBalance != tt.wantBalance {
				t.Fatalf("balance = %d, want %d", updatedBalance, tt.wantBalance)
			}
		})
	}
}

func TestServiceGetUserCredit(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo)

	user := modelstesting.GenerateUser("credit_user", 0)
	user.AccountBalance = 200
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	credit, err := service.GetUserCredit(ctx, user.Username, 500)
	if err != nil {
		t.Fatalf("GetUserCredit returned error: %v", err)
	}
	if credit != 700 {
		t.Fatalf("credit = %d, want 700", credit)
	}

	credit, err = service.GetUserCredit(ctx, "missing_user", 500)
	if err != nil {
		t.Fatalf("expected no error for missing user, got %v", err)
	}
	if credit != 500 {
		t.Fatalf("credit for missing user = %d, want 500", credit)
	}
}
