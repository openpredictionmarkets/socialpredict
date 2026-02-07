package usershandlers

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// ChangeDisplayNameHandler returns an HTTP handler that delegates display name updates to the users service.
func ChangeDisplayNameHandler(svc dusers.ServiceInterface, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		user, httperr := auth.RequireUser(r)
		if httperr != nil {
			http.Error(w, "Invalid token: "+httperr.Error(), httperr.StatusCode)
			return
		}

		var request dto.ChangeDisplayNameRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		updated, err := svc.UpdateDisplayName(r.Context(), user.Username, request.DisplayName)
		if err != nil {
			writeProfileError(w, err, "display name")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(toPrivateUserResponse(updated)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
