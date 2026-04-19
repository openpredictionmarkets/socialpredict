package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"socialpredict/internal/domain/boundary"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/security"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// login and validation stuff
// getJWTKey returns the JWT signing key, checking environment variable at runtime
func getJWTKey() []byte {
	return []byte(strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY")))
}

// LoginUserRepository exposes only the user lookup required for login.
type LoginUserRepository interface {
	FindAuthenticatedUser(ctx context.Context, username string) (*boundary.AuthenticatedUser, error)
}

// UserClaims represents the expected structure of the JWT claims
type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func LoginHandler(users LoginUserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}

		securityService := security.NewSecurityService()

		req, err := decodeLoginRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		req, err = validateAndSanitizeLogin(securityService, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if users == nil {
			http.Error(w, "Error accessing database", http.StatusInternalServerError)
			return
		}

		user, loginErr := authenticateUser(r.Context(), users, req)
		if loginErr != nil {
			http.Error(w, loginErr.message, loginErr.statusCode)
			return
		}

		jwtKey := getJWTKey()
		if len(jwtKey) == 0 {
			http.Error(w, "Error creating token", http.StatusInternalServerError)
			return
		}

		tokenString, err := generateJWT(user.Username, jwtKey)
		if err != nil {
			http.Error(w, "Error creating token", http.StatusInternalServerError)
			return
		}

		writeLoginResponse(w, user, tokenString)
	}
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
	message    string
	statusCode int
}

func authenticateUser(ctx context.Context, users LoginUserRepository, req loginRequest) (boundary.AuthenticatedUser, *loginError) {
	if users == nil {
		return boundary.AuthenticatedUser{}, &loginError{message: "Error accessing database", statusCode: http.StatusInternalServerError}
	}

	user, err := findUserByUsername(ctx, users, req.Username)
	if err != nil {
		if errors.Is(err, dusers.ErrUserNotFound) {
			return boundary.AuthenticatedUser{}, &loginError{message: "Invalid Credentials", statusCode: http.StatusUnauthorized}
		}
		return boundary.AuthenticatedUser{}, &loginError{message: "Error accessing database", statusCode: http.StatusInternalServerError}
	}

	if !user.CheckPasswordHash(req.Password) {
		return boundary.AuthenticatedUser{}, &loginError{message: "Invalid Credentials", statusCode: http.StatusUnauthorized}
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

func writeLoginResponse(w http.ResponseWriter, user boundary.AuthenticatedUser, tokenString string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	responseData := map[string]interface{}{
		"token":              tokenString,
		"username":           user.Username,
		"usertype":           user.UserType,
		"mustChangePassword": user.MustChangePassword,
	}
	_ = json.NewEncoder(w).Encode(responseData)
}
