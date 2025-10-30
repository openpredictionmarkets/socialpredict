package middleware

import (
	"net/http"
	"strings"

	"socialpredict/models"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// ValidateTokenAndGetUserFromDB retains the legacy DB-backed authentication path.
func ValidateTokenAndGetUserFromDB(r *http.Request, db *gorm.DB) (*models.User, *HTTPError) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Authorization header is required"}
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})
	if err != nil {
		return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Invalid token"}
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		var user models.User
		result := db.Where("username = ?", claims.Username).First(&user)
		if result.Error != nil {
			return nil, &HTTPError{StatusCode: http.StatusNotFound, Message: "User not found"}
		}
		return &user, nil
	}
	return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Invalid token"}
}

// ValidateUserAndEnforcePasswordChangeGetUserFromDB mirrors the legacy helper but keeps the DB dependency isolated here.
func ValidateUserAndEnforcePasswordChangeGetUserFromDB(r *http.Request, db *gorm.DB) (*models.User, *HTTPError) {
	user, httpErr := ValidateTokenAndGetUserFromDB(r, db)
	if httpErr != nil {
		return nil, httpErr
	}

	if user.MustChangePassword {
		return nil, &HTTPError{StatusCode: http.StatusForbidden, Message: "Password change required"}
	}

	return user, nil
}
