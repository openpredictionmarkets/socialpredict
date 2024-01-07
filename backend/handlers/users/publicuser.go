package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"socialpredict/models"
	"socialpredict/util"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// PublicUserResponse is a struct for user data that is safe to send to the client for Profiles
type PublicUserResponse struct {
	Username              string  `json:"username"`
	DisplayName           string  `json:"displayname" gorm:"unique;not null"`
	UserType              string  `json:"usertype"`
	InitialAccountBalance float64 `json:"initialAccountBalance"`
	AccountBalance        float64 `json:"accountBalance"`
	PersonalEmoji         string  `json:"personalEmoji,omitempty"`
	Description           string  `json:"description,omitempty"`
	PersonalLink1         string  `json:"personalink1,omitempty"`
	PersonalLink2         string  `json:"personalink2,omitempty"`
	PersonalLink3         string  `json:"personalink3,omitempty"`
	PersonalLink4         string  `json:"personalink4,omitempty"`
}

func GetPublicUserResponse(w http.ResponseWriter, r *http.Request) {
	// Extract the username from the URL
	vars := mux.Vars(r)
	username := vars["username"]

	db := util.GetDB()
	var user models.User

	// Fetch user data from the database
	result := db.Where("Username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching user data", http.StatusInternalServerError)
		}
		return
	}

	// Convert to PublicUserResponse
	response := PublicUserResponse{
		Username:              user.Username,
		DisplayName:           user.DisplayName,
		UserType:              user.UserType,
		InitialAccountBalance: user.InitialAccountBalance,
		AccountBalance:        user.AccountBalance, // Added AccountBalance
		PersonalEmoji:         user.PersonalEmoji,
		Description:           user.Description,
		PersonalLink1:         user.PersonalLink1,
		PersonalLink2:         user.PersonalLink2,
		PersonalLink3:         user.PersonalLink3,
		PersonalLink4:         user.PersonalLink4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
