package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"socialpredict/models"
	"socialpredict/util"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

// login and validation stuff
var jwtKey = []byte(os.Getenv("JWT_SIGNING_KEY"))

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

	// Parse the request body
	type loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Use database connection
	db := util.GetDB()

	// Find user by username
	var user models.User
	result := db.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error accessing database", http.StatusInternalServerError)
		return
	}

	// Check password
	if !user.CheckPasswordHash(req.Password) {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	// Create UserClaim
	claims := &UserClaims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	// Log for debugging
	log.Printf("Token issued for user: %s", user.Username)
	log.Printf("Tokenstring: %s", tokenString)

	// Send token, username, and usertype in the response
	responseData := map[string]interface{}{
		"token":              tokenString,
		"username":           user.Username,
		"usertype":           user.UserType,
		"mustChangePassword": user.MustChangePassword,
	}
	json.NewEncoder(w).Encode(responseData)
}
