package privateuser

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

func GetPrivateProfileHandler(svc dusers.ServiceInterface, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, httperr := auth.RequireUser(r)
		if httperr != nil {
			http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
			return
		}

		profile, err := svc.GetPrivateProfile(r.Context(), user.Username)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to fetch user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(privateProfileResponse(profile)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func privateProfileResponse(profile *dusers.PrivateProfile) dto.PrivateUserResponse {
	if profile == nil {
		return dto.PrivateUserResponse{}
	}

	return dto.PrivateUserResponse{
		ID:                    profile.ID,
		Username:              profile.Username,
		DisplayName:           profile.DisplayName,
		UserType:              profile.UserType,
		InitialAccountBalance: profile.InitialAccountBalance,
		AccountBalance:        profile.AccountBalance,
		PersonalEmoji:         profile.PersonalEmoji,
		Description:           profile.Description,
		PersonalLink1:         profile.PersonalLink1,
		PersonalLink2:         profile.PersonalLink2,
		PersonalLink3:         profile.PersonalLink3,
		PersonalLink4:         profile.PersonalLink4,
		Email:                 profile.Email,
		APIKey:                profile.APIKey,
		MustChangePassword:    profile.MustChangePassword,
	}
}
