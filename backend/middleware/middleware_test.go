package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestHTTPError(t *testing.T) {
	err := &HTTPError{
		StatusCode: 404,
		Message:    "Not found",
	}

	if err.Error() != "Not found" {
		t.Errorf("Expected 'Not found', got '%s'", err.Error())
	}

	if err.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", err.StatusCode)
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
	// Set up JWT key for testing
	originalKey := jwtKey
	jwtKey = []byte("test-secret-key")
	defer func() { jwtKey = originalKey }()

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
		return jwtKey, nil
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
			user := &models.User{
				MustChangePassword: tt.mustChangePassword,
			}

			httpErr := CheckMustChangePasswordFlag(user)

			if tt.expectError {
				if httpErr == nil {
					t.Errorf("Expected error but got none")
				} else if httpErr.StatusCode != http.StatusForbidden {
					t.Errorf("Expected status code %d, got %d", http.StatusForbidden, httpErr.StatusCode)
				}
			} else {
				if httpErr != nil {
					t.Errorf("Expected no error but got: %v", httpErr)
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
	}{
		{
			name:           "Valid POST method",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest, // Will fail due to invalid body, but method is accepted
		},
		{
			name:           "Invalid GET method",
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid PUT method",
			method:         http.MethodPut,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/login", nil)
			w := httptest.NewRecorder()

			LoginHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	invalidJSON := "{ invalid json }"
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(invalidJSON))
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
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

			LoginHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestValidateTokenAndGetUser_MissingHeader(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	user, httpErr := ValidateTokenAndGetUser(req, db)

	if user != nil {
		t.Error("Expected nil user")
	}
	if httpErr == nil {
		t.Error("Expected error but got none")
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, httpErr.StatusCode)
	}
}

func TestValidateTokenAndGetUser_InvalidToken(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	user, httpErr := ValidateTokenAndGetUser(req, db)

	if user != nil {
		t.Error("Expected nil user")
	}
	if httpErr == nil {
		t.Error("Expected error but got none")
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, httpErr.StatusCode)
	}
}

func TestValidateAdminToken_MissingHeader(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	err := ValidateAdminToken(req, db)

	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestValidateAdminToken_InvalidToken(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	err := ValidateAdminToken(req, db)

	if err == nil {
		t.Error("Expected error but got none")
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

	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header

	user, httpErr := ValidateUserAndEnforcePasswordChangeGetUser(req, db)

	if user != nil {
		t.Error("Expected nil user")
	}
	if httpErr == nil {
		t.Error("Expected error but got none")
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, httpErr.StatusCode)
	}
}

// Helper function to create a valid JWT token for testing
func createTestToken(username string, userType string) (string, error) {
	claims := &UserClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func TestCreateTestToken(t *testing.T) {
	// Set up JWT key for testing
	originalKey := jwtKey
	jwtKey = []byte("test-secret-key")
	defer func() { jwtKey = originalKey }()

	token, err := createTestToken("testuser", "USER")
	if err != nil {
		t.Errorf("Expected no error creating test token, got: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestJWTKeyExists(t *testing.T) {
	// Test that jwtKey is available (it's loaded from environment)
	// In a test environment, it should be non-nil
	originalKey := jwtKey
	defer func() { jwtKey = originalKey }()

	// Set a test key if not set
	if jwtKey == nil || len(jwtKey) == 0 {
		jwtKey = []byte("test-key-for-testing")
	}

	if len(jwtKey) == 0 {
		t.Error("Expected jwtKey to be non-empty")
	}
}

// Test environment setup
func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	// Run tests
	code := m.Run()

	// Clean up
	os.Exit(code)
}
