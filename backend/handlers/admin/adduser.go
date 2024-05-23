package adminhandlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"regexp"
	"socialpredict/models"
	"socialpredict/util"

	"github.com/brianvoe/gofakeit"
)

func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	// Decode JSON from request body
	var req struct {
		Username string `json:"username"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// Validate username
	if match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", req.Username); !match {
		http.Error(w, "Username must only contain letters and numbers", http.StatusBadRequest)
		return
	}

	// Create user model instance
	user := models.User{
		Username:              req.Username,
		DisplayName:           gofakeit.Name(),
		UserType:              "REGULAR",
		InitialAccountBalance: 0,
		Email:                 "",
		ApiKey:                "",
		PersonalEmoji:         randomEmoji(),
		Description:           "",
		PersonalLink1:         "",
		PersonalLink2:         "",
		PersonalLink3:         "",
		PersonalLink4:         "",
	}

	// Generate a random password
	password := gofakeit.Password(true, true, true, false, false, 12)
	err = user.HashPassword(password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	db := util.GetDB() // Assuming you have a way to retrieve your DB connection
	result := db.Create(&user)
	if result.Error != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Prepare and send the response
	responseData := map[string]interface{}{
		"message":  "User created successfully",
		"username": user.Username,
		"password": password,
		"usertype": user.UserType,
	}
	json.NewEncoder(w).Encode(responseData)
}

func randomEmoji() string {
	emojis := []string{"😀", "😃", "😄", "😁", "😆"}
	return emojis[rand.Intn(len(emojis))]
}
