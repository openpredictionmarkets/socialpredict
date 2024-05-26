package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"socialpredict/models"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

// ValidateAdminToken checks if the authenticated user is an admin
// It returns error if not an admin or if any validation fails
func ValidateAdminToken(r *http.Request, db *gorm.DB) error {
	tokenString, err := extractTokenFromHeader(r)
	if err != nil {
		return err
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	}

	token, err := parseToken(tokenString, keyFunc)
	if err != nil {
		return errors.New("invalid token")
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		var user models.User
		result := db.Where("username = ?", claims.Username).First(&user)
		if result.Error != nil {
			return fmt.Errorf("user not found")
		}
		if user.UserType != "ADMIN" {
			return fmt.Errorf("access denied for non-ADMIN users")
		}

		return nil
	}

	return errors.New("invalid token")
}
