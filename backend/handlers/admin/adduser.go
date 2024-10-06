package adminhandlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"

	"github.com/brianvoe/gofakeit"
	"gorm.io/gorm"
)

func AddUserHandler(loadEconConfig setup.EconConfigLoader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			log.Printf("AddUserHandler: %v", err)
			return
		}

		if match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", req.Username); !match {
			err := fmt.Errorf("username %s must only contain letters and numbers", req.Username)
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("AddUserHandler: %v", err)
			return
		}

		db := util.GetDB()

		// validate that the user performing this function is indeed admin
		middleware.ValidateAdminToken(r, db)

		appConfig := loadEconConfig()
		user := models.User{
			PublicUser: models.PublicUser{
				Username:              req.Username,
				DisplayName:           util.UniqueDisplayName(db),
				UserType:              "REGULAR",
				InitialAccountBalance: appConfig.Economics.User.InitialAccountBalance,
				AccountBalance:        appConfig.Economics.User.InitialAccountBalance,
				PersonalEmoji:         randomEmoji(),
			},
			PrivateUser: models.PrivateUser{
				Email:  util.UniqueEmail(db),
				APIKey: util.GenerateUniqueApiKey(db),
			},
			MustChangePassword: true,
		}

		// Check uniqueness of username, displayname, and email
		if err := checkUniqueFields(db, &user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("AddUserHandler: %v", err)
			return
		}

		password := gofakeit.Password(true, true, true, false, false, 12)
		if err := user.HashPassword(password); err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			log.Printf("AddUserHandler: %v", err)
			return
		}

		if result := db.Create(&user); result.Error != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			log.Printf("AddUserHandler: %v", result.Error)
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
}

func checkUniqueFields(db *gorm.DB, user *models.User) error {
	// Check for existing users with the same username, display name, email, or API key.
	var count int64
	db.Model(&models.User{}).Where(
		"username = ? OR display_name = ? OR email = ? OR api_key = ?",
		user.Username, user.DisplayName, user.Email, user.APIKey,
	).Count(&count)

	if count > 0 {
		return fmt.Errorf("username, display name, email, or API key already in use")
	}

	return nil
}

func randomEmoji() string {
	emojis := []string{"😀", "😃", "😄", "😁", "😆"}
	return emojis[rand.Intn(len(emojis))]
}
