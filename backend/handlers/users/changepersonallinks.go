package usershandlers

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/middleware"
	"socialpredict/util"
)

// ChangePersonalLinksHandler returns an HTTP handler that delegates personal link updates to the users service.
func ChangePersonalLinksHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		db := util.GetDB()
		user, httperr := middleware.ValidateTokenAndGetUser(r, db)
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

		user.PersonalLink1 = updated.PersonalLink1
		user.PersonalLink2 = updated.PersonalLink2
		user.PersonalLink3 = updated.PersonalLink3
		user.PersonalLink4 = updated.PersonalLink4

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
