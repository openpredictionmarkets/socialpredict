package usercredit_test

import (
	usercredit "socialpredict/handlers/users/credit"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup/setuptesting"
	"testing"
)

type UserPublicInfo struct {
	AccountBalance int
}

func TestCalculateUserCredit(t *testing.T) {

	db := modelstesting.NewFakeDB(t)
	if err := db.AutoMigrate(&models.Bet{}, &models.User{}); err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	appConfig := setuptesting.MockEconomicConfig()
	user := &models.User{Username: "testuser", AccountBalance: 1000, ApiKey: "unique_api_key_1"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to save user to database: %v", err)
	}

	userCredit := usercredit.CalculateUserCredit(db, "testuser")

	expectedCredit := appConfig.Economics.User.MaximumDebtAllowed + 1000

	if userCredit != expectedCredit {
		t.Errorf("Expected user credit to be %d, got %d", expectedCredit, userCredit)
	}
}
