package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	dusers "socialpredict/internal/domain/users"
)

// --- Test doubles ---

type fakeIdentity struct {
	user error // error returned by UserFromToken
	usr  *dusers.User
	adm  error // error returned by EnsureAdmin
}

func (f *fakeIdentity) UserFromToken(_ context.Context, _ string) (*dusers.User, error) {
	if f.user != nil {
		return nil, f.user
	}
	return f.usr, nil
}

func (f *fakeIdentity) EnsureAdmin(_ *dusers.User) error {
	return f.adm
}

type fakePasswords struct {
	mustChange    bool
	mustChangeErr error
	changeErr     error
}

func (f *fakePasswords) MustChangePassword(_ context.Context, _ string) (bool, error) {
	return f.mustChange, f.mustChangeErr
}

func (f *fakePasswords) ChangePassword(_ context.Context, _, _, _ string) error {
	return f.changeErr
}

type fakeUserReader struct {
	user *dusers.User
	err  error
}

func (r *fakeUserReader) GetUser(_ context.Context, _ string) (*dusers.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.user, nil
}

// --- UserFromToken tests ---

func TestUserFromToken_NilUsers(t *testing.T) {
	svc := NewIdentityService(nil)

	_, err := svc.UserFromToken(context.Background(), "some-token")
	if !errors.Is(err, dusers.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData for nil users, got %v", err)
	}
}

func TestUserFromToken_InvalidToken(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	reader := &fakeUserReader{user: &dusers.User{Username: "alice"}}
	svc := NewIdentityService(reader)

	_, err := svc.UserFromToken(context.Background(), "totally.invalid.token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestUserFromToken_EmptyToken(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	reader := &fakeUserReader{user: &dusers.User{Username: "alice"}}
	svc := NewIdentityService(reader)

	_, err := svc.UserFromToken(context.Background(), "")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for empty token, got %v", err)
	}
}

func TestUserFromToken_BlankUsernameClaim(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	reader := &fakeUserReader{user: &dusers.User{Username: "alice"}}
	svc := NewIdentityService(reader)

	// Create a token with blank username
	token, err := createTestToken("   ", "USER")
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	_, err = svc.UserFromToken(context.Background(), token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for blank username claim, got %v", err)
	}
}

func TestUserFromToken_Success(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	expectedUser := &dusers.User{Username: "alice", UserType: "regular"}
	reader := &fakeUserReader{user: expectedUser}
	svc := NewIdentityService(reader)

	token, err := createTestToken("alice", "regular")
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	user, err := svc.UserFromToken(context.Background(), token)
	if err != nil {
		t.Fatalf("UserFromToken returned error: %v", err)
	}
	if user.Username != "alice" {
		t.Fatalf("expected username alice, got %s", user.Username)
	}
}

func TestUserFromToken_UserNotFound(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	reader := &fakeUserReader{err: dusers.ErrUserNotFound}
	svc := NewIdentityService(reader)

	token, err := createTestToken("deleted_user", "regular")
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	_, err = svc.UserFromToken(context.Background(), token)
	if !errors.Is(err, dusers.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

// --- mapIdentityError tests ---

func TestMapIdentityError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"invalid token", ErrInvalidToken, http.StatusUnauthorized, "Invalid token"},
		{"admin required", ErrAdminPrivilegesRequired, http.StatusForbidden, "admin privileges required"},
		{"user not found", dusers.ErrUserNotFound, http.StatusNotFound, "User not found"},
		{"invalid user data", dusers.ErrInvalidUserData, http.StatusInternalServerError, "Authentication service misconfigured"},
		{"unknown error", errors.New("something else"), http.StatusInternalServerError, "Failed to load user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpErr := mapIdentityError(tt.err)
			if httpErr.StatusCode != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, httpErr.StatusCode)
			}
			if httpErr.Message != tt.wantMsg {
				t.Fatalf("expected message %q, got %q", tt.wantMsg, httpErr.Message)
			}
		})
	}
}

// --- AuthService RequireUser tests ---

func TestRequireUser_NilIdentity(t *testing.T) {
	svc := NewAuthServiceWithDependencies(nil, &fakePasswords{})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer some-token")

	_, httpErr := svc.RequireUser(req)
	if httpErr == nil {
		t.Fatalf("expected error for nil identity")
	}
	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", httpErr.StatusCode)
	}
}

func TestRequireUser_MissingHeader(t *testing.T) {
	identity := &fakeIdentity{usr: &dusers.User{Username: "alice"}}
	svc := NewAuthServiceWithDependencies(identity, &fakePasswords{})
	req := httptest.NewRequest("GET", "/test", nil)

	_, httpErr := svc.RequireUser(req)
	if httpErr == nil {
		t.Fatalf("expected error for missing auth header")
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", httpErr.StatusCode)
	}
}

