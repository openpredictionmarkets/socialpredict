package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"

	"gorm.io/gorm"
)

// PrivateUserResponse is a struct for user data that is safe to send to the client for login
type PrivateUserResponse struct {
	Email  string `json:"email"`
	ApiKey string `json:"apiKey,omitempty"`
}

type CombinedUserResponse struct {
	// Private fields
	Email  string `json:"email"`
	ApiKey string `json:"apiKey,omitempty"`
	// Public fields
	Username              string  `json:"username"`
	DisplayName           string  `json:"displayname"`
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

func GetPrivateProfileUserResponse(w http.ResponseWriter, r *http.Request) {
	// accept get requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Use database connection
	db := util.GetDB()

	// Validate the token and get the user
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// The username is extracted from the token
	username := user.Username

	publicInfo := GetPublicUserInfo(db, username)
	privateInfo := GetPrivateUserInfo(db, username)

	response := CombinedUserResponse{
		// Private fields
		Email:  privateInfo.Email,
		ApiKey: privateInfo.ApiKey,
		// Public fields
		Username:              publicInfo.Username,
		DisplayName:           publicInfo.DisplayName,
		UserType:              publicInfo.UserType,
		InitialAccountBalance: publicInfo.InitialAccountBalance,
		AccountBalance:        publicInfo.AccountBalance,
		PersonalEmoji:         publicInfo.PersonalEmoji,
		Description:           publicInfo.Description,
		PersonalLink1:         publicInfo.PersonalLink1,
		PersonalLink2:         publicInfo.PersonalLink2,
		PersonalLink3:         publicInfo.PersonalLink3,
		PersonalLink4:         publicInfo.PersonalLink4,
	}

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
