package usershandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/middleware"
)

// UserMarketPositionHandlerWithService returns an HTTP handler that resolves the authenticated
// user's position in the specified market via the markets service.
func UserMarketPositionHandlerWithService(marketSvc dmarkets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		user, httperr := middleware.ValidateTokenAndGetUser(r, usersSvc)
		if httperr != nil {
			http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		marketIDStr := vars["marketId"]
		if marketIDStr == "" {
			http.Error(w, "Market ID is required", http.StatusBadRequest)
			return
		}

		marketID, err := strconv.ParseInt(marketIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid market ID", http.StatusBadRequest)
			return
		}

		position, err := marketSvc.GetUserPositionInMarket(r.Context(), marketID, user.Username)
		if err != nil {
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
			case dmarkets.ErrUserNotFound:
				http.Error(w, "User not found", http.StatusNotFound)
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			default:
				http.Error(w, "Failed to fetch user position", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(position); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
