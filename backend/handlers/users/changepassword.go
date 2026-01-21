package usershandlers

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/logger"
)

// ChangePasswordHandler returns an HTTP handler that delegates password changes to the users service.
func ChangePasswordHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		logger.LogInfo("ChangePassword", "ChangePassword", "ChangePassword handler called")

		user, httperr := authsvc.ValidateTokenAndGetUser(r, svc)
		if httperr != nil {
			http.Error(w, "Invalid token: "+httperr.Error(), httperr.StatusCode)
			logger.LogError("ChangePassword", "ValidateTokenAndGetUser", httperr)
			return
		}

		var req dto.ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			logger.LogError("ChangePassword", "DecodeRequestBody", err)
			return
		}

		if err := svc.ChangePassword(r.Context(), user.Username, req.CurrentPassword, req.NewPassword); err != nil {
			writeProfileError(w, err, "password")
			logger.LogError("ChangePassword", "ChangePassword", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Password changed successfully")); err != nil {
			logger.LogError("ChangePassword", "WriteResponse", err)
		}
		logger.LogInfo("ChangePassword", "ChangePassword", "Password changed successfully for user "+user.Username)
	}
}
