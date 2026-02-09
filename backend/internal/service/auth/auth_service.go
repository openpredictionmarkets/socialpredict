package auth

import (
	"context"
	"errors"
	"net/http"

	dauth "socialpredict/internal/domain/auth"
	dusers "socialpredict/internal/domain/users"
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

// RequestAuthenticator defines HTTP-facing identity/authz operations.
type RequestAuthenticator interface {
	CurrentUser(r *http.Request) (*dusers.User, *HTTPError)
	RequireUser(r *http.Request) (*dusers.User, *HTTPError)
	RequireAdmin(r *http.Request) (*dusers.User, *HTTPError)
}

// PasswordManager defines password and credential lifecycle operations.
type PasswordManager interface {
	ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error
	MustChangePassword(ctx context.Context, username string) (bool, error)
}

// IdentityResolver defines token-to-user and role authorization operations.
type IdentityResolver interface {
	UserFromToken(ctx context.Context, tokenString string) (*dusers.User, error)
	EnsureAdmin(user *dusers.User) error
}

// Authenticator is the backwards-compatible façade used by handlers.
// It intentionally embeds the HTTP adapter and password management contracts.
type Authenticator interface {
	RequestAuthenticator
	PasswordManager
}

// AuthService adapts HTTP requests to the split identity and password services.
type AuthService struct {
	identity  IdentityResolver
	passwords PasswordManager
}

// NewAuthService wires the default identity and password implementations.
func NewAuthService(users UserReader, repo CredentialRepository, sanitizer PasswordSanitizer) *AuthService {
	return NewAuthServiceWithDependencies(
		NewIdentityService(users),
		NewPasswordService(repo, sanitizer),
	)
}

// NewAuthServiceWithDependencies allows injecting custom identity/password services.
func NewAuthServiceWithDependencies(identity IdentityResolver, passwords PasswordManager) *AuthService {
	return &AuthService{
		identity:  identity,
		passwords: passwords,
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
	if a.identity == nil {
		return nil, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "Authentication service misconfigured"}
	}

	tokenString, err := extractTokenFromHeader(r)
	if err != nil {
		return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Authorization header is required"}
	}

	user, err := a.identity.UserFromToken(r.Context(), tokenString)
	if err != nil {
		return nil, mapIdentityError(err)
	}
	return user, nil
}

// RequireAdmin ensures the current user is authenticated and has admin privileges.
func (a *AuthService) RequireAdmin(r *http.Request) (*dusers.User, *HTTPError) {
	user, err := a.RequireUser(r)
	if err != nil {
		return nil, err
	}

	if err := a.identity.EnsureAdmin(user); err != nil {
		return nil, mapIdentityError(err)
	}

	return user, nil
}

// MustChangePassword reports whether the specified user is required to change their password.
func (a *AuthService) MustChangePassword(ctx context.Context, username string) (bool, error) {
	if a.passwords == nil {
		return false, dusers.ErrInvalidUserData
	}
	return a.passwords.MustChangePassword(ctx, username)
}

// ChangePassword validates credentials and persists a new hashed password.
func (a *AuthService) ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error {
	if a.passwords == nil {
		return dusers.ErrInvalidUserData
	}
	return a.passwords.ChangePassword(ctx, username, currentPassword, newPassword)
}

func mapIdentityError(err error) *HTTPError {
	switch {
	case errors.Is(err, ErrInvalidToken):
		return &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Invalid token"}
	case errors.Is(err, ErrAdminPrivilegesRequired):
		return &HTTPError{StatusCode: http.StatusForbidden, Message: "admin privileges required"}
	case errors.Is(err, dusers.ErrUserNotFound):
		return &HTTPError{StatusCode: http.StatusNotFound, Message: "User not found"}
	case errors.Is(err, dusers.ErrInvalidUserData):
		return &HTTPError{StatusCode: http.StatusInternalServerError, Message: "Authentication service misconfigured"}
	default:
		return &HTTPError{StatusCode: http.StatusInternalServerError, Message: "Failed to load user"}
	}
}

var _ Authenticator = (*AuthService)(nil)
