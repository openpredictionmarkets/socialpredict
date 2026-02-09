package auth

import (
	"net/http"

	dusers "socialpredict/internal/domain/users"
)

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Here you would verify the JWT token or session
		// If it's valid, call next.ServeHTTP to pass to the next handler
		// Otherwise, return an error
	})
}

// ValidateTokenAndGetUser checks that the user is who they claim to be, and returns their information for use.
// Accepts any UserReader (including dusers.ServiceInterface implementors).
func ValidateTokenAndGetUser(r *http.Request, users UserReader) (*dusers.User, *HTTPError) {
	tokenString, err := extractTokenFromHeader(r)
	if err != nil {
		return nil, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Authorization header is required"}
	}

	identity := NewIdentityService(users)
	user, err := identity.UserFromToken(r.Context(), tokenString)
	if err != nil {
		return nil, mapIdentityError(err)
	}
	return user, nil
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
