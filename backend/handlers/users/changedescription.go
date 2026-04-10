package usershandlers

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// ChangeDescriptionHandler returns an HTTP handler that delegates description updates to the users service.
func ChangeDescriptionHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeProfileJSONError(w, http.StatusMethodNotAllowed, "Method is not supported.")
			return
		}

		user, httperr := authsvc.ValidateTokenAndGetUser(r, svc)
		if httperr != nil {
			writeProfileJSONError(w, httperr.StatusCode, "Invalid token: "+httperr.Error())
			return
		}

		var request dto.ChangeDescriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeProfileJSONError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		updated, err := svc.UpdateDescription(r.Context(), user.Username, request.Description)
		if err != nil {
			writeProfileError(w, err, "description")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(toPrivateUserResponse(updated)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
