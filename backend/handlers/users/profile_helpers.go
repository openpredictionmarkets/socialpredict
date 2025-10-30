package usershandlers

import (
	"errors"
	"net/http"
	"strings"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
)

func writeProfileError(w http.ResponseWriter, err error, field string) {
	switch {
	case errors.Is(err, dusers.ErrUserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
	case errors.Is(err, dusers.ErrInvalidUserData):
		http.Error(w, "Invalid user data", http.StatusBadRequest)
	case errors.Is(err, dusers.ErrInvalidCredentials):
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
	default:
		message := err.Error()
		if isValidationError(message) {
			http.Error(w, message, http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to update "+field+": "+message, http.StatusInternalServerError)
	}
}

func isValidationError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "invalid") ||
		strings.Contains(lower, "exceeds") ||
		strings.Contains(lower, "must") ||
		strings.Contains(lower, "cannot") ||
		strings.Contains(lower, "required")
}

func toPrivateUserResponse(user *dusers.User) dto.PrivateUserResponse {
	if user == nil {
		return dto.PrivateUserResponse{}
	}

	return dto.PrivateUserResponse{
		ID:                    user.ID,
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
		Email:                 user.Email,
		APIKey:                user.APIKey,
		MustChangePassword:    user.MustChangePassword,
	}
}
