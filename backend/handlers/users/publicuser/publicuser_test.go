package publicuser

import (
	"net/http"
	"net/http/httptest"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetPublicUserInfo(t *testing.T) {

	db := modelstesting.NewFakeDB(t)

	user := models.User{
		PublicUser: models.PublicUser{
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
		},
		PrivateUser: models.PrivateUser{
			Email:    "testuser@example.com",
			APIKey:   "whatever123",
			Password: "whatever123",
		},
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to save user to database: %v", err)
	}

	retrievedUser := GetPublicUserInfo(db, "testuser")

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

	if retrievedUser != expectedUser {
		t.Errorf("GetPublicUserInfo(db, 'testuser') = %+v, want %+v", retrievedUser, expectedUser)
	}
}

func TestGetPublicUserResponse_NotFound(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	req, _ := http.NewRequest("GET", "/v0/userinfo/ghostuser", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "ghostuser"})
	rr := httptest.NewRecorder()

	GetPublicUserResponse(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown user, got %d", rr.Code)
	}
}

func TestGetPublicUserResponse_Found(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	user := modelstesting.GenerateUser("knownuser", 500)
	db.Create(&user)

	req, _ := http.NewRequest("GET", "/v0/userinfo/knownuser", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "knownuser"})
	rr := httptest.NewRecorder()

	GetPublicUserResponse(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for existing user, got %d", rr.Code)
	}
}
