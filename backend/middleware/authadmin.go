package middleware

import (
	"errors"
	"fmt"
	"net/http"

	dusers "socialpredict/internal/domain/users"

	"github.com/golang-jwt/jwt/v4"
)

// ValidateAdminToken checks if the authenticated user is an admin
// It returns error if not an admin or if any validation fails
func ValidateAdminToken(r *http.Request, svc dusers.ServiceInterface) error {
	tokenString, err := extractTokenFromHeader(r)
	if err != nil {
		return err
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	}

	token, err := parseToken(tokenString, keyFunc)
	if err != nil {
		return errors.New("invalid token")
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		user, err := svc.GetUser(r.Context(), claims.Username)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				return fmt.Errorf("user not found")
			}
			return fmt.Errorf("failed to load user")
		}
		if user.UserType != "ADMIN" {
			return fmt.Errorf("access denied for non-ADMIN users")
		}

		return nil
	}

	return errors.New("invalid token")
}
