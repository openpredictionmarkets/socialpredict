package publicuser_test

import (
	"socialpredict/handlers/users/publicuser"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"testing"
)

// Test for GetPublicUserInfo using an in-memory SQLite database
func TestGetPublicUserInfo(t *testing.T) {
	// Set up the in-memory SQLite database using modelstesting.NewFakeDB
	db := modelstesting.NewFakeDB(t)

	// Migrate the User model
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("Failed to migrate user model: %v", err)
	}

	// Create a test user in the database
	user := models.PublicUser{
		Username:              "testuser",
		DisplayName:           "Test User",
		UserType:              "regular",
		InitialAccountBalance: 1000,
		AccountBalance:        500,
		PersonalEmoji:         "😊",
		Description:           "Test description",
		PersonalLink1:         "http://link1.com",
		PersonalLink2:         "http://link2.com",
		PersonalLink3:         "http://link3.com",
		PersonalLink4:         "http://link4.com",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to save user to database: %v", err)
	}

	// Call GetPublicUserInfo to retrieve the user
	retrievedUser := publicuser.GetPublicUserInfo(db, "testuser")

	// Expected result
	expectedUser := models.PublicUser{
		Username:              "testuser",
		DisplayName:           "Test User",
		UserType:              "regular",
		InitialAccountBalance: 1000,
		AccountBalance:        500,
		PersonalEmoji:         "😊",
		Description:           "Test description",
		PersonalLink1:         "http://link1.com",
		PersonalLink2:         "http://link2.com",
		PersonalLink3:         "http://link3.com",
		PersonalLink4:         "http://link4.com",
	}

	// Compare the retrieved user with the expected result
	if retrievedUser != expectedUser {
		t.Errorf("GetPublicUserInfo(db, 'testuser') = %+v, want %+v", retrievedUser, expectedUser)
	}
}
