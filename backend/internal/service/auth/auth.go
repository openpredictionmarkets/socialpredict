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
func ValidateUserAndEnforcePasswordChangeGetUser(r *http.Request, svc dusers.ServiceInterface) (*dusers.User, *AuthError) {
	user, authErr := ValidateTokenAndGetUser(r, svc)
	if authErr != nil {
		return nil, authErr
	}

	if authErr := CheckMustChangePasswordFlag(user); authErr != nil {
		return nil, authErr
	}

	return user, nil
}

// ValidateTokenAndGetUser checks that the user is who they claim to be, and returns their information for use
func ValidateTokenAndGetUser(r *http.Request, svc dusers.ServiceInterface) (*dusers.User, *AuthError) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, newAuthError(ErrorKindMissingToken)
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})
	if err != nil {
		return nil, newAuthError(ErrorKindInvalidToken)
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		user, err := svc.GetUser(r.Context(), claims.Username)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				return nil, newAuthError(ErrorKindUserNotFound)
			}
			return nil, newAuthError(ErrorKindUserLoadFailed)
		}
		return user, nil
	}
	return nil, newAuthError(ErrorKindInvalidToken)
}

// CheckMustChangePasswordFlag checks if the user needs to change their password
func CheckMustChangePasswordFlag(user *dusers.User) *AuthError {
	if user.MustChangePassword {
		return newAuthError(ErrorKindPasswordChangeRequired)
	}
	return nil
}
