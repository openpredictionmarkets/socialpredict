package usershandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	dmarkets "socialpredict/internal/domain/markets"
	authsvc "socialpredict/internal/service/auth"
)

// UserMarketPositionHandlerWithService returns an HTTP handler that resolves the authenticated
// user's position in the specified market via the markets service.
func UserMarketPositionHandlerWithService(marketSvc dmarkets.PositionsService, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		user, httperr := auth.RequireUser(r)
		if httperr != nil {
			http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
			return
		}

		marketID, err := parseMarketID(mux.Vars(r)["marketId"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		position, err := marketSvc.GetUserPositionInMarket(r.Context(), marketID, user.Username)
		if err != nil {
			writeUserPositionError(w, marketID, user.Username, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(position); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func parseMarketID(marketIDStr string) (int64, error) {
	if marketIDStr == "" {
		return 0, errors.New("Market ID is required")
	}

	marketID, err := strconv.ParseInt(marketIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("Invalid market ID")
	}
	return marketID, nil
}

func writeUserPositionError(w http.ResponseWriter, marketID int64, username string, err error) {
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
}
