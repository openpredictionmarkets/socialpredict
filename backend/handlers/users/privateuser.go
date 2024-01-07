package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"

	"gorm.io/gorm"
)

// PrivateUserResponse is a struct for user data that is safe to send to the client for login
type PrivateUserResponse struct {
	ID                    uint    `json:"id"`
	Username              string  `json:"username"`
	DisplayName           string  `json:"displayname" gorm:"unique;not null"`
	Email                 string  `json:"email"`
	UserType              string  `json:"usertype"`
	InitialAccountBalance float64 `json:"initialAccountBalance"`
	AccountBalance        float64 `json:"accountBalance"`
	ApiKey                string  `json:"apiKey,omitempty" gorm:"unique"`
	PersonalEmoji         string  `json:"personalEmoji,omitempty"`
	Description           string  `json:"description,omitempty"`
	PersonalLink1         string  `json:"personalink1,omitempty"`
	PersonalLink2         string  `json:"personalink2,omitempty"`
	PersonalLink3         string  `json:"personalink3,omitempty"`
	PersonalLink4         string  `json:"personalink4,omitempty"`
}

func PrivateUserResponseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not supported", http.StatusNotFound)
		return
	}

	// Extract userID from query parameters
	queryValues := r.URL.Query()
	userID, err := strconv.Atoi(queryValues.Get("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db := util.GetDB()
	var user models.User

	// Fetch the user from the database
	result := db.First(&user, userID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error fetching user: %v", result.Error)
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}

	// Create a response object excluding sensitive data
	userResponse := PrivateUserResponse{
		ID:                    user.ID,
		Username:              user.Username,
		Email:                 user.Email,
		UserType:              user.UserType,
		InitialAccountBalance: user.InitialAccountBalance,
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Encode the userResponse to JSON and send it as a response
	if err := json.NewEncoder(w).Encode(userResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getMarketUsers returns the number of unique users for a given market
func GetMarketUsers(bets []models.Bet) int {
	userMap := make(map[string]bool)
	for _, bet := range bets {
		userMap[bet.Username] = true
	}

	return len(userMap)
}
