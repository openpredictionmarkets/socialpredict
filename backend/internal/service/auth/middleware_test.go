package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"socialpredict/handlers"
	"socialpredict/models/modelstesting"
	"testing"
	"time"

	dusers "socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	"socialpredict/security"

	"github.com/golang-jwt/jwt/v4"
)

func TestAuthError(t *testing.T) {
	err := &AuthError{
		Kind:    ErrorKindUserNotFound,
		Message: "Not found",
	}

	if err.Error() != "Not found" {
		t.Errorf("Expected 'Not found', got '%s'", err.Error())
	}

	if err.Kind != ErrorKindUserNotFound {
		t.Errorf("Expected kind %q, got %q", ErrorKindUserNotFound, err.Kind)
	}
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name          string
		header        string
		expectedToken string
		expectError   bool
	}{
		{
			name:          "Valid Bearer token",
			header:        "Bearer abc123token",
			expectedToken: "abc123token",
			expectError:   false,
		},
		{
			name:        "Missing Authorization header",
			header:      "",
			expectError: true,
		},
		{
			name:          "Token without Bearer prefix",
			header:        "abc123token",
			expectedToken: "abc123token",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			token, err := extractTokenFromHeader(req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if token != tt.expectedToken {
					t.Errorf("Expected token '%s', got '%s'", tt.expectedToken, token)
				}
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	// Set up JWT key for testing via environment variable
	originalKey := os.Getenv("JWT_SIGNING_KEY")
	os.Setenv("JWT_SIGNING_KEY", "test-secret-key")
	defer func() { os.Setenv("JWT_SIGNING_KEY", originalKey) }()

	tests := []struct {
		name        string
		tokenString string
		expectError bool
	}{
		{
			name:        "Invalid token",
			tokenString: "invalid.token.here",
			expectError: true,
		},
		{
			name:        "Empty token",
			tokenString: "",
			expectError: true,
		},
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseToken(tt.tokenString, keyFunc)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestCheckMustChangePasswordFlag(t *testing.T) {
	tests := []struct {
		name               string
		mustChangePassword bool
		expectError        bool
	}{
		{
			name:               "Password change required",
			mustChangePassword: true,
			expectError:        true,
		},
		{
			name:               "Password change not required",
			mustChangePassword: false,
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &dusers.User{
				MustChangePassword: tt.mustChangePassword,
			}

			authErr := CheckMustChangePasswordFlag(user)

			if tt.expectError {
				if authErr == nil {
					t.Errorf("Expected error but got none")
				} else if authErr.Kind != ErrorKindPasswordChangeRequired {
					t.Errorf("Expected kind %q, got %q", ErrorKindPasswordChangeRequired, authErr.Kind)
				}
			} else {
				if authErr != nil {
					t.Errorf("Expected no error but got: %v", authErr)
				}
			}
		})
	}
}

func TestLoginHandler_MethodValidation(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedReason handlers.FailureReason
	}{
		{
			name:           "Valid POST method",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest, // Will fail due to invalid body, but method is accepted
			expectedReason: handlers.ReasonInvalidRequest,
		},
		{
			name:           "Invalid GET method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedReason: handlers.ReasonMethodNotAllowed,
		},
		{
			name:           "Invalid PUT method",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedReason: handlers.ReasonMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/login", nil)
			w := httptest.NewRecorder()

			LoginHandler(nil, security.NewSecurityService())(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			var response handlers.FailureEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode failure response: %v", err)
			}
			if response.OK || response.Reason != string(tt.expectedReason) {
				t.Fatalf("expected reason %q, got %+v", tt.expectedReason, response)
			}
		})
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	invalidJSON := "{ invalid json }"
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(invalidJSON))
	w := httptest.NewRecorder()

	LoginHandler(rusers.NewGormRepository(modelstesting.NewFakeDB(t)), security.NewSecurityService())(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response handlers.FailureEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode failure response: %v", err)
	}
	if response.OK || response.Reason != string(handlers.ReasonInvalidRequest) {
		t.Fatalf("expected reason %q, got %+v", handlers.ReasonInvalidRequest, response)
	}
}

func TestLoginHandler_RejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"testuser","password":"password123","extra":"nope"}`))
	w := httptest.NewRecorder()

	LoginHandler(rusers.NewGormRepository(modelstesting.NewFakeDB(t)), security.NewSecurityService())(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response handlers.FailureEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode failure response: %v", err)
	}
	if response.OK || response.Reason != string(handlers.ReasonInvalidRequest) {
		t.Fatalf("expected reason %q, got %+v", handlers.ReasonInvalidRequest, response)
	}
}

func TestLoginHandler_ValidationFailure(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "Empty username",
			username: "",
			password: "password123",
		},
		{
			name:     "Empty password",
			username: "testuser",
			password: "",
		},
		{
			name:     "Invalid username format",
			username: "Test@User",
			password: "password123",
		},
		{
			name:     "Username with only whitespace",
			username: "   ",
			password: "password123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := modelstesting.NewFakeDB(t)
			repo := rusers.NewGormRepository(db)
			loginReq := map[string]string{
				"username": tt.username,
				"password": tt.password,
			}
			jsonData, _ := json.Marshal(loginReq)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonData))
			w := httptest.NewRecorder()

			LoginHandler(repo, security.NewSecurityService())(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
			}

			var response handlers.FailureEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode failure response: %v", err)
			}
			if response.OK || response.Reason != string(handlers.ReasonValidationFailed) {
				t.Fatalf("expected reason %q, got %+v", handlers.ReasonValidationFailed, response)
			}
		})
	}
}

func TestLoginHandler_MissingDB(t *testing.T) {
	loginReq := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(loginReq)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonData))
	w := httptest.NewRecorder()

	LoginHandler(nil, security.NewSecurityService())(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response handlers.FailureEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode failure response: %v", err)
	}
	if response.OK || response.Reason != string(handlers.ReasonInternalError) {
		t.Fatalf("expected reason %q, got %+v", handlers.ReasonInternalError, response)
	}
}

func TestLoginHandler_TrimsUsernameWhitespace(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)

	testUser := modelstesting.GenerateUser("testuser", 1000)
	if err := testUser.HashPassword("password123"); err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	if err := db.Create(&testUser).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test that usernames with leading/trailing spaces are trimmed before DB lookup
	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "Username with leading space",
			username: " testuser",
			password: "password123",
		},
		{
			name:     "Username with trailing space",
			username: "testuser ",
			password: "password123",
		},
		{
			name:     "Username with both leading and trailing spaces",
			username: " testuser ",
			password: "password123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginReq := map[string]string{
				"username": tt.username,
				"password": tt.password,
			}
			jsonData, _ := json.Marshal(loginReq)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonData))
			w := httptest.NewRecorder()

			LoginHandler(repo, security.NewSecurityService())(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Usernames should be trimmed before DB lookup. Expected status code %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
			}
		})
	}
}

func TestLoginHandler_SuccessResponseEnvelope(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	testUser := modelstesting.GenerateUser("testuser", 1000)
	if err := testUser.HashPassword("password123"); err != nil {
		t.Fatalf("hash password: %v", err)
	}
	testUser.MustChangePassword = true
	if err := db.Create(&testUser).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"testuser","password":"password123"}`))
	w := httptest.NewRecorder()

	LoginHandler(repo, security.NewSecurityService())(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response struct {
		OK     bool `json:"ok"`
		Result struct {
			Token              string `json:"token"`
			Username           string `json:"username"`
			UserType           string `json:"usertype"`
			MustChangePassword bool   `json:"mustChangePassword"`
		} `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !response.OK {
		t.Fatalf("expected ok=true")
	}
	if response.Result.Token == "" {
		t.Fatalf("expected token in response")
	}
	if response.Result.Username != testUser.Username {
		t.Fatalf("expected username %q, got %q", testUser.Username, response.Result.Username)
	}
	if !response.Result.MustChangePassword {
		t.Fatalf("expected mustChangePassword=true")
	}
}

func TestLoginHandler_InvalidCredentialsReturnsFailureEnvelope(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)

	testUser := modelstesting.GenerateUser("testuser", 1000)
	if err := testUser.HashPassword("password123"); err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := db.Create(&testUser).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"testuser","password":"wrong-password"}`))
	w := httptest.NewRecorder()

	LoginHandler(repo, security.NewSecurityService())(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}

	var response struct {
		OK     bool   `json:"ok"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.OK {
		t.Fatalf("expected ok=false")
	}
	if response.Reason != string(handlers.ReasonAuthorizationDenied) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonAuthorizationDenied, response.Reason)
	}
}

func TestValidateTokenAndGetUser_MissingHeader(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	svc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)

	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	user, authErr := ValidateTokenAndGetUser(req, svc)

	if user != nil {
		t.Error("Expected nil user")
	}
	if authErr == nil {
		t.Error("Expected error but got none")
	}
	if authErr.Kind != ErrorKindMissingToken {
		t.Errorf("Expected kind %q, got %q", ErrorKindMissingToken, authErr.Kind)
	}
}

func TestValidateTokenAndGetUser_InvalidToken(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	svc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	user, authErr := ValidateTokenAndGetUser(req, svc)

	if user != nil {
		t.Error("Expected nil user")
	}
	if authErr == nil {
		t.Error("Expected error but got none")
	}
	if authErr.Kind != ErrorKindInvalidToken {
		t.Errorf("Expected kind %q, got %q", ErrorKindInvalidToken, authErr.Kind)
	}
}

func TestValidateAdminToken_MissingHeader(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	svc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	auth := NewAuthService(svc)

	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	err := ValidateAdminToken(req, auth)

	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestValidateAdminToken_InvalidToken(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	svc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	auth := NewAuthService(svc)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	err := ValidateAdminToken(req, auth)

	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestAuthServiceRequireAdmin_EnforcesPasswordChange(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")
	ConfigureJWTSigningKey([]byte("test-secret-key"))

	admin := modelstesting.GenerateUser("admin-needs-reset", 1000)
	admin.UserType = "ADMIN"
	admin.MustChangePassword = true
	if err := admin.HashPassword("password123"); err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}

	svc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	auth := NewAuthService(svc)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	token, err := generateJWT(admin.Username, getJWTKey())
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	user, authErr := auth.RequireAdmin(req)

	if user != nil {
		t.Fatalf("expected nil user when password change is required")
	}
	if authErr == nil {
		t.Fatalf("expected password-change enforcement error")
	}
	if authErr.Kind != ErrorKindPasswordChangeRequired {
		t.Fatalf("expected kind %q, got %q", ErrorKindPasswordChangeRequired, authErr.Kind)
	}
}

func TestUserClaims(t *testing.T) {
	// Test UserClaims struct creation
	claims := &UserClaims{
		Username: "testuser",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	if claims.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", claims.Username)
	}

	if claims.ExpiresAt == 0 {
		t.Error("Expected ExpiresAt to be set")
	}
}

func TestAuthenticate_MiddlewareStructure(t *testing.T) {
	// Test that Authenticate returns a proper http.Handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Authenticate(testHandler)

	if middleware == nil {
		t.Error("Expected middleware to return a handler")
	}

	// Test that the middleware can be called
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// The middleware currently doesn't implement any logic, so this is just testing structure
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestValidateUserAndEnforcePasswordChangeGetUser_MissingToken(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	svc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)

	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	user, authErr := ValidateUserAndEnforcePasswordChangeGetUser(req, svc)

	if user != nil {
		t.Error("Expected nil user")
	}
	if authErr == nil {
		t.Error("Expected error but got none")
	}
	if authErr.Kind != ErrorKindMissingToken {
		t.Errorf("Expected kind %q, got %q", ErrorKindMissingToken, authErr.Kind)
	}
}

func TestValidateUserAndEnforcePasswordChangeGetUser_PasswordChangeRequired(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	testUser := modelstesting.GenerateUser("testuser", 1000)
	testUser.MustChangePassword = true
	if err := testUser.HashPassword("password123"); err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := db.Create(&testUser).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	svc := dusers.NewService(rusers.NewGormRepository(db), nil, security.NewSecurityService().Sanitizer)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	token, err := generateJWT(testUser.Username, getJWTKey())
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	user, authErr := ValidateUserAndEnforcePasswordChangeGetUser(req, svc)

	if user != nil {
		t.Fatalf("expected nil user when password change is required")
	}
	if authErr == nil {
		t.Fatalf("expected password-change enforcement error")
	}
	if authErr.Kind != ErrorKindPasswordChangeRequired {
		t.Fatalf("expected kind %q, got %q", ErrorKindPasswordChangeRequired, authErr.Kind)
	}
	if authErr.Message != "Password change required" {
		t.Fatalf("expected password change message, got %q", authErr.Message)
	}
}

// Helper function to create a valid JWT token for testing
func createTestToken(username string, userType string) (string, error) {
	return generateJWT(username, getJWTKey())
}

func TestCreateTestToken(t *testing.T) {
	// Set up JWT key for testing via environment variable
	originalKey := os.Getenv("JWT_SIGNING_KEY")
	os.Setenv("JWT_SIGNING_KEY", "test-secret-key")
	defer func() { os.Setenv("JWT_SIGNING_KEY", originalKey) }()

	token, err := createTestToken("testuser", "USER")
	if err != nil {
		t.Errorf("Expected no error creating test token, got: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestJWTKeyExists(t *testing.T) {
	// Test that JWT key is available from environment
	// In a test environment, it should be non-nil
	originalKey := os.Getenv("JWT_SIGNING_KEY")
	defer func() { os.Setenv("JWT_SIGNING_KEY", originalKey) }()

	// Set a test key if not set
	if os.Getenv("JWT_SIGNING_KEY") == "" {
		os.Setenv("JWT_SIGNING_KEY", "test-key-for-testing")
	}

	jwtKey := getJWTKey()
	if len(jwtKey) == 0 {
		t.Error("Expected JWT key to be non-empty")
	}
}

func TestGenerateJWT_RequiresSigningKey(t *testing.T) {
	token, err := generateJWT("testuser", nil)
	if err == nil {
		t.Fatal("expected error when JWT key is empty")
	}
	if token != "" {
		t.Fatalf("expected empty token, got %q", token)
	}
}

// Test environment setup
func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	ConfigureJWTSigningKey([]byte("test-secret-key-for-testing"))

	// Run tests
	code := m.Run()

	// Clean up
	os.Exit(code)
}
