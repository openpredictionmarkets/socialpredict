package auth

import (
	"context"
	"errors"
	"strings"

	dusers "socialpredict/internal/domain/users"

	"github.com/golang-jwt/jwt/v4"
)

var (
	// ErrInvalidToken indicates a malformed, expired, or unparseable token.
	ErrInvalidToken = errors.New("invalid token")
	// ErrAdminPrivilegesRequired indicates the authenticated user is not an admin.
	ErrAdminPrivilegesRequired = errors.New("admin privileges required")
)

// IdentityService resolves users from tokens and enforces role-based authorization.
type IdentityService struct {
	users UserReader
}

// NewIdentityService constructs the token/identity service.
func NewIdentityService(users UserReader) *IdentityService {
	return &IdentityService{users: users}
}

// UserFromToken validates the token and resolves its user.
func (s *IdentityService) UserFromToken(ctx context.Context, tokenString string) (*dusers.User, error) {
	if s.users == nil {
		return nil, dusers.ErrInvalidUserData
	}

	token, err := parseToken(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || claims == nil || !token.Valid || strings.TrimSpace(claims.Username) == "" {
		return nil, ErrInvalidToken
	}

	return s.users.GetUser(ctx, claims.Username)
}

// EnsureAdmin verifies that the given user has admin privileges.
func (s *IdentityService) EnsureAdmin(user *dusers.User) error {
	if user == nil {
		return ErrInvalidToken
	}
	if strings.ToUpper(user.UserType) != "ADMIN" {
		return ErrAdminPrivilegesRequired
	}
	return nil
}

var _ IdentityResolver = (*IdentityService)(nil)
