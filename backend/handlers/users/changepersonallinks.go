package usershandlers

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/middleware"
)

// ChangePersonalLinksHandler returns an HTTP handler that delegates personal link updates to the users service.
func ChangePersonalLinksHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		user, httperr := middleware.ValidateTokenAndGetUser(r, svc)
		if httperr != nil {
			http.Error(w, "Invalid token: "+httperr.Error(), httperr.StatusCode)
			return
		}

		var request dto.ChangePersonalLinksRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		updated, err := svc.UpdatePersonalLinks(r.Context(), user.Username, dusers.PersonalLinks{
			PersonalLink1: request.PersonalLink1,
			PersonalLink2: request.PersonalLink2,
			PersonalLink3: request.PersonalLink3,
			PersonalLink4: request.PersonalLink4,
		})
		if err != nil {
			writeProfileError(w, err, "personal links")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(toPrivateUserResponse(updated)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
