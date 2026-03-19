package publicuser

import (
	"encoding/json"
	"errors"
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

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user.PublicUser)
}

// Function to get the users public info From the Database
func GetPublicUserInfo(db *gorm.DB, username string) models.PublicUser {
	var user models.User
	db.Where("username = ?", username).First(&user)

	return user.PublicUser
}
