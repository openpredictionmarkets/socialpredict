package sellbetshandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"socialpredict/handlers/bets/dto"
	bets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/middleware"
)

// SellPositionHandler returns an HTTP handler that delegates sales to the bets service.
func SellPositionHandler(betsSvc bets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		user, httpErr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, usersSvc)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode)
			return
		}

		var req dto.SellBetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		result, err := betsSvc.Sell(r.Context(), bets.SellRequest{
			Username: user.Username,
			MarketID: req.MarketID,
			Amount:   req.Amount,
			Outcome:  req.Outcome,
		})
		if err != nil {
			if dustErr, ok := err.(bets.ErrDustCapExceeded); ok {
				http.Error(w, dustErr.Error(), http.StatusUnprocessableEntity)
				return
			}

			switch {
			case errors.Is(err, bets.ErrInvalidOutcome), errors.Is(err, bets.ErrInvalidAmount):
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case errors.Is(err, bets.ErrMarketClosed):
				http.Error(w, err.Error(), http.StatusConflict)
				return
			case errors.Is(err, bets.ErrNoPosition), errors.Is(err, bets.ErrInsufficientShares):
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			case errors.Is(err, dmarkets.ErrMarketNotFound):
				http.Error(w, "Market not found", http.StatusNotFound)
				return
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		response := dto.SellBetResponse{
			Username:      result.Username,
			MarketID:      result.MarketID,
			SharesSold:    result.SharesSold,
			SaleValue:     result.SaleValue,
			Dust:          result.Dust,
			Outcome:       result.Outcome,
			TransactionAt: result.TransactionAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
