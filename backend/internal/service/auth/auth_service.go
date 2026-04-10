package auth

import (
	"net/http"
	"strings"

	dusers "socialpredict/internal/domain/users"
)

// Authenticator exposes the authentication operations used by HTTP handlers.
type Authenticator interface {
	CurrentUser(r *http.Request) (*dusers.User, *HTTPError)
	RequireUser(r *http.Request) (*dusers.User, *HTTPError)
	RequireAdmin(r *http.Request) (*dusers.User, *HTTPError)
}

// AuthService provides a façade over the authentication helpers so callers can
// depend on a single injected object rather than package-level functions.
type AuthService struct {
	users dusers.ServiceInterface
}

// NewAuthService constructs a façade that uses the provided users service for
// token validation and password-change enforcement.
func NewAuthService(users dusers.ServiceInterface) *AuthService {
	return &AuthService{users: users}
}

// CurrentUser returns the authenticated user, ensuring any password-change
// requirements are enforced.
func (a *AuthService) CurrentUser(r *http.Request) (*dusers.User, *HTTPError) {
	return ValidateUserAndEnforcePasswordChangeGetUser(r, a.users)
}

// RequireUser resolves the authenticated user without checking the
// must-change-password flag.
func (a *AuthService) RequireUser(r *http.Request) (*dusers.User, *HTTPError) {
	return ValidateTokenAndGetUser(r, a.users)
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
