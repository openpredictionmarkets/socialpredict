package usershandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers"
	dusers "socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	"socialpredict/models/modelstesting"
	"socialpredict/security"
)

func TestChangeDescriptionHandler_InvalidTokenReturnsFailureEnvelope(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v0/profilechange/description", bytes.NewBufferString(`{"description":"updated"}`))
	rec := httptest.NewRecorder()

	ChangeDescriptionHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var response handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.OK || response.Reason != string(handlers.ReasonInvalidToken) {
		t.Fatalf("expected invalid token envelope, got %+v", response)
	}
}

func TestChangeDescriptionHandler_SuccessReturnsEnvelope(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	user := modelstesting.GenerateUser("alice", 100)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Model(&user).Update("must_change_password", false).Error; err != nil {
		t.Fatalf("update must_change_password: %v", err)
	}

	svc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	req := httptest.NewRequest(http.MethodPost, "/v0/profilechange/description", bytes.NewBufferString(`{"description":"updated bio"}`))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT(user.Username))
	rec := httptest.NewRecorder()

	ChangeDescriptionHandler(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response handlers.SuccessEnvelope[map[string]any]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK {
		t.Fatalf("expected success envelope, got %+v", response)
	}
	if response.Result["description"] != "updated bio" {
		t.Fatalf("expected updated description, got %+v", response.Result)
	}
}

func TestWriteProfileError_SanitizesFailureReasons(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeProfileError(rec, errors.New("description exceeds maximum length"), "description")

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}

		var response handlers.FailureEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.OK || response.Reason != string(handlers.ReasonValidationFailed) {
			t.Fatalf("expected validation envelope, got %+v", response)
		}
	})

	t.Run("internal", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeProfileError(rec, errors.New("database connection string leaked"), "display name")

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}

		var response handlers.FailureEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.OK || response.Reason != "DISPLAY_NAME_UPDATE_FAILED" {
			t.Fatalf("expected sanitized display-name failure, got %+v", response)
		}
		if bytes.Contains(rec.Body.Bytes(), []byte("database connection string leaked")) {
			t.Fatalf("expected raw error details to be removed: %s", rec.Body.String())
		}
	})
}
