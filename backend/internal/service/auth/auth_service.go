package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	dauth "socialpredict/internal/domain/auth"
	dusers "socialpredict/internal/domain/users"
	usermodels "socialpredict/internal/domain/users/models"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// UserReader is the narrow view of the users service that AuthService needs.
type UserReader interface {
	GetUser(ctx context.Context, username string) (*dusers.User, error)
}

// PasswordSanitizer validates and sanitizes password input.
type PasswordSanitizer interface {
	SanitizePassword(string) (string, error)
}

// CredentialRepository provides access to authentication credentials.
type CredentialRepository interface {
	GetCredentials(ctx context.Context, username string) (*dauth.Credentials, error)
	UpdatePassword(ctx context.Context, username string, hashedPassword string, mustChange bool) error
}

// Authenticator exposes the authentication operations used by HTTP handlers.
type Authenticator interface {
	CurrentUser(r *http.Request) (*dusers.User, *HTTPError)
	RequireUser(r *http.Request) (*dusers.User, *HTTPError)
	RequireAdmin(r *http.Request) (*dusers.User, *HTTPError)
	ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error
	MustChangePassword(ctx context.Context, username string) (bool, error)
}

// AuthService provides a façade over the authentication helpers so callers can
// depend on a single injected object rather than package-level functions.
type AuthService struct {
	users     UserReader
	repo      CredentialRepository
	sanitizer PasswordSanitizer
}

// NewAuthService constructs a façade that uses the provided dependencies for
// token validation, password-change enforcement, and credential management.
func NewAuthService(users UserReader, repo CredentialRepository, sanitizer PasswordSanitizer) *AuthService {
	return &AuthService{
		users:     users,
		repo:      repo,
		sanitizer: sanitizer,
	}
}

// CurrentUser returns the authenticated user, ensuring any password-change
// requirements are enforced.
func (a *AuthService) CurrentUser(r *http.Request) (*dusers.User, *HTTPError) {
	user, httpErr := a.RequireUser(r)
	if httpErr != nil {
		return nil, httpErr
	}

	mustChange, err := a.MustChangePassword(r.Context(), user.Username)
	if err != nil {
		return nil, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "Failed to check password status"}
	}

	if mustChange {
		return nil, &HTTPError{
			StatusCode: http.StatusForbidden,
			Message:    "Password change required",
		}
	}

	return user, nil
}

// RequireUser resolves the authenticated user without checking the
// must-change-password flag.
func (a *AuthService) RequireUser(r *http.Request) (*dusers.User, *HTTPError) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Authorization header is required"}
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := parseToken(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})
	if err != nil {
		return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Invalid token"}
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		user, err := a.users.GetUser(r.Context(), claims.Username)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				return nil, &HTTPError{StatusCode: http.StatusNotFound, Message: "User not found"}
			}
			return nil, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "Failed to load user"}
		}
		return user, nil
	}
	return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Invalid token"}
}

// RequireAdmin ensures the current user is authenticated and has admin privileges.
func (a *AuthService) RequireAdmin(r *http.Request) (*dusers.User, *HTTPError) {
	user, err := a.RequireUser(r)
	if err != nil {
		return nil, err
	}

	if strings.ToUpper(user.UserType) != "ADMIN" {
		return nil, &HTTPError{
			StatusCode: http.StatusForbidden,
			Message:    "admin privileges required",
		}
	}

	return user, nil
}

// MustChangePassword reports whether the specified user is required to change their password.
func (a *AuthService) MustChangePassword(ctx context.Context, username string) (bool, error) {
	creds, err := a.repo.GetCredentials(ctx, username)
	if err != nil {
		return false, err
	}
	return creds.MustChangePassword, nil
}

// ChangePassword validates credentials and persists a new hashed password.
func (a *AuthService) ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error {
	if username == "" {
		return dusers.ErrInvalidUserData
	}
	if currentPassword == "" {
		return fmt.Errorf("current password is required")
	}
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}
	if a.sanitizer == nil {
		return dusers.ErrInvalidUserData
	}

	creds, err := a.repo.GetCredentials(ctx, username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(currentPassword)); err != nil {
		return dauth.ErrInvalidCredentials
	}

	sanitized, err := a.sanitizer.SanitizePassword(newPassword)
	if err != nil {
		return fmt.Errorf("new password does not meet security requirements: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(sanitized)); err == nil {
		return fmt.Errorf("new password must differ from the current password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(sanitized), usermodels.PasswordHashCost())
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	return a.repo.UpdatePassword(ctx, username, string(hashed), false)
}

var _ Authenticator = (*AuthService)(nil)
