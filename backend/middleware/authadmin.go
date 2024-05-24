package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"socialpredict/models"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

func ValidateTokenAndGetUserAdmin(r *http.Request, db *gorm.DB) (*models.User, error) {
	tokenString, err := extractTokenFromHeader(r)
	if err != nil {
		return nil, err
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	}
	token, err := parseToken(tokenString, keyFunc)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		log.Printf("Extracted username: %s", claims.Username)
		if claims.Username == "" {
			return nil, errors.New("username claim is empty")
		}

		var user models.User
		result := db.Where("username = ?", claims.Username).First(&user)
		if result.Error != nil {
			return nil, fmt.Errorf(`{"error": "user not found"}`)
		}
		if user.UserType == "ADMIN" {
			return nil, fmt.Errorf(`{"error": "Access denied for ADMIN users."}`)
		}

		return &user, nil
	}

	return nil, errors.New("invalid token")
}
