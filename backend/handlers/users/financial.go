package usershandlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	dusers "socialpredict/internal/domain/users"
)

// GetUserFinancialHandler returns an HTTP handler that responds with comprehensive user financials.
func GetUserFinancialHandler(svc dusers.ServiceInterface) http.HandlerFunc {
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

		snapshot, err := svc.GetUserFinancials(r.Context(), username)
		if err != nil {
			switch err {
			case dusers.ErrUserNotFound:
				http.Error(w, "user not found", http.StatusNotFound)
			default:
				http.Error(w, "failed to generate financial snapshot", http.StatusInternalServerError)
			}
			return
		}

		if snapshot == nil {
			snapshot = make(map[string]int64)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"financial": snapshot}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
