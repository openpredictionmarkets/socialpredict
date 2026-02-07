package privateuser

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers/users/dto"
	"socialpredict/internal/app"
	"socialpredict/models/modelstesting"
)

func TestGetPrivateProfileUserResponse_Success(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	user := modelstesting.GenerateUser("alice", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token := modelstesting.GenerateValidJWT(user.Username)

	req := httptest.NewRequest("GET", "/v0/privateprofile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	config := modelstesting.GenerateEconomicConfig()
	container := app.BuildApplication(db, config)

	handler := GetPrivateProfileHandler(container.GetUsersService(), container.GetAuthService())
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp dto.PrivateUserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Username != user.Username {
		t.Fatalf("expected username %q, got %q", user.Username, resp.Username)
	}
	if resp.Email != user.PrivateUser.Email {
		t.Fatalf("expected email %q, got %q", user.PrivateUser.Email, resp.Email)
	}
}

func TestGetPrivateProfileUserResponse_Unauthorized(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	req := httptest.NewRequest("GET", "/v0/privateprofile", nil)
	rec := httptest.NewRecorder()

	config := modelstesting.GenerateEconomicConfig()
	container := app.BuildApplication(db, config)

	handler := GetPrivateProfileHandler(container.GetUsersService(), container.GetAuthService())
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
