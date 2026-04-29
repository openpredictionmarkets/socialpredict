package usershandlers

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// ChangeDescriptionHandler returns an HTTP handler that delegates description updates to the users service.
func ChangeDescriptionHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		user, authErr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, svc)
		if authErr != nil {
			_ = authhttp.WriteFailure(w, authErr)
			return
		}

		var request dto.ChangeDescriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		updated, err := svc.UpdateDescription(r.Context(), user.Username, request.Description)
		if err != nil {
			writeProfileError(w, err, "description")
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, toPrivateUserResponse(updated)); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}
