package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"socialpredict/handlers"

	"socialpredict/internal/domain/boundary"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/security"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// LoginUserRepository exposes only the user lookup required for login.
type LoginUserRepository interface {
	FindAuthenticatedUser(ctx context.Context, username string) (*boundary.AuthenticatedUser, error)
}

// UserClaims represents the expected structure of the JWT claims
type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type loginResponse struct {
	Token              string `json:"token"`
	Username           string `json:"username"`
	UserType           string `json:"usertype"`
	MustChangePassword bool   `json:"mustChangePassword"`
}

func LoginHandler(users LoginUserRepository, securityService *security.SecurityService, jwtSigningKey ...[]byte) http.HandlerFunc {
	key := currentJWTSigningKey()
	if len(jwtSigningKey) > 0 {
		key = jwtSigningKey[0]
	}
	key = cloneJWTKey(key)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = writeLoginFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		req, err := decodeLoginRequest(r)
		if err != nil {
			_ = writeLoginFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		req, err = validateAndSanitizeLogin(securityService, req)
		if err != nil {
			_ = writeLoginFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
			return
		}

		if users == nil {
			_ = writeLoginFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		user, loginErr := authenticateUser(r.Context(), users, req)
		if loginErr != nil {
			_ = writeLoginFailure(w, loginErr.statusCode, loginFailureReason(loginErr.statusCode))
			return
		}

		if len(key) == 0 {
			_ = writeLoginFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		tokenString, err := generateJWT(user.Username, key)
		if err != nil {
			_ = writeLoginFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		_ = writeLoginResponse(w, user, tokenString)
	}
}

func cloneJWTKey(jwtSigningKey []byte) []byte {
	if len(jwtSigningKey) == 0 {
		return nil
	}
	return append([]byte(nil), jwtSigningKey...)
}

type loginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Password string `json:"password" validate:"required,min=1"`
}

func decodeLoginRequest(r *http.Request) (loginRequest, error) {
	var req loginRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return loginRequest{}, fmt.Errorf("error reading request body")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return loginRequest{}, fmt.Errorf("error reading request body")
	}
	return req, nil
}

func validateAndSanitizeLogin(securityService *security.SecurityService, req loginRequest) (loginRequest, error) {
	if securityService == nil {
		return req, fmt.Errorf("security service unavailable")
	}

	sanitizedUsername, err := securityService.Sanitizer.SanitizeUsername(req.Username)
	if err != nil {
		return req, fmt.Errorf("invalid input")
	}
	req.Username = sanitizedUsername

	if err := securityService.Validator.ValidateStruct(req); err != nil {
		return req, fmt.Errorf("invalid input")
	}

	return req, nil
}

type loginError struct {
	statusCode int
}

func authenticateUser(ctx context.Context, users LoginUserRepository, req loginRequest) (boundary.AuthenticatedUser, *loginError) {
	if users == nil {
		return boundary.AuthenticatedUser{}, &loginError{statusCode: http.StatusInternalServerError}
	}

	user, err := findUserByUsername(ctx, users, req.Username)
	if err != nil {
		if errors.Is(err, dusers.ErrUserNotFound) {
			return boundary.AuthenticatedUser{}, &loginError{statusCode: http.StatusUnauthorized}
		}
		return boundary.AuthenticatedUser{}, &loginError{statusCode: http.StatusInternalServerError}
	}

	if !user.CheckPasswordHash(req.Password) {
		return boundary.AuthenticatedUser{}, &loginError{statusCode: http.StatusUnauthorized}
	}

	return user, nil
}

func findUserByUsername(ctx context.Context, users LoginUserRepository, username string) (boundary.AuthenticatedUser, error) {
	if users == nil {
		return boundary.AuthenticatedUser{}, fmt.Errorf("database connection is not initialized")
	}

	user, err := users.FindAuthenticatedUser(ctx, username)
	if err != nil {
		return boundary.AuthenticatedUser{}, err
	}
	if user == nil {
		return boundary.AuthenticatedUser{}, dusers.ErrUserNotFound
	}
	return *user, nil
}

func generateJWT(username string, jwtKey []byte) (string, error) {
	if len(jwtKey) == 0 {
		return "", fmt.Errorf("missing JWT signing key")
	}

	claims := &UserClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func writeLoginResponse(w http.ResponseWriter, user boundary.AuthenticatedUser, tokenString string) error {
	return handlers.WriteResult(w, http.StatusOK, loginResponse{
		Token:              tokenString,
		Username:           user.Username,
		UserType:           user.UserType,
		MustChangePassword: user.MustChangePassword,
	})
}

func writeLoginFailure(w http.ResponseWriter, statusCode int, reason handlers.FailureReason) error {
	return handlers.WriteFailure(w, statusCode, reason)
}

func loginFailureReason(statusCode int) handlers.FailureReason {
	switch statusCode {
	case http.StatusUnauthorized:
		return handlers.ReasonAuthorizationDenied
	default:
		return handlers.ReasonInternalError
	}
}
