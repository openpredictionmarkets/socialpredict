package usershandlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	dusers "socialpredict/internal/domain/users"
)

// GetUserFinancialHandler returns an HTTP handler that responds with comprehensive user financials.
func GetUserFinancialHandler(svc dusers.ServiceInterface) http.HandlerFunc {
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

		snapshot, err := svc.GetUserFinancials(r.Context(), username)
		if err != nil {
			switch err {
			case dusers.ErrUserNotFound:
				_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
			default:
				_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			}
			return
		}

		if snapshot == nil {
			snapshot = make(map[string]int64)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"financial": snapshot}); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}
