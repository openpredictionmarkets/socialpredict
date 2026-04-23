package usershandlers

import (
	"errors"
	"net/http"
	"strings"

	"socialpredict/handlers"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

func writeProfileError(w http.ResponseWriter, err error, field string) {
	switch {
	case errors.Is(err, dusers.ErrUserNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
	case errors.Is(err, dusers.ErrInvalidUserData):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, "INVALID_USER_DATA")
	case errors.Is(err, dusers.ErrInvalidCredentials):
		_ = handlers.WriteFailure(w, http.StatusUnauthorized, "INVALID_CREDENTIALS")
	default:
		message := err.Error()
		if isValidationError(message) {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
			return
		}
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, profileFailureReason(field))
	}
}

func writeProfileJSONError(w http.ResponseWriter, statusCode int, message string) {
	_ = handlers.WriteFailure(w, statusCode, handlers.FailureReason(message))
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

func profileFailureReason(field string) handlers.FailureReason {
	switch field {
	case "display name":
		return "DISPLAY_NAME_UPDATE_FAILED"
	case "description":
		return "DESCRIPTION_UPDATE_FAILED"
	case "emoji":
		return "EMOJI_UPDATE_FAILED"
	case "personal links":
		return "PERSONAL_LINKS_UPDATE_FAILED"
	default:
		return handlers.ReasonInternalError
	}
}

func profileAuthFailureReason(err *authsvc.HTTPError) handlers.FailureReason {
	if err == nil {
		return handlers.ReasonInternalError
	}

	switch err.Message {
	case "Authorization header is required", "Invalid token":
		return handlers.ReasonInvalidToken
	case "Password change required":
		return handlers.ReasonPasswordChangeRequired
	case "User not found":
		return handlers.ReasonUserNotFound
	default:
		if err.StatusCode >= http.StatusInternalServerError {
			return handlers.ReasonInternalError
		}
		return handlers.ReasonInvalidToken
	}
}
