package usercredit

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
)

// GetUserCreditHandler returns an HTTP handler that responds with the user's available credit.
func GetUserCreditHandler(svc dusers.ServiceInterface, maximumDebtAllowed int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		username := mux.Vars(r)["username"]
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}

		credit, err := svc.GetUserCredit(r.Context(), username, maximumDebtAllowed)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				// Maintain legacy behavior: treat missing users as zero-account and return max debt.
				credit = maximumDebtAllowed
			} else {
				http.Error(w, "failed to calculate user credit", http.StatusInternalServerError)
				return
			}
		}

		response := dto.UserCreditResponse{Credit: credit}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
