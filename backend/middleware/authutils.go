package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return e.Message
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
