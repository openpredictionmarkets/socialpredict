package usershandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/logger"
)

type changePasswordResult struct {
	Message string `json:"message"`
}

// ChangePasswordHandler returns an HTTP handler that delegates password changes to the users service.
func ChangePasswordHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.FailureReason("METHOD_NOT_ALLOWED"))
			return
		}

		logger.LogInfo("ChangePassword", "ChangePassword", "ChangePassword handler called")

		user, httperr := authsvc.ValidateTokenAndGetUser(r, svc)
		if httperr != nil {
			_ = handlers.WriteFailure(w, httperr.StatusCode, handlers.FailureReason("INVALID_TOKEN"))
			logger.LogError("ChangePassword", "ValidateTokenAndGetUser", httperr)
			return
		}

		var req dto.ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.FailureReason("INVALID_REQUEST"))
			logger.LogError("ChangePassword", "DecodeRequestBody", err)
			return
		}

		if err := svc.ChangePassword(r.Context(), user.Username, req.CurrentPassword, req.NewPassword); err != nil {
			writeChangePasswordError(w, err)
			logger.LogError("ChangePassword", "ChangePassword", err)
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, changePasswordResult{Message: "Password changed successfully"}); err != nil {
			logger.LogError("ChangePassword", "WriteResponse", err)
		}
		logger.LogInfo("ChangePassword", "ChangePassword", "Password changed successfully for user "+user.Username)
	}
}

func writeChangePasswordError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dusers.ErrUserNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.FailureReason("USER_NOT_FOUND"))
	case errors.Is(err, dusers.ErrInvalidCredentials):
		_ = handlers.WriteFailure(w, http.StatusUnauthorized, handlers.FailureReason("INVALID_CREDENTIALS"))
	default:
		if isValidationError(err.Error()) {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.FailureReason("VALIDATION_FAILED"))
			return
		}
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}
