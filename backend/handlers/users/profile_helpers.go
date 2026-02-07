package usershandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"socialpredict/handlers/users/dto"
	"socialpredict/internal/domain/auth"
	dusers "socialpredict/internal/domain/users"
)

func writeProfileError(w http.ResponseWriter, err error, field string) {
	switch {
	case errors.Is(err, dusers.ErrUserNotFound):
		writeProfileJSONError(w, http.StatusNotFound, "User not found")
	case errors.Is(err, dusers.ErrInvalidUserData):
		writeProfileJSONError(w, http.StatusBadRequest, "Invalid user data")
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeProfileJSONError(w, http.StatusUnauthorized, "Current password is incorrect")
	default:
		message := err.Error()
		if isValidationError(message) {
			writeProfileJSONError(w, http.StatusBadRequest, message)
			return
		}
		writeProfileJSONError(w, http.StatusInternalServerError, "Failed to update "+field+": "+message)
	}
}

func writeProfileJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(dto.ErrorResponse{Error: message}); err != nil {
		http.Error(w, message, statusCode)
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
	}
}
