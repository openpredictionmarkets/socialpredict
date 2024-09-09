package usershandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/models"
	"socialpredict/repository"
	"socialpredict/util"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// PublicUserType is a struct for user data that is safe to send to the client for Profiles
type PublicUserType struct {
	Username              string `json:"username"`
	DisplayName           string `json:"displayname" gorm:"unique;not null"`
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

func GetPublicUserResponse(w http.ResponseWriter, r *http.Request) {
	// Extract the username from the URL
	vars := mux.Vars(r)
	username := vars["username"]

	db := util.GetDB()

	response := GetPublicUserInfo(db, username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Function to get the Info From the Database
func GetPublicUserInfo(db *gorm.DB, username string) PublicUserType {

	var user models.User
	db.Where("username = ?", username).First(&user)

	return PublicUserType{
		Username:              user.Username,
		DisplayName:           user.DisplayName,
		UserType:              user.UserType,
		InitialAccountBalance: user.InitialAccountBalance,
		AccountBalance:        user.AccountBalance,
		PersonalEmoji:         user.PersonalEmoji,
		Description:           user.Description,
		PersonalLink1:         user.PersonalLink1,
		PersonalLink2:         user.PersonalLink2,
		PersonalLink3:         user.PersonalLink3,
		PersonalLink4:         user.PersonalLink4,
	}
}

// GetAllPublicUsers fetches all users from the database and converts them to PublicUserType.
func GetAllPublicUsers(db *gorm.DB) ([]PublicUserType, error) {
	var users []models.User
	var publicUsers []PublicUserType

	// Fetch all users from the database
	repo := repository.NewUserRepository(db)
	users, err := repo.GetAllUsers()
	if err != nil {
		log.Fatalf("Failed to get all users: %v", err)
	}

	// Convert each User model to PublicUserType
	for _, user := range users {
		publicUser := PublicUserType{
			Username:              user.Username,
			DisplayName:           user.DisplayName,
			UserType:              user.UserType,
			InitialAccountBalance: user.InitialAccountBalance,
			AccountBalance:        user.AccountBalance,
			PersonalEmoji:         user.PersonalEmoji,
			Description:           user.Description,
			PersonalLink1:         user.PersonalLink1,
			PersonalLink2:         user.PersonalLink2,
			PersonalLink3:         user.PersonalLink3,
			PersonalLink4:         user.PersonalLink4,
		}
		publicUsers = append(publicUsers, publicUser)
	}

	return publicUsers, nil
}
