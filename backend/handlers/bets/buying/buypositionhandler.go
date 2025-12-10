package buybetshandlers

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers/bets/dto"
	dbets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// PlaceBetHandler returns an HTTP handler that delegates bet placement to the bets domain service.
func PlaceBetHandler(betsSvc dbets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		user, httpErr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, usersSvc)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode)
			return
		}

		var req dto.PlaceBetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		placedBet, err := betsSvc.Place(r.Context(), dbets.PlaceRequest{
			Username: user.Username,
			MarketID: req.MarketID,
			Amount:   req.Amount,
			Outcome:  req.Outcome,
		})
		if err != nil {
			switch err {
			case dbets.ErrInvalidOutcome, dbets.ErrInvalidAmount:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case dbets.ErrMarketClosed:
				http.Error(w, err.Error(), http.StatusConflict)
				return
			case dbets.ErrInsufficientBalance:
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
				return
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		response := dto.PlaceBetResponse{
			Username: placedBet.Username,
			MarketID: placedBet.MarketID,
			Amount:   placedBet.Amount,
			Outcome:  placedBet.Outcome,
			PlacedAt: placedBet.PlacedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
