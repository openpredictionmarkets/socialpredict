package middleware

import (
	"net/http"
	"socialpredict/models"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Here you would verify the JWT token or session
		// If it's valid, call next.ServeHTTP to pass to the next handler
		// Otherwise, return an error
	})
}

// ValidateUserAndEnforcePasswordChange performs user validation and checks if a password change is required.
// It returns the user and any errors encountered.
func ValidateUserAndEnforcePasswordChangeGetUser(r *http.Request, db *gorm.DB) (*models.User, *HTTPError) {
	user, httpErr := ValidateTokenAndGetUser(r, db)
	if httpErr != nil {
		return nil, httpErr
	}

	// Check if a password change is required
	if httpErr := CheckMustChangePasswordFlag(user); httpErr != nil {
		return nil, httpErr
	}

	return user, nil
}

// ValidateTokenAndGetUser checks that the user is who they claim to be, and returns their information for use
func ValidateTokenAndGetUser(r *http.Request, db *gorm.DB) (*models.User, *HTTPError) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Authorization header is required"}
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})
	if err != nil {
		return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Error parsing token: " + err.Error()} // fix: this will be reached if the current password is incorrect, but the "new password" field value and "confirm new password" field value are matching (includes empty values). this is probably accurate and meaningful to the user. But is it a security risk to say "the current password is incorrect"?
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		var user models.User
		result := db.Where("username = ?", claims.Username).First(&user)

		if result.Error != nil {
			return nil, &HTTPError{StatusCode: http.StatusNotFound, Message: "User not found"}
		}

		if user.UserType == "ADMIN" {
			return nil, &HTTPError{StatusCode: http.StatusForbidden, Message: "Access denied for ADMIN users"}
		}

		return &user, nil
	}

	return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Invalid token"}
}

// CheckMustChangePasswordFlag checks if the user needs to change their password
func CheckMustChangePasswordFlag(user *models.User) *HTTPError {
	if user.MustChangePassword {
		return &HTTPError{
			StatusCode: http.StatusForbidden,
			Message:    "Password change required",
		}
	}
	return nil
}
