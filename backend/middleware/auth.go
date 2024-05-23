package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"socialpredict/models"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Here you would verify the JWT token or session
		// If it's valid, call next.ServeHTTP to pass to the next handler
		// Otherwise, return an error
	})
}

// validateTokenAndGetUser validates the JWT token and returns the user info
func ValidateTokenAndGetUser(r *http.Request, db *gorm.DB) (*models.User, error) {
	// Extract the token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("authorization header is required")
	}

	// Typically, the Authorization header is in the format "Bearer {token}"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Define a function to handle token parsing and claims extraction
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Here, you would return your JWT signing key, used to validate the token
		return jwtKey, nil
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		log.Printf("claims.Username is %s", claims.Username)
		if claims.Username == "" {
			return nil, errors.New("username claim is empty")
		}
		log.Printf("Extracted username: %s", claims.Username)
		var user models.User
		result := db.Where("username = ?", claims.Username).First(&user)
		if result.Error != nil {
			// Format the error message as JSON
			return nil, fmt.Errorf(`{"error": "user not found"}`)
		}
		// Check if the user is an admin and restrict access
		if user.UserType == "ADMIN" {
			return nil, &HTTPError{
				StatusCode: http.StatusForbidden,
				Message:    "Access denied for ADMIN users.",
			}
		}

		return &user, nil
	}

	return nil, errors.New("invalid token")
}
