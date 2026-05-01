package usercredit

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
)

// GetUserCreditHandler returns an HTTP handler that responds with the user's available credit.
func GetUserCreditHandler(svc dusers.ServiceInterface, maximumDebtAllowed int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		username := mux.Vars(r)["username"]
		if username == "" {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		credit, err := svc.GetUserCredit(r.Context(), username, maximumDebtAllowed)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				// Maintain legacy behavior: treat missing users as zero-account and return max debt.
				credit = maximumDebtAllowed
			} else {
				_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
				return
			}
		}

		response := dto.UserCreditResponse{Credit: credit}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}
