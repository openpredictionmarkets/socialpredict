package privateuser

import (
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

func GetPrivateProfileHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, authErr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, svc)
		if authErr != nil {
			_ = authhttp.WriteFailure(w, authErr)
			return
		}

		profile, err := svc.GetPrivateProfile(r.Context(), user.Username)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
				return
			}
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		_ = handlers.WriteResult(w, http.StatusOK, privateProfileResponse(profile))
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
