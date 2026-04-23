package usershandlers

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// ChangePersonalLinksHandler returns an HTTP handler that delegates personal link updates to the users service.
func ChangePersonalLinksHandler(svc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		user, httperr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, svc)
		if httperr != nil {
			_ = handlers.WriteFailure(w, httperr.StatusCode, profileAuthFailureReason(httperr))
			return
		}

		var request dto.ChangePersonalLinksRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
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

		if err := handlers.WriteResult(w, http.StatusOK, toPrivateUserResponse(updated)); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}
