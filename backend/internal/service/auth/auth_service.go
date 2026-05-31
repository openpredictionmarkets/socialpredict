package auth

import (
	"context"
	"net/http"
	"strings"

	dusers "socialpredict/internal/domain/users"
)

// Authenticator exposes the authentication operations used by HTTP handlers.
type Authenticator interface {
	CurrentUser(r *http.Request) (*dusers.User, *AuthError)
	RequireUser(r *http.Request) (*dusers.User, *AuthError)
	RequireAdmin(r *http.Request) (*dusers.User, *AuthError)
}

// AuthService provides a façade over the authentication helpers so callers can
// depend on a single injected object rather than package-level functions.
type AuthService struct {
	users         dusers.ServiceInterface
	jwtSigningKey []byte
}

// NewAuthService constructs a façade that uses the provided users service for
// token validation and password-change enforcement.
func NewAuthService(users dusers.ServiceInterface, jwtSigningKey ...[]byte) *AuthService {
	var key []byte
	if len(jwtSigningKey) > 0 && len(jwtSigningKey[0]) > 0 {
		key = append([]byte(nil), jwtSigningKey[0]...)
	} else {
		key = currentJWTSigningKey()
	}
	return &AuthService{users: users, jwtSigningKey: key}
}

// CurrentUser returns the authenticated user, ensuring any password-change
// requirements are enforced.
func (a *AuthService) CurrentUser(r *http.Request) (*dusers.User, *AuthError) {
	tokenString, authErr := tokenFromRequest(r)
	if authErr != nil {
		return nil, authErr
	}
	return a.CurrentUserFromToken(r.Context(), tokenString)
}

// RequireUser resolves the authenticated user without checking the
// must-change-password flag.
func (a *AuthService) RequireUser(r *http.Request) (*dusers.User, *AuthError) {
	tokenString, authErr := tokenFromRequest(r)
	if authErr != nil {
		return nil, authErr
	}
	return a.RequireUserFromToken(r.Context(), tokenString)
}

// RequireAdmin ensures the current user is authenticated and has admin privileges.
func (a *AuthService) RequireAdmin(r *http.Request) (*dusers.User, *AuthError) {
	tokenString, authErr := tokenFromRequest(r)
	if authErr != nil {
		return nil, authErr
	}
	return a.RequireAdminFromToken(r.Context(), tokenString)
}

// CurrentUserFromToken resolves the authenticated user from an extracted token
// and enforces password-change requirements without depending on HTTP request shape.
func (a *AuthService) CurrentUserFromToken(ctx context.Context, tokenString string) (*dusers.User, *AuthError) {
	return ValidateUserAndEnforcePasswordChangeFromToken(ctx, tokenString, a.users, a.jwtSigningKey)
}

// RequireUserFromToken resolves an authenticated user from an extracted token.
func (a *AuthService) RequireUserFromToken(ctx context.Context, tokenString string) (*dusers.User, *AuthError) {
	return ValidateTokenAndGetUserFromToken(ctx, tokenString, a.users, a.jwtSigningKey)
}

// RequireAdminFromToken resolves an authenticated admin from an extracted token.
func (a *AuthService) RequireAdminFromToken(ctx context.Context, tokenString string) (*dusers.User, *AuthError) {
	user, err := a.CurrentUserFromToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if strings.ToUpper(user.UserType) != "ADMIN" {
		return nil, newAuthError(ErrorKindAdminRequired)
	}

	return user, nil
}
