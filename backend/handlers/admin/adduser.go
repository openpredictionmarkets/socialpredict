package adminhandlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"socialpredict/models"
	"socialpredict/util"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	if match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", req.Username); !match {
		http.Error(w, "Username must only contain letters and numbers", http.StatusBadRequest)
		return
	}

	db := util.GetDB()
	var user models.User

	user = models.User{
		Username:              req.Username,
		DisplayName:           uniqueDisplayName(db),
		Email:                 uniqueEmail(db),
		UserType:              "REGULAR",
		InitialAccountBalance: 0,
		PersonalEmoji:         randomEmoji(),
		ApiKey:                generateUniqueApiKey(db),
	}

	// Check uniqueness of username, displayname, and email
	if err := checkUniqueFields(db, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	password := gofakeit.Password(true, true, true, false, false, 12)
	if err := user.HashPassword(password); err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	if result := db.Create(&user); result.Error != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	responseData := map[string]interface{}{
		"message":  "User created successfully",
		"username": user.Username,
		"password": password,
		"usertype": user.UserType,
	}
	json.NewEncoder(w).Encode(responseData)
}

func checkUniqueFields(db *gorm.DB, user *models.User) error {
	// Check for existing users with the same username, display name, email, or API key.
	var count int64
	db.Model(&models.User{}).Where(
		"username = ? OR display_name = ? OR email = ? OR api_key = ?",
		user.Username, user.DisplayName, user.Email, user.ApiKey,
	).Count(&count)

	if count > 0 {
		return fmt.Errorf("username, display name, email, or API key already in use")
	}

	return nil
}

func generateUniqueApiKey(db *gorm.DB) string {
	for {
		apiKey := uuid.NewString()
		if count := countByField(db, "api_key", apiKey); count == 0 {
			return apiKey
		}
	}
}

func uniqueDisplayName(db *gorm.DB) string {
	// Generate display name and check for uniqueness
	for {
		name := gofakeit.Name()
		if count := countByField(db, "display_name", name); count == 0 {
			return name
		}
	}
}

func uniqueEmail(db *gorm.DB) string {
	// Generate email and check for uniqueness
	for {
		email := gofakeit.Email()
		if count := countByField(db, "email", email); count == 0 {
			return email
		}
	}
}

func countByField(db *gorm.DB, field, value string) int64 {
	var count int64
	db.Model(&models.User{}).Where(fmt.Sprintf("%s = ?", field), value).Count(&count)
	return count
}

func randomEmoji() string {
	emojis := []string{"😀", "😃", "😄", "😁", "😆"}
	return emojis[rand.Intn(len(emojis))]
}
