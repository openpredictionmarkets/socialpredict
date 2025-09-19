package privateuser

import (
	"encoding/json"
	"net/http"
	"socialpredict/handlers/users/publicuser"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
)

type CombinedUserResponse struct {
	// Private fields
	models.PrivateUser
	// Public fields
	Username              string `json:"username"`
	DisplayName           string `json:"displayname"`
	UserType              string `json:"usertype"`
	InitialAccountBalance int64  `json:"initialAccountBalance"`
	AccountBalance        int64  `json:"accountBalance"`
	PersonalEmoji         string `json:"personalEmoji,omitempty"`
	Description           string `json:"description,omitempty"`
	PersonalLink1         string `json:"personalink1,omitempty"`
	PersonalLink2         string `json:"personalink2,omitempty"`
	PersonalLink3         string `json:"personalink3,omitempty"`
	PersonalLink4         string `json:"personalink4,omitempty"`
}

func GetPrivateProfileUserResponse(w http.ResponseWriter, r *http.Request) {
	// Use database connection
	db := util.GetDB()

	// Validate the token and get the user
	user, httpErr := middleware.ValidateTokenAndGetUser(r, db)
	if httpErr != nil {
		http.Error(w, "Invalid token: "+httpErr.Error(), http.StatusUnauthorized)
		return
	}

	// The username is extracted from the token
	username := user.Username

	publicInfo := publicuser.GetPublicUserInfo(db, username)

	response := CombinedUserResponse{
		// Private fields
		PrivateUser: user.PrivateUser,
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
