package usershandlers

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/middleware"
	"socialpredict/util"
)

// ChangeEmojiHandler returns an HTTP handler that delegates emoji updates to the users service.
func ChangeEmojiHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		db := util.GetDB()
		user, httperr := middleware.ValidateTokenAndGetUser(r, db)
		if httperr != nil {
			http.Error(w, "Invalid token: "+httperr.Error(), httperr.StatusCode)
			return
		}

		var request dto.ChangeEmojiRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		updated, err := svc.UpdateEmoji(r.Context(), user.Username, request.Emoji)
		if err != nil {
			writeProfileError(w, err, "emoji")
			return
		}

		user.PersonalEmoji = updated.PersonalEmoji

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