func TestRequireUser_InvalidToken(t *testing.T) {
	identity := &fakeIdentity{user: ErrInvalidToken}
	svc := NewAuthServiceWithDependencies(identity, &fakePasswords{})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer bad-token")

	_, httpErr := svc.RequireUser(req)
	if httpErr == nil {
		t.Fatalf("expected error for invalid token")
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", httpErr.StatusCode)
	}
}

func TestRequireUser_Success(t *testing.T) {
	identity := &fakeIdentity{usr: &dusers.User{Username: "alice", UserType: "regular"}}
	svc := NewAuthServiceWithDependencies(identity, &fakePasswords{})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	user, httpErr := svc.RequireUser(req)
	if httpErr != nil {
		t.Fatalf("expected no error, got %v", httpErr)
	}
	if user.Username != "alice" {
		t.Fatalf("expected alice, got %s", user.Username)
	}
}

// --- AuthService RequireAdmin tests ---

func TestRequireAdmin_NotAdmin(t *testing.T) {
	identity := &fakeIdentity{
		usr: &dusers.User{Username: "alice", UserType: "regular"},
		adm: ErrAdminPrivilegesRequired,
	}
	svc := NewAuthServiceWithDependencies(identity, &fakePasswords{})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	_, httpErr := svc.RequireAdmin(req)
	if httpErr == nil {
		t.Fatalf("expected error for non-admin")
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", httpErr.StatusCode)
	}
}

func TestRequireAdmin_Success(t *testing.T) {
	identity := &fakeIdentity{
		usr: &dusers.User{Username: "admin", UserType: "ADMIN"},
		adm: nil,
	}
	svc := NewAuthServiceWithDependencies(identity, &fakePasswords{})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	user, httpErr := svc.RequireAdmin(req)
	if httpErr != nil {
		t.Fatalf("expected no error, got %v", httpErr)
	}
	if user.Username != "admin" {
		t.Fatalf("expected admin, got %s", user.Username)
	}
}

// --- AuthService CurrentUser tests ---

func TestCurrentUser_PasswordChangeRequired(t *testing.T) {
	identity := &fakeIdentity{usr: &dusers.User{Username: "alice"}}
	passwords := &fakePasswords{mustChange: true}
	svc := NewAuthServiceWithDependencies(identity, passwords)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	_, httpErr := svc.CurrentUser(req)
	if httpErr == nil {
		t.Fatalf("expected error for password change required")
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", httpErr.StatusCode)
	}
	if httpErr.Message != "Password change required" {
		t.Fatalf("expected 'Password change required', got %q", httpErr.Message)
	}
}

func TestCurrentUser_PasswordCheckError(t *testing.T) {
	identity := &fakeIdentity{usr: &dusers.User{Username: "alice"}}
	passwords := &fakePasswords{mustChangeErr: errors.New("db error")}
	svc := NewAuthServiceWithDependencies(identity, passwords)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	_, httpErr := svc.CurrentUser(req)
	if httpErr == nil {
		t.Fatalf("expected error for password check failure")
	}
	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", httpErr.StatusCode)
	}
}

func TestCurrentUser_Success(t *testing.T) {
	identity := &fakeIdentity{usr: &dusers.User{Username: "alice"}}
	passwords := &fakePasswords{mustChange: false}
	svc := NewAuthServiceWithDependencies(identity, passwords)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	user, httpErr := svc.CurrentUser(req)
	if httpErr != nil {
		t.Fatalf("expected no error, got %v", httpErr)
	}
	if user.Username != "alice" {
		t.Fatalf("expected alice, got %s", user.Username)
	}
}

// --- AuthService delegated ChangePassword/MustChangePassword ---

func TestAuthService_ChangePassword_NilPasswords(t *testing.T) {
	svc := NewAuthServiceWithDependencies(&fakeIdentity{}, nil)

	err := svc.ChangePassword(context.Background(), "alice", "old", "new")
	if !errors.Is(err, dusers.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData, got %v", err)
	}
}

func TestAuthService_MustChangePassword_NilPasswords(t *testing.T) {
	svc := NewAuthServiceWithDependencies(&fakeIdentity{}, nil)

	_, err := svc.MustChangePassword(context.Background(), "alice")
	if !errors.Is(err, dusers.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData, got %v", err)
	}
}

// Ensure TestMain from middleware_test.go sets JWT key for UserFromToken tests.
// If running this file in isolation, set the env var.
func init() {
	if os.Getenv("JWT_SIGNING_KEY") == "" {
		os.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
	}
}
