package privateuser

import (
	"encoding/json"
	"net/http"

	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"

	dusers "socialpredict/internal/domain/users"
)

type CombinedUserResponse struct {
	// Private fields
	models.PrivateUser
	// Public fields
	Username              string `json:"username"`
	DisplayName           string `json:"displayname"`
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

func GetPrivateProfileHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := util.GetDB()

		user, httperr := middleware.ValidateTokenAndGetUser(r, db)
		if httperr != nil {
			http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
			return
		}

		publicInfo, err := svc.GetPublicUser(r.Context(), user.Username)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				http.Error(w, "user not found", http.StatusNotFound)
			} else {
				http.Error(w, "failed to fetch user", http.StatusInternalServerError)
			}
			return
		}

		response := CombinedUserResponse{
			PrivateUser:           user.PrivateUser,
			Username:              publicInfo.Username,
			DisplayName:           publicInfo.DisplayName,
			UserType:              publicInfo.UserType,
			InitialAccountBalance: publicInfo.InitialAccountBalance,
			AccountBalance:        publicInfo.AccountBalance,
			PersonalEmoji:         publicInfo.PersonalEmoji,
			Description:           publicInfo.Description,
			PersonalLink1:         publicInfo.PersonalLink1,
			PersonalLink2:         publicInfo.PersonalLink2,
			PersonalLink3:         publicInfo.PersonalLink3,
			PersonalLink4:         publicInfo.PersonalLink4,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
