package usershandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
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

	db := util.GetDB()
	user, httperr := middleware.ValidateTokenAndGetUser(r, db)
	if httperr != nil {
		http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
		return
	}

	var request ChangeEmojiRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if request.Emoji == "" {
		http.Error(w, "Emoji cannot be blank", http.StatusBadRequest)
		return
	}

	user.PersonalEmoji = request.Emoji
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update emoji: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
