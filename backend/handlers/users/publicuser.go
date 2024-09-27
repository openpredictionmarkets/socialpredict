package usershandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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

	// Call GetPublicUserInfo and handle the error
	response, err := GetPublicUserInfo(db, username)
	if err != nil {
		// Check if the error is because the user was not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User not found, return 404 Not Found
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			// Internal server error
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			// Optionally, log the error for debugging purposes
			log.Printf("Failed to get user info: %v", err)
		}
		return
	}

	// If no error, write the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Function to get the Info From the Database
func GetPublicUserInfo(db *gorm.DB, username string) (PublicUserType, error) {

	gormDatabase := &repository.GormDatabase{DB: db}
	repo := repository.NewUserRepository(gormDatabase)

	user, err := repo.GetUserByUsername(username)
	if err != nil {
		return PublicUserType{}, fmt.Errorf("failed to get user by username: %w", err)
	}

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
	}, nil
}

func GetAllPublicUsers(db *gorm.DB) ([]PublicUserType, error) {
	var publicUsers []PublicUserType

	gormDatabase := &repository.GormDatabase{DB: db}
	repo := repository.NewUserRepository(gormDatabase)

	users, err := repo.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	// Convert each User model to PublicUserType to keep sensitive info secure
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
