package usercredit

import (
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"testing"
)

type UserPublicInfo struct {
	AccountBalance int
}

func TestCalculateUserCredit(t *testing.T) {

	db := modelstesting.NewFakeDB(t)
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	tests := []struct {
		username       string
		email          string
		apiKey         string
		accountBalance int64
		maxDebt        int64
		expectedCredit int64
	}{
		// Test with maximum debt of 500
		{"testuser1", "testuser1@example.com", "api_key_testuser1", -100, 500, 400},
		{"testuser2", "testuser2@example.com", "api_key_testuser2", 0, 500, 500},
		{"testuser3", "testuser3@example.com", "api_key_testuser3", 100, 500, 500 + 100},

		// Test with maximum debt of 5000
		{"testuser4", "testuser4@example.com", "api_key_testuser4", -100, 5000, 4900},
		{"testuser5", "testuser5@example.com", "api_key_testuser5", 0, 5000, 5000},
		{"testuser6", "testuser6@example.com", "api_key_testuser6", 100, 5000, 5100},
	}

	for _, test := range tests {
		// Clear the users table before each test run
		if err := db.Exec("DELETE FROM users").Error; err != nil {
			t.Fatalf("Failed to clear users table: %v", err)
		}

		user := &models.PublicUser{
			Username:       test.username,
			AccountBalance: test.accountBalance,
			Password:       "testpassword",
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("Failed to save user to database: %v", err)
		}

		userCredit := calculateUserCredit(db, test.username, test.maxDebt)

		if userCredit != test.expectedCredit {
			t.Errorf("For %s, with max debt %d, expected user credit to be %d, got %d", test.username, test.maxDebt, test.expectedCredit, userCredit)
		}
	}
}
