package usershandlers

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestApplyTransactionToUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Arrange: create a user with a starting balance
	startingBalance := int64(100)
	user := modelstesting.GenerateUser("testuser", startingBalance)

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Act: apply a WIN transaction of 50
	err := ApplyTransactionToUser(user.Username, 50, db, TransactionWin)
	if err != nil {
		t.Errorf("unexpected error applying WIN transaction: %v", err)
	}

	// Assert: balance should be incremented
	var updated models.User
	if err := db.Where("username = ?", user.Username).First(&updated).Error; err != nil {
		t.Fatalf("failed to fetch user after update: %v", err)
	}

	expectedBalance := startingBalance + 50
	if updated.AccountBalance != expectedBalance {
		t.Errorf("expected balance %d, got %d", expectedBalance, updated.AccountBalance)
	}
}
