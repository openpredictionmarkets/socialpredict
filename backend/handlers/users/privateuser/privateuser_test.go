package privateuser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers"
	"socialpredict/handlers/users/dto"
	"socialpredict/internal/app"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/models/modelstesting"
)

func TestGetPrivateProfileUserResponse_Success(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	user := modelstesting.GenerateUser("alice", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	if err := db.Model(&user).Update("must_change_password", false).Error; err != nil {
		t.Fatalf("clear must_change_password: %v", err)
	}

	token := modelstesting.GenerateValidJWT(user.Username)

	req := httptest.NewRequest("GET", "/v0/privateprofile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	config := modelstesting.GenerateEconomicConfig()
	container := app.BuildApplicationWithConfigService(db, configsvc.NewStaticService(config))

	handler := GetPrivateProfileHandler(container.GetUsersService())
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var envelope handlers.SuccessEnvelope[dto.PrivateUserResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	resp := envelope.Result

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
	container := app.BuildApplicationWithConfigService(db, configsvc.NewStaticService(config))

	handler := GetPrivateProfileHandler(container.GetUsersService())
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
	var envelope handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode failure envelope: %v", err)
	}
	if envelope.OK || envelope.Reason != string(handlers.ReasonInvalidToken) {
		t.Fatalf("expected invalid token envelope, got %+v", envelope)
	}
}

func TestGetPrivateProfileUserResponse_RequiresPasswordChange(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	user := modelstesting.GenerateUser("needsreset", 0)
	user.MustChangePassword = true
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/privateprofile", nil)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT(user.Username))
	rec := httptest.NewRecorder()

	config := modelstesting.GenerateEconomicConfig()
	container := app.BuildApplicationWithConfigService(db, configsvc.NewStaticService(config))

	handler := GetPrivateProfileHandler(container.GetUsersService())
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", rec.Code, rec.Body.String())
	}
	var envelope handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode failure envelope: %v", err)
	}
	if envelope.OK || envelope.Reason != string(handlers.ReasonPasswordChangeRequired) {
		t.Fatalf("expected password change envelope, got %+v", envelope)
	}
}
