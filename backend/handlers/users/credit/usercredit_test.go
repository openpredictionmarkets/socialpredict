package usercredit

import (
	"fmt"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"testing"
)

func TestCalculateUserCredit(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	testCases := []struct {
		username       string
		displayName    string
		accountBalance int64
		maximumDebt    int64
		expectedCredit int64
	}{
		{"user1", "Test User 1", -100, 500, 400},
		{"user2", "Test User 2", 0, 500, 500},
		{"user3", "Test User 3", 100, 500, 600},
		{"user4", "Test User 4", -100, 5000, 4900},
		{"user5", "Test User 5", 0, 5000, 5000},
		{"user6", "Test User 6", 100, 5000, 5100},
	}

	for _, tc := range testCases {
		user := models.User{
			PublicUser: models.PublicUser{
				Username:       tc.username,
				DisplayName:    tc.displayName,
				UserType:       "REGULAR",
				AccountBalance: tc.accountBalance,
			},
			PrivateUser: models.PrivateUser{
				Email:    tc.username + "@example.com",
				Password: "password123",
				APIKey:   "apikey-" + tc.username,
			},
		}

		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("Failed to save user %s to database: %v", tc.username, err)
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Username=%s_AccountBalance=%d_MaximumDebt=%d", tc.username, tc.accountBalance, tc.maximumDebt), func(t *testing.T) {
			credit := calculateUserCredit(db, tc.username, tc.maximumDebt)
			if credit != tc.expectedCredit {
				t.Errorf(
					"calculateUserCredit(db, username=%s, maximumDebt=%d) = %d; want %d",
					tc.username, tc.maximumDebt, credit, tc.expectedCredit,
				)
			}
		})
	}
}
