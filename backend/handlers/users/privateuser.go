package handlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/models"
	"socialpredict/util"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// PrivateUserResponse is a struct for user data that is safe to send to the client for login
type PrivateUserResponse struct {
	Email  string `json:"email"`
	ApiKey string `json:"apiKey,omitempty"`
}

func PrivateUserResponseHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the username from the URL
	vars := mux.Vars(r)
	username := vars["username"]

	db := util.GetDB()

	response := GetPrivateUserInfo(db, username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Function to get the Info From the Database
func GetPrivateUserInfo(db *gorm.DB, username string) PrivateUserResponse {
	var user models.User
	db.Where("username = ?", username).First(&user)

	return PrivateUserResponse{
		Email:  user.Email,
		ApiKey: user.ApiKey,
	}
}
