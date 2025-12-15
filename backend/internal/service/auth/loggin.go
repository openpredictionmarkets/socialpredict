package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"socialpredict/models"
	"socialpredict/security"
	"socialpredict/util"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// login and validation stuff
// getJWTKey returns the JWT signing key, checking environment variable at runtime
func getJWTKey() []byte {
	return []byte(os.Getenv("JWT_SIGNING_KEY"))
}

// UserClaims represents the expected structure of the JWT claims
type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	user, loginErr := authenticateUser(req)
	if loginErr != nil {
		http.Error(w, loginErr.message, loginErr.statusCode)
		return
	}

	tokenString, err := generateJWT(user.Username)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	writeLoginResponse(w, user, tokenString)
}

type loginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Password string `json:"password" validate:"required,min=1"`
}

func decodeLoginRequest(r *http.Request) (loginRequest, error) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return loginRequest{}, fmt.Errorf("Error reading request body")
	}
	return req, nil
}

func validateAndSanitizeLogin(securityService *security.SecurityService, req loginRequest) (loginRequest, error) {
	if err := securityService.Validator.ValidateStruct(req); err != nil {
		return req, fmt.Errorf("Invalid input: %w", err)
	}

	sanitizedUsername, err := securityService.Sanitizer.SanitizeUsername(req.Username)
	if err != nil {
		return req, fmt.Errorf("Invalid username format")
	}
	req.Username = sanitizedUsername
	return req, nil
}

type loginError struct {
	message    string
	statusCode int
}

func authenticateUser(req loginRequest) (models.User, *loginError) {
	user, err := findUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, &loginError{message: "Invalid Credentials", statusCode: http.StatusUnauthorized}
		}
		return models.User{}, &loginError{message: "Error accessing database", statusCode: http.StatusInternalServerError}
	}

	if !user.CheckPasswordHash(req.Password) {
		return models.User{}, &loginError{message: "Invalid Credentials", statusCode: http.StatusUnauthorized}
	}

	return user, nil
}

func findUserByUsername(username string) (models.User, error) {
	db := util.GetDB()
	var user models.User
	result := db.Where("username = ?", username).First(&user)
	return user, result.Error
}

func generateJWT(username string) (string, error) {
	claims := &UserClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTKey())
}

func writeLoginResponse(w http.ResponseWriter, user models.User, tokenString string) {
	w.Header().Set("Content-Type", "application/json")

	responseData := map[string]interface{}{
		"token":              tokenString,
		"username":           user.Username,
		"usertype":           user.UserType,
		"mustChangePassword": user.MustChangePassword,
	}
	_ = json.NewEncoder(w).Encode(responseData)
}
