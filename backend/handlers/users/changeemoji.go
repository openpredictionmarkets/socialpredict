package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/security"
	"socialpredict/util"
)

type ChangeEmojiRequest struct {
	Emoji string `json:"emoji"`
}

func ChangeEmoji(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Initialize security service
	securityService := security.NewSecurityService()

	db := util.GetDB()
	user, httpErr := middleware.ValidateTokenAndGetUser(r, db)
	if httpErr != nil {
		http.Error(w, "Invalid token: "+httpErr.Error(), http.StatusUnauthorized)
		return
	}

	var request ChangeEmojiRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate emoji length and content
	if len(request.Emoji) > 20 {
		http.Error(w, "Emoji exceeds maximum length of 20 characters", http.StatusBadRequest)
		return
	}

	if request.Emoji == "" {
		http.Error(w, "Emoji cannot be blank", http.StatusBadRequest)
		return
	}

	// Sanitize the emoji to prevent XSS
	sanitizedEmoji, err := securityService.Sanitizer.SanitizeEmoji(request.Emoji)
	if err != nil {
		http.Error(w, "Invalid emoji: "+err.Error(), http.StatusBadRequest)
		return
	}

	user.PersonalEmoji = sanitizedEmoji
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update emoji: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
