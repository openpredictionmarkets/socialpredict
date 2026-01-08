package usershandlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
)

// GetPublicUserHandler returns an HTTP handler that fetches public user information via the users service.
func GetPublicUserHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := mux.Vars(r)["username"]
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}

		user, err := svc.GetPublicUser(r.Context(), username)
		if err != nil {
			switch err {
			case dusers.ErrUserNotFound:
				http.Error(w, "user not found", http.StatusNotFound)
			default:
				http.Error(w, "failed to fetch user", http.StatusInternalServerError)
			}
			return
		}

		response := dto.PublicUserResponse{
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

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
