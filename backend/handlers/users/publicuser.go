package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/models"
	"socialpredict/util"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func GetPublicUserResponse(w http.ResponseWriter, r *http.Request) {
	// Extract the username from the URL
	vars := mux.Vars(r)
	username := vars["username"]

	db := util.GetDB()

	response := GetPublicUserInfo(db, username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Function to get the users public info From the Database
func GetPublicUserInfo(db *gorm.DB, username string) models.PublicUser {
	var user models.User
	db.Where("username = ?", username).First(&user)

	return user.PublicUser
}
