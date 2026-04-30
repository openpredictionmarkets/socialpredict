package adminhandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers"
	dusers "socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	authsvc "socialpredict/internal/service/auth"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/security"

	"gorm.io/gorm"
)

func buildAddUserTestHandler(t *testing.T) (http.HandlerFunc, *gorm.DB) {
	t.Helper()

	db := modelstesting.NewFakeDB(t)
	usersService := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	authService := authsvc.NewAuthService(usersService)
	configService := configsvc.NewStaticService(modelstesting.GenerateEconomicConfig())

	return AddUserHandler(usersService, configService, authService), db
}

func createAddUserAuthSubject(t *testing.T, db *gorm.DB, username, userType string, mustChangePassword bool) string {
	t.Helper()

	user := modelstesting.GenerateUser(username, 1000)
	user.UserType = userType
	user.MustChangePassword = mustChangePassword
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create auth subject: %v", err)
	}
	if err := db.Model(&user).Update("must_change_password", mustChangePassword).Error; err != nil {
		t.Fatalf("persist must_change_password=%t: %v", mustChangePassword, err)
	}

	return modelstesting.GenerateValidJWT(username)
}

func TestAddUserHandler_ReturnsFailureEnvelopes(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	tests := []struct {
		name       string
		seedUser   bool
		username   string
		userType   string
		mustChange bool
		authHeader string
		body       string
		wantStatus int
		wantReason handlers.FailureReason
	}{
		{
			name:       "invalid token",
			authHeader: "Bearer invalid.token",
			body:       `{"username":"freshuser"}`,
			wantStatus: http.StatusUnauthorized,
			wantReason: handlers.ReasonInvalidToken,
		},
		{
			name:       "missing auth user",
			authHeader: "Bearer " + modelstesting.GenerateValidJWT("missing-admin"),
			body:       `{"username":"freshuser"}`,
			wantStatus: http.StatusNotFound,
			wantReason: handlers.ReasonUserNotFound,
		},
		{
			name:       "password change required",
			seedUser:   true,
			username:   "adminneedsreset",
			userType:   "ADMIN",
			mustChange: true,
			body:       `{"username":"freshuser"}`,
			wantStatus: http.StatusForbidden,
			wantReason: handlers.ReasonPasswordChangeRequired,
		},
		{
			name:       "non-admin caller",
			seedUser:   true,
			username:   "regularuser",
			userType:   "REGULAR",
			body:       `{"username":"freshuser"}`,
			wantStatus: http.StatusForbidden,
			wantReason: handlers.ReasonAuthorizationDenied,
		},
		{
			name:       "invalid username",
			seedUser:   true,
			username:   "admincreator",
			userType:   "ADMIN",
			body:       `{"username":"BadName"}`,
			wantStatus: http.StatusBadRequest,
			wantReason: handlers.ReasonValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, db := buildAddUserTestHandler(t)

			authHeader := tt.authHeader
			if tt.seedUser {
				authHeader = "Bearer " + createAddUserAuthSubject(t, db, tt.username, tt.userType, tt.mustChange)
			}

			req := httptest.NewRequest(http.MethodPost, "/v0/admin/createuser", bytes.NewBufferString(tt.body))
			req.Header.Set("Authorization", authHeader)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}

			var response handlers.FailureEnvelope
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode failure response: %v", err)
			}
			if response.OK || response.Reason != string(tt.wantReason) {
				t.Fatalf("expected reason %q, got %+v", tt.wantReason, response)
			}
		})
	}
}

func TestAddUserHandler_CreatesUserWithPlainSuccessResponse(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	handler, db := buildAddUserTestHandler(t)
	token := createAddUserAuthSubject(t, db, "admincreator", "ADMIN", false)

	req := httptest.NewRequest(http.MethodPost, "/v0/admin/createuser", bytes.NewBufferString(`{"username":"freshuser"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response struct {
		Message  string `json:"message"`
		Username string `json:"username"`
		Password string `json:"password"`
		UserType string `json:"usertype"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode success response: %v", err)
	}

	if response.Message != "User created successfully" {
		t.Fatalf("expected success message, got %q", response.Message)
	}
	if response.Username != "freshuser" {
		t.Fatalf("expected username freshuser, got %q", response.Username)
	}
	if response.Password == "" {
		t.Fatalf("expected generated password in response")
	}
	if response.UserType != "REGULAR" {
		t.Fatalf("expected REGULAR user type, got %q", response.UserType)
	}

	var created models.User
	if err := db.Where("username = ?", "freshuser").First(&created).Error; err != nil {
		t.Fatalf("load created user: %v", err)
	}
	if !created.MustChangePassword {
		t.Fatalf("expected created user to require password change")
	}
}
