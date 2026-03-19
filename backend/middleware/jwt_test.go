package middleware

// Additional JWT-focused tests for issue #244.
// The core middleware_test.go covers error paths; this file adds happy-path and
// edge-case coverage for the JWT lifecycle: sign → transmit → validate → deny.

import (
	"net/http/httptest"
	"socialpredict/models/modelstesting"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// signToken is a test helper that signs a UserClaims payload using the
// same key that middleware.getJWTKey() returns (driven by JWT_SIGNING_KEY).
func signToken(claims *UserClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTKey())
}

// TestParseToken_ValidToken verifies that a freshly-signed token round-trips
// through parseToken without error.
func TestParseToken_ValidToken(t *testing.T) {
	claims := &UserClaims{
		Username: "roundtripuser",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}
	tokenStr, err := signToken(claims)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	}
	parsed, err := parseToken(tokenStr, keyFunc)
	if err != nil {
		t.Fatalf("parseToken returned error for valid token: %v", err)
	}
	if !parsed.Valid {
		t.Error("expected parsed token to be valid")
	}
	got := parsed.Claims.(*UserClaims).Username
	if got != "roundtripuser" {
		t.Errorf("claims.Username = %q, want %q", got, "roundtripuser")
	}
}

// TestParseToken_ExpiredToken verifies that a token whose expiry is in the past
// is rejected.
func TestParseToken_ExpiredToken(t *testing.T) {
	claims := &UserClaims{
		Username: "expireduser",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(), // already expired
		},
	}
	tokenStr, err := signToken(claims)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	}
	_, err = parseToken(tokenStr, keyFunc)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

// TestParseToken_WrongKey verifies that a token signed with a different key is
// rejected.
func TestParseToken_WrongKey(t *testing.T) {
	claims := &UserClaims{
		Username: "tampereduser",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}
	wrongKey := []byte("this-is-not-the-right-key")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(wrongKey)
	if err != nil {
		t.Fatalf("failed to sign with wrong key: %v", err)
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil // correct key — signature won't match
	}
	_, err = parseToken(tokenStr, keyFunc)
	if err == nil {
		t.Error("expected error for token signed with wrong key, got nil")
	}
}

// TestValidateTokenAndGetUser_ValidToken_UserExists is the success path:
// valid JWT + user present in DB → returns the user.
func TestValidateTokenAndGetUser_ValidToken_UserExists(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	user := modelstesting.GenerateUser("jwtuser", 1000)
	db.Create(&user)

	tokenStr := modelstesting.GenerateValidJWT("jwtuser")
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	got, httpErr := ValidateTokenAndGetUser(req, db)
	if httpErr != nil {
		t.Fatalf("expected no error, got: %v", httpErr)
	}
	if got == nil {
		t.Fatal("expected non-nil user")
	}
	if got.Username != "jwtuser" {
		t.Errorf("Username = %q, want %q", got.Username, "jwtuser")
	}
}

// TestValidateTokenAndGetUser_ValidToken_UserNotFound: valid JWT, but the
// username in the claims does not exist in the DB → 404.
func TestValidateTokenAndGetUser_ValidToken_UserNotFound(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	// DB is empty — no user created

	tokenStr := modelstesting.GenerateValidJWT("ghostuser")
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	got, httpErr := ValidateTokenAndGetUser(req, db)
	if got != nil {
		t.Error("expected nil user")
	}
	if httpErr == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
	if httpErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", httpErr.StatusCode)
	}
}

// TestValidateTokenAndGetUser_ExpiredToken: expired JWT → 401.
func TestValidateTokenAndGetUser_ExpiredToken(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	// Build an expired token using the same key as the middleware.
	claims := &UserClaims{
		Username: "expiryuser",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
		},
	}
	tokenStr, err := signToken(claims)
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	got, httpErr := ValidateTokenAndGetUser(req, db)
	if got != nil {
		t.Error("expected nil user for expired token")
	}
	if httpErr == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if httpErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", httpErr.StatusCode)
	}
}

// TestValidateUserAndEnforcePasswordChange_MustChange: valid token, user exists,
// but MustChangePassword is true → 403.
func TestValidateUserAndEnforcePasswordChange_MustChange(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	user := modelstesting.GenerateUser("pwchangeuser", 1000)
	user.MustChangePassword = true
	db.Create(&user)

	tokenStr := modelstesting.GenerateValidJWT("pwchangeuser")
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	got, httpErr := ValidateUserAndEnforcePasswordChangeGetUser(req, db)
	if got != nil {
		t.Error("expected nil user when password change required")
	}
	if httpErr == nil {
		t.Fatal("expected error when MustChangePassword=true")
	}
	if httpErr.StatusCode != 403 {
		t.Errorf("StatusCode = %d, want 403", httpErr.StatusCode)
	}
}

// TestValidateUserAndEnforcePasswordChange_Success: valid token, user exists,
// MustChangePassword is false → returns the user.
func TestValidateUserAndEnforcePasswordChange_Success(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	user := modelstesting.GenerateUser("normaluser", 1000)
	user.MustChangePassword = false
	db.Create(&user)

	tokenStr := modelstesting.GenerateValidJWT("normaluser")
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	got, httpErr := ValidateUserAndEnforcePasswordChangeGetUser(req, db)
	if httpErr != nil {
		t.Fatalf("expected no error, got: %v", httpErr)
	}
	if got == nil {
		t.Fatal("expected non-nil user")
	}
	if got.Username != "normaluser" {
		t.Errorf("Username = %q, want %q", got.Username, "normaluser")
	}
}

// TestJWTSigningMethod verifies we only accept HMAC-signed tokens (not "alg:none").
func TestJWTSigningMethod_AlgNoneRejected(t *testing.T) {
	// Build a token with no signing (alg=none) — should be rejected at key validation.
	// jwt.UnsafeAllowNoneSignatureType is needed to construct such a token; if the
	// library refuses to even sign it we treat that as a pass.
	claims := jwt.StandardClaims{
		Subject:   "algnoneuser",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		// Some versions of the library disallow this outright — that's fine.
		t.Skipf("library disallows alg:none signing: %v", err)
	}

	db := modelstesting.NewFakeDB(t)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	got, httpErr := ValidateTokenAndGetUser(req, db)
	if got != nil {
		t.Error("expected nil user for alg:none token")
	}
	if httpErr == nil {
		t.Error("expected error for alg:none token, got nil")
	}
}

// TestMain is defined in middleware_test.go and sets JWT_SIGNING_KEY for the
// entire package, so all tests in this file inherit the correct signing key.
