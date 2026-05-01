package auth

import (
	"net/http"
	"os"
	"strings"
	"sync"

	dusers "socialpredict/internal/domain/users"

	"github.com/golang-jwt/jwt/v4"
)

var (
	processJWTKeyMu sync.RWMutex
	processJWTKey   []byte
)

// ConfigureJWTSigningKey injects the runtime/bootstrap-owned JWT signing key for legacy helper call sites.
func ConfigureJWTSigningKey(jwtSigningKey []byte) {
	processJWTKeyMu.Lock()
	defer processJWTKeyMu.Unlock()
	processJWTKey = cloneJWTKey(jwtSigningKey)
}

func currentJWTSigningKey() []byte {
	processJWTKeyMu.RLock()
	defer processJWTKeyMu.RUnlock()
	if len(processJWTKey) > 0 {
		return cloneJWTKey(processJWTKey)
	}
	return []byte(strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY")))
}

func getJWTKey() []byte {
	return currentJWTSigningKey()
}

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
	return ValidateUserAndEnforcePasswordChangeGetUserWithSigningKey(r, svc, currentJWTSigningKey())
}

// ValidateUserAndEnforcePasswordChangeGetUserWithSigningKey performs user validation with an injected JWT key.
func ValidateUserAndEnforcePasswordChangeGetUserWithSigningKey(r *http.Request, svc dusers.ServiceInterface, jwtSigningKey []byte) (*dusers.User, *AuthError) {
	user, authErr := validateTokenAndGetUser(r, svc, jwtSigningKey)
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
	return validateTokenAndGetUser(r, svc, currentJWTSigningKey())
}

// ValidateTokenAndGetUserWithSigningKey checks that the user is who they claim to be using an injected JWT key.
func ValidateTokenAndGetUserWithSigningKey(r *http.Request, svc dusers.ServiceInterface, jwtSigningKey []byte) (*dusers.User, *AuthError) {
	return validateTokenAndGetUser(r, svc, jwtSigningKey)
}

func validateTokenAndGetUser(r *http.Request, svc dusers.ServiceInterface, jwtSigningKey []byte) (*dusers.User, *AuthError) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, newAuthError(ErrorKindMissingToken)
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if len(strings.Split(tokenString, ".")) != 3 {
		return nil, newAuthError(ErrorKindInvalidToken)
	}

	if len(jwtSigningKey) == 0 {
		return nil, newAuthError(ErrorKindServiceUnavailable)
	}

	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSigningKey, nil
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
