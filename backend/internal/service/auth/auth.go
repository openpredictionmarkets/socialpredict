package auth

import (
	"net/http"
	"strings"

	dusers "socialpredict/internal/domain/users"

	"github.com/golang-jwt/jwt/v4"
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
func ValidateUserAndEnforcePasswordChangeGetUser(r *http.Request, svc dusers.ServiceInterface) (*dusers.User, *HTTPError) {
	user, httpErr := ValidateTokenAndGetUser(r, svc)
	if httpErr != nil {
		return nil, httpErr
	}

	mustChange, err := svc.MustChangePassword(r.Context(), user.Username)
	if err != nil {
		return nil, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "Failed to check password status"}
	}

	if httpErr := CheckMustChangePasswordFlag(mustChange); httpErr != nil {
		return nil, httpErr
	}

	return user, nil
}

// ValidateTokenAndGetUser checks that the user is who they claim to be, and returns their information for use
func ValidateTokenAndGetUser(r *http.Request, svc dusers.ServiceInterface) (*dusers.User, *HTTPError) {
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
		user, err := svc.GetUser(r.Context(), claims.Username)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				return nil, &HTTPError{StatusCode: http.StatusNotFound, Message: "User not found"}
			}
			return nil, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "Failed to load user"}
		}
		return user, nil
	}
	return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Invalid token"}
}

// CheckMustChangePasswordFlag checks if a password change is required.
func CheckMustChangePasswordFlag(mustChange bool) *HTTPError {
	if mustChange {
		return &HTTPError{
			StatusCode: http.StatusForbidden,
			Message:    "Password change required",
		}
	}
	return nil
}
