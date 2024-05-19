package usershandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/util"
)

type ChangeEmojiRequest struct {
	Emoji string `json:"emoji"`
}

func ChangeEmoji(w http.ResponseWriter, r *http.Request) {
	log.Println("ChangeEmoji endpoint hit") // Initial log to confirm the endpoint is reached

	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		log.Printf("Error: Method %s not allowed", r.Method)
		return
	}

	db := util.GetDB()
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		log.Printf("Token validation failed: %s", err.Error()) // Log specific token validation failure
		return
	}

	var request ChangeEmojiRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		log.Printf("Error decoding request: %s", err.Error()) // Log specific decoding error
		return
	}

	log.Printf("Decoded request for user %s: %+v", user.Username, request) // Log the request along with the username

	user.PersonalEmoji = request.Emoji
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update emoji: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error saving emoji for user %s: %s", user.Username, err.Error()) // Log specific database save error
		return
	}

	log.Printf("Emoji updated successfully for user %s: %s", user.Username, user.PersonalEmoji) // Log success message

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
