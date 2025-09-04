package usershandlers

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestApplyTransactionToUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	startingBalance := int64(100)
	user := modelstesting.GenerateUser("testuser", startingBalance)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	type testCase struct {
		txType        string
		amount        int64
		expectBalance int64
		expectErr     bool
	}

	testCases := []testCase{
		{TransactionWin, 50, 150, false},
		{TransactionRefund, 25, 175, false},
		{TransactionSale, 20, 195, false},
		{TransactionBuy, 40, 155, false},
		{TransactionFee, 10, 145, false},
		{"UNKNOWN", 10, 145, true}, // balance should not change
	}

	for _, tc := range testCases {
		err := ApplyTransactionToUser(user.Username, tc.amount, db, tc.txType)
		var updated models.User
		if err := db.Where("username = ?", user.Username).First(&updated).Error; err != nil {
			t.Fatalf("failed to fetch user after update: %v", err)
		}
		if tc.expectErr {
			if err == nil {
				t.Errorf("expected error for type %s but got nil", tc.txType)
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error for type %s: %v", tc.txType, err)
		}
		if updated.AccountBalance != tc.expectBalance {
			t.Errorf("after %s, expected balance %d, got %d", tc.txType, tc.expectBalance, updated.AccountBalance)
		}
	}
}
