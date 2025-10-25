package usershandlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

func TestGetPublicUserInfoReturnsPublicProfile(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	user := modelstesting.GenerateUser("public_user", 0)
	user.PublicUser.DisplayName = "Public Name"
	user.PublicUser.UserType = "regular"
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	public := GetPublicUserInfo(db, user.Username)
	if public.Username != user.Username {
		t.Fatalf("expected username %s, got %s", user.Username, public.Username)
	}
	if public.DisplayName != "Public Name" {
		t.Fatalf("expected display name %q, got %q", "Public Name", public.DisplayName)
	}
	if public.UserType != "regular" {
		t.Fatalf("expected user type regular, got %s", public.UserType)
	}
}

func TestGetPublicUserResponseWritesJSON(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	orig := util.DB
	util.DB = db
	t.Cleanup(func() { util.DB = orig })

	user := modelstesting.GenerateUser("public_user_handler", 0)
	user.PublicUser.DisplayName = "Handler Name"
	user.PublicUser.UserType = "regular"
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest("GET", "/users/public/"+user.Username, nil)
	req = mux.SetURLVars(req, map[string]string{"username": user.Username})
	rec := httptest.NewRecorder()

	GetPublicUserResponse(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body models.PublicUser
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body.Username != user.Username || body.DisplayName != user.DisplayName {
		t.Fatalf("unexpected body: %+v", body)
	}
}
