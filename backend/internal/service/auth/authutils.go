package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type ErrorKind string

const (
	ErrorKindMissingToken           ErrorKind = "missing_token"
	ErrorKindInvalidToken           ErrorKind = "invalid_token"
	ErrorKindUserNotFound           ErrorKind = "user_not_found"
	ErrorKindUserLoadFailed         ErrorKind = "user_load_failed"
	ErrorKindPasswordChangeRequired ErrorKind = "password_change_required"
	ErrorKindAdminRequired          ErrorKind = "admin_required"
	ErrorKindServiceUnavailable     ErrorKind = "service_unavailable"
)

type AuthError struct {
	Kind    ErrorKind
	Message string
}

func (e *AuthError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func newAuthError(kind ErrorKind) *AuthError {
	return &AuthError{Kind: kind, Message: kind.defaultMessage()}
}

func (k ErrorKind) defaultMessage() string {
	switch k {
	case ErrorKindMissingToken:
		return "Authorization header is required"
	case ErrorKindInvalidToken:
		return "Invalid token"
	case ErrorKindUserNotFound:
		return "User not found"
	case ErrorKindUserLoadFailed:
		return "Failed to load user"
	case ErrorKindPasswordChangeRequired:
		return "Password change required"
	case ErrorKindAdminRequired:
		return "admin privileges required"
	case ErrorKindServiceUnavailable:
		return "authentication service unavailable"
	default:
		return "authentication failed"
	}
}

// ExtractTokenFromHeader extracts the JWT token from the Authorization header of the request
func extractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

// ParseToken parses the JWT token and returns the claims
func parseToken(tokenString string, keyFunc jwt.Keyfunc) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &UserClaims{}, keyFunc)
}
